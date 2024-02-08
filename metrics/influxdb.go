package metrics

import (
	"fmt"
	influxdb_client "github.com/influxdata/influxdb1-client/v2"
	"time"
)

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
	p, _ := influxdb_client.NewPoint(m.Name,
		nil,
		map[string]interface{}{"value": m.Value},
		time.Now())
	bp, _ := influxdb_client.NewBatchPoints(influxdb_client.BatchPointsConfig{
		Database:  flxDB.database,
		Precision: "ns",
	})
	bp.AddPoint(p)
	// Fire and forget
	flxDB.client.Write(bp)
	fmt.Printf("EmitSingle: metric sent...")
}
