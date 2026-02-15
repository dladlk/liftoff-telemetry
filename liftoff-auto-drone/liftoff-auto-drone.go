package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
)

// Manual calibration:
/*
Yaw	- Channel 0, left.x
Throttle - Channel 1, left.y
Roll - Channel 2, right.x
Pitch - Channel 3, right.y
*/

type Joystick struct {
	x int8
	y int8
}

func (this *Joystick) Reset() {
	this.x = 0
	this.y = 0
}

// Declare a custom type for the enum
type RunMode int

// Define the enum values using iota
const (
	Keyboard RunMode = iota
	PlanMode
	TrackMode
	TrackStreamMode
)

const posPerDirection = 8
const convertRatio int16 = math.MaxInt16 / posPerDirection

func convert(v int8) int16 {
	return convertRatio * int16(v)
}

const VERSION = "0.0.1"

var drone IDrone

func main() {

	flag.Usage = func() {
		fmt.Printf("Version %s\n", VERSION)
		fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}
	doHelp := flag.Bool("help", false, "Prints help")
	doDryRun := flag.Bool("dry-run", false, "Dry-run - without starting joystick simulation")

	flag.Parse()
	if *doHelp {
		flag.Usage()
		os.Exit(1)
	}

	noopDrone := *doDryRun

	telemetryListener := TelemetryListener{}

	left := Joystick{}
	right := Joystick{}
	var step int8 = 1

	if noopDrone {
		drone = &NoopDrone{}
		fmt.Printf("Start without joystick simulation\n")
	} else {
		drone = &Drone{}
	}
	drone.Init()

	trackRunStopChannel := make(chan bool)
	trackRunning := false

	fmt.Println("Left: WSAD, Right: ↑↓←→, Q quit, R reset, U udp toggle, SPACE telemtry")
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		if key.Code == keys.Null {
			// Ignore Ctrl-@ which is sent by debugger...
			return false, nil
		}
		if key.Code == keys.CtrlC {
			return true, nil
		}

		runMode := Keyboard
		modeIndex := 0

		switch key.Code {
		case keys.Up:
			if right.y < posPerDirection {
				right.y += step
			}
		case keys.Down:
			if right.y > -posPerDirection {
				right.y -= step
			}
		case keys.Left:
			if right.x > -posPerDirection {
				right.x -= step
			}
		case keys.Right:
			if right.x < posPerDirection {
				right.x += step
			}
		case keys.Space:
			// Just do nothing
		case keys.RuneKey:
			switch key.String() {
			case "w":
				if left.y < posPerDirection {
					left.y += step
				}
			case "s":
				if left.y > -posPerDirection {
					left.y -= step
				}
			case "a":
				if left.x > -posPerDirection {
					left.x -= step
				}
			case "d":
				if left.x < posPerDirection {
					left.x += step
				}
			case "t": // Terminate if track running
				if trackRunning {
					trackRunStopChannel <- true
					trackRunning = false
				}
			case "r": // RESET
				left.Reset()
				right.Reset()
			case "q":
				return true, nil
			case "u": // Toggle telemetry listener
				telemetryListener.Toggle()
			default:
				if key.String() >= "0" && key.String() < "5" {
					runMode = PlanMode
					modeIndex, _ = strconv.Atoi(key.String())
				}
				if key.String() == "5" {
					runMode = TrackMode
					modeIndex, _ = strconv.Atoi(key.String())
				}
				if key.String() >= "6" && key.String() <= "9" {
					runMode = TrackStreamMode
					modeIndex, _ = strconv.Atoi(key.String())
				}
			}

		}

		switch runMode {
		case PlanMode:
			var plan *Plan
			plan, err := ReadPlan(fmt.Sprintf("plan_%d.txt", modeIndex))
			if err != nil {
				fmt.Printf("\nFailed to read plan %d: %v", modeIndex, err)
				return false, nil
			}

			fmt.Printf("\nPlan %d '%s' Start                              \r\n", modeIndex, plan.Name)
			// Sleep 3 second to switch to Liftoff
			p(3000)
			for _, c := range plan.List {
				u(c.Update[0], c.Update[1], c.Update[2], c.Update[3])
				p(c.Duration)
			}
			fmt.Printf("Plan %d Done                              \r\n", modeIndex)
		case TrackMode:
			var telemetry *Telemetry
			telemetry, err := ReadTelemetry(fmt.Sprintf("telemetry_%d.txt", modeIndex))
			if err != nil {
				fmt.Printf("\nFailed to read track %d: %v", modeIndex, err)
				return false, nil
			}

			fmt.Printf("\nTrack %d '%s' Start                              \r\n", modeIndex, telemetry.Name)
			// Sleep 3 second to switch to Liftoff
			p(3000)
			for _, c := range telemetry.Records {
				fmt.Printf("\r\n %+v", c)
				drone.UpdateByInput(&c.Input)
				p(100)
			}
			fmt.Printf("Plan %d Done                              \r\n", modeIndex)
		case TrackStreamMode:
			sleepTime := 3000
			if noopDrone {
				sleepTime = 0
			}
			fmt.Printf("\rYou have %d seconds to switch to Liftoff window to process track. Press T to terminate running\r\n", sleepTime/1000)
			p(sleepTime)
			filepath := fmt.Sprintf("track_%d.bin", modeIndex)
			trackRunning = true
			go func() (bool, error) {
				track := Track{}
				startTime := time.Now()
				err := track.Open(filepath)
				if err != nil {
					fmt.Printf("Failed to read bin track %d: %v\n", modeIndex, err)
					return false, err
				}
				skipFramesCount := 0
				progressPrint := func(prefix string, i int) string {
					durationSec := float64(time.Since(startTime).Round(time.Millisecond).Milliseconds()) / 1000.0
					simulationDurationSec := track.List[i].Timestamp - track.minTs
					diff := durationSec - float64(simulationDurationSec)
					progressPercent := float32(i+1) / float32(len(track.List)) * 100.0
					return fmt.Sprintf("%s %d of %d - %.0f%% in %.2f s, track dur %.2f s, diff %.2f s, skipped %d frames", prefix, i+1, len(track.List), progressPercent, durationSec, simulationDurationSec, diff, skipFramesCount)
				}

				for i, c := range track.List {
					if i == 0 {
						drone.UpdateByInput(&c.Input)
						continue
					}
					select {
					case <-trackRunStopChannel:
						fmt.Printf("%s\r\n", progressPrint("\rTerminated on", i))
						return false, nil
					default:
						durationSec := float32(time.Since(startTime).Microseconds()) / 1000_000
						simulationDurationSec := c.Timestamp - track.minTs
						diff := simulationDurationSec - durationSec
						if diff > 0.000001 {
							time.Sleep(time.Duration(diff*1000_000) * time.Microsecond)
						}
						if diff < 0 {
							skipFramesCount += 1
						} else {
							drone.UpdateByInput(&c.Input)
						}
						if i%100 == 0 {
							fmt.Print(progressPrint("\rDone", i))
						}
					}
				}
				fmt.Printf("%s\r\n", progressPrint("\rFinished", len(track.List)-1))
				return false, nil
			}()

		default:
			lastTelemetry := "not listening for telemtry"
			if telemetryListener.running {
				d, datagramIndex, ok := telemetryListener.LastDatagram()
				if ok {
					lastTelemetry = fmt.Sprintf("[%d] %.6f %.6f %.6f %.6f", datagramIndex, d.Input[0], d.Input[1], d.Input[2], d.Input[3])
				} else {
					lastTelemetry = "Nothing"
				}
			}
			fmt.Printf("\r'%s'\t: %+v %+v %s %-30s", key.String(), left, right, lastTelemetry, "")
			drone.UpdateLeftRight(left, right)
		}
		return false, nil
	})

	fmt.Println("\r\nUnregister, disconnect and release")

	drone.Close()
}

func p(millis int) {
	//fmt.Printf("Sleep %v millis\r\n", millis)
	time.Sleep(time.Duration(millis) * time.Millisecond)
}
func u(lx int8, ly int8, rx int8, ry int8) {
	//fmt.Printf("Update %d %d %d %d\r\n", lx, ly, rx, ry)
	drone.Update(lx, ly, rx, ry)
}
