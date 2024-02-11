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

func (flxDB *InfluxDB) EmitSingle(m SingleMetric) {
	var tags map[string]string
	var additionalFields map[string]interface{}

	if m.Tags == nil {
		tags = make(map[string]string)
		tags["host"], _ = os.Hostname()
	}
	_, ok := tags["host"]
	if !ok {
		tags["host"], _ = os.Hostname()
	}
	// additionalFields can not be empty
	if additionalFields == nil {
		additionalFields = make(map[string]interface{})
	}
	additionalFields["value"] = m.Value
	point, _ := influxdb_client.NewPoint(m.Name,
		tags,
		additionalFields,
		time.Now())
	bp, _ := influxdb_client.NewBatchPoints(influxdb_client.BatchPointsConfig{
		Database:  flxDB.database,
		Precision: "ns",
	})
	bp.AddPoint(point)
	// Fire and forget
	flxDB.client.Write(bp)
}
