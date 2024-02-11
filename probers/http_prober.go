package probers

import (
	"context"
	"fmt"
	"inspector/metrics"
	"inspector/mylogger"
	"net"
	"net/http"
	"time"
)

type HTTPProber struct {
	Interval   time.Duration
	Url        string
	Method     string
	Parameters map[string]string
	client     *http.Client
}

func (httpProber *HTTPProber) Initialize() error {
	return nil
}

func (httpProber *HTTPProber) Connect(c chan metrics.SingleMetric) error {
	//TODO: handle https urls in httpProber
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			start := time.Now()
			conn, err := net.Dial(network, addr)
			if err != nil {
				mylogger.MainLogger.Errorf("Connection Failed for URL %s. Method: %s. Error: %s",
					httpProber.Url, httpProber.Method, err)
				return nil, err
			}
			c <- metrics.SingleMetric{
				Name:  "connect_time",
				Value: time.Since(start).Milliseconds(),
			}
			return conn, nil
		},
		DisableKeepAlives: true,
	}
	httpProber.client = &http.Client{
		//TODO: move http prober timeout to config
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	return nil
}

func (httpProber *HTTPProber) RunOnce(c chan metrics.SingleMetric) error {
	var response *http.Response
	var err error
	var start time.Time
	if httpProber.Method == "GET" {
		start = time.Now()
		response, err = httpProber.client.Get(httpProber.Url)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported method: %s", httpProber.Method)
	}
	c <- metrics.SingleMetric{
		Name:  "response_time",
		Value: time.Since(start).Milliseconds(),
	}
	c <- metrics.SingleMetric{
		Name:  "status",
		Value: int64(response.StatusCode),
	}
	response.Body.Close()
	return nil
}

func (httpProber *HTTPProber) TearDown() error {
	httpProber.client.CloseIdleConnections()
	return nil
}
