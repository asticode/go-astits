package astits

import (
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
)

var tot = &TOTData{
	Descriptors: descriptors,
	UTCTime:     dvbTime,
}

func totBytes() []byte {
	w := astibinary.New()
	w.Write(dvbTimeBytes) // UTC time
	w.Write("0000")       // Reserved
	descriptorsBytes(w)   // Service #1 descriptors
	return w.Bytes()
}

func TestParseTOTSection(t *testing.T) {
	var offset int
	d := parseTOTSection(totBytes(), &offset)
	assert.Equal(t, d, tot)
}
