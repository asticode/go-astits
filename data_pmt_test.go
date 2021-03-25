package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
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
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("111")                       // Reserved bits
	w.Write("1010101010101")             // PCR PID
	w.Write("1111")                      // Reserved
	descriptorsBytes(w)                  // Program descriptors
	w.Write(uint8(StreamTypeMPEG1Audio)) // Stream #1 stream type
	w.Write("111")                       // Stream #1 reserved
	w.Write("0101010101010")             // Stream #1 PID
	w.Write("1111")                      // Stream #1 reserved
	descriptorsBytes(w)                  // Stream #1 descriptors
	return buf.Bytes()
}

func TestParsePMTSection(t *testing.T) {
	var b = pmtBytes()
	d, err := parsePMTSection(astikit.NewBytesIterator(b), len(b), uint16(1))
	assert.Equal(t, d, pmt)
	assert.NoError(t, err)
}

func TestWritePMTSection(t *testing.T) {
	buf := bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
	n, err := writePMTSection(w, pmt)
	assert.NoError(t, err)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, pmtBytes(), buf.Bytes())
}

func BenchmarkParsePMTSection(b *testing.B) {
	b.ReportAllocs()
	bs := pmtBytes()

	for i := 0; i < b.N; i++ {
		parsePMTSection(astikit.NewBytesIterator(bs), len(bs), uint16(1))
	}
}

func BenchmarkWritePMTSection(b *testing.B) {
	b.ReportAllocs()

	bw := &bytes.Buffer{}
	bw.Grow(1024)
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: bw})

	for i := 0; i < b.N; i++ {
		bw.Reset()
		writePMTSection(w, pmt)
	}
}
