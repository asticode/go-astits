package astits

import (
	"testing"

	"github.com/asticode/go-astitools/binary"
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
	w := astibinary.New()
	w.Write("0000")         // Reserved for future use
	descriptorsBytes(w)     // Network descriptors
	w.Write("0000")         // Reserved for future use
	w.Write("000000001001") // Transport stream loop length
	w.Write(uint16(2))      // Transport stream #1 id
	w.Write(uint16(3))      // Transport stream #1 original network id
	w.Write("0000")         // Transport stream #1 reserved for future use
	descriptorsBytes(w)     // Transport stream #1 descriptors
	return w.Bytes()
}

func TestParseNITSection(t *testing.T) {
	var offset int
	var b = nitBytes()
	d := parseNITSection(b, &offset, uint16(1))
	assert.Equal(t, d, nit)
}
