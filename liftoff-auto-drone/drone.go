package main

import (
	"math"

	"github.com/dladlk/liftoff-auto-drone/vigem"
	X360 "github.com/dladlk/liftoff-auto-drone/x360"
)

type IDrone interface {
	Init()
	Update(lx int8, ly int8, rx int8, ry int8)
	UpdateByInput(Input *[4]float32)
	UpdateLeftRight(left Joystick, right Joystick)
	Close()
}

type Drone struct {
	client *vigem.ClientImpl
	x360   *X360.Gamepad
}

// Input     [4]float32 `desc:"throttle, yaw, pitch, roll"`
// But update: yaw, throttle, roll, pitch
func (t Drone) UpdateByInput(Input *[4]float32) {
	//fmt.Printf("Update with %v", u)
	t.x360.LeftJoystick(float32ToInt16(Input[1]), float32ToInt16(Input[0]))
	t.x360.RightJoystick(-float32ToInt16(Input[3]), float32ToInt16(Input[2]))
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

func (t *Drone) UpdateLeftRight(left Joystick, right Joystick) {
	t.Update(left.x, left.y, right.x, right.y)
}

func (t *Drone) Update(lx int8, ly int8, rx int8, ry int8) {
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
