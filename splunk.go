package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

	// Create raw json for debug output, otherwide slog will
	// re-encode the json with escape characters
	if debug {
		var raw json.RawMessage
		_ = json.Unmarshal(jsondata, &raw)
		logger.Debug("Data for HEC prepared", "payload", raw)
	}

	return jsondata, nil
}

// Send JSON data to Splunk HEC using the http package
// Data is send as a POST request in a certain format
// https://docs.splunk.com/Documentation/Splunk/latest/Data/FormateventsforHTTPEventCollector
func sendToSplunkHec(data []SNMPData) error {
	splunk := config.Splunk

	// Check if http or https is used and configure the http client accordingly
	u, err := url.Parse(splunk.SplunkHecUrl)
	if err != nil {
		return fmt.Errorf("error parsing URL: %v", err)
	}

	// Create request data
	jsondata, err := createHecEvent(data)
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

	var client *http.Client

	switch u.Scheme {
	case "http":
		logger.Debug("using HTTP scheme")
		client = &http.Client{}
	case "https":
		logger.Debug("using HTTPS scheme. Certificate check disabled")
		// Disable certificate check
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	default:
		return fmt.Errorf("url scheme not supported: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read the response body to get the error message
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("POST request failed with %s and failed to read response body: %v", resp.Status, readErr)
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("POST request failed with status %s and message: %s", resp.Status, bodyString)
	}
	logger.Info("data successfully sent to Splunk")

	return nil
}
