package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Plan struct {
	Name    string
	List    []Command
	Changed time.Time
}

func (t *Plan) Add(lx int8, ly int8, rx int8, ry int8, during int64) Command {
	command := Command{Update: []int8{lx, ly, rx, ry}, Duration: int(during)}
	t.List = append(t.List, command)
	return command
}

type Command struct {
	Update   []int8
	Duration int
}

func ReadPlan(path string) (*Plan, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	plan := Plan{}

	plan.Changed = fileInfo.ModTime()
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
		if plan.Name == "" {
			plan.Name = line
		} else {
			parts := strings.Split(line, "\t")
			if len(parts) != 5 {
				return nil, fmt.Errorf("Line %d is invalid, expected 5 values separated by TAB, but found: %s", lineIndex, line)
			}
			c := [4]int8{}
			for i := range c {
				val, err := strconv.Atoi(parts[i])
				if err != nil {
					return nil, fmt.Errorf("Line %d is invalid, value %d is not a valid integer: %s", lineIndex, (i + 1), parts[i])
				}
				c[i] = int8(val)
			}
			durationStr := parts[4]
			var duration int64
			if strings.HasSuffix(durationStr, "s") {
				val, err := strconv.Atoi(durationStr[:len(durationStr)-1])
				if err != nil {
					return nil, fmt.Errorf("Line %d is invalid, duration ends on 's' but does not start with a valid integer: %s", lineIndex, durationStr)
				}
				duration = int64(val) * 1000
			} else {
				val, err := strconv.Atoi(durationStr)
				if err != nil {
					return nil, fmt.Errorf("Line %d is invalid, duration a valid integer of milliseconds: %s", lineIndex, durationStr)
				}
				duration = int64(val)
			}
			plan.Add(c[0], c[1], c[2], c[3], duration)
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	return &plan, nil
}
