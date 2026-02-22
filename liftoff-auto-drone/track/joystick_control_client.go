package track

import (
	"encoding/binary"
	"fmt"
	"net"
)

type JoystickControlClient struct {
	conn  net.Conn
	debug bool
}

func (j *JoystickControlClient) Start(serverAddress string, debug bool) error {
	conn, err := net.Dial("udp", serverAddress)
	if err != nil {
		return err
	}
	j.conn = conn
	j.debug = debug
	return nil
}

func (j *JoystickControlClient) Close() {
	j.conn.Close()
}

func (j *JoystickControlClient) Send(lv, lh, rv, rh int16) error {
	// Prepare 4 int16 values (8 bytes total)
	data := []int16{lv, lh, rv, rh}
	buf := make([]byte, len(data)*2)
	for i, v := range data {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}
	n, err := j.conn.Write(buf)
	if n != len(buf) {
		return fmt.Errorf("Expected to write %d bytes but written %d", len(buf), n)
	}
	if j.debug {
		if err == nil {
			fmt.Printf("Sent %d %d %d %d as %d bytes\n", lv, lh, rv, rh, n)
		} else {
			fmt.Printf("Error on sending: %v\n", err)
		}
	}
	return err
}
