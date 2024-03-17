package knocker

import (
	"crypto/sha1"
	"encoding/binary"
	"time"
)

type StaticSequenceProvider struct {
	Sequence []uint16
}

func (p *StaticSequenceProvider) GetSequence(srcIP string, timestamp time.Time) []uint16 {
	return p.Sequence
}

type TimeBasedSequenceProvider struct {
	Salt string
}

func (p *TimeBasedSequenceProvider) GetSequence(srcIP string, timestamp time.Time) []uint16 {
	// Use the current hour and a salt to generate a hash
	hasher := sha1.New()
	data := []byte(p.Salt + srcIP + timestamp.Format("2006010215"))
	hasher.Write(data)
	hashSum := hasher.Sum(nil)

	// Generate a sequence from the hash
	// Here, we're simplistically picking parts of the hash to form port numbers
	// This is just a demonstration and might need adjustments for real applications
	sequence := make([]uint16, 4)
	for i := 0; i < 4; i++ {
		// Each port is derived from two bytes of the hash, converted to a uint16
		// and then made to fit into the dynamic port range (49152–65535, for example)
		minPort := uint16(49152)
		maxPort := uint16(65535)
		sequence[i] = minPort + binary.BigEndian.Uint16(hashSum[i*2:i*2+2])%(maxPort-minPort)
	}

	return sequence
}
