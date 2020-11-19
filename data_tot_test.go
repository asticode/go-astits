package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

var tot = &TOTData{
	Descriptors: descriptors,
	UTCTime:     dvbTime,
}

func totBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(dvbTimeBytes) // UTC time
	w.Write("0000")       // Reserved
	descriptorsBytes(w)   // Service #1 descriptors
	return buf.Bytes()
}

func TestParseTOTSection(t *testing.T) {
	d, err := parseTOTSection(astikit.NewBytesIterator(totBytes()))
	removeOriginalBytesFromData(&Data{TOT: d})
	assert.Equal(t, d, tot)
	assert.NoError(t, err)
}
