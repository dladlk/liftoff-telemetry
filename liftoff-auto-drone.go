package main

import (
	"fmt"
	"math"
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

const posPerDirection = 8
const convertRatio int16 = math.MaxInt16 / posPerDirection

func convert(v int8) int16 {
	return convertRatio * int16(v)
}

var drone Drone

func main() {
	left := Joystick{}
	right := Joystick{}
	var step int8 = 1

	drone = Drone{}
	drone.Init()

	fmt.Println("Left joystick: WSAD, Right joystick: ↑↓←→, Q quit, R reset")
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		if key.Code == keys.CtrlC {
			return true, nil
		}

		planMode := false
		planIndex := 0

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
			case "r": // RESET
				left.Reset()
				right.Reset()
			case "q":
				return true, nil
			default:
				if key.String() >= "0" && key.String() <= "9" {
					planMode = true
					planIndex, _ = strconv.Atoi(key.String())
				}
			}

		}

		if planMode {
			var plan *Plan
			plan, err := ReadPlan(fmt.Sprintf("plan_%d.txt", planIndex))
			if err != nil {
				fmt.Printf("\nFailed to read plan %d: %v", planIndex, err)
				return false, nil
			}

			fmt.Printf("\nPlan %d '%s' Start                              \r\n", planIndex, plan.Name)
			// Sleep 3 second to switch to Liftoff
			p(3000)
			for _, c := range plan.List {
				u(c.Update[0], c.Update[1], c.Update[2], c.Update[3])
				p(c.Duration)
			}
			fmt.Printf("Plan %d Done                              \r\n", planIndex)
		} else {
			fmt.Printf("\r%s      \t: %+v %+v    ", key.String(), left, right)
			drone.Update(left, right)
		}
		return false, nil
	})

	fmt.Println("Unregister, disconnect and release")

	drone.Close()
}

func p(millis int) {
	fmt.Printf("Sleep %v millis\r\n", millis)
	time.Sleep(time.Duration(millis) * time.Millisecond)
}
func u(lx int8, ly int8, rx int8, ry int8) {
	fmt.Printf("Update %d %d %d %d\r\n", lx, ly, rx, ry)
	drone.Update2(lx, ly, rx, ry)
}
