package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Telemetry struct {
	Name    string
	Records []TelemetryRecord
}

func (t *Telemetry) Add(input [4]float32, timestamp float32) TelemetryRecord {
	command := TelemetryRecord{Input: input, Timestamp: timestamp}
	t.Records = append(t.Records, command)
	return command
}

type TelemetryRecord struct {
	Timestamp float32
	Input     [4]float32
}

func ReadTelemetry(path string) (*Telemetry, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	telemetry := Telemetry{Name: path}

	lineIndex := -1
	for scanner.Scan() {
		lineIndex++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Skip comments
			continue
		}
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 9 {
			return nil, fmt.Errorf("Line %d is invalid, expected 9 values separated by COMMA, but found: %s", lineIndex, line)
		}

		timestamp, err := strconv.ParseFloat(parts[2], 32)
		if err != nil {
			return nil, fmt.Errorf("Line %d is invalid, value %d is not a valid float32: %s", lineIndex, 2, parts[2])
		}

		inputValue := [4]float32{}

		trimmed := strings.Trim(parts[7], "[]")
		inputParts := strings.Split(trimmed, " ")

		for i := range inputValue {
			val, err := strconv.ParseFloat(inputParts[i], 32)
			if err != nil {
				return nil, fmt.Errorf("Line %d is invalid, value of inputs at index %d is not a valid integer: %s", lineIndex, (i + 1), parts[7])
			}
			inputValue[i] = float32(val)
		}
		telemetry.Add(inputValue, float32(timestamp))
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	return &telemetry, nil
}
