package astits

import (
	"bytes"
	"context"
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
)

func TestDemuxerAutoDetectPacketSize(t *testing.T) {
	// Packet should start with a sync byte
	w := astibinary.New()
	w.Write(uint8(2))
	w.Write(byte(syncByte))
	dmx := New(context.Background(), bytes.NewReader(w.Bytes()))
	err := dmx.autoDetectPacketSize()
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
	dmx = New(context.Background(), r)
	err = dmx.autoDetectPacketSize()
	assert.NoError(t, err)
	assert.Equal(t, 188, dmx.PacketSize)
	assert.Equal(t, 4, r.Len())
}

func TestDemuxerNextPacket(t *testing.T) {
	// Ctx error
	ctx, cancel := context.WithCancel(context.Background())
	dmx := New(ctx, bytes.NewReader([]byte{}))
	cancel()
	_, err := dmx.NextPacket()
	assert.Error(t, err)

	// Valid
	w := astibinary.New()
	b1, p1 := packet(*packetHeader, *packetAdaptationField, []byte("1"))
	w.Write(b1)
	b2, p2 := packet(*packetHeader, *packetAdaptationField, []byte("2"))
	w.Write(b2)
	dmx = New(context.Background(), bytes.NewReader(w.Bytes()))

	// First packet
	p, err := dmx.NextPacket()
	assert.NoError(t, err)
	assert.Equal(t, p1, p)
	assert.Equal(t, 192, dmx.PacketSize)

	// Second packet
	p, err = dmx.NextPacket()
	assert.NoError(t, err)
	assert.Equal(t, p2, p)

	// EOF
	_, err = dmx.NextPacket()
	assert.EqualError(t, err, ErrNoMorePackets.Error())
}

func TestDemuxerNextData(t *testing.T) {
	// Init
	w := astibinary.New()
	b := psiBytes()
	b1, _ := packet(PacketHeader{ContinuityCounter: uint8(0), PayloadUnitStartIndicator: true, PID: PIDPAT}, PacketAdaptationField{}, b[:147])
	w.Write(b1)
	b2, _ := packet(PacketHeader{ContinuityCounter: uint8(1), PID: PIDPAT}, PacketAdaptationField{}, b[147:])
	w.Write(b2)
	dmx := New(context.Background(), bytes.NewReader(w.Bytes()))

	// Next data
	var ds []*Data
	for _, s := range psi.Sections {
		if s.Header.TableType != PSITableTypeUnknown {
			d, err := dmx.NextData()
			assert.NoError(t, err)
			ds = append(ds, d)
		}
	}
	assert.Equal(t, psi.toData(PIDPAT), ds)
	assert.Equal(t, map[uint16]uint16{0x3: 0x2, 0x5: 0x4}, dmx.programMap.p)

	// No more packets
	_, err := dmx.NextData()
	assert.EqualError(t, err, ErrNoMorePackets.Error())
}

func TestDemuxerRewind(t *testing.T) {
	r := bytes.NewReader([]byte("content"))
	dmx := New(context.Background(), r)
	dmx.packetBuffer.add(&Packet{Header: &PacketHeader{PID: 1}})
	dmx.dataBuffer = append(dmx.dataBuffer, &Data{})
	b := make([]byte, 2)
	_, err := r.Read(b)
	assert.NoError(t, err)
	n, err := dmx.Rewind()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, 7, r.Len())
	assert.Equal(t, 0, len(dmx.dataBuffer))
	assert.Equal(t, 0, len(dmx.packetBuffer.b))
}
