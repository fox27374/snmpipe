package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	configFile  = "config.json"
	snmpTimeout = 5
)

var (
	pollEnabled  = false
	pollInterval = 60
	trapEnabled  = false
	trapPort     = 8162
)

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Errorf("Error loading config: %w", err)
	}

	// Start trap receiver
	if trapEnabled {
		trapReceiver(config.Trap)
	}

	// Start polling devices
	if pollEnabled {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		for {
			pollAndSend(*config)
			<-ticker.C
		}
	}
}

func pollAndSend(config Config) {
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
				fmt.Printf("Received error: %v\n", err)
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
