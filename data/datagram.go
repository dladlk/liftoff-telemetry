package lot_config

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
)

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

func (d Datagram) DistanceFrom(firstEvent *Datagram) float64 {
	a := firstEvent.Position
	b := d.Position

	return math.Sqrt(math.Pow(float64(a[0]-b[0]), 2) + math.Pow(float64(a[2]-b[2]), 2))
}

func (d *Datagram) ZeroPosition() bool {
	return d.Position[0] == 0 && d.Position[1] == 0 && d.Position[2] == 0
}

func (cur *Datagram) ParseDatagram(reader *bytes.Reader, fields *[]StreamDataType) {
	order := binary.LittleEndian

	for _, dataType := range *fields {
		switch dataType {
		case Timestamp:
			if err := binary.Read(reader, order, &cur.Timestamp); err != nil {
				log.Fatalf("Failed to read Timestamp as float: %s\n", err)
			}
		case Position:
			if err := binary.Read(reader, order, &cur.Position); err != nil {
				log.Fatalf("Failed to read Position as float[3]: %s\n", err)
			}
		case Attitude:
			if err := binary.Read(reader, order, &cur.Attitude); err != nil {
				log.Fatalf("Failed to read Attitude as float[4]: %s\n", err)
			}
		case Velocity:
			if err := binary.Read(reader, order, &cur.Velocity); err != nil {
				log.Fatalf("Failed to read Velocity as float[3]: %s\n", err)
			}
		case Gyro:
			if err := binary.Read(reader, order, &cur.Gyro); err != nil {
				log.Fatalf("Failed to read Gyro as float[3]: %s\n", err)
			}
		case Input:
			if err := binary.Read(reader, order, &cur.Input); err != nil {
				log.Fatalf("Failed to read Input as float[4]: %s\n", err)
			}
		case Battery:
			if err := binary.Read(reader, order, &cur.Battery); err != nil {
				log.Fatalf("Failed to read Battery as float[2]: %s\n", err)
			}
		case MotorRPM:
			if err := binary.Read(reader, order, &cur.Motors); err != nil {
				log.Fatalf("Failed to read Motors as byte: %s\n", err)
			}
			cur.MotorRPM = make([]float32, cur.Motors)
			if err := binary.Read(reader, order, &cur.MotorRPM); err != nil {
				log.Fatalf("Failed to read MotorRPM as float[%d]: %s\n", cur.Motors, err)
			}
		}
	}
}
