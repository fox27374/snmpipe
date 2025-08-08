package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	configFile  = "config.json"
	snmpTimeout = 5
)

// TODO: Pass debug variable as an environmental variable or set it in the config file
var (
	debug          = false
	config         Config
	logger         *slog.Logger
	pollEnabled    = false
	pollInterval   = 60
	trapEnabled    = false
	trapListenAddr = "0.0.0.0"
	trapPort       = 8162
)

func init() {
	// Check if the "DEBUG" environment variable is set
	_, ok := os.LookupEnv("DEBUG")
	if ok {
		debug = true
	}

	var logLevel slog.Level
	var addSource bool

	if debug {
		logLevel = slog.LevelDebug
		addSource = true
	} else {
		logLevel = slog.LevelInfo
		addSource = false
	}

	// Create and configure log handler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: addSource,
	})

	// Set the global logger variable
	logger = slog.New(handler)
	// Set as default logger
	slog.SetDefault(logger)
}

func main() {
	logger.Info("Application started")
	err := loadConfig()
	if err != nil {
		logger.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Create channels for trap communication
	trapDataChan := make(chan SNMPData)
	trapErrChan := make(chan error)

	// Create a channel to listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start trap receiver
	logger.Info("Trap receiver", slog.Any("enabled", trapEnabled))
	if trapEnabled {
		logger.Info(fmt.Sprintf("Trap receiver enabled, starting on port %d", trapPort))
		go trapReceiver(config.Trap, trapDataChan, trapErrChan)
	}

	// Start polling devices
	logger.Info("Poller", slog.Any("enabled", pollEnabled))
	if pollEnabled {
		logger.Info(fmt.Sprintf("Device polling enabled, starting with an interval of %d seconds", pollInterval))
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

		go func() {
			for {
				select {
				case <-ticker.C:
					err := pollAndSend()
					if err != nil {
						logger.Error("device polling failed", slog.Any("error", err))
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	slog.Info("Application started, waiting for traps or polling events")

	// Use a for-select loop to continuously read from trap channels
	for {
		select {
		case data := <-trapDataChan:
			slog.Info("New trap data received", slog.Any("host", data["ip"]))
			s := []SNMPData{data}
			err := sendToSplunkHec(s)
			if err != nil {
				slog.Error(fmt.Sprintf("sending data to Splunk failed: %v", err))
			}
		case err := <-trapErrChan:
			slog.Error("Trap receiver error", slog.Any("error", err))
		case <-quit:
			slog.Info("Received shutdown signal, stopping...")
			return
		}
	}
}
