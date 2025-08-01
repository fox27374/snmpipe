package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	g "github.com/gosnmp/gosnmp"
)

func trapReceiver(tc TrapConfig) {
	secParamsList := []*g.UsmSecurityParameters{
		{
			UserName:                 tc.TrapUser,
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: tc.TrapAuthPassphrase,
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        tc.TrapPrivPassphrase,
		},
	}

	tl := g.NewTrapListener()
	tl.OnNewTrap = trapHandler

	usmTable := g.NewSnmpV3SecurityParametersTable(g.NewLogger(log.New(os.Stdout, "", 0)))
	for _, sp := range secParamsList {
		err := usmTable.Add(sp.UserName, sp)
		if err != nil {
			usmTable.Logger.Print(err)
		}
	}

	gs := &g.GoSNMP{
		Port:                        uint16(trapPort),
		Transport:                   "udp",
		Version:                     g.Version3, // Always using version3 for traps, only option that works with all SNMP versions simultaneously
		SecurityModel:               g.UserSecurityModel,
		SecurityParameters:          &g.UsmSecurityParameters{AuthoritativeEngineID: "12345"}, // Use for server's engine ID
		TrapSecurityParametersTable: usmTable,
	}
	tl.Params = gs
	// tl.Params.Logger = g.NewLogger(log.New(os.Stdout, "", 0))

	listenErr := tl.Listen("0.0.0.0:" + tc.TrapPort)
	if listenErr != nil {
		log.Panicf("error in listen: %s", listenErr)
	}
}

func trapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	s := make(SNMPData)
	t := make(map[string]string)

	// Add trap metadata to the map.
	s["ip"] = addr.IP.String()

	if len(packet.Variables) > 0 {
		// Get the last variable binding, which contains the main message
		v := packet.Variables[len(packet.Variables)-1]

		// Add the OID to the map.
		t["oid"] = v.Name

		// Convert the value to a string
		var valueAsString string
		switch v.Type {
		case g.OctetString:
			// Convert OctetString (byte array) to string
			valueAsString = string(v.Value.([]byte))
		case g.ObjectIdentifier:
			// Convert OIDs to string
			valueAsString = v.Value.(string)
		default:
			// Convert all other types to string
			valueAsString = fmt.Sprintf("%v", v.Value)
		}

		t["value"] = valueAsString
	}

	s["values"] = t

	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	fmt.Println(string(jsonData))

	// Put data in a slice to be compatible with the poll functions
	sendData := []SNMPData{s}

	err = sendToSplunkHec(sendData)
	if err != nil {
		fmt.Println(err)
	}
}
