package main

import (
	"fmt"
	"sync"
)

const (
	configFile  = "config.json"
	snmpTimeout = 5
)

type Config struct {
	Splunk  SplunkConfig   `json:"splunk"`
	SNMP    SNMPConfig     `json:"snmp"`
	Devices []DeviceConfig `json:"devices"`
}

type SplunkConfig struct {
	SplunkHecUrl     string `json:"splunk_hec_url"`
	SplunkHectoken   string `json:"splunk_hec_token"`
	SplunkIndex      string `json:"splunk_index"`
	SplunkSourcetype string `json:"splunk_sourcetype"`
}

type SNMPConfig struct {
	SNMPPort           string `json:"snmp_port"`
	SNMPVersion        string `json:"snmp_version"`
	SNMPCommunity      string `json:"snmp_community"`
	SNMPUser           string `json:"snmp_user"`
	SNMPAuthProtocol   string `json:"snmp_auth_protocol"`
	SNMPAuthPassphrase string `json:"snmp_auth_passphrase"`
	SNMPPrivProtocol   string `json:"snmp_priv_protocol"`
	SNMPPrivPassphrase string `json:"snmp_priv_passphrase"`
}

type DeviceConfig struct {
	IP                 string            `json:"ip"`
	Name               string            `json:"name"`
	SNMPPort           string            `json:"snmp_port"`
	SNMPVersion        string            `json:"snmp_version"`
	SNMPCommunity      string            `json:"snmp_community"`
	SNMPUser           string            `json:"snmp_user"`
	SNMPAuthProtocol   string            `json:"snmp_auth_protocol"`
	SNMPAuthPassphrase string            `json:"snmp_auth_passphrase"`
	SNMPPrivProtocol   string            `json:"snmp_priv_protocol"`
	SNMPPrivPassphrase string            `json:"snmp_priv_passphrase"`
	OIDs               map[string]string `json:"oids"`
}

type PollResult map[string]any

type SplunkHecEvent struct {
	Index      string       `json:"index"`
	Sourcetype string       `json:"sourcetype"`
	Event      []PollResult `json:"event"`
}

func main() {
	dataChan := make(chan PollResult)
	errChan := make(chan error)
	var wg sync.WaitGroup
	var pollResults []PollResult

	config := loadConfig()

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
	err := sendToSplunkHec(config.Splunk, pollResults)
	if err != nil {
		fmt.Println(err)
	}
}
