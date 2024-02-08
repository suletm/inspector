package metrics

import (
	"fmt"
	"inspector/config"
)

type SingleMetric struct {
	Name  string
	Value int64
}

type MetricsDB interface {
	InitializeClient(addr string, port int, database string) error
	EmitSingle(m SingleMetric)
}

// NewMetricsDB initializes a metrics database specified by the config. It returns an object that implements the MetricsDB
// interface. Currently only InfluxDB is supported.
func NewMetricsDB(c config.MetricsDBSubConfig) (MetricsDB, error) {
	var mdb MetricsDB
	if c.InfluxDBSubConfig != nil {
		mdb = &InfluxDB{}
		err := mdb.InitializeClient(c.InfluxDBSubConfig.DatabaseURL, c.InfluxDBSubConfig.Port, c.InfluxDBSubConfig.DatabaseName)
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Printf("Specified metrics database is not supported in config: %v", c)
		return nil, fmt.Errorf("MetricsDB defiend in configuration is not supported: %s", c)
	}
	return mdb, nil
}
