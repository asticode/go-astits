package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var tot = &TOTData{
	Descriptors: descriptors,
	UTCTime:     dvbTime,
}

func totBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.Write(dvbTimeBytes)  // UTC time.
	WriteBinary(w, "0000") // Reserved.
	descriptorsBytes(w)    // Service #1 descriptors.
	return buf.Bytes()
}

func TestParseTOTSection(t *testing.T) {
	r := bitio.NewCountReader(bytes.NewReader(totBytes()))
	d, err := parseTOTSection(r)
	assert.Equal(t, d, tot)
	assert.NoError(t, err)
}
