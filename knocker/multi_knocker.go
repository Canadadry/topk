package knocker

import "time"

type MultiSequenceTracker struct {
	trackers map[string]*SequenceTracker
}

func NewMulti(providers map[string]PortSequenceProvider, timeout, minInterval time.Duration) *MultiSequenceTracker {
	ms := &MultiSequenceTracker{
		trackers: map[string]*SequenceTracker{},
	}
	for name, provider := range providers {
		ms.trackers[name] = New(provider, timeout, minInterval)
	}
	return ms
}

func (ms *MultiSequenceTracker) CheckSequence(srcIP string, port uint16, timestamp time.Time) (string, bool) {
	for name, tracker := range ms.trackers {
		if tracker.CheckSequence(srcIP, port, timestamp) {
			return name, true
		}
	}
	return "", false
}
