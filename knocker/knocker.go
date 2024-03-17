package knocker

// sequenceTracker keeps track of the sequence of ports knocked
type sequenceTracker struct {
	sequence []uint16
	hits     map[string][]int
	valid    func(string)
}

// newSequenceTracker creates a new sequence tracker with the given port sequence
func New(sequence []uint16) *sequenceTracker {
	return &sequenceTracker{
		sequence: sequence,
		hits:     make(map[string][]int),
	}
}

// checkSequence updates the sequence tracker with the given source IP and port, and returns true if the sequence is completed
func (s *sequenceTracker) CheckSequence(srcIP string, port uint16) bool {
	for ip, seq := range s.hits {
		if ip == srcIP {
			nextIndex := len(seq)
			if nextIndex < len(s.sequence) && s.sequence[nextIndex] == port {
				s.hits[srcIP] = append(s.hits[srcIP], nextIndex)
				if len(s.hits[srcIP]) == len(s.sequence) {
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
	if len(s.sequence) > 0 && s.sequence[0] == port {
		s.hits[srcIP] = []int{0} // Start sequence
	}
	return false
}
