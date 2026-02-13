package main_test

import (
	"fmt"
	"testing"

	main "github.com/dladlk/liftoff-auto-drone"
)

func TestTrack_Open(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		wantErr bool
	}{
		{name: "Read file 6", path: "track_6.bin", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := main.Track{}
			list, gotErr := tr.Open(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Open() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Open() succeeded unexpectedly")
			}
			if len(list) != 10327 {
				t.Fatalf("Wrong number of rows read: %d", len(list))
			}
			fmt.Printf("First:\t %+v", list[0])
			fmt.Printf("Last:\t %+v", list[len(list)-1])
		})
	}
}
