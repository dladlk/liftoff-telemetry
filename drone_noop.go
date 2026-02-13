package main

type NoopDrone struct {
}

func (t *NoopDrone) Init()                                         {}
func (t *NoopDrone) Update(lx int8, ly int8, rx int8, ry int8)     {}
func (t *NoopDrone) UpdateByInput(Input [4]float32)                {}
func (t *NoopDrone) UpdateLeftRight(left Joystick, right Joystick) {}
func (t *NoopDrone) Close()                                        {}
