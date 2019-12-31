package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
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
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("0000")         // Reserved for future use
	descriptorsBytes(w)     // Network descriptors
	w.Write("0000")         // Reserved for future use
	w.Write("000000001001") // Transport stream loop length
	w.Write(uint16(2))      // Transport stream #1 id
	w.Write(uint16(3))      // Transport stream #1 original network id
	w.Write("0000")         // Transport stream #1 reserved for future use
	descriptorsBytes(w)     // Transport stream #1 descriptors
	return buf.Bytes()
}

func TestParseNITSection(t *testing.T) {
	var b = nitBytes()
	d, err := parseNITSection(astikit.NewBytesIterator(b), uint16(1))
	assert.Equal(t, d, nit)
	assert.NoError(t, err)
}
