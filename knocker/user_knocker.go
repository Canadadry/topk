package knocker

import "time"

type UserKnockerConfig struct {
	Users                    map[string]string
	SequenceLen              int
	MinPort                  uint16
	MaxPort                  uint16
	SequenceIntervalInSecond int
	Timeout                  time.Duration
	MinInterval              time.Duration
}

type UserKnocker struct {
	trackers map[string]Sequencer
}

func NewUserKnocker(conf UserKnockerConfig) *UserKnocker {
	uk := UserKnocker{
		trackers: map[string]Sequencer{},
	}
	for user, secret := range conf.Users {
		provider := &TimeBasedSequenceProvider{
			MinPort:          conf.MinPort,
			MaxPort:          conf.MaxPort,
			Salt:             secret,
			IntervalInSecond: conf.SequenceIntervalInSecond,
			SequenceLen:      conf.SequenceLen,
		}
		uk.trackers[user] = NewKnocker(provider, conf.Timeout, conf.MinInterval)
	}
	return &uk
}

func (uk *UserKnocker) CheckSequence(srcIP string, port uint16, timestamp time.Time) (string, bool) {
	return CheckMultiSequence(uk.trackers, srcIP, port, timestamp)
}
