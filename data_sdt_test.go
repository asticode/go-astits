package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var sdt = &SDTData{
	OriginalNetworkID: 2,
	Services: []*SDTDataService{{
		Descriptors:            descriptors,
		HasEITPresentFollowing: true,
		HasEITSchedule:         true,
		HasFreeCSAMode:         true,
		RunningStatus:          5,
		ServiceID:              3,
	}},
	TransportStreamID: 1,
}

func sdtBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteBits(uint64(2), 16) // Original network ID
	w.WriteByte(uint8(0))      // Reserved for future use
	w.WriteBits(uint64(3), 16) // Service #1 id
	WriteBinary(w, "000000")   // Service #1 reserved for future use
	WriteBinary(w, "1")        // Service #1 EIT schedule flag
	WriteBinary(w, "1")        // Service #1 EIT present/following flag
	WriteBinary(w, "101")      // Service #1 running status
	WriteBinary(w, "1")        // Service #1 free CA mode
	descriptorsBytes(w)        // Service #1 descriptors
	return buf.Bytes()
}

func TestParseSDTSection(t *testing.T) {
	b := sdtBytes()
	r := bitio.NewCountReader(bytes.NewReader(b))
	d, err := parseSDTSection(r, int64(len(b)*8), uint16(1))
	assert.Equal(t, d, sdt)
	assert.NoError(t, err)
}
