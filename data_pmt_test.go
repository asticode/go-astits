package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var pmt = &PMTData{
	ElementaryStreams: []*PMTElementaryStream{{
		ElementaryPID:               2730,
		ElementaryStreamDescriptors: descriptors,
		StreamType:                  StreamTypeMPEG1Audio,
	}},
	PCRPID:             5461,
	ProgramDescriptors: descriptors,
	ProgramNumber:      1,
}

func pmtBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	WriteBinary(w, "111")                    // Reserved bits
	WriteBinary(w, "1010101010101")          // PCR PID
	WriteBinary(w, "1111")                   // Reserved
	descriptorsBytes(w)                      // Program descriptors
	w.WriteByte(uint8(StreamTypeMPEG1Audio)) // Stream #1 stream type
	WriteBinary(w, "111")                    // Stream #1 reserved
	WriteBinary(w, "0101010101010")          // Stream #1 PID
	WriteBinary(w, "1111")                   // Stream #1 reserved
	descriptorsBytes(w)                      // Stream #1 descriptors
	return buf.Bytes()
}

func TestParsePMTSection(t *testing.T) {
	b := pmtBytes()
	r := bitio.NewCountReader(bytes.NewReader(b))
	d, err := parsePMTSection(r, int64(len(b)*8), uint16(1))
	assert.Equal(t, d, pmt)
	assert.NoError(t, err)
}

func TestWritePMTSection(t *testing.T) {
	buf := bytes.Buffer{}
	w := bitio.NewWriter(&buf)
	n, err := writePMTSection(w, pmt)
	assert.NoError(t, err)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, pmtBytes(), buf.Bytes())
}

func BenchmarkParsePMTSection(b *testing.B) {
	b.ReportAllocs()
	bs := pmtBytes()

	for i := 0; i < b.N; i++ {
		r := bitio.NewCountReader(bytes.NewReader(bs))
		parsePMTSection(r, int64(len(bs)), uint16(1))
	}
}

func BenchmarkWritePMTSection(b *testing.B) {
	b.ReportAllocs()

	bw := &bytes.Buffer{}
	bw.Grow(1024)
	w := bitio.NewWriter(bw)

	for i := 0; i < b.N; i++ {
		bw.Reset()
		writePMTSection(w, pmt)
	}
}
