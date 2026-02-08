package main

import (
	"fmt"
	"math"

	"github.com/dladlk/liftoff-auto-drone/vigem"
	X360 "github.com/dladlk/liftoff-auto-drone/x360"
)

type Drone struct {
	client *vigem.ClientImpl
	x360   *X360.Gamepad
}

// Input     [4]float32 `desc:"throttle, yaw, pitch, roll"`
// But update: yaw, throttle, roll, pitch
func (t Drone) UpdateByTelemetryRecord(v TelemetryRecord) {
	u := []int16{float32ToInt16(v.Input[1]), float32ToInt16(v.Input[0]), float32ToInt16(v.Input[3]), float32ToInt16(v.Input[2])}
	fmt.Printf("Update with %v", u)
	t.x360.LeftJoystick(u[0], u[1])
	t.x360.RightJoystick(u[2], u[3])
	t.x360.Update()
}

func float32ToInt16(f float32) int16 {
	return int16(f * math.MaxInt16)
}

func (t *Drone) Init() {
	t.client = vigem.NewClient()
	t.x360 = X360.NewGamepad(t.client)
	t.x360.Connect()
}

func (t *Drone) Update(left Joystick, right Joystick) {
	t.Update2(left.x, left.y, right.x, right.y)
}

func (t *Drone) Update2(lx int8, ly int8, rx int8, ry int8) {
	t.x360.LeftJoystick(convert(lx), convert(ly))
	t.x360.RightJoystick(convert(rx), convert(ry))
	t.x360.Update()
}

func (t *Drone) Close() {
	t.x360.UnregisterNotification()
	t.x360.Disconnect()
	t.x360.Release()

	t.client.Release()
}
