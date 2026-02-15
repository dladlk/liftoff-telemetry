package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

const CIRCLE_DISTANCE_TO_START = 3

type Trip struct {
	Type            string
	Start           time.Time
	End             time.Time
	DurationSeconds int
	Events          int32
	Index           int
	MaxDistance     float64
	MaxVelocity     float32
	TripDistance    float64
}

func (this *Trip) Report() {
	this.End = time.Now()
	duration := this.End.Sub(this.Start)
	this.DurationSeconds = int(duration.Seconds())
	log.Printf("%s #%d: %v (%ds), %d events, total %.1f, max velocity: %.2f m/s, max from start: %.1f",
		this.Type, this.Index, duration.Round(time.Second), this.DurationSeconds, this.Events, this.TripDistance, this.MaxVelocity, this.MaxDistance)
}

func main() {
	log.SetPrefix("")
	log.SetFlags(log.Ltime | log.Ldate)

	config, err := LoadConfig("liftoff-telemetry.toml.ini")
	if err != nil {
		log.Fatalf("Failed to read app config file liftoff-telemetry.toml.ini: %v", err)
	}
	log.Printf("Liftoff Telemetry Listener config: %+v", config)

	debug := config.Log.Debug

	if config.Log.LogToFile {
		logFile, err := os.OpenFile(os.Args[0]+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		defer logFile.Close() // Ensure the file is closed when the program exits
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}

	var curSession Trip
	var curCircle Trip

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

	lotConfig, err := lot_config.ReadLiftoffTelemetryConfig()
	if err != nil {
		log.Fatalf("Failed to read telemetry configuration: %v", err)
	}
	log.Printf("Found Liftoff Telemetry Config: %+v \n", lotConfig)

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

	binFormat := config.General.Format == "bin"

	writeLogToFileExtension := ".csv"
	if binFormat {
		writeLogToFileExtension = ".bin"
	}
	writeLogToFile := fmt.Sprintf("liftoff_telemetry_%s", time.Now().Format("20060102_150405")+writeLogToFileExtension)
	logFile, err := os.OpenFile(writeLogToFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file %s: %v", writeLogToFile, err)
	}
	defer logFile.Close()
	writeHeader(lotConfig, logFile)
	binWriteBuf := new(bytes.Buffer)

	buffer := make([]byte, 1024)
	var prev *lot_config.Datagram

	curSession = Trip{Type: "Race", Start: time.Now(), Index: 1}
	curCircle = Trip{Type: "Circle", Start: time.Now(), Index: 1}
	defer curSession.Report()
	curSessionReported := false
	var firstEvent *lot_config.Datagram = nil

	expectedBlockLength := int(lot_config.CalculateBlockLength(lotConfig.StreamFormats))

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error from %s: %v\n", clientAddr, err)
			continue
		}
		if debug {
			log.Printf("Received %d bytes from %s\n", n, clientAddr)
		}
		if n == expectedBlockLength {
			buf := bytes.NewReader(buffer[:n])
			curSession.Events++
			curCircle.Events++

			if config.General.SaveEachNth > 0 && (curSession.Events-1)%config.General.SaveEachNth != 0 {
				continue
			}

			cur := lot_config.Datagram{}
			cur.ParseDatagram(buf, &lotConfig.StreamFormats)

			if cur.ZeroPosition() {
				if curSessionReported {
					// Ignore telemetry with zero position after reporting - wait for restart
					continue
				}
			} else {
				if curSessionReported {
					curSessionReported = false
					// Discard previous session data - it was an empty, fake session after race finished until new started
					curSession = Trip{Type: curSession.Type, Start: time.Now(), Index: curSession.Index}
				}
			}

			var distance float64

			if firstEvent == nil {
				firstEvent = &cur
			} else {
				if lotConfig.HasPosition() {
					distance = cur.DistanceFrom(firstEvent)
					if distance > curSession.MaxDistance {
						curSession.MaxDistance = distance
					}

					if distance > curCircle.MaxDistance {
						curCircle.MaxDistance = distance
					}
				}
			}

			if lotConfig.HasVelocity() {
				for i := range cur.Velocity {
					if curSession.MaxVelocity < cur.Velocity[i] {
						curSession.MaxVelocity = cur.Velocity[i]
					}
					if curCircle.MaxVelocity < cur.Velocity[i] {
						curCircle.MaxVelocity = cur.Velocity[i]
					}
				}
			}

			if prev != nil {
				if lotConfig.HasPosition() {
					curSession.TripDistance += cur.DistanceFrom(prev)
					curCircle.TripDistance += cur.DistanceFrom(prev)
				}

				// When we restart race - get timestamp less than before
				if prev.Timestamp > cur.Timestamp ||
					//	When race is finished, zero position is constantly sent
					(lotConfig.HasPosition() && prev.ZeroPosition() && cur.ZeroPosition()) {
					curSession.Report()
					curSessionReported = true
					curSession = Trip{Type: "Race", Start: time.Now(), Index: curSession.Index + 1}
					curCircle = Trip{Type: "Circle", Start: time.Now(), Index: 1}
					firstEvent = &cur
				}

				if lotConfig.HasPosition() {
					// Let's say that we did a circle if distance from start point is less than some value AND current cicle max distance is bigger then current 50 times
					if distance < CIRCLE_DISTANCE_TO_START && curCircle.TripDistance > 100 && (curCircle.TripDistance/curCircle.MaxDistance+0.1) > 2 {
						curCircle.End = time.Now()
						curCircle.DurationSeconds = int(curCircle.End.Sub(curCircle.Start).Round(time.Second))
						curCircle.Report()

						curCircle = Trip{Type: curCircle.Type, Start: time.Now(), Index: curCircle.Index + 1}
					}
				}
			}

			if binFormat {
				for _, f := range lotConfig.StreamFormats {
					binWriteBuf.Reset()
					switch f {
					case lot_config.Timestamp:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Timestamp)
					case lot_config.Position:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Position)
					case lot_config.Attitude:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Attitude)
					case lot_config.Velocity:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Velocity)
					case lot_config.Gyro:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Gyro)
					case lot_config.Input:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Input)
					case lot_config.Battery:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Battery)
					case lot_config.MotorRPM:
						binary.Write(binWriteBuf, binary.LittleEndian, cur.Motors)
						binary.Write(binWriteBuf, binary.LittleEndian, cur.MotorRPM)
					}
					logFile.Write(binWriteBuf.Bytes())
				}
			} else {
				fmt.Fprintf(logFile, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n", curSession.Index, curSession.Events, cur.Timestamp, cur.Position, cur.Attitude, cur.Velocity, cur.Gyro, cur.Input, cur.MotorRPM)
			}

			if debug {
				log.Printf("%+v", cur)
			}
			prev = &cur
		} else {
			log.Fatalf("Received unexpected UDP block length %d instead of %d", n, expectedBlockLength)
		}
	}
}

func writeHeader(lotConfig *lot_config.LiftoffTelemetryConfig, logFile *os.File) {
	var headerBuffer bytes.Buffer // Declare a bytes.Buffer
	for i, name := range lotConfig.StreamFormatNames {
		if i > 0 {
			headerBuffer.WriteString(",")
		}
		headerBuffer.WriteString(name)
	}
	headerBuffer.WriteString("\n")
	logFile.Write(headerBuffer.Bytes())
}
