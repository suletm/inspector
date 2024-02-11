package main

import (
	"flag"
	glogger "github.com/google/logger"
	"inspector/config"
	"inspector/metrics"
	"inspector/mylogger"
	"inspector/probers"
	"io"
	"math/rand"
	"os"
	"time"
)

var METRIC_CHANNEL_POLL_INTERVAL = 2 * time.Second
var TARGET_LIST_SCAN_WAIT_INTERVAL = 5 * time.Second
var PROBER_RESTART_INTERVAL_JITTER_RANGE = 2
var METRIC_CHANNEL_SIZE = 1000

func main() {

	var configPath = flag.String("config_path", "", "Path to the configuration file. Mandatory argument.")
	var logFilePath = flag.String("log_path", "", "A file where to write logs. Optional argument, defaults to stdout")

	flag.Parse()

	if *logFilePath != "" {
		logFile, err := os.OpenFile(*logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			glogger.Fatalf("Failed to open log file path: %s, error: %s", logFilePath, err)
		}
		defer logFile.Close()
		mylogger.MainLogger = glogger.Init("InspectorLogger", false, true, logFile)
	} else {
		mylogger.MainLogger = glogger.Init("InspectorLogger", true, true, io.Discard)
	}

	if *configPath == "" {
		mylogger.MainLogger.Errorf("Missing a mandatory argument: config_path. Try -help option for the list of supported" +
			"arguments")
		os.Exit(1)
	}
	c, err := config.NewConfig(*configPath)
	if err != nil {
		mylogger.MainLogger.Infof("Error reading config: %s", err)
		os.Exit(1)
	}
	mylogger.MainLogger.Infof("Config parsed: %v", c.TimeSeriesDB[0])

	//TODO: enable support for multiple time series databases. For now only the first one is used from config.
	mdb, err := metrics.NewMetricsDB(c.TimeSeriesDB[0])
	if err != nil {
		mylogger.MainLogger.Infof("Failed initializing metrics db client with error: %s", err)
		os.Exit(1)
	}
	mylogger.MainLogger.Infof("Initialized metrics database...")

	// TODO: determine what should the size of the channel be ?
	metricsChannel := make(chan metrics.SingleMetric, METRIC_CHANNEL_SIZE)

	go func() {
		for {
			select {
			case m := <-metricsChannel:
				mdb.EmitSingle(m)
			default:
				mylogger.MainLogger.Infof("Metrics channel is empty. Waiting some more...")
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
						mylogger.MainLogger.Errorf("Failed creating new prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					err = prober.Initialize()
					if err != nil {
						mylogger.MainLogger.Errorf("Failed initializing prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					mylogger.MainLogger.Infof("Successfully initialized prober: %s for target: %s", proberSubConfig.Name, target.Name)
					err = prober.RunOnce(metricsChannel)
					if err != nil {
						mylogger.MainLogger.Errorf("Failed running prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					err = prober.TearDown()
					if err != nil {
						mylogger.MainLogger.Errorf("Failed tearing down prober: %s for target: %s, error: %s", proberSubConfig.Name,
							target.Name, err)
						return
					}
					mylogger.MainLogger.Infof("Successfully torn down prober: %s for target: %s", proberSubConfig.Name, target.Name)
				}()

				jitter := rand.Intn(PROBER_RESTART_INTERVAL_JITTER_RANGE)
				time.Sleep(time.Duration(jitter) * time.Second)
			}
		}
		// Wait before scanning through the targets from scratch
		time.Sleep(TARGET_LIST_SCAN_WAIT_INTERVAL)
	}
}
