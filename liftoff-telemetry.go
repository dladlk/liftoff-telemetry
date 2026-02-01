package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// UDP Server to get Litfoff Telemtry
// https://steamcommunity.com/sharedfiles/filedetails/?id=3160488434

type Datagram struct {
	Timestamp float32    `desc:"seconds"`
	Position  [3]float32 `desc:"3d coordinate, X, Y, Z"`
	Attitude  [4]float32 `desc:"X, Y, Z, W"`
	Velocity  [3]float32 `desc:"X, Y, Z (world space, https://steamcommunity.com/linkfilter/?u=https%3A%2F%2Fmath.stackexchange.com%2Fa%2F3209449 )"`
	Gyro      [3]float32 `desc:"angular velocity rates - pitch, roll, yaw in degrees/second"`
	Input     [4]float32 `desc:"throttle, yaw, pitch, roll"`
	Battery   [2]float32 `desc:"remaining voltage and charge percentage"`
	Motors    byte       `desc:"number of motors"`
	MotorRPM  []float32  `desc:"rpm per each motor"`
}

func (d *Datagram) ZeroPosition() bool {
	return d.Position[0] == 0 && d.Position[1] == 0 && d.Position[2] == 0
}

type Session struct {
	Start           time.Time
	End             time.Time
	DurationSeconds int
	Events          int32
	Attempt         int
}

func (this *Session) Report() {
	this.End = time.Now()
	this.DurationSeconds = int(this.End.Sub(this.Start).Seconds())
	log.Printf("Finished attempt %d after %d seconds and %d events", this.Attempt, this.DurationSeconds, this.Events)
}

func main() {
	const debug = false
	const logToFile = true
	const logEachNth = 100

	log.SetPrefix("")
	log.SetFlags(log.Ltime)

	if logToFile {
		logFile, err := os.OpenFile(os.Args[0]+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		defer logFile.Close() // Ensure the file is closed when the program exits
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}

	address, err := net.ResolveUDPAddr("udp", ":9001")
	if err != nil {
		log.Fatal("Error resolving UDP address:", err)
	}
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatal("Error listening: ", err)
	}
	defer conn.Close() // Ensure the connection is closed when the function exits

	log.Printf("Liftoff Telemetry UDP server listening on %s\n", conn.LocalAddr().String())

	writeLogToFile := fmt.Sprintf("liftoff_telemetry_%s", time.Now().Format("20060102_150405")+".csv")
	logFile, err := os.OpenFile(writeLogToFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file %s: %v", writeLogToFile, err)
	}
	defer logFile.Close()

	buffer := make([]byte, 1024)
	var prev *Datagram

	curSession := Session{Start: time.Now(), Attempt: 1}
	defer curSession.Report()
	curSessionReported := false

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error from %s: %v\n", clientAddr, err)
			continue
		}
		if debug {
			log.Printf("Received %d bytes from %s\n", n, clientAddr)
		}
		if n == 97 {
			buf := bytes.NewReader(buffer[:n])
			curSession.Events++

			if logEachNth > 0 && (curSession.Events-1)%logEachNth != 0 {
				continue
			}

			cur := Datagram{}
			order := binary.LittleEndian
			if err := binary.Read(buf, order, &cur.Timestamp); err != nil {
				log.Fatalf("Failed to read Timestamp as float: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Position); err != nil {
				log.Fatalf("Failed to read Position as float[3]: %s\n", err)
			}

			if cur.ZeroPosition() {
				if curSessionReported {
					// Ignore telemetry with zero position after reporting - wait for restart
					continue
				}
			} else {
				if curSessionReported {
					curSessionReported = false
				}
			}

			if err := binary.Read(buf, order, &cur.Attitude); err != nil {
				log.Fatalf("Failed to read Attitude as float[4]: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Velocity); err != nil {
				log.Fatalf("Failed to read Velocity as float[3]: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Gyro); err != nil {
				log.Fatalf("Failed to read Gyro as float[3]: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Input); err != nil {
				log.Fatalf("Failed to read Input as float[4]: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Battery); err != nil {
				log.Fatalf("Failed to read Battery as float[2]: %s\n", err)
			}
			if err := binary.Read(buf, order, &cur.Motors); err != nil {
				log.Fatalf("Failed to read Motors as byte: %s\n", err)
			}
			cur.MotorRPM = make([]float32, cur.Motors)
			if err := binary.Read(buf, order, &cur.MotorRPM); err != nil {
				log.Fatalf("Failed to read MotorRPM as float[%d]: %s\n", cur.Motors, err)
			}

			if prev != nil {
				if prev.Timestamp > cur.Timestamp ||
					//	When race is finished, zero position is constantly sent
					(prev.ZeroPosition() && cur.ZeroPosition()) {
					curSession.Report()
					curSessionReported = true
					curSession = Session{Start: time.Now(), Attempt: curSession.Attempt + 1}
				} else {
					if curSession.Events%1000 == 0 {
						log.Printf("Received %v events in session #%v", curSession.Events, curSession.Attempt)
					}
				}
			}

			fmt.Fprintf(logFile, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n", curSession.Attempt, curSession.Events, cur.Timestamp, cur.Position, cur.Attitude, cur.Velocity, cur.Gyro, cur.Input, cur.MotorRPM)

			if debug {
				log.Printf("%+v", cur)
			}
			prev = &cur
		}
	}
}
