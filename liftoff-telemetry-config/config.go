package lot_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

type LiftoffTelemetryConfig struct {
	Endpoint          string   `json:"EndPoint"`
	StreamFormatNames []string `json:"StreamFormat"`
	StreamFormats     []StreamDataType
	StreamFormatsMap  map[StreamDataType]string
}

func (t LiftoffTelemetryConfig) HasStreamDataType(dataType StreamDataType) bool {
	_, ok := t.StreamFormatsMap[dataType]
	return ok
}

func (t LiftoffTelemetryConfig) HasPosition() bool {
	return t.HasStreamDataType(Position)
}
func (t LiftoffTelemetryConfig) HasVelocity() bool {
	return t.HasStreamDataType(Velocity)
}

func (t *LiftoffTelemetryConfig) UpdateStreamFormats() {
	t.StreamFormats = make([]StreamDataType, len(t.StreamFormatNames))
	t.StreamFormatsMap = map[StreamDataType]string{}
	for i, name := range t.StreamFormatNames {
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
		t.StreamFormats[i] = ts
		t.StreamFormatsMap[ts] = name
	}
}

func ReadLiftoffTelemetryConfig() (*LiftoffTelemetryConfig, error) {
	userProfile := os.Getenv("USERPROFILE") // For Windows, use "HOME" or similar for Unix
	if userProfile == "" {
		return nil, errors.New("Cannot resolve %USERPROFILE% env variable")
	}
	telemetryConfigurationPath := fmt.Sprintf(`%s\AppData\LocalLow\LuGus Studios\Liftoff\TelemetryConfiguration.json`, userProfile)
	file, err := os.Open(telemetryConfigurationPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config LiftoffTelemetryConfig
	json.Unmarshal(bytes, &config)
	config.UpdateStreamFormats()

	return &config, nil
}
