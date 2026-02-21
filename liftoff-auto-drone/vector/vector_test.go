package vector_test

import (
	"fmt"
	"testing"

	vector "github.com/dladlk/liftoff-auto-drone/vector"
)

func TestVelocityWorldSpaceToLocalSpace(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		velocity       [3]float32
		attitude       [4]float32
		expectedResult string
	}{
		{name: "Zero", velocity: [3]float32{0, 0, 0}, attitude: [4]float32{0, 0, 0, 1}, expectedResult: "[0 0 0]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vl := vector.VelocityWorldSpaceToLocalSpace(tt.velocity, tt.attitude)
			actualResult := fmt.Sprintf("%+v", vl)
			if actualResult != tt.expectedResult {
				t.Fatalf("Expected %v, got %v", tt.expectedResult, actualResult)
			}
		})
	}
}
