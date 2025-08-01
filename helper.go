package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Splunk  SplunkConfig   `json:"splunk"`
	Trap    TrapConfig     `json:"snmp_trap"`
	Poll    PollConfig     `json:"snmp_poll"`
	Devices []DeviceConfig `json:"devices"`
}

type SplunkConfig struct {
	SplunkHecUrl     string `json:"splunk_hec_url"`
	SplunkHectoken   string `json:"splunk_hec_token"`
	SplunkIndex      string `json:"splunk_index"`
	SplunkSourcetype string `json:"splunk_sourcetype"`
}

type TrapConfig struct {
	Enabled            string `json:"enabled"`
	TrapPort           string `json:"trap_port"`
	TrapUser           string `json:"trap_user"`
	TrapAuthProtocol   string `json:"trap_auth_protocol"`
	TrapAuthPassphrase string `json:"trap_auth_passphrase"`
	TrapPrivProtocol   string `json:"trap_priv_protocol"`
	TrapPrivPassphrase string `json:"trap_priv_passphrase"`
}

type PollConfig struct {
	Enabled            string `json:"enabled"`
	Interval           string `json:"interval"`
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

type SNMPData map[string]any

type SplunkHecEvent struct {
	Index      string     `json:"index"`
	Sourcetype string     `json:"sourcetype"`
	Event      []SNMPData `json:"event"`
}

// Loads the configuration from the config file
// and sets the default values for every device
// if the value in the device config does not exist
func loadConfig() (*Config, error) {
	var config Config
	err := json.Unmarshal(readConfigFile(), &config)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
		return nil, err
	}

	// Check if mandatory values are set
	// Splunk
	if config.Splunk.SplunkHecUrl == "" {
		log.Fatalf("Splun HEC URL is not set")
	}
	if config.Splunk.SplunkHectoken == "" {
		log.Fatalf("Splun HEC token is not set")
	}

	// Trap
	if strings.ToLower(config.Trap.Enabled) == "true" {
		trapEnabled = true
	}
	if strings.ToLower(config.Trap.Enabled) == "true" && config.Trap.TrapPort == "" {
		log.Fatalf("Trap port is not set")
	}

	// Poll
	if strings.ToLower(config.Poll.Enabled) == "true" {
		pollEnabled = true
	}

	// Convert string values from config to int
	pollInterval, err = strconv.Atoi(config.Poll.Interval)
	if err != nil {
		fmt.Errorf("String to int convert: %w", err)
		return nil, err
	}

	trapPort, err = strconv.Atoi(config.Trap.TrapPort)
	if err != nil {
		fmt.Errorf("String to int convert: %w", err)
		return nil, err
	}

	// Apply default values from global SNMP config to devices
	for d := range config.Devices {
		device := &config.Devices[d] // Get a pointer to modify the slice element

		if device.SNMPPort == "" {
			device.SNMPPort = config.Poll.SNMPPort
		}
		if device.SNMPCommunity == "" {
			device.SNMPCommunity = config.Poll.SNMPCommunity
		}
		if device.SNMPVersion == "" {
			device.SNMPVersion = config.Poll.SNMPVersion
		}
		if device.SNMPUser == "" {
			device.SNMPUser = config.Poll.SNMPUser
		}
		if device.SNMPAuthProtocol == "" {
			device.SNMPAuthProtocol = config.Poll.SNMPAuthProtocol
		}
		if device.SNMPAuthPassphrase == "" {
			device.SNMPAuthPassphrase = config.Poll.SNMPAuthPassphrase
		}
		if device.SNMPPrivProtocol == "" {
			device.SNMPPrivProtocol = config.Poll.SNMPPrivProtocol
		}
		if device.SNMPPrivPassphrase == "" {
			device.SNMPPrivPassphrase = config.Poll.SNMPPrivPassphrase
		}
	}

	return &config, nil
}

// Loads the configuration from the config file
func readConfigFile() []byte {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	return configData
}
