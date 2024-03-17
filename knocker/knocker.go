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
	if lastTimestamp, ok := s.timestamps[srcIP]; ok {
		if timestamp.Before(lastTimestamp.Add(s.minInterval)) {
			// Step too quick, reset the sequence for this IP
			delete(s.hits, srcIP)
			delete(s.timestamps, srcIP)
			// As we're here because of a quick follow-up, we don't want to immediately start a new sequence
			// Return false to indicate the sequence is not completed
			return false
		} else if !timestamp.Before(lastTimestamp.Add(s.timeout)) {
			// Step too slow, also reset the sequence
			delete(s.hits, srcIP)
			delete(s.timestamps, srcIP)
			// Although it's a reset, the step might be the start of a new sequence, so don't return here
		}
	}

	currentSequence := s.provider.GetSequence(srcIP, timestamp)
	currentHits, ok := s.hits[srcIP]

	// New IP or starting over
	if !ok {
		if len(currentSequence) > 0 && currentSequence[0] == port {
			s.hits[srcIP] = []int{0} // Start sequence
			s.timestamps[srcIP] = timestamp
			return false
		}
	} else {
		// Existing sequence, check next step
		nextIndex := len(currentHits)
		if nextIndex < len(currentSequence) && currentSequence[nextIndex] == port {
			s.hits[srcIP] = append(s.hits[srcIP], nextIndex)
			s.timestamps[srcIP] = timestamp
			if len(s.hits[srcIP]) == len(currentSequence) {
				// Sequence completed
				delete(s.hits, srcIP) // Reset for this IP
				delete(s.timestamps, srcIP)
				return true
			}
			return false
		} else {
			// Incorrect port for the next step in the sequence, reset
			delete(s.hits, srcIP)
			delete(s.timestamps, srcIP)
		}
	}

	// If we reach here, the step did not complete a valid sequence.
	return false
}
