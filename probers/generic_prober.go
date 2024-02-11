package probers

import (
	"fmt"
	"inspector/config"
	"inspector/metrics"
)

type Prober interface {
	Initialize() error
	Connect(chan metrics.SingleMetric) error
	RunOnce(chan metrics.SingleMetric) error
	TearDown() error
}

func NewProber(c config.ProberSubConfig) (Prober, error) {
	var newProber Prober
	switch c.Name {
	case "basic_http_prober":
		newProber = &HTTPProber{
			Url:    c.Context.Url,
			Method: c.Context.Method,
		}
		break
	default:
		return nil, fmt.Errorf("unsupported prober type: %s", c.Name)
	}
	return newProber, nil
}
