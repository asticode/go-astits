package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var eit = &EITData{
	Events: []*EITDataEvent{{
		Descriptors:    descriptors,
		Duration:       dvbDurationSeconds,
		EventID:        6,
		HasFreeCSAMode: true,
		RunningStatus:  7,
		StartTime:      dvbTime,
	}},
	LastTableID:              5,
	OriginalNetworkID:        3,
	SegmentLastSectionNumber: 4,
	ServiceID:                1,
	TransportStreamID:        2,
}

func eitBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteBits(uint64(2), 16)       // Transport stream ID
	w.WriteBits(uint64(3), 16)       // Original network ID
	w.WriteByte(uint8(4))            // Segment last section number
	w.WriteByte(uint8(5))            // Last table id
	w.WriteBits(uint64(6), 16)       // Event #1 id
	w.Write(dvbTimeBytes)            // Event #1 start time
	w.Write(dvbDurationSecondsBytes) // Event #1 duration
	WriteBinary(w, "111")            // Event #1 running status
	w.WriteBool(true)                // Event #1 free CA mode
	descriptorsBytes(w)              // Event #1 descriptors
	return buf.Bytes()
}

func TestParseEITSection(t *testing.T) {
	b := eitBytes()
	r := bitio.NewCountReader(bytes.NewReader(b))
	d, err := parseEITSection(r, int64(len(b)*8), uint16(1))
	assert.Equal(t, d, eit)
	assert.NoError(t, err)
}
