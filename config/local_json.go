package config

import (
	"encoding/json"
	"inspector/mylogger"
	"io/ioutil"
)

type InfluxDBSubConfig struct {
	DatabaseURL  string `json:"database_url"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Protocol     string `json:"transport_protocol"`
}

type MySQLDBSubConfig struct {
	DatabaseURL  string `json:"database_url"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
}

type MetricsDBSubConfig struct {
	*InfluxDBSubConfig `json:"influxdb,omitempty"`
	*MySQLDBSubConfig  `json:"mysqldb,omitempty"`
}

type ProberContextSubConfig struct {
	Url        string      `json:"url"`
	Method     string      `json:"method"`
	Parameters interface{} `json:"parameters"`
}

type ProberSubConfig struct {
	Name    string                 `json:"name"`
	Context ProberContextSubConfig `json:"context"`
}

type TargetSubConfig struct {
	Id      int64             `json:"id"`
	Name    string            `json:"name"`
	Probers []ProberSubConfig `json:"probers"`
}

type Config struct {
	TimeSeriesDB []MetricsDBSubConfig `json:"metrics_db"`
	Targets      []TargetSubConfig    `json:"targets"`
}

func NewConfig(path string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data Config
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		mylogger.MainLogger.Errorf("Failed parsing config at path: %s with error: %s", path, err)
		return nil, err
	}
	return &data, err
}
