package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

type ResponseMetrics struct {
	URL             string
	ResponseTime    time.Duration
	ConnectionTime  time.Duration
	StatusCode      int
	CertificateDays int
}

// Constructors
func NewResponseMetrics(url string) *ResponseMetrics {
	return &ResponseMetrics{URL: url}
}

// ResponseMetrics methods
func (m *ResponseMetrics) Measure() error {
	client, err := m.createClient()
	if err != nil {
		fmt.Printf("[%v] [ERROR] Failed to Create Client for URL '%s': %s\n", time.Now().Format(time.RFC3339), m.URL, err)
		return err
	}

	err = m.makeRequest(client)
	if err != nil {
		fmt.Printf("[%v] [ERROR] Request Failed for URL '%s': %s\n", time.Now().Format(time.RFC3339), m.URL, err)
		return err
	}

	return nil
}

func (m *ResponseMetrics) createClient() (*http.Client, error) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			start := time.Now()
			conn, err := net.Dial(network, addr)
			if err != nil {
				fmt.Printf("[%v] [ERROR] Connection Failed for URL '%s': %s\n", time.Now().Format(time.RFC3339), m.URL, err)
				return nil, err
			}
			m.ConnectionTime = time.Since(start)
			return conn, nil
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}, nil
}

func (m *ResponseMetrics) makeRequest(client *http.Client) error {
	start := time.Now()
	resp, err := client.Get(m.URL)
	if err != nil {
		fmt.Printf("[%v] [ERROR] Error Making Request to URL '%s': %s\n", time.Now().Format(time.RFC3339), m.URL, err)
		return err
	}
	defer resp.Body.Close()

	m.ResponseTime = time.Since(start)
	m.StatusCode = resp.StatusCode

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		m.CertificateDays = int(time.Until(resp.TLS.PeerCertificates[0].NotAfter).Hours() / 24)
	}

	return nil
}
