package lot_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type LiftoffTelemetryConfig struct {
	EndPoint     string   `json:"EndPoint"`
	StreamFormat []string `json:"StreamFormat"`
}

func ReadConfig() (*LiftoffTelemetryConfig, error) {
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

	return &config, nil
}
