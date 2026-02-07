package main

import (
	"github.com/dladlk/liftoff-auto-drone/vigem"
	X360 "github.com/dladlk/liftoff-auto-drone/x360"
)

type Drone struct {
	client *vigem.ClientImpl
	x360   *X360.Gamepad
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
