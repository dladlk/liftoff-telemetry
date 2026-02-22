package track_test

import (
	"fmt"
	"testing"

	track "github.com/dladlk/liftoff-auto-drone/track"
)

func TestTrack_Open(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		wantErr bool
	}{
		{name: "Read file 6", path: "track_test.bintest", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := track.Track{}
			gotErr := tr.Open(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Open() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Open() succeeded unexpectedly")
			}
			if len(tr.List) != 225 {
				t.Fatalf("Wrong number of rows read: %d", len(tr.List))
			}
			fmt.Printf("First:\t %+v\n", tr.List[0])
			fmt.Printf("Last:\t %+v\n", tr.List[len(tr.List)-1])
		})
	}
}
