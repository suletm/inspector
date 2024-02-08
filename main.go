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
	for {
		for _, target := range c.Targets {
			for _, proberSubConfig := range target.Probers {
				prober, err := probers.NewProber(proberSubConfig)
				if err != nil {
					fmt.Printf("Failed creating new prober: %s for target: %s, error: %s", proberSubConfig.Name,
						target.Name, err)
					continue
				}
				err = prober.Initialize()
				if err != nil {
					fmt.Printf("Failed initializing prober: %s for target: %s, error: %s", proberSubConfig.Name,
						target.Name, err)
					continue
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
				jitter := rand.Intn(2)
				time.Sleep(time.Duration(jitter) * time.Second)
			}
		}
		// For each target, spawn the set of defined probers in an infinite loop.

		select {
		case m := <-metricsChannel:
			mdb.EmitSingle(m)
		default:
			fmt.Printf("Drained all metrics from the channel...")
			break
		}
		time.Sleep(5 * time.Second)
	}

	//inspector := NewInspector("https://sumnotes.net/", dbConfig, []string{"ResponseTime", "ConnectionTime", "StatusCode", "CertificateDays"}, 60*time.Second)

	//	inspector.MeasureAndLog()
}
