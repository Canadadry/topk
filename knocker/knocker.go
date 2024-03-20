package knocker

import (
	"time"
)

type PortSequenceProvider interface {
	GetSequence(srcIP string, timestamp time.Time) []uint16
}

type sequenceInfo struct {
	consecutiveMistakes int
	hits                int
	lastAttemptAt       time.Time
	Ip                  string
}

func (s *sequenceInfo) resetSequence() {
	s.hits = 0
}

type SequenceTracker struct {
	provider                   PortSequenceProvider
	sequences                  map[string]*sequenceInfo
	timeout                    time.Duration
	minInterval                time.Duration
	MaxAllowedConsecutiveError int
}

func NewKnocker(provider PortSequenceProvider, timeout, minInterval time.Duration) *SequenceTracker {
	return &SequenceTracker{
		provider:                   provider,
		sequences:                  map[string]*sequenceInfo{},
		timeout:                    timeout,
		minInterval:                minInterval,
		MaxAllowedConsecutiveError: 2,
	}
}

func (s *SequenceTracker) CheckSequence(srcIP string, port uint16, timestamp time.Time) bool {
	info := s.getIpInfo(srcIP, timestamp)

	validTiming := s.isValidTiming(info, timestamp)
	info.lastAttemptAt = timestamp

	if !validTiming {
		return false
	}

	return s.processPort(info, port)
}

func (s *SequenceTracker) getIpInfo(srcIP string, timestamp time.Time) *sequenceInfo {
	info, ok := s.sequences[srcIP]
	if !ok {
		info = &sequenceInfo{
			lastAttemptAt: timestamp.Add(-s.minInterval),
			Ip:            srcIP,
		}
		s.sequences[srcIP] = info
	}
	return info
}

func (s *SequenceTracker) isValidTiming(info *sequenceInfo, timestamp time.Time) bool {
	isTooQuick := func(lastTimestamp, timestamp time.Time) bool {
		return timestamp.Before(lastTimestamp.Add(s.minInterval))
	}

	isTooSlow := func(lastTimestamp, timestamp time.Time) bool {
		return !timestamp.Before(lastTimestamp.Add(s.timeout))
	}

	if isTooQuick(info.lastAttemptAt, timestamp) || isTooSlow(info.lastAttemptAt, timestamp) {
		info.resetSequence()
		return false
	}

	return true
}
func (s *SequenceTracker) processPort(info *sequenceInfo, port uint16) bool {
	sequence := s.provider.GetSequence(info.Ip, info.lastAttemptAt)

	if info.hits >= len(sequence) {
		return false
	}

	if sequence[info.hits] == port {
		info.hits++
		info.consecutiveMistakes = 0
		return len(sequence) == info.hits
	}

	info.consecutiveMistakes++
	if info.consecutiveMistakes > s.MaxAllowedConsecutiveError {
		info.resetSequence()
	}
	return false

}
