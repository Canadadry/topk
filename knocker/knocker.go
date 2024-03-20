package knocker

import (
	"time"
)

type PortSequenceProvider interface {
	GetSequence(srcIP string, timestamp time.Time) []uint16
}

type sequenceInfo struct {
	consecutiveMistakes int
	expectedNextPorts   []uint16
	lastAttemptAt       time.Time
}

func (s *sequenceInfo) resetSequence() {
	s.expectedNextPorts = nil
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

	if !s.isValidTiming(info, timestamp) {
		return false
	}

	info.lastAttemptAt = timestamp

	return s.processPort(info, port)
}

func (s *SequenceTracker) getIpInfo(srcIP string, timestamp time.Time) *sequenceInfo {
	info, ok := s.sequences[srcIP]
	if !ok {
		info = &sequenceInfo{
			lastAttemptAt: timestamp.Add(-s.minInterval),
		}
		s.sequences[srcIP] = info
	}
	if len(info.expectedNextPorts) == 0 {
		info.expectedNextPorts = s.provider.GetSequence(srcIP, timestamp)
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
	if len(info.expectedNextPorts) == 0 {
		return false
	}

	if info.expectedNextPorts[0] != port {
		info.consecutiveMistakes++
	}

	if info.consecutiveMistakes > s.MaxAllowedConsecutiveError {
		info.resetSequence()
		return false
	}

	info.expectedNextPorts = info.expectedNextPorts[1:]
	info.consecutiveMistakes = 0

	return len(info.expectedNextPorts) == 0
}
