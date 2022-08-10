package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var nit = &NITData{
	NetworkDescriptors: descriptors,
	NetworkID:          1,
	TransportStreams: []*NITDataTransportStream{{
		OriginalNetworkID:    3,
		TransportDescriptors: descriptors,
		TransportStreamID:    2,
	}},
}

func nitBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	WriteBinary(w, "0000")         // Reserved for future use
	descriptorsBytes(w)            // Network descriptors
	WriteBinary(w, "0000")         // Reserved for future use
	WriteBinary(w, "000000001001") // Transport stream loop length
	w.WriteBits(uint64(2), 16)     // Transport stream #1 id
	w.WriteBits(uint64(3), 16)     // Transport stream #1 original network id
	WriteBinary(w, "0000")         // Transport stream #1 reserved for future use
	descriptorsBytes(w)            // Transport stream #1 descriptors
	return buf.Bytes()
}

func TestParseNITSection(t *testing.T) {
	b := nitBytes()
	r := bitio.NewCountReader(bytes.NewReader(b))
	d, err := parseNITSection(r, uint16(1))
	assert.Equal(t, d, nit)
	assert.NoError(t, err)
}
