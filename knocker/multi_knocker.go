package knocker

import "time"

type Sequencer interface {
	CheckSequence(srcIP string, port uint16, timestamp time.Time) bool
}

func CheckMultiSequence(trackers map[string]Sequencer, srcIP string, port uint16, timestamp time.Time) (string, bool) {
	for name, tracker := range trackers {
		if tracker.CheckSequence(srcIP, port, timestamp) {
			return name, true
		}
	}
	return "", false
}
