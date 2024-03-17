package knocker

import (
	"testing"
	"time"
)

func TestGetSequenceCases(t *testing.T) {
	provider := TimeBasedSequenceProvider{
		Salt:             "TestSalt",
		MinPort:          49152,
		MaxPort:          65535,
		IntervalInSecond: 15,
		SequenceLen:      4,
	}

	timestamp := time.Date(2022, 10, 31, 15, 0, 0, 0, time.UTC)
	ip := []string{"192.168.1.1", "192.168.1.2"}

	testCases := map[string]struct {
		srcIP            string
		timestamp        time.Time
		expectedSequence []uint16
		sameSequence     bool
	}{
		"Fixed timestamp reproducibility": {
			srcIP:            ip[0],
			timestamp:        timestamp,
			expectedSequence: provider.GetSequence(ip[0], timestamp),
			sameSequence:     true,
		},
		"Variable timestamp for difference": {
			srcIP:            ip[0],
			timestamp:        timestamp,
			expectedSequence: provider.GetSequence(ip[0], timestamp.Add(time.Hour)),
			sameSequence:     false,
		},
		"Same sequence for less than 15 second timestamp difference": {
			srcIP:            ip[0],
			timestamp:        timestamp,
			expectedSequence: provider.GetSequence(ip[0], timestamp.Add(10*time.Second)),
			sameSequence:     true,
		},
		"Different IPs, same timestamp": {
			srcIP:            ip[0],
			timestamp:        timestamp,
			expectedSequence: provider.GetSequence(ip[1], timestamp),
			sameSequence:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			sequence := provider.GetSequence(tc.srcIP, tc.timestamp)
			for i, port := range sequence {
				if port < provider.MinPort || port > provider.MaxPort {
					t.Fatalf("Port %d out of range: got %d, want between %d and %d", i, port, provider.MinPort, provider.MaxPort)
				}
			}
			if len(tc.expectedSequence) != len(sequence) {
				t.Fatalf("For sequence len  want %d  got %d", len(tc.expectedSequence), len(sequence))
			}

			if tc.sameSequence {
				for i := range sequence {
					if sequence[i] != tc.expectedSequence[i] {
						t.Errorf("Sequence port %d got %d want %d", i, sequence[i], tc.expectedSequence[i])
					}
				}
			} else {
				match := 0
				for i := range sequence {
					if sequence[i] == tc.expectedSequence[i] {
						match++
					}
				}
				if match == len(sequence) {
					t.Errorf("Sequence match given, should have been different")
				}
			}
		})
	}
}

func TestPortDistribution(t *testing.T) {
	provider := TimeBasedSequenceProvider{
		Salt:             "TestSalt",
		MinPort:          49152,
		MaxPort:          65535,
		IntervalInSecond: 15,
		SequenceLen:      4,
	}
	now := time.Date(2022, 10, 31, 15, 0, 0, 0, time.UTC)

	portCount := make(map[uint16]int)
	totalPorts := int(provider.MaxPort - provider.MinPort + 1)
	attempts := 10000000
	for i := 0; i < attempts; i++ {
		timestamp := now.Add(time.Duration(i*provider.IntervalInSecond) * time.Second)
		sequence := provider.GetSequence("192.168.1.1", timestamp)
		for _, port := range sequence {
			portCount[port]++
		}
	}

	percentOfTolerance := 0.05
	minCountThatShouldMatch := totalPorts - int(float64(totalPorts)*percentOfTolerance)

	avgAppearances := attempts * provider.SequenceLen / len(portCount)
	avgTolerance := int(float64(avgAppearances) * percentOfTolerance)
	minAppearances := avgAppearances - avgTolerance
	maxAppearances := avgAppearances + avgTolerance

	countOfMatchingPort := 0
	for _, count := range portCount {
		if count >= minAppearances && count <= maxAppearances {
			countOfMatchingPort++
		}
	}
	if countOfMatchingPort < minCountThatShouldMatch {
		t.Fatalf("%d ports match of wanted range of occurence want %d", countOfMatchingPort, minCountThatShouldMatch)
	}
}
