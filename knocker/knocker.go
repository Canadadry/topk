package knocker

import "time"

type PortSequenceProvider interface {
	GetSequence(srcIP string, timestamp time.Time) []uint16
}

type SequenceTracker struct {
	provider    PortSequenceProvider
	hits        map[string][]int
	timestamps  map[string]time.Time
	timeout     time.Duration
	minInterval time.Duration
}

func New(provider PortSequenceProvider, timeout, minInterval time.Duration) *SequenceTracker {
	return &SequenceTracker{
		provider:    provider,
		hits:        make(map[string][]int),
		timestamps:  make(map[string]time.Time), // Initialize the map
		timeout:     timeout,
		minInterval: minInterval,
	}
}

func (s *SequenceTracker) CheckSequence(srcIP string, port uint16, timestamp time.Time) bool {
	if !s.isValidTiming(srcIP, port, timestamp) {
		return false
	}

	return s.processPort(srcIP, port, timestamp)
}

func (s *SequenceTracker) isValidTiming(srcIP string, port uint16, timestamp time.Time) bool {
	lastTimestamp, exists := s.timestamps[srcIP]
	if !exists {
		return true
	}

	if timestamp.Before(lastTimestamp.Add(s.minInterval)) {
		s.resetSequence(srcIP) // Too quick
		return false
	}

	if !timestamp.Before(lastTimestamp.Add(s.timeout)) {
		s.resetSequence(srcIP) // Too slow
		// Do not return here as it could be the start of a new sequence
	}

	return true
}

// processPort checks if the current port is part of the ongoing sequence or starts a new one
func (s *SequenceTracker) processPort(srcIP string, port uint16, timestamp time.Time) bool {
	currentSequence := s.provider.GetSequence(srcIP, timestamp)
	currentHits := s.hits[srcIP]

	if len(currentHits) == 0 && len(currentSequence) > 0 && currentSequence[0] == port {
		s.startSequence(srcIP, timestamp)
		return false
	}

	return s.continueOrResetSequence(srcIP, port, currentSequence, timestamp)
}

// resetSequence resets the tracking for a given IP
func (s *SequenceTracker) resetSequence(srcIP string) {
	delete(s.hits, srcIP)
	delete(s.timestamps, srcIP)
}

// startSequence initializes a sequence for an IP
func (s *SequenceTracker) startSequence(srcIP string, timestamp time.Time) {
	s.hits[srcIP] = []int{0} // Start with the first port hit
	s.timestamps[srcIP] = timestamp
}

// continueOrResetSequence handles the logic to continue an existing sequence or reset it based on the current port
func (s *SequenceTracker) continueOrResetSequence(srcIP string, port uint16, currentSequence []uint16, timestamp time.Time) bool {
	nextIndex := len(s.hits[srcIP])
	if nextIndex < len(currentSequence) && currentSequence[nextIndex] == port {
		s.hits[srcIP] = append(s.hits[srcIP], nextIndex)
		s.timestamps[srcIP] = timestamp
		if len(s.hits[srcIP]) == len(currentSequence) {
			s.resetSequence(srcIP) // Sequence completed
			return true
		}
		return false
	}

	s.resetSequence(srcIP) // Incorrect port, reset
	return false
}
