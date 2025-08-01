package main

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	g "github.com/gosnmp/gosnmp"
)

// Poll an SNMP device using the gosnmp package
// Supports SNMP version 1, 2c and 3 as well as different AUTH and PRIV protocols
// Poll results and errors are passed into separate channels
func pollDevice(device DeviceConfig, dataChan chan<- SNMPData, errChan chan<- error) {

	p := make(SNMPData)
	o := make(map[string]any)

	p["ip"] = device.IP
	p["name"] = device.Name

	snmpPort, portErr := strconv.Atoi(device.SNMPPort)
	if portErr != nil {
		errChan <- fmt.Errorf("INVALID SNMP-PORT %w FOR %s (IP: %s)", portErr, device.Name, device.IP)
		return
	}

	// Configure gosnmp
	params := &g.GoSNMP{
		Target:  device.IP,
		Port:    uint16(snmpPort),
		Timeout: time.Duration(snmpTimeout) * time.Second,
	}

	// Configure version specific settings
	switch device.SNMPVersion {
	case "2":
		params.Version = g.Version2c
		params.Community = device.SNMPCommunity

	case "3":
		params.Version = g.Version3
		params.SecurityModel = g.UserSecurityModel
		params.MsgFlags = g.AuthPriv
		params.SecurityParameters = &g.UsmSecurityParameters{
			UserName:                 device.SNMPUser,
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: device.SNMPAuthPassphrase,
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        device.SNMPPrivPassphrase,
		}
	default:
		errChan <- fmt.Errorf("SNMP-VERSION %v NOT SUPPORTED FOR %s (IP: %s)", device.SNMPVersion, device.Name, device.IP)
		dataChan <- nil
		return
	}

	// Try connecting to the device
	ConnErr := params.Connect()
	if ConnErr != nil {
		errChan <- fmt.Errorf("Connect() err for device %s (IP: %s): %w", device.Name, device.IP, ConnErr)
		dataChan <- nil
		return
	}
	defer params.Conn.Close()

	// Create slice of OIDs to pass to the Get function
	oids := slices.Collect(maps.Keys(device.OIDs))

	// Get the devices SNMP values
	result, getErr := params.Get(oids)
	if getErr != nil {
		errChan <- fmt.Errorf("Get() err for device %s (IP: %s): %w", device.Name, device.IP, getErr)
		dataChan <- nil
		return
	}

	// Write the SNMP oid name and value to the oid map
	for _, variable := range result.Variables {
		o[device.OIDs[variable.Name]] = variable.Value
	}

	// Attach the oids map to the SNMPData struct
	p["values"] = o

	errChan <- nil
	dataChan <- p
}
