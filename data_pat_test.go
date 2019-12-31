package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
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
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint16(2))       // Program #1 number
	w.Write("111")           // Program #1 reserved bits
	w.Write("0000000000011") // Program #1 map ID
	w.Write(uint16(4))       // Program #2 number
	w.Write("111")           // Program #2 reserved bits
	w.Write("0000000000101") // Program #3 map ID
	return buf.Bytes()
}

func TestParsePATSection(t *testing.T) {
	var b = patBytes()
	d, err := parsePATSection(astikit.NewBytesIterator(b), len(b), uint16(1))
	assert.Equal(t, d, pat)
	assert.NoError(t, err)
}
