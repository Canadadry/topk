package knocker

import "time"

type PortSequenceProvider interface {
	GetSequence(srcIP string, timestamp time.Time) []uint16
}

type StaticSequenceProvider struct {
	Sequence []uint16
}

func (p *StaticSequenceProvider) GetSequence(srcIP string, timestamp time.Time) []uint16 {
	return p.Sequence
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

// Modify checkSequence to include a timestamp parameter
func (s *SequenceTracker) CheckSequence(srcIP string, port uint16, timestamp time.Time) bool {
	currentSequence := s.provider.GetSequence(srcIP, timestamp)

	if lastTimestamp, ok := s.timestamps[srcIP]; ok {
		// Check if the step is too quick after the last one
		if timestamp.Before(lastTimestamp.Add(s.minInterval)) {
			// Too quick, reset the sequence for this IP
			delete(s.hits, srcIP)
			delete(s.timestamps, srcIP)
			return false
		}
		// If the sequence is too slow, also reset
		if !timestamp.Before(lastTimestamp.Add(s.timeout)) {
			delete(s.hits, srcIP)
			delete(s.timestamps, srcIP)
		}
	}

	// If the IP is not new but the sequence is too slow, reset
	if lastTimestamp, ok := s.timestamps[srcIP]; ok && !timestamp.Before(lastTimestamp.Add(s.timeout)) {
		delete(s.hits, srcIP)
		delete(s.timestamps, srcIP)
	}

	for ip, seq := range s.hits {
		if ip == srcIP {
			nextIndex := len(seq)
			if nextIndex < len(currentSequence) && currentSequence[nextIndex] == port {
				s.hits[srcIP] = append(s.hits[srcIP], nextIndex)

				s.timestamps[srcIP] = timestamp
				if len(s.hits[srcIP]) == len(currentSequence) {
					// Sequence completed
					delete(s.hits, srcIP) // Reset for this IP
					return true
				}
				return false
			} else {
				// Sequence broken, start over for this IP
				delete(s.hits, srcIP)
			}
		}
	}
	// New IP or starting over
	if len(currentSequence) > 0 && currentSequence[0] == port {
		s.hits[srcIP] = []int{0} // Start sequence
	}
	return false
}
