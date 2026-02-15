package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

type TelemetryListener struct {
	running        bool
	lotConfig      lot_config.LiftoffTelemetryConfig
	conn           *net.UDPConn
	expected       int
	lastBytes      []byte
	lastBytesIndex int
	mu             sync.Mutex
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
		t.conn = conn

		fmt.Printf("\r\nStarted telemetry listener on %v by config %+v\n", lotConfig.Endpoint, lotConfig)

		expectedBlockLength := int(lot_config.CalculateBlockLength(lotConfig.StreamFormats))
		t.expected = expectedBlockLength
		t.running = true

		go func() {
			buffer := make([]byte, expectedBlockLength)

			for {
				if !t.running {
					break
				}
				n, clientAddr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					log.Fatalf("Read error from %s: %v\n", clientAddr, err)
				}
				if n == expectedBlockLength {
					copiedBytes := make([]byte, expectedBlockLength)
					copied := copy(copiedBytes, buffer)
					if copied != n {
						log.Fatalf("Expected to copy %d, but copied only %d", n, copied)
					}

					t.mu.Lock()
					t.lastBytes = copiedBytes
					t.lastBytesIndex++

					if t.lastBytesIndex%10 == 0 {
						var input [4]float32
						if err := binary.Read(bytes.NewReader(copiedBytes), binary.LittleEndian, &input); err != nil {
							log.Fatalf("Failed to read Input as float[4]: %s\n", err)
						}
						inputStr := fmt.Sprintf("[%d] %.6f %.6f %.6f %.6f", t.lastBytesIndex, input[0], input[1], input[2], input[3])
						fmt.Printf("\r\n%s", inputStr)
					}

					t.mu.Unlock()
				} else {
					log.Fatalf("Unexpected block length %d, expected %d", n, expectedBlockLength)
				}
			}

		}()
	} else {
		t.running = false
		t.conn.Close()
		t.lastBytes = nil
		t.lastBytesIndex = 0
		fmt.Printf("\r\nStopped telemetry listener\n")
	}
}

func (t *TelemetryListener) LastDatagram() (*lot_config.Datagram, int, bool) {
	if t.running {
		if t.lastBytesIndex > 0 {
			var lastBytes []byte = make([]byte, t.expected)
			t.mu.Lock()
			index := t.lastBytesIndex
			copy(lastBytes, t.lastBytes)
			t.mu.Unlock()

			//fmt.Printf("\r\nParsed %d: %v\n", t.lastBytesIndex, lastBytes)

			res := &lot_config.Datagram{}
			res.ParseDatagram(bytes.NewReader(lastBytes), &t.lotConfig.StreamFormats)
			return res, index, true
		}
	}
	return nil, 0, false
}
