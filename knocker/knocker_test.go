package knocker

import (
	"testing"
)

// TestCheckSequence tests the checkSequence method of sequenceTracker
func TestCheckSequence(t *testing.T) {
	type step struct {
		srcIP string
		port  uint16
		valid bool
	}
	tests := map[string]struct {
		name     string
		sequence []uint16
		steps    []step
	}{
		"Valid sequence": {
			sequence: []uint16{1000, 2000, 3000, 4000},
			steps: []step{
				{"192.168.1.1", 1000, false},
				{"192.168.1.1", 2000, false},
				{"192.168.1.1", 3000, false},
				{"192.168.1.1", 4000, true},
			},
		},
		"Invalid sequence": {
			sequence: []uint16{1000, 2000, 3000, 4000},
			steps: []step{
				{"192.168.1.2", 1000, false},
				{"192.168.1.2", 3000, false}, // Out of order
				{"192.168.1.2", 2000, false},
				{"192.168.1.2", 4000, false},
			},
		},
		"Reset on invalid step": {
			sequence: []uint16{1000, 2000, 3000, 4000},
			steps: []step{
				{"192.168.1.3", 1000, false},
				{"192.168.1.3", 2000, false},
				{"192.168.1.3", 5000, false}, // Invalid step
				{"192.168.1.3", 1000, false}, // Start over
				{"192.168.1.3", 2000, false},
				{"192.168.1.3", 3000, false},
				{"192.168.1.3", 4000, true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := New(tt.sequence)
			for _, step := range tt.steps {
				if got := tracker.CheckSequence(step.srcIP, step.port); got != step.valid {
					t.Fatalf("checkSequence() got = %v, want %v for step srcIP %s port %d", got, step.valid, step.srcIP, step.port)
				}
			}
		})
	}
}
