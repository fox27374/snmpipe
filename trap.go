package main

import (
	"fmt"
	"log/slog"
	"net"

	g "github.com/gosnmp/gosnmp"
)

// Implement gosnmp logger interface in order to use the projects
// slog logger interface.
//
//	type LoggerInterface interface {
//		Print(v ...interface{})
//		Printf(format string, v ...interface{})
//	}
type SlogLogger struct {
	logger *slog.Logger
}

func (l *SlogLogger) Print(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *SlogLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}

func trapReceiver(tc TrapConfig, dataChan chan<- SNMPData, errChan chan<- error) {
	gosnmpLogger := &SlogLogger{logger: logger}
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

	// Use anon function for the trapHandler in order to pass the errorChan channel
	tl.OnNewTrap = func(packet *g.SnmpPacket, addr *net.UDPAddr) {
		sendData, err := trapHandler(packet, addr)
		if err != nil {
			errChan <- fmt.Errorf("trap handling failed: %v", err)
		}
		dataChan <- sendData
	}

	usmTable := g.NewSnmpV3SecurityParametersTable(g.NewLogger(gosnmpLogger))
	for _, sp := range secParamsList {
		err := usmTable.Add(sp.UserName, sp)
		if err != nil {
			errChan <- fmt.Errorf("failed to set security parameter: %v", err)
			return
		}
	}

	gs := &g.GoSNMP{
		Port:                        uint16(trapPort),
		Transport:                   "udp",
		Version:                     g.Version3, // Always using version3 for traps, only option that works with all SNMP versions simultaneously
		SecurityModel:               g.UserSecurityModel,
		SecurityParameters:          &g.UsmSecurityParameters{AuthoritativeEngineID: "snmpipe"},
		TrapSecurityParametersTable: usmTable,
	}
	tl.Params = gs
	tl.Params.Logger = g.NewLogger(gosnmpLogger)

	// Start listener - hard blocking function
	err := tl.Listen(trapListenAddr + ":" + tc.TrapPort)
	if err != nil {
		errChan <- fmt.Errorf("error creating listener: %v", err)
	}
}

func trapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) (SNMPData, error) {
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

	return s, nil
}
