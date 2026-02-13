package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
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
}

// UDP Server to get Litfoff Telemtry
// https://steamcommunity.com/sharedfiles/filedetails/?id=3160488434

type Datagram struct {
	Timestamp float32    `desc:"seconds"`
	Position  [3]float32 `desc:"3d coordinate, X, Y, Z"`
	Attitude  [4]float32 `desc:"X, Y, Z, W"`
	Velocity  [3]float32 `desc:"meters/second, X, Y, Z (world space, https://steamcommunity.com/linkfilter/?u=https%3A%2F%2Fmath.stackexchange.com%2Fa%2F3209449 )"`
	Gyro      [3]float32 `desc:"angular velocity rates - pitch, roll, yaw in degrees/second"`
	Input     [4]float32 `desc:"throttle, yaw, pitch, roll"`
	Battery   [2]float32 `desc:"remaining voltage and charge percentage"`
	Motors    byte       `desc:"number of motors"`
	MotorRPM  []float32  `desc:"rpm per each motor"`
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

func (t *Track) Open(path string) ([]Datagram, error) {
	t.path = path

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	header, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return nil, errors.New("Failed to read first line as header")
	}
	header = strings.TrimSpace(header)
	t.fields = parseFormats(strings.Split(header, ","))

	blockLength := CalculateBlockLength(t.fields)

	fmt.Printf("File header: %s, block length: %d\n", header, blockLength)

	buffer := make([]byte, blockLength)

	blocks := 0

	var minTs float32 = math.MaxFloat32
	var maxTs float32 = 0

	list := []Datagram{}

	for {
		n, err := io.ReadFull(reader, buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		if n != int(blockLength) {
			return nil, fmt.Errorf("Expected to read %d bytes, but read only %d", blockLength, n)
		}
		blocks++
		data := Datagram{}
		readDatagram(buffer, n, t, &data)

		list = append(list, data)

		if data.Timestamp > maxTs {
			maxTs = data.Timestamp
		}
		if data.Timestamp < minTs {
			maxTs = data.Timestamp
		}

	}
	fmt.Printf("Loaded %d blocks, min ts %v, max ts %v\n", blocks, minTs, maxTs)

	return list, nil
}

func readDatagram(buffer []byte, n int, t *Track, cur *Datagram) {
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
