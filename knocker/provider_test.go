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
