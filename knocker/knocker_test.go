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
	timeout := 10 * time.Second
	lowerThanMinInterval := 500 * time.Millisecond
	minInterval := 2 * lowerThanMinInterval

	staticProvider := StaticSequenceProvider{
		Sequence: []uint16{1000, 2000, 3000, 4000},
	}
	tests := map[string]struct {
		steps []step
	}{
		"Valid sequence": {
			steps: []step{
				{"192.168.1.1", 1000, now, false},
				{"192.168.1.1", 2000, now.Add(1 * minInterval), false},
				{"192.168.1.1", 3000, now.Add(2 * minInterval), false},
				{"192.168.1.1", 4000, now.Add(3 * minInterval), true},
			},
		},
		"Invalid sequence": {
			steps: []step{
				{"192.168.1.2", 1000, now, false},
				{"192.168.1.2", 3000, now.Add(1 * minInterval), false}, // Out of order
				{"192.168.1.2", 2000, now.Add(2 * minInterval), false},
				{"192.168.1.2", 4000, now.Add(3 * minInterval), false},
			},
		},
		"Reset on invalid step": {
			steps: []step{
				{"192.168.1.3", 1000, now, false},
				{"192.168.1.3", 2000, now.Add(1 * minInterval), false},
				{"192.168.1.3", 5000, now.Add(2 * minInterval), false}, // Invalid step
				{"192.168.1.3", 1000, now.Add(3 * minInterval), false}, // Start over
				{"192.168.1.3", 2000, now.Add(4 * minInterval), false},
				{"192.168.1.3", 3000, now.Add(5 * minInterval), false},
				{"192.168.1.3", 4000, now.Add(6 * minInterval), true},
			},
		},
		"Sequence with timeout": {
			steps: []step{
				{"192.168.1.4", 1000, time.Now(), false},
				{"192.168.1.4", 2000, time.Now().Add(1 * minInterval), false},  // Within timeout
				{"192.168.1.4", 3000, time.Now().Add(14 * minInterval), false}, // Exceeds timeout
				{"192.168.1.4", 1000, time.Now().Add(15 * minInterval), false}, // Start over due to timeout
				{"192.168.1.4", 2000, time.Now().Add(16 * minInterval), false},
				{"192.168.1.4", 3000, time.Now().Add(17 * minInterval), false},
				{"192.168.1.4", 4000, time.Now().Add(18 * minInterval), true}, // Should succeed now
			},
		},
		"Sequence too quick": {
			steps: []step{
				{"192.168.1.5", 1000, now, false},
				{"192.168.1.5", 2000, now.Add(lowerThanMinInterval), false},                 // Too quick
				{"192.168.1.5", 3000, now.Add(lowerThanMinInterval + minInterval), false},   // Proper interval, but sequence reset
				{"192.168.1.5", 4000, now.Add(lowerThanMinInterval + minInterval*2), false}, // Proper interval, but sequence reset
				{"192.168.1.5", 1000, now.Add(lowerThanMinInterval + minInterval*3), false}, // Start over, proper interval
				{"192.168.1.5", 2000, now.Add(lowerThanMinInterval + minInterval*4), false},
				{"192.168.1.5", 3000, now.Add(lowerThanMinInterval + minInterval*5), false},
				{"192.168.1.5", 4000, now.Add(lowerThanMinInterval + minInterval*6), true}, // Sequence completes successfully
			},
		},
		"Immediate repeated attempts": {
			steps: []step{
				{"192.168.1.6", 1000, now, false},
				{"192.168.1.6", 1000, now.Add(lowerThanMinInterval), false}, // Immediate repeat
				{"192.168.1.6", 2000, now.Add(lowerThanMinInterval + minInterval), false},
				{"192.168.1.6", 3000, now.Add(lowerThanMinInterval + minInterval*2), false},
				{"192.168.1.6", 4000, now.Add(lowerThanMinInterval + minInterval*3), false},
			},
		},
		"Sequence attempt with long pause": {
			steps: []step{
				{"192.168.1.7", 1000, now, false},
				{"192.168.1.7", 2000, now.Add(1 * time.Second), false},
				// Long pause before the next step
				{"192.168.1.7", 3000, now.Add(timeout + (1 * time.Second)), false}, // Should reset here
				{"192.168.1.7", 1000, now.Add(timeout + (2 * time.Second)), false}, // New sequence start
				{"192.168.1.7", 2000, now.Add(timeout + (3 * time.Second)), false},
				{"192.168.1.7", 3000, now.Add(timeout + (4 * time.Second)), false},
				{"192.168.1.7", 4000, now.Add(timeout + (5 * time.Second)), true},
			},
		},
		"Correct sequence with exact minimum interval timing": {
			steps: []step{
				{"192.168.1.8", 1000, now, false},
				{"192.168.1.8", 2000, now.Add(minInterval), false}, // Exactly at minInterval
				{"192.168.1.8", 3000, now.Add(minInterval * 2), false},
				{"192.168.1.8", 4000, now.Add(minInterval * 3), true},
			},
		},
		"Different IPs independent sequences": {
			steps: []step{
				// IP 1 starts a sequence
				{"192.168.1.9", 1000, now, false},
				// IP 2 starts and completes its sequence amidst IP 1's attempts
				{"192.168.2.9", 1000, now.Add(minInterval), false},
				{"192.168.2.9", 2000, now.Add(2 * minInterval), false},
				{"192.168.2.9", 3000, now.Add(3 * minInterval), false},
				{"192.168.2.9", 4000, now.Add(4 * minInterval), true},
				// IP 1 completes its sequence
				{"192.168.1.9", 2000, now.Add(5 * minInterval), false},
				{"192.168.1.9", 3000, now.Add(6 * minInterval), false},
				{"192.168.1.9", 4000, now.Add(7 * minInterval), true},
			},
		},
		"Invalid port before starting a valid sequence": {
			steps: []step{
				{"192.168.1.10", 5000, now, false}, // Invalid port
				{"192.168.1.10", 1000, now.Add(1 * time.Second), false},
				{"192.168.1.10", 2000, now.Add(2 * time.Second), false},
				{"192.168.1.10", 3000, now.Add(3 * time.Second), false},
				{"192.168.1.10", 4000, now.Add(4 * time.Second), true},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tracker := New(&staticProvider, timeout, minInterval)
			for i, step := range tt.steps {
				if got := tracker.CheckSequence(step.srcIP, step.port, step.timestamp); got != step.valid {
					t.Fatalf("[%d] got = %v, want %v for step srcIP %s port %d", i, got, step.valid, step.srcIP, step.port)
				}
			}
		})
	}
}
