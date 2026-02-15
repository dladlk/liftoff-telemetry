package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

type Writer struct {
	logFile     *os.File
	binWriteBuf *bytes.Buffer
	binFormat   bool
	config      *Config
	lotConfig   *lot_config.LiftoffTelemetryConfig
}

func (t *Writer) Start(config *Config, lotConfig *lot_config.LiftoffTelemetryConfig) {
	t.binFormat = config.General.Format == "bin"
	t.config = config
	t.lotConfig = lotConfig

	writeLogToFileExtension := ".csv"
	if t.binFormat {
		writeLogToFileExtension = ".bin"
	}
	writeLogToFile := fmt.Sprintf("liftoff_telemetry_%s", time.Now().Format("20060102_150405")+writeLogToFileExtension)
	logFile, err := os.OpenFile(writeLogToFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file %s: %v", writeLogToFile, err)
	}
	t.logFile = logFile
	t.writeHeader()
	t.binWriteBuf = new(bytes.Buffer)
}

func (t *Writer) Restart() {
	t.Close()
	t.Start(t.config, t.lotConfig)
}

func (t *Writer) Close() {
	log.Printf("Session is written to file %s", t.logFile.Name())
	t.logFile.Close()
}

func (t *Writer) writeHeader() {
	var headerBuffer bytes.Buffer // Declare a bytes.Buffer
	for i, name := range t.lotConfig.StreamFormatNames {
		if i > 0 {
			headerBuffer.WriteString(",")
		}
		headerBuffer.WriteString(name)
	}
	headerBuffer.WriteString("\n")
	t.logFile.Write(headerBuffer.Bytes())
}

func (t *Writer) Write(cur *lot_config.Datagram, curSession *Trip) {
	if t.binFormat {
		for _, f := range t.lotConfig.StreamFormats {
			t.binWriteBuf.Reset()
			switch f {
			case lot_config.Timestamp:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Timestamp)
			case lot_config.Position:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Position)
			case lot_config.Attitude:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Attitude)
			case lot_config.Velocity:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Velocity)
			case lot_config.Gyro:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Gyro)
			case lot_config.Input:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Input)
			case lot_config.Battery:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Battery)
			case lot_config.MotorRPM:
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.Motors)
				binary.Write(t.binWriteBuf, binary.LittleEndian, cur.MotorRPM)
			}
			t.logFile.Write(t.binWriteBuf.Bytes())
		}
	} else {
		fmt.Fprintf(t.logFile, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n", curSession.Index, curSession.Events, cur.Timestamp, cur.Position, cur.Attitude, cur.Velocity, cur.Gyro, cur.Input, cur.MotorRPM)
	}

}
