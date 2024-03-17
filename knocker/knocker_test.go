package knocker

import (
	"testing"
	"time"
)

// TestCheckSequence tests the checkSequence method of sequenceTracker
func TestCheckSequence(t *testing.T) {
	type step struct {
		srcIP     string
		port      uint16
		timestamp time.Time
		valid     bool
	}
	now := time.Now()
	sequence := []uint16{1000, 2000, 3000, 4000}
	tests := map[string]struct {
		steps []step
	}{
		"Valid sequence": {
			steps: []step{
				{"192.168.1.1", 1000, now, false},
				{"192.168.1.1", 2000, now.Add(1 * time.Second), false},
				{"192.168.1.1", 3000, now.Add(2 * time.Second), false},
				{"192.168.1.1", 4000, now.Add(3 * time.Second), true},
			},
		},
		"Invalid sequence": {
			steps: []step{
				{"192.168.1.2", 1000, now, false},
				{"192.168.1.2", 3000, now.Add(1 * time.Second), false}, // Out of order
				{"192.168.1.2", 2000, now.Add(2 * time.Second), false},
				{"192.168.1.2", 4000, now.Add(3 * time.Second), false},
			},
		},
		"Reset on invalid step": {
			steps: []step{
				{"192.168.1.3", 1000, now, false},
				{"192.168.1.3", 2000, now.Add(1 * time.Second), false},
				{"192.168.1.3", 5000, now.Add(2 * time.Second), false}, // Invalid step
				{"192.168.1.3", 1000, now.Add(3 * time.Second), false}, // Start over
				{"192.168.1.3", 2000, now.Add(4 * time.Second), false},
				{"192.168.1.3", 3000, now.Add(5 * time.Second), false},
				{"192.168.1.3", 4000, now.Add(6 * time.Second), true},
			},
		},
		"Sequence with timeout": {
			steps: []step{
				{"192.168.1.4", 1000, time.Now(), false},
				{"192.168.1.4", 2000, time.Now().Add(1 * time.Second), false},  // Within timeout
				{"192.168.1.4", 3000, time.Now().Add(14 * time.Second), false}, // Exceeds timeout
				{"192.168.1.4", 1000, time.Now().Add(15 * time.Second), false}, // Start over due to timeout
				{"192.168.1.4", 2000, time.Now().Add(16 * time.Second), false},
				{"192.168.1.4", 3000, time.Now().Add(17 * time.Second), false},
				{"192.168.1.4", 4000, time.Now().Add(18 * time.Second), true}, // Should succeed now
			},
		},
		"Sequence too quick": {
			steps: []step{
				{"192.168.1.5", 1000, now, false},
				{"192.168.1.5", 2000, now.Add(500 * time.Millisecond), false},  // Too quick
				{"192.168.1.5", 3000, now.Add(1500 * time.Millisecond), false}, // Proper interval, but sequence reset
				{"192.168.1.5", 4000, now.Add(2500 * time.Millisecond), false}, // Proper interval, but sequence reset
				{"192.168.1.5", 1000, now.Add(3000 * time.Millisecond), false}, // Start over, proper interval
				{"192.168.1.5", 2000, now.Add(4000 * time.Millisecond), false},
				{"192.168.1.5", 3000, now.Add(5000 * time.Millisecond), false},
				{"192.168.1.5", 4000, now.Add(6000 * time.Millisecond), true}, // Sequence completes successfully
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tracker := New(sequence, 10*time.Second, 1*time.Second)
			for i, step := range tt.steps {
				if got := tracker.CheckSequence(step.srcIP, step.port, step.timestamp); got != step.valid {
					t.Fatalf("[%d] got = %v, want %v for step srcIP %s port %d", i, got, step.valid, step.srcIP, step.port)
				}
			}
		})
	}
}
