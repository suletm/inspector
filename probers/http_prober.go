package probers

import (
	"context"
	"crypto/tls"
	"fmt"
	"inspector/metrics"
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
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := net.Dial(network, addr)
			if err != nil {
				fmt.Printf("[%v] [ERROR] Connection Failed for URL '%s', method: %s, error: %s", time.Now().Format(time.RFC3339),
					httpProber.Url, httpProber.Method, err)
				return nil, err
			}
			return conn, nil
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpProber.client = &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       10 * time.Second,
	}
	return nil
}

func (httpProber *HTTPProber) RunOnce(c chan metrics.SingleMetric) error {
	var response *http.Response
	var err error
	if httpProber.Method == "GET" {
		response, err = httpProber.client.Get(httpProber.Url)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported method: %s", httpProber.Method)
	}
	c <- metrics.SingleMetric{
		Name:  "status",
		Value: int64(response.StatusCode),
	}
	return nil
}

func (httpProber *HTTPProber) TearDown() error {
	httpProber.client.CloseIdleConnections()
	return nil
}
