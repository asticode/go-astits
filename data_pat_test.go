package astits

import (
	"testing"

	"github.com/asticode/go-astitools/binary"
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
	w := astibinary.New()
	w.Write(uint16(2))       // Program #1 number
	w.Write("111")           // Program #1 reserved bits
	w.Write("0000000000011") // Program #1 map ID
	w.Write(uint16(4))       // Program #2 number
	w.Write("111")           // Program #2 reserved bits
	w.Write("0000000000101") // Program #3 map ID
	return w.Bytes()
}

func TestParsePATSection(t *testing.T) {
	var offset int
	var b = patBytes()
	d := parsePATSection(b, &offset, len(b), uint16(1))
	assert.Equal(t, d, pat)
}
