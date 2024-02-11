package metrics

import (
	"fmt"
	influxdb_client "github.com/influxdata/influxdb1-client/v2"
	"os"
	"time"
)

/*
 * Implementation of metrics collectors.
 * Currently only InfluxDB is supported.
 */

type InfluxDB struct {
	client   influxdb_client.Client
	addr     string
	port     int
	database string
}

// InitializeClient created a new UDP based influx db client. The more secure way of doing would be to use http client
// with batching to avoid lock-ups. For now, UDP client is easier to implement with the useful non-blocking property.
// TODO: use HTTP client with batching instead of UDP.
func (flxDB *InfluxDB) InitializeClient(addr string, port int, database string) error {
	var err error
	flxDB.client, err = influxdb_client.NewUDPClient(influxdb_client.UDPConfig{
		Addr:        fmt.Sprintf("%s:%d", addr, port),
		PayloadSize: 512,
	})
	flxDB.port = port
	flxDB.addr = addr
	flxDB.database = database
	if err != nil {
		return err
	}
	return nil
}

// EmitSingle sends a single metric out using the current influx db client. It implicitly adds the source host where
// the metric is coming from. This operation is non-blocking since we use the UDP client.
func (flxDB *InfluxDB) EmitSingle(m SingleMetric) {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
		m.Tags["host"], _ = os.Hostname()
	}
	_, ok := m.Tags["host"]
	if !ok {
		m.Tags["host"], _ = os.Hostname()
	}
	// additionalFields can not be empty
	if m.AdditionalFields == nil {
		m.AdditionalFields = make(map[string]interface{})
	}
	m.AdditionalFields["value"] = m.Value
	point, _ := influxdb_client.NewPoint(m.Name,
		m.Tags,
		m.AdditionalFields,
		time.Now())
	bp, _ := influxdb_client.NewBatchPoints(influxdb_client.BatchPointsConfig{
		Database:  flxDB.database,
		Precision: "ns",
	})
	bp.AddPoint(point)
	// Fire and forget
	flxDB.client.Write(bp)
}
