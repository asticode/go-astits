package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

func TestAutoDetectPacketSize(t *testing.T) {
	// Packet should start with a sync byte
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(uint8(2))
	w.WriteByte(byte(syncByte))
	_, err := autoDetectPacketSize(bytes.NewReader(buf.Bytes()))
	assert.ErrorIs(t, err, ErrPacketStartSyncByte)

	// Valid packet size
	buf.Reset()
	w.WriteByte(byte(syncByte))
	w.Write(make([]byte, 20))
	w.WriteByte(byte(syncByte))
	w.Write(make([]byte, 166))
	w.WriteByte(byte(syncByte))
	w.Write(make([]byte, 187))
	w.Write([]byte("test"))
	r := bytes.NewReader(buf.Bytes())
	p, err := autoDetectPacketSize(r)
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, p)
	assert.Equal(t, 380, r.Len())
}
