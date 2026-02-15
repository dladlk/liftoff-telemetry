package main

import (
	"fmt"
	"testing"
)

func Test_float32ToInt16(t *testing.T) {
	tests := []struct {
		// Named input parameters for target function.
		f    float32
		want int16
	}{
		{-1.0, -32767},
		{0, 0},
		{1.0, 32767},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Row %d", i), func(t *testing.T) {
			got := float32ToInt16(tt.f)
			// TODO: update the condition below to compare got with tt.want.
			if tt.want != got {
				t.Errorf("float32ToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}
