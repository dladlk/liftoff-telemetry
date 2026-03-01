package vector_test

import (
	"fmt"
	"math"
	"testing"

	vector "github.com/dladlk/liftoff-telemetry/vector"
)

func TestVelocityWorldSpaceToLocalSpace(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		velocity       [3]float32
		attitude       [4]float32
		expectedResult string
	}{
		{name: "All zero", velocity: [3]float32{0, 0, 0}, attitude: [4]float32{0, 0, 0, 1}, expectedResult: "[0.000 0.000 0.000]"},
		{name: "Zero speed x 1m", velocity: [3]float32{1, 0, 0}, attitude: [4]float32{0, 0, 0, 1}, expectedResult: "[1.000 0.000 0.000]"},
		{name: "Zero speed 10m all", velocity: [3]float32{10, 10, 10}, attitude: [4]float32{0, 0, 0, 1}, expectedResult: "[10.000 10.000 10.000]"},

		{name: "45deg all directions and speed x 1M",
			velocity: [3]float32{1, 0, 0}, attitude: vector.EulerAnglesToQuaternion(math.Pi/4, math.Pi/4, math.Pi/4).ToAttitude(), expectedResult: "[0.500 0.500 -0.707]"},

		{name: "Yaw 45deg x 1m",
			velocity: [3]float32{1, 0, 0}, attitude: vector.EulerAnglesToQuaternion(0, 0, math.Pi/4).ToAttitude(), expectedResult: "[0.707 0.707 0.000]"},

		{name: "All 360deg x 1m",
			velocity: [3]float32{1, 0, 0}, attitude: vector.EulerAnglesToQuaternion(math.Pi*2, math.Pi*2, math.Pi*2).ToAttitude(), expectedResult: "[1.000 -0.000 0.000]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vl := vector.VelocityWorldSpaceToLocalSpace(tt.velocity, tt.attitude)
			actualResult := fmt.Sprintf("[%.3f %.3f %.3f]", vl[0], vl[1], vl[2])
			if actualResult != tt.expectedResult {
				t.Fatalf("Expected %v, got %v", tt.expectedResult, actualResult)
			}
		})
	}
}
