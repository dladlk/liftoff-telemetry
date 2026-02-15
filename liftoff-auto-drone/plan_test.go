package main_test

import (
	"testing"

	main "github.com/dladlk/liftoff-auto-drone"
)

func TestReadPlan(t *testing.T) {
	plan0 := main.Plan{}
	plan0.Name = "Test plan"
	plan0.Add(0, -8, 0, 0, 2000)
	plan0.Add(0, -2, 0, 0, 3000)
	plan0.Add(0, -2, 0, 2, 500)
	plan0.Add(0, -2, 0, -2, 500)
	plan0.Add(0, -5, 0, 0, 3000)
	plan0.Add(0, -8, 0, 0, 5000)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		want    *main.Plan
		wantErr bool
	}{
		{name: "Test plan", path: "plan_test.txt", want: &plan0, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := main.ReadPlan(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ReadPlan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ReadPlan() succeeded unexpectedly")
			}

			ok := true
			if got.Name != tt.want.Name || len(got.List) != len(tt.want.List) {
				ok = false
			}
			if ok {
				for i := range got.List {
					c1 := got.List[i]
					c2 := tt.want.List[i]
					if c1.Duration != c2.Duration {
						ok = false
						break
					}
					if len(c1.Update) != len(c2.Update) {
						ok = false
						break
					}
					for j := range c1.Update {
						if c1.Update[j] != c2.Update[j] {
							ok = false
							break
						}
					}
				}
			}

			if !ok {
				t.Errorf("ReadPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}
