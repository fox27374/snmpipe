package main

import (
	"encoding/json"
	"log"
	"os"
)

// Loads the configuration from the config file
// and sets the default values for every device
// if the value in the device config does not exist
func loadConfig() *Config {
	var config Config
	err := json.Unmarshal(readConfigFile(), &config)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	// Apply default values from global SNMP config to devices
	for d := range config.Devices {
		device := &config.Devices[d] // Get a pointer to modify the slice element

		if device.SNMPPort == "" {
			device.SNMPPort = config.SNMP.SNMPPort
		}
		if device.SNMPCommunity == "" {
			device.SNMPCommunity = config.SNMP.SNMPCommunity
		}
		if device.SNMPVersion == "" {
			device.SNMPVersion = config.SNMP.SNMPVersion
		}
		if device.SNMPUser == "" {
			device.SNMPUser = config.SNMP.SNMPUser
		}
		if device.SNMPAuthProtocol == "" {
			device.SNMPAuthProtocol = config.SNMP.SNMPAuthProtocol
		}
		if device.SNMPAuthPassphrase == "" {
			device.SNMPAuthPassphrase = config.SNMP.SNMPAuthPassphrase
		}
		if device.SNMPPrivProtocol == "" {
			device.SNMPPrivProtocol = config.SNMP.SNMPPrivProtocol
		}
		if device.SNMPPrivPassphrase == "" {
			device.SNMPPrivPassphrase = config.SNMP.SNMPPrivPassphrase
		}
	}

	return &config
}

// Loads the configuration from the config file
func readConfigFile() []byte {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	return configData
}
