package main_test

import (
	"math"
	"testing"

	main "github.com/dladlk/liftoff-auto-drone"
)

func TestReadTelemetry(t *testing.T) {
	plan0 := main.Telemetry{Name: "telemetry_test.txt"}

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		want    *main.Telemetry
		wantErr bool
	}{
		{name: "Test telemetry", path: "telemetry_test.txt", want: &plan0, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := main.ReadTelemetry(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ReadTelemetry() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ReadTelemetry() succeeded unexpectedly")
			}

			ok := true

			if got.Name != tt.want.Name {
				ok = false
			}
			if ok {
				if len(got.Records) != 35 {
					ok = false
				}
			}
			if ok {
				if math.Abs(float64(got.Records[0].Timestamp)) > 0.0000001 {
					ok = false
				}
			}
			if ok {
				if got.Records[0].Input[0] != -1 {
					ok = false
				}
			}

			if !ok {
				t.Errorf("ReadTelemetry() = %v, want %v", got, tt.want)
			}
		})
	}
}
