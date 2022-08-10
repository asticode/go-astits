package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var pat = &PATData{
	Programs: []*PATProgram{
		{ProgramMapID: 3, ProgramNumber: 2},
		{ProgramMapID: 5, ProgramNumber: 4},
	},
	TransportStreamID: 1,
}

func patBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteBits(uint64(2), 16)      // Program #1 number
	WriteBinary(w, "111")           // Program #1 reserved bits
	WriteBinary(w, "0000000000011") // Program #1 map ID
	w.WriteBits(uint64(4), 16)      // Program #2 number
	WriteBinary(w, "111")           // Program #2 reserved bits
	WriteBinary(w, "0000000000101") // Program #3 map ID
	return buf.Bytes()
}

func TestParsePATSection(t *testing.T) {
	b := patBytes()
	r := bitio.NewCountReader(bytes.NewReader(b))
	d, err := parsePATSection(r, int64(len(b)*8), uint16(1))
	assert.Equal(t, d, pat)
	assert.NoError(t, err)
}

func TestWritePATSection(t *testing.T) {
	bw := &bytes.Buffer{}
	w := bitio.NewWriter(bw)
	n, err := writePATSection(w, pat)
	assert.NoError(t, err)
	assert.Equal(t, n, 8)
	assert.Equal(t, n, bw.Len())
	assert.Equal(t, patBytes(), bw.Bytes())
}

func BenchmarkParsePATSection(b *testing.B) {
	b.ReportAllocs()
	bs := patBytes()

	for i := 0; i < b.N; i++ {
		r := bitio.NewCountReader(bytes.NewReader(bs))
		parsePATSection(r, int64(len(bs)), uint16(1))
	}
}

func BenchmarkWritePATSection(b *testing.B) {
	b.ReportAllocs()

	bw := &bytes.Buffer{}
	bw.Grow(1024)
	w := bitio.NewWriter(bw)

	for i := 0; i < b.N; i++ {
		bw.Reset()
		writePATSection(w, pat)
	}
}
