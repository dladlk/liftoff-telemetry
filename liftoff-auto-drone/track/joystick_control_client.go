package track

import (
	"encoding/binary"
	"net"
)

type JoystickControlClient struct {
	conn net.Conn
}

func (j *JoystickControlClient) Start(serverAddress string) error {
	conn, err := net.Dial("udp", serverAddress)
	if err != nil {
		return err
	}
	j.conn = conn
	return nil
}

func (j *JoystickControlClient) Close() {
	j.conn.Close()
}

func (j *JoystickControlClient) Send(lv, lh, rv, rh int16) error {
	// 2. Prepare 4 int16 values (8 bytes total)
	data := []int16{lv, lv, rv, rh}

	// 8 bytes = 4 * 2 bytes per int16
	buf := make([]byte, len(data)*2)

	for i, v := range data {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}

	// 4. Send the bytes
	_, err := j.conn.Write(buf)
	return err
}
