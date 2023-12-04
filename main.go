package main

import "time"

func main() {
	dbConfig := DBConfig{
		DatabaseURL:  "http://inspector-influxdb",
		Port:         "8086",
		DatabaseName: "inspector",
	}

	inspector := NewInspector("https://sumnotes.net/", dbConfig, []string{"ResponseTime", "ConnectionTime", "StatusCode", "CertificateDays"}, 60*time.Second)

	inspector.MeasureAndLog()
}
