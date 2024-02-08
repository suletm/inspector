package main

import (
	"fmt"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

type Inspector struct {
	website   string
	dbConfig  DBConfig
	whatToLog []string
	Period    time.Duration
}

type DBConfig struct {
	DatabaseURL  string
	Port         string
	Username     string
	Password     string
	DatabaseName string
}

func NewInspector(website string, dbConfig DBConfig, whatToLog []string, period time.Duration) *Inspector {
	fmt.Printf("[%v]\n[%v] [INFO] Inspector Initiated for Website: '%s'\n", time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), website)
	return &Inspector{website, dbConfig, whatToLog, period}
}

func (ins *Inspector) MeasureAndLog() {
	ticker := time.NewTicker(ins.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := NewResponseMetrics(ins.website)
			err := metrics.Measure()
			if err != nil {
				fmt.Printf("[%v] [ERROR] Metrics Measurement Failed for '%s': %s\n", time.Now().Format(time.RFC3339), ins.website, err)
				continue
			}

			fmt.Printf("[%v] [INFO] URL: '%s', Response: %.2fs, Connect: %.2fs, Status: %d, Cert Days: %d\n",
				time.Now().Format(time.RFC3339), metrics.URL, metrics.ResponseTime.Seconds(), metrics.ConnectionTime.Seconds(), metrics.StatusCode, metrics.CertificateDays)

			if err := ins.logToInfluxDB(metrics); err != nil {
				fmt.Printf("[%v] [ERROR] InfluxDBSubConfig Logging Failed for '%s': %s\n", time.Now().Format(time.RFC3339), ins.website, err)
			}
		}
	}
}

func (ins *Inspector) logToInfluxDB(metrics *ResponseMetrics) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("%s:%s", ins.dbConfig.DatabaseURL, ins.dbConfig.Port),
		Username: ins.dbConfig.Username,
		Password: ins.dbConfig.Password,
	})
	if err != nil {
		return fmt.Errorf("error creating InfluxDBSubConfig client for '%s': %w", ins.website, err)
	}
	defer c.Close()

	createDBQuery := fmt.Sprintf("CREATE DATABASE \"%s\"", ins.dbConfig.DatabaseName)
	response, err := c.Query(client.NewQuery(createDBQuery, "", ""))
	if err != nil {
		return fmt.Errorf("error executing query to create database for '%s': %w", ins.website, err)
	}
	if response.Error() != nil {
		return fmt.Errorf("error in response when creating database for '%s': %w", ins.website, response.Error())
	}

	return ins.writeMetrics(c, metrics)
}

func (ins *Inspector) writeMetrics(c client.Client, metrics *ResponseMetrics) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ins.dbConfig.DatabaseName,
		Precision: "s",
	})
	if err != nil {
		return fmt.Errorf("error creating new batch points for '%s': %w", ins.website, err)
	}

	measurementName := strings.Split(ins.website, "//")[1]
	fields := map[string]interface{}{
		"response_time":    metrics.ResponseTime.Seconds(),
		"connection_time":  metrics.ConnectionTime.Seconds(),
		"status_code":      metrics.StatusCode,
		"certificate_days": metrics.CertificateDays,
	}

	pt, err := client.NewPoint(measurementName, nil, fields, time.Now())
	if err != nil {
		return fmt.Errorf("error creating new point for '%s': %w", ins.website, err)
	}
	bp.AddPoint(pt)

	if err := c.Write(bp); err != nil {
		return fmt.Errorf("error writing batch points for '%s': %w", ins.website, err)
	}

	return nil
}
