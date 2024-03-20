package knocker

import (
	"testing"
	"time"
)

type StaticSequenceProvider struct {
	Sequence   [][]uint16
	SequenceId int
}

func (p *StaticSequenceProvider) GetSequence(srcIP string, timestamp time.Time) []uint16 {
	return p.Sequence[p.SequenceId]
}

func TestCheckSequence(t *testing.T) {
	type step struct {
		srcIP          string
		port           uint16
		timestamp      time.Time
		valid          bool
		sequenceNumber int
	}
	now := time.Date(2022, 10, 31, 15, 0, 0, 0, time.UTC)
	timeout := 10 * time.Second
	lowerThanMinInterval := 500 * time.Millisecond
	minInterval := 2 * lowerThanMinInterval

	staticProvider := StaticSequenceProvider{
		Sequence: [][]uint16{
			{1000, 2000, 3000, 4000},
			{4000, 3000, 2000, 1000},
		},
	}
	tests := map[string]struct {
		steps []step
	}{
		"Valid sequence": {
			steps: []step{
				{"192.168.1.1", 1000, now, false, 0},
				{"192.168.1.1", 2000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.1", 3000, now.Add(2 * minInterval), false, 0},
				{"192.168.1.1", 4000, now.Add(3 * minInterval), true, 0},
			},
		},
		"Out of order sequence": {
			steps: []step{
				{"192.168.1.2", 1000, now, false, 0},
				{"192.168.1.2", 3000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.2", 2000, now.Add(2 * minInterval), false, 0},
				{"192.168.1.2", 4000, now.Add(3 * minInterval), false, 0},
			},
		},
		"Valid sequence with interspersed invalid attempts": {
			steps: []step{
				{"192.168.1.11", 1000, now, false, 0},
				{"192.168.1.11", 5000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.11", 2000, now.Add(2 * minInterval), false, 0},
				{"192.168.1.11", 5001, now.Add(3 * minInterval), false, 0},
				{"192.168.1.11", 3000, now.Add(4 * minInterval), false, 0},
				{"192.168.1.11", 4000, now.Add(5 * minInterval), true, 0},
			},
		},
		"Sequence invalidated by three consecutive wrong packets": {
			steps: []step{
				{"192.168.1.12", 1000, now, false, 0},
				{"192.168.1.12", 5000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.12", 5001, now.Add(2 * minInterval), false, 0},
				{"192.168.1.12", 5002, now.Add(3 * minInterval), false, 0},
				{"192.168.1.12", 2000, now.Add(4 * minInterval), false, 0},
				{"192.168.1.12", 3000, now.Add(5 * minInterval), false, 0},
				{"192.168.1.12", 4000, now.Add(6 * minInterval), false, 0},
			},
		},
		"Sequence with timeout": {
			steps: []step{
				{"192.168.1.4", 1000, now, false, 0},
				{"192.168.1.4", 2000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.4", 3000, now.Add(2*minInterval + timeout), false, 0},
				{"192.168.1.4", 1000, now.Add(3*minInterval + timeout), false, 0},
				{"192.168.1.4", 2000, now.Add(4*minInterval + timeout), false, 0},
				{"192.168.1.4", 3000, now.Add(5*minInterval + timeout), false, 0},
				{"192.168.1.4", 4000, now.Add(6*minInterval + timeout), true, 0},
			},
		},
		"Sequence too quick": {
			steps: []step{
				{"192.168.1.5", 1000, now, false, 0},
				{"192.168.1.5", 2000, now.Add(lowerThanMinInterval), false, 0},
				{"192.168.1.5", 3000, now.Add(lowerThanMinInterval + minInterval), false, 0},
				{"192.168.1.5", 4000, now.Add(lowerThanMinInterval + minInterval*2), false, 0},
				{"192.168.1.5", 1000, now.Add(lowerThanMinInterval + minInterval*3), false, 0},
				{"192.168.1.5", 2000, now.Add(lowerThanMinInterval + minInterval*4), false, 0},
				{"192.168.1.5", 3000, now.Add(lowerThanMinInterval + minInterval*5), false, 0},
				{"192.168.1.5", 4000, now.Add(lowerThanMinInterval + minInterval*6), true, 0},
			},
		},
		"Immediate repeated attempts": {
			steps: []step{
				{"192.168.1.6", 1000, now, false, 0},
				{"192.168.1.6", 1000, now.Add(lowerThanMinInterval), false, 0},
				{"192.168.1.6", 2000, now.Add(lowerThanMinInterval + minInterval), false, 0},
				{"192.168.1.6", 3000, now.Add(lowerThanMinInterval + minInterval*2), false, 0},
				{"192.168.1.6", 4000, now.Add(lowerThanMinInterval + minInterval*3), false, 0},
			},
		},
		"Sequence attempt with long pause": {
			steps: []step{
				{"192.168.1.7", 1000, now, false, 0},
				{"192.168.1.7", 2000, now.Add(1 * minInterval), false, 0},
				{"192.168.1.7", 3000, now.Add(1*minInterval + timeout), false, 0},
				{"192.168.1.7", 1000, now.Add(2*minInterval + timeout), false, 0},
				{"192.168.1.7", 2000, now.Add(3*minInterval + timeout), false, 0},
				{"192.168.1.7", 3000, now.Add(4*minInterval + timeout), false, 0},
				{"192.168.1.7", 4000, now.Add(5*minInterval + timeout), true, 0},
			},
		},
		"Correct sequence with exact minimum interval timing": {
			steps: []step{
				{"192.168.1.8", 1000, now, false, 0},
				{"192.168.1.8", 2000, now.Add(minInterval), false, 0},
				{"192.168.1.8", 3000, now.Add(minInterval * 2), false, 0},
				{"192.168.1.8", 4000, now.Add(minInterval * 3), true, 0},
			},
		},
		"Different IPs independent sequences": {
			steps: []step{
				{"192.168.1.9", 1000, now, false, 0},
				{"192.168.2.9", 1000, now.Add(minInterval), false, 0},
				{"192.168.2.9", 2000, now.Add(2 * minInterval), false, 0},
				{"192.168.2.9", 3000, now.Add(3 * minInterval), false, 0},
				{"192.168.2.9", 4000, now.Add(4 * minInterval), true, 0},
				{"192.168.1.9", 2000, now.Add(5 * minInterval), false, 0},
				{"192.168.1.9", 3000, now.Add(6 * minInterval), false, 0},
				{"192.168.1.9", 4000, now.Add(7 * minInterval), true, 0},
			},
		},
		"Invalid port before starting a valid sequence": {
			steps: []step{
				{"192.168.1.10", 5000, now, false, 0},
				{"192.168.1.10", 1000, now.Add(1 * time.Second), false, 0},
				{"192.168.1.10", 2000, now.Add(2 * time.Second), false, 0},
				{"192.168.1.10", 3000, now.Add(3 * time.Second), false, 0},
				{"192.168.1.10", 4000, now.Add(4 * time.Second), true, 0},
			},
		},
		"sequence change in middle ": {
			steps: []step{
				{"192.168.1.20", 1000, now, false, 0},
				{"192.168.1.20", 2000, now.Add(1 * time.Second), false, 0},
				{"192.168.1.20", 3000, now.Add(2 * time.Second), false, 1},
				{"192.168.1.20", 4000, now.Add(3 * time.Second), false, 1},
			},
		},
		"valid when client and serve change sequence at the same time ": {
			steps: []step{
				{"192.168.1.20", 1000, now, false, 0},
				{"192.168.1.20", 2000, now.Add(1 * time.Second), false, 0},
				{"192.168.1.20", 2000, now.Add(2 * time.Second), false, 1},
				{"192.168.1.20", 1000, now.Add(3 * time.Second), true, 1},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tracker := NewKnocker(&staticProvider, timeout, minInterval)
			tracker.MaxAllowedConsecutiveError = 2
			for i, step := range tt.steps {
				staticProvider.SequenceId = step.sequenceNumber
				if got := tracker.CheckSequence(step.srcIP, step.port, step.timestamp); got != step.valid {
					t.Fatalf("[%d] got = %v, want %v for step srcIP %s port %d", i, got, step.valid, step.srcIP, step.port)
				}
			}
		})
	}
}
