package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
func loadConfig(configFileLocation string) error {
	// Load config file and unmarshal it
	configFileData, err := readConfigFile(configFileLocation)
	if err != nil {
		return fmt.Errorf("cannot not read config file: %w", err)
	}
	err = json.Unmarshal(configFileData, &config)
	if err != nil {
		return fmt.Errorf("cannot unmarshal config file: %w", err)
	}

	// Check if mandatory values are set
	// Splunk
	if config.Splunk.SplunkHecUrl == "" {
		return fmt.Errorf("splunk HEC URL is not set")
	}
	if config.Splunk.SplunkHectoken == "" {
		return fmt.Errorf("splunk HEC token is not set")
	}

	// Trap
	if strings.ToLower(config.Trap.Enabled) == "true" {
		trapEnabled = true
	}
	if strings.ToLower(config.Trap.Enabled) == "true" && config.Trap.TrapPort == "" {
		return fmt.Errorf("trap port is not set")
	}

	// Poll
	if strings.ToLower(config.Poll.Enabled) == "true" {
		pollEnabled = true
	}

	// Convert string values from config to int
	if config.Poll.Interval != "" {
		pollInterval, err = strconv.Atoi(config.Poll.Interval)
		if err != nil {
			return fmt.Errorf("unable to convert poll interval string to int")
		}
	}

	trapPort, err = strconv.Atoi(config.Trap.TrapPort)
	if err != nil {
		return fmt.Errorf("unable to convert trap port string to int")
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

	return nil
}

// Loads the configuration from the config file
func readConfigFile(configFileLocation string) ([]byte, error) {
	configData, err := os.ReadFile(configFileLocation)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file", slog.Any("error", err))
	}

	return configData, nil
}

// Check if the config file exists either in the same folder
// or in /etc/snmpipe to maintain compatibility between the local
// installation and the containerized version
func checkConfigFile() (string, error) {
	// Check for the config file in the standard system path first.
	if _, err := os.Stat(configFile); err == nil {
		return configFile, nil
	}

	// If not found, check for the config file in the current directory.
	if _, err := os.Stat("./config.json"); err == nil {
		return "./config.json", nil
	}

	// If neither file exists, return a clear and specific error.
	return "", fmt.Errorf("Config file not found in '%s' or local folder", configFile)
}
