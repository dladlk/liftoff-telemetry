package main

type NoopDrone struct {
}

func (t *NoopDrone) Init()                                              {}
func (t *NoopDrone) UpdatePrimitive(lx int8, ly int8, rx int8, ry int8) {}
func (t *NoopDrone) UpdateDirect(lx, ly, rx, ry int16)                  {}
func (t *NoopDrone) UpdateByInput(Input *[4]float32)                    {}
func (t *NoopDrone) UpdateLeftRight(left Joystick, right Joystick)      {}
func (t *NoopDrone) Close()                                             {}
