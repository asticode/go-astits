package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
)

func TestAutoDetectPacketSize(t *testing.T) {
	// Packet should start with a sync byte
	w := astibinary.New()
	w.Write(uint8(2))
	w.Write(byte(syncByte))
	_, err := autoDetectPacketSize(bytes.NewReader(w.Bytes()))
	assert.EqualError(t, err, ErrPacketMustStartWithASyncByte.Error())

	// Valid packet size
	w.Reset()
	w.Write(byte(syncByte))
	w.Write(make([]byte, 20))
	w.Write(byte(syncByte))
	w.Write(make([]byte, 166))
	w.Write(byte(syncByte))
	w.Write(make([]byte, 187))
	w.Write([]byte("test"))
	r := bytes.NewReader(w.Bytes())
	p, err := autoDetectPacketSize(r)
	assert.NoError(t, err)
	assert.Equal(t, 188, p)
	assert.Equal(t, 380, r.Len())
}
