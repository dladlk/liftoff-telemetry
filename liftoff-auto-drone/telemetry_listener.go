package main

import (
	"bytes"
	"fmt"
	"log"
	"net"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

type TelemetryListener struct {
	running        bool
	lotConfig      lot_config.LiftoffTelemetryConfig
	conn           *net.UDPConn
	lastBytes      *[]byte
	lastBytesIndex int
}

func (t *TelemetryListener) Toggle() {
	if !t.running {
		lotConfig, err := lot_config.ReadLiftoffTelemetryConfig()
		if err != nil {
			log.Fatalf("Failed to read telemetry configuration: %v", err)
		}
		t.lotConfig = *lotConfig

		address, err := net.ResolveUDPAddr("udp", lotConfig.Endpoint)
		if err != nil {
			log.Fatal("Error resolving UDP address:", err)
		}
		conn, err := net.ListenUDP("udp", address)
		if err != nil {
			log.Fatal("Error listening: ", err)
		}

		fmt.Printf("\r\nStarted telemetry listener on %v by config %+v\n", lotConfig.Endpoint, lotConfig)

		expectedBlockLength := int(lot_config.CalculateBlockLength(lotConfig.StreamFormats))
		empty := make([]byte, expectedBlockLength)
		t.lastBytes = &empty

		go func() {
			buffer := make([]byte, 1024)

			for {
				n, clientAddr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					log.Fatalf("Read error from %s: %v\n", clientAddr, err)
					continue
				}
				if n == expectedBlockLength {
					copy(*t.lastBytes, buffer[:n])
					t.lastBytesIndex++
				}
			}

		}()
	} else {
		t.conn.Close()
		t.lastBytes = nil
		t.lastBytesIndex = 0
		fmt.Printf("\r\nStopped telemetry listener\n")
	}
	t.running = !t.running
}

func (t *TelemetryListener) LastDatagram() (*lot_config.Datagram, bool) {
	if t.running {
		if t.lastBytesIndex > 0 {
			res := &lot_config.Datagram{}
			res.ParseDatagram(bytes.NewReader(*t.lastBytes), &t.lotConfig.StreamFormats)
			return res, true
		}
	}
	return nil, false
}
