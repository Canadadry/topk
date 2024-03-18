package knocker

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"time"
)

type TimeBasedSequenceProvider struct {
	MinPort          uint16
	MaxPort          uint16
	Salt             string
	IntervalInSecond int
	SequenceLen      int
}

func (p *TimeBasedSequenceProvider) GetSequence(srcIP string, timestamp time.Time) []uint16 {
	adjustedTime := timestamp.Unix() / int64(p.IntervalInSecond)
	data := []byte(p.Salt + srcIP + fmt.Sprintf("%d", adjustedTime))

	hasher := sha1.New()
	hasher.Write(data)
	hashSum := hasher.Sum(nil)

	sequence := make([]uint16, p.SequenceLen)
	for i := 0; i < p.SequenceLen; i++ {
		sequence[i] = p.MinPort + binary.BigEndian.Uint16(hashSum[i*2:i*2+2])%(p.MaxPort-p.MinPort)
	}

	return sequence
}
