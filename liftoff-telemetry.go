package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
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

type Session struct {
	Start           time.Time
	End             time.Time
	DurationSeconds int
	Events          int32
	Attempt         int
	MaxDistance     float64
	MaxVelocity     float32
	TripDistance    float64
}

func (this *Session) Report() {
	this.End = time.Now()
	duration := this.End.Sub(this.Start)
	this.DurationSeconds = int(duration.Seconds())
	log.Printf("Session #%d: %v (%ds), %d events, total trip %.1f, max velocity: %.2f m/s, max from start: %.1f",
		this.Attempt, duration.Round(time.Second), this.DurationSeconds, this.Events, this.TripDistance, this.MaxVelocity, this.MaxDistance)
}

func main() {
	const debug = false
	const logToFile = true
	const logEachNth = 100

	log.SetPrefix("")
	log.SetFlags(log.Ltime | log.Ldate)

	if logToFile {
		logFile, err := os.OpenFile(os.Args[0]+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		defer logFile.Close() // Ensure the file is closed when the program exits
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}

	var curSession Session

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	// Notify the channel of SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle the signal
	go func() {
		<-signalChan // Block until a signal is received
		curSession.Report()
		os.Exit(0) // Exit gracefully after the command finishes
	}()

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

	curSession = Session{Start: time.Now(), Attempt: 1}
	defer curSession.Report()
	curSessionReported := false
	var firstEvent *Datagram = nil

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
					// Discard previous session data - it was an empty, fake session after race finished until new started
					curSession = Session{Start: time.Now(), Attempt: curSession.Attempt}
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

			if firstEvent == nil {
				firstEvent = &cur
			} else {
				distance := cur.DistanceFrom(firstEvent)
				if distance > curSession.MaxDistance {
					curSession.MaxDistance = distance
				}
			}

			for i := range cur.Velocity {
				if curSession.MaxVelocity < cur.Velocity[i] {
					curSession.MaxVelocity = cur.Velocity[i]
				}
			}

			if prev != nil {
				curSession.TripDistance += cur.DistanceFrom(prev)

				// When we restart race - get timestamp less than before
				if prev.Timestamp > cur.Timestamp ||
					//	When race is finished, zero position is constantly sent
					(prev.ZeroPosition() && cur.ZeroPosition()) {
					curSession.Report()
					curSessionReported = true
					curSession = Session{Start: time.Now(), Attempt: curSession.Attempt + 1}
					firstEvent = &cur
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
