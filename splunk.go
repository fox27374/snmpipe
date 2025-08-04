package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Create JSON data that can be send to the Splunk HEC
// Data has to be in a specific format in order to be processed by Splunk
// https://docs.splunk.com/Documentation/Splunk/latest/Data/FormateventsforHTTPEventCollector
func createHecEvent(data []SNMPData) ([]byte, error) {
	splunk := config.Splunk
	var splunkHecEvent SplunkHecEvent

	// Add additional Splunk data if configured in the config file
	if splunk.SplunkIndex != "" {
		splunkHecEvent.Index = splunk.SplunkIndex
	}

	if splunk.SplunkSourcetype != "" {
		splunkHecEvent.Sourcetype = splunk.SplunkSourcetype
	} else {
		splunkHecEvent.Sourcetype = "_json"
	}

	// Add poll results to SplunkHecEvent struct
	splunkHecEvent.Event = data

	// transform the SplunkHecEvent struct into a byte slice
	jsondata, err := json.Marshal(splunkHecEvent)
	if err != nil {
		return nil, fmt.Errorf("creating json data failed: %v", err)
	}

	return jsondata, nil
}

// Send JSON data to Splunk HEC using the http package
// Data is send as a POST request in a certain format
// https://docs.splunk.com/Documentation/Splunk/latest/Data/FormateventsforHTTPEventCollector
func sendToSplunkHec(data []SNMPData) error {
	splunk := config.Splunk
	client := &http.Client{}

	// Create request data
	jsondata, err := createHecEvent(data)
	logger.Debug("Data for HEC prepared", slog.String("data", string(jsondata)))
	if err != nil {
		return fmt.Errorf("creating request data failed: %v", err)
	}

	req, err := http.NewRequest("POST", splunk.SplunkHecUrl, bytes.NewBuffer(jsondata))
	if err != nil {
		return fmt.Errorf("creating request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+splunk.SplunkHectoken)

	logger.Debug("splunk HEC url", slog.String("url", splunk.SplunkHecUrl))
	logger.Debug("splunk authorization token", slog.String("token", splunk.SplunkHectoken))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST request failed with %s", resp.Status)
	}

	return nil
}
