package knocker

import "time"

type PortSequenceProvider interface {
	GetSequence(srcIP string, timestamp time.Time) []uint16
}

type SequenceTracker struct {
	provider    PortSequenceProvider
	hits        map[string]int
	timestamps  map[string]time.Time
	timeout     time.Duration
	minInterval time.Duration
}

func New(provider PortSequenceProvider, timeout, minInterval time.Duration) *SequenceTracker {
	return &SequenceTracker{
		provider:    provider,
		hits:        make(map[string]int),
		timestamps:  make(map[string]time.Time), // Initialize the map
		timeout:     timeout,
		minInterval: minInterval,
	}
}

func (s *SequenceTracker) CheckSequence(srcIP string, port uint16, timestamp time.Time) bool {
	if !s.isValidTiming(srcIP, timestamp) {
		return false
	}

	return s.processPort(srcIP, port, timestamp)
}

func (s *SequenceTracker) isValidTiming(srcIP string, timestamp time.Time) bool {

	isTooQuick := func(lastTimestamp, timestamp time.Time) bool {
		return timestamp.Before(lastTimestamp.Add(s.minInterval))
	}

	isTooSlow := func(lastTimestamp, timestamp time.Time) bool {
		return !timestamp.Before(lastTimestamp.Add(s.timeout))
	}

	lastTimestamp, exists := s.timestamps[srcIP]
	if !exists {
		return true
	}

	if isTooQuick(lastTimestamp, timestamp) {
		s.resetSequence(srcIP)
		return false
	}

	if isTooSlow(lastTimestamp, timestamp) {
		s.resetSequence(srcIP)
		return true
	}

	return true
}

func (s *SequenceTracker) processPort(srcIP string, port uint16, timestamp time.Time) bool {
	currentSequence := s.provider.GetSequence(srcIP, timestamp)
	nextIndex := s.hits[srcIP]

	if nextIndex >= len(currentSequence) {
		s.resetSequence(srcIP)
		return false
	}

	if currentSequence[nextIndex] != port {
		s.resetSequence(srcIP)
		return false
	}

	s.hits[srcIP] += 1
	s.timestamps[srcIP] = timestamp

	if s.hits[srcIP] == len(currentSequence) {
		s.resetSequence(srcIP)
		return true
	}
	return false

}

func (s *SequenceTracker) resetSequence(srcIP string) {
	delete(s.hits, srcIP)
	delete(s.timestamps, srcIP)
}
