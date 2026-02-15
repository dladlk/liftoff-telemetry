package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

type StreamDataType int

const (
	Timestamp StreamDataType = iota
	Position
	Attitude
	Velocity
	Gyro
	Input
	Battery
	MotorRPM
	Unknown
)

func parseFormats(formatNames []string) []StreamDataType {
	streamFormats := make([]StreamDataType, len(formatNames))
	for i, name := range formatNames {
		var ts StreamDataType
		switch name {
		case "Timestamp":
			ts = Timestamp
		case "Position":
			ts = Position
		case "Attitude":
			ts = Attitude
		case "Velocity":
			ts = Velocity
		case "Gyro":
			ts = Gyro
		case "Input":
			ts = Input
		case "Battery":
			ts = Battery
		case "MotorRPM":
			ts = MotorRPM
		default:
			ts = Unknown
		}
		streamFormats[i] = ts
	}
	return streamFormats
}

type Track struct {
	path   string
	fields []StreamDataType
	List   []lot_config.Datagram
	minTs  float32
	maxTs  float32
}

const (
	FLOAT32 int8 = 4
	BYTE    int8 = 1
)

func CalculateBlockLength(fields []StreamDataType) int8 {
	var length int8 = 0
	for _, field := range fields {
		var fieldLength int8
		switch field {
		case Timestamp:
			fieldLength = FLOAT32
		case Position, Velocity, Gyro:
			fieldLength = 3 * FLOAT32
		case Attitude, Input:
			fieldLength = 4 * FLOAT32
		case Battery:
			fieldLength = 2 * FLOAT32
		case MotorRPM:
			fieldLength = BYTE + 4*FLOAT32
		}
		length += fieldLength
	}
	return length
}

func (t *Track) Open(path string) error {
	t.path = path

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	header, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return errors.New("Failed to read first line as header")
	}
	header = strings.TrimSpace(header)
	t.fields = parseFormats(strings.Split(header, ","))

	blockLength := CalculateBlockLength(t.fields)

	fmt.Printf("File header: %s, block length: %d\r\n", header, blockLength)

	buffer := make([]byte, blockLength)

	blocks := 0

	for {
		n, err := io.ReadFull(reader, buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if n != int(blockLength) {
			return fmt.Errorf("Expected to read %d bytes, but read only %d", blockLength, n)
		}
		blocks++
		data := lot_config.Datagram{}
		readDatagram(buffer, n, t, &data)

		t.List = append(t.List, data)
	}
	t.minTs = t.List[0].Timestamp
	t.maxTs = t.List[len(t.List)-1].Timestamp
	fmt.Printf("Loaded %d blocks, min ts %.2f sec, max ts %.2f sec\n", blocks, t.minTs, t.maxTs)

	return nil
}

func readDatagram(buffer []byte, n int, t *Track, cur *lot_config.Datagram) {
	buf := bytes.NewReader(buffer[:n])

	order := binary.LittleEndian

	for _, dataType := range t.fields {
		switch dataType {
		case Timestamp:
			if err := binary.Read(buf, order, &cur.Timestamp); err != nil {
				log.Fatalf("Failed to read Timestamp as float: %s\n", err)
			}
		case Position:
			if err := binary.Read(buf, order, &cur.Position); err != nil {
				log.Fatalf("Failed to read Position as float[3]: %s\n", err)
			}
		case Attitude:
			if err := binary.Read(buf, order, &cur.Attitude); err != nil {
				log.Fatalf("Failed to read Attitude as float[4]: %s\n", err)
			}
		case Velocity:
			if err := binary.Read(buf, order, &cur.Velocity); err != nil {
				log.Fatalf("Failed to read Velocity as float[3]: %s\n", err)
			}
		case Gyro:
			if err := binary.Read(buf, order, &cur.Gyro); err != nil {
				log.Fatalf("Failed to read Gyro as float[3]: %s\n", err)
			}
		case Input:
			if err := binary.Read(buf, order, &cur.Input); err != nil {
				log.Fatalf("Failed to read Input as float[4]: %s\n", err)
			}
		case Battery:
			if err := binary.Read(buf, order, &cur.Battery); err != nil {
				log.Fatalf("Failed to read Battery as float[2]: %s\n", err)
			}
		case MotorRPM:
			if err := binary.Read(buf, order, &cur.Motors); err != nil {
				log.Fatalf("Failed to read Motors as byte: %s\n", err)
			}
			cur.MotorRPM = make([]float32, cur.Motors)
			if err := binary.Read(buf, order, &cur.MotorRPM); err != nil {
				log.Fatalf("Failed to read MotorRPM as float[%d]: %s\n", cur.Motors, err)
			}
		}
	}
}
