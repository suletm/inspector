package main

import (
	"flag"
	"fmt"
	"inspector/config"
	"inspector/metrics"
	"inspector/probers"
	"math/rand"
	"os"
	"time"
)

var METRIC_CHANNEL_POLL_INTERVAL = 2 * time.Second
var TARGET_LIST_SCAN_WAIT_INTERVAL = 5 * time.Second
var PROBER_RESTART_INTERVAL_JITTER_RANGE = 2

func main() {
	var configPath = flag.String("config_path", "", "Path to the configuration file. Mandatory argument.")
	flag.Parse()
	if *configPath == "" {
		fmt.Printf("Missing a mandatory argument: config_path. Try -help option for the list of supported" +
			"arguments")
		os.Exit(1)
	}
	c, err := config.NewConfig(*configPath)
	if err != nil {
		fmt.Printf("Error reading config: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Config parsed: %v", c.TimeSeriesDB[0])

	//TODO: enable support for multiple time series databases. For now only the first one is used from config.
	mdb, err := metrics.NewMetricsDB(c.TimeSeriesDB[0])
	if err != nil {
		fmt.Printf("Failed initializing metrics db client with error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Initialized metrics database...")

	metricsChannel := make(chan metrics.SingleMetric, 1000)

	go func() {
		for {
			select {
			case m := <-metricsChannel:
				mdb.EmitSingle(m)
			default:
				fmt.Printf("Metrics channel is empty. Waiting some more...")
				time.Sleep(METRIC_CHANNEL_POLL_INTERVAL)
			}
		}
	}()

	for {
		for _, target := range c.Targets {
			for _, proberSubConfig := range target.Probers {
				go func() {
					prober, err := probers.NewProber(proberSubConfig)
					if err != nil {
						fmt.Printf("Failed creating new prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					err = prober.Initialize()
					if err != nil {
						fmt.Printf("Failed initializing prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					fmt.Printf("Successfully initialized prober: %s for target: %s", proberSubConfig.Name, target.Name)
					err = prober.RunOnce(metricsChannel)
					if err != nil {
						fmt.Printf("Failed running prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
					}
					err = prober.TearDown()
					if err != nil {
						fmt.Printf("Failed tearing down prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
					}
					fmt.Printf("Successfully torn down prober: %s for target: %s", proberSubConfig.Name, target.Name)
				}()

				jitter := rand.Intn(PROBER_RESTART_INTERVAL_JITTER_RANGE)
				time.Sleep(time.Duration(jitter) * time.Second)
			}
		}
		// Wait before scanning through the targets from scratch
		time.Sleep(TARGET_LIST_SCAN_WAIT_INTERVAL)
	}
}
