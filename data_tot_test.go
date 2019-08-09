package astits

import (
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
	"github.com/asticode/go-astitools/byte"
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
	d, err := parseTOTSection(astibyte.NewIterator(totBytes()))
	assert.Equal(t, d, tot)
	assert.NoError(t, err)
}
