package main_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	main "github.com/dladlk/liftoff-drone-navigator"
	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

func TestCalculate(t *testing.T) {
	cfg := main.NewControllerConfig()
	controller := main.NewController(cfg, &TestTelemetryProvider{}, &TestSetpointProvider{}, &TestJoystick{})

	zeroDatagram := main.DatagramToTelemetry(&lot_config.Datagram{}, 1)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tel  main.Telemetry
		sp   main.Setpoint
		want main.JoystickPosition
	}{
		{"Expect non-zero thrust joystick if we have zero and want zero - because of gravity G",
			zeroDatagram, main.Setpoint{}, main.JoystickPosition{-12767, 0, 0, 0}},

		{"Expect zero joystick if we have 1 on pos and want 1 on pos",
			main.DatagramToTelemetry(&lot_config.Datagram{Position: [3]float32{1, 1, 1}}, 1), main.Setpoint{PositionDesired: main.Vec3{1, 1, 1}}, main.JoystickPosition{-12767, 0, 0, 0}},

		{"Expect non-zero joystick if we want to move 1m in direction x and stop",
			zeroDatagram, main.Setpoint{PositionDesired: main.Vec3{1, 0, 0}}, main.JoystickPosition{-12521, 0, 1638, 0}},

		{"Expect non-zero joystick if we want to move 10m in direction x and stop",
			zeroDatagram, main.Setpoint{PositionDesired: main.Vec3{10, 0, 0}}, main.JoystickPosition{-9491, 0, 1638, 0}},

		{"Expect non-zero joystick if we want to move 100m in direction x and stop",
			zeroDatagram, main.Setpoint{PositionDesired: main.Vec3{100, 0, 0}}, main.JoystickPosition{-9491, 0, 1638, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If we want to get 0 throttle for 0 movement, we should also imitate situation when throttle is in "hover" position
			controller.ResetState(cfg.Throttle.HoverStick)
			got := main.CalculateJoysticksPosition(controller, tt.tel, tt.sp)
			if got != tt.want {
				t.Errorf("CalculateJoysticksPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestTelemetryProvider struct{}

func (t *TestTelemetryProvider) ReadTelemetry() (main.Telemetry, bool, error) {
	return main.Telemetry{}, true, nil
}

type TestJoystick struct{}

func (a *TestJoystick) SendJoystick(joystickPosition main.JoystickPosition) error {
	return nil
}

type TestSetpointProvider struct{}

func (m *TestSetpointProvider) GetDesiredSetpoint(now time.Time, tel main.Telemetry) (main.Setpoint, error) {
	return main.Setpoint{}, nil
}

func TestToInt16Uncentered01(t *testing.T) {
	tests := []struct {
		x    float64
		want int16
	}{
		{0, math.MinInt16 + 1},
		{0.05, math.MinInt16 + 2 + int16(math.Round(math.MaxInt16)*0.1)},
		{0.5, 0},
		{1, math.MaxInt16},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%f", tt.x), func(t *testing.T) {
			got := main.ToInt16Uncentered01(tt.x)
			if got != tt.want {
				t.Errorf("ToInt16Uncentered01() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToInt16Signed(t *testing.T) {
	tests := []struct {
		x    float64
		want int16
	}{
		{-1, -32767},
		{0, 0},
		{1, 32767},
		{0.5, 16384},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%f", tt.x), func(t *testing.T) {
			got := main.ToInt16Signed(tt.x)
			if got != tt.want {
				t.Errorf("ToInt16Signed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttitudeAndThrustFromAccelYaw(t *testing.T) {
	cfg := main.NewControllerConfig()
	t.Run("Test", func(t *testing.T) {
		a_c := main.Vec3{-13, 0.000, -9.9}
		a_c = main.Vec3{1, 0.000, 0}
		Rd, thrust := main.AttitudeAndThrustFromAccelYaw(a_c, cfg.G, 0, cfg.Mass)
		t.Errorf("%f %s %s\n", thrust, main.RToEulerDegreeVec(Rd), Rd)
	})
}
