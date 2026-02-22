package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func RunJoystickControlServer(address string, drone IDrone, debug bool) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Couldn't resolve address: %v", err)
	}

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	// Notify the channel of SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle the signal
	go func() {
		<-signalChan // Block until a signal is received
		// TODO: Shutdown gracefully somehow...
		os.Exit(0) // Exit gracefully after the command finishes
	}()

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Listen failed: %v", err)
	}
	defer conn.Close() // Ensure the connection is closed when the function exits

	fmt.Printf("UDP server listening on %s\n", conn.LocalAddr().String())

	// Buffer for incoming data
	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Read error: %v", err)
			continue
		}

		reader := bytes.NewReader(buffer[:n])
		order := binary.LittleEndian
		joystickPosition := [4]int16{}
		err = binary.Read(reader, order, &joystickPosition)
		if err != nil {
			fmt.Printf("Failed to parse joystick position from client %v: %v\n", clientAddr, err)
		}
		if debug {
			fmt.Printf("Received %d bytes of joystick position: %v", n, joystickPosition)
		}

		// Receive lv, lh, rv, rh - so should swap pairs to get lx, ly, rx, ry
		// TODO: Maybe we need to change sing of rx - like it was done when sending Input
		drone.UpdateDirect(joystickPosition[1], joystickPosition[0], joystickPosition[3], joystickPosition[2])
	}
}
