package main_test

import (
	"context"
	"testing"
	"time"

	main "github.com/dladlk/liftoff-drone-navigator"
	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

func TestCalculate(t *testing.T) {
	controller := main.NewController(main.NewControllerConfig(), &TestTelemetryProvider{}, &TestSetpointProvider{}, &TestJoystick{})

	zeroDatagram := main.DatagramToTelemetry(&lot_config.Datagram{})

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tel  main.Telemetry
		sp   main.Setpoint
		want main.JoystickPosition
	}{
		{"Expect zero joystick if we have zero and want zero",
			zeroDatagram, main.Setpoint{}, main.JoystickPosition{}},

		{"Expect zero joystick if we have 1 on pos and want 1 on pos",
			main.DatagramToTelemetry(&lot_config.Datagram{Position: [3]float32{1, 1, 1}}), main.Setpoint{PositionDesired: main.Vec3{1, 1, 1}}, main.JoystickPosition{}},

		{"Expect non-zero joystick if we want to move 1m in direction x and stop",
			zeroDatagram, main.Setpoint{PositionDesired: main.Vec3{1, 0, 0}}, main.JoystickPosition{0, 0, 1638, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := main.CalculateJoysticksPosition(controller, tt.tel, tt.sp)
			if got != tt.want {
				t.Errorf("CalculateJoysticksPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestTelemetryProvider struct{}

func (t *TestTelemetryProvider) Read(ctx context.Context) (main.Telemetry, bool, error) {
	return main.Telemetry{}, true, nil
}

type TestJoystick struct{}

func (a *TestJoystick) SendJoystick(ctx context.Context, joystickPosition main.JoystickPosition) error {
	return nil
}

type TestSetpointProvider struct{}

func (m *TestSetpointProvider) Desired(ctx context.Context, now time.Time) (main.Setpoint, error) {
	return main.Setpoint{}, nil
}
