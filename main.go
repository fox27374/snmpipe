package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	configFile  = "config.json"
	snmpTimeout = 5
)

var (
	config       Config
	logger       *slog.Logger
	pollEnabled  = false
	pollInterval = 60
	trapEnabled  = false
	trapPort     = 8162
)

func init() {
	debug := false
	var logLevel slog.Level
	if debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	// Create and configure the handler.
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Set the global logger variable.
	logger = slog.New(handler)
	// You can also set it as the default logger for the entire application.
	slog.SetDefault(logger)
}

func main() {
	logger.Info("Application started")
	err := loadConfig()
	if err != nil {
		logger.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Start trap receiver
	logger.Info("Trap receiver", slog.Any("enabled", trapEnabled))
	if trapEnabled {
		logger.Info(fmt.Sprintf("Trap receiver enabled, starting on port %d", trapPort))
		go trapReceiver(config.Trap)
	}

	// Start polling devices
	logger.Info("Poller", slog.Any("enabled", pollEnabled))
	if pollEnabled {
		logger.Info(fmt.Sprintf("Device polling enabled, starting with an interval of %d seconds", pollInterval))
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		for {
			pollAndSend()
			<-ticker.C
		}
	}
}

// func trapAndSend() {
// 	dataChan := make(chan SNMPData)
// 	errChan := make(chan error)
// }

func pollAndSend() {
	dataChan := make(chan SNMPData)
	errChan := make(chan error)
	var wg sync.WaitGroup
	var pollResults []SNMPData

	// Launch goroutines for polling
	for _, device := range config.Devices {
		wg.Add(1)
		go func(d DeviceConfig) {
			defer wg.Done()
			pollDevice(d, dataChan, errChan)
		}(device)
	}

	// Close channels once all workers are done
	go func() {
		wg.Wait() // Wait for all pollDevice goroutines to complete
		close(dataChan)
		close(errChan)
	}()

	// Read from errChan
	go func() {
		for err := range errChan {
			if err != nil {
				logger.Error("Error received from go routine", slog.Any("error", err))
			}
		}
	}()

	// Read from dataChan and fill the pollResults slice
	for res := range dataChan {
		if res != nil {
			pollResults = append(pollResults, res)
		}
	}

	// Send data to Splunk
	err := sendToSplunkHec(pollResults)
	if err != nil {
		fmt.Println(err)
	}
}
