package astits

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func TestDemuxerNew(t *testing.T) {
	ps := 1
	pp := func(ps []*Packet) (ds []*DemuxerData, skip bool, err error) { return }
	dmx := NewDemuxer(context.Background(), nil, DemuxerOptPacketSize(ps), DemuxerOptPacketsParser(pp))
	assert.Equal(t, ps, dmx.optPacketSize)
	assert.Equal(t, fmt.Sprintf("%p", pp), fmt.Sprintf("%p", dmx.optPacketsParser))
}

func TestDemuxerNextPacket(t *testing.T) {
	// Ctx error
	ctx, cancel := context.WithCancel(context.Background())
	dmx := NewDemuxer(ctx, bytes.NewReader([]byte{}))
	cancel()
	_, err := dmx.NextPacket()
	assert.Error(t, err)

	// Valid
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	b1, p1 := packet(*packetHeader, *packetAdaptationField, []byte("1"), true)
	w.Write(b1)
	b2, p2 := packet(*packetHeader, *packetAdaptationField, []byte("2"), true)
	w.Write(b2)
	dmx = NewDemuxer(context.Background(), bytes.NewReader(buf.Bytes()))

	// First packet
	p, err := dmx.NextPacket()
	assert.NoError(t, err)
	assert.Equal(t, p1, p)
	assert.Equal(t, 192, dmx.packetBuffer.packetSize)

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
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	b := psiBytes()
	b1, _ := packet(PacketHeader{ContinuityCounter: uint8(0), PayloadUnitStartIndicator: true, PID: PIDPAT}, PacketAdaptationField{}, b[:147], true)
	w.Write(b1)
	b2, _ := packet(PacketHeader{ContinuityCounter: uint8(1), PID: PIDPAT}, PacketAdaptationField{}, b[147:], true)
	w.Write(b2)
	dmx := NewDemuxer(context.Background(), bytes.NewReader(buf.Bytes()))
	p, err := dmx.NextPacket()
	assert.NoError(t, err)
	_, err = dmx.Rewind()
	assert.NoError(t, err)

	// Next data
	var ds []*DemuxerData
	for _, s := range psi.Sections {
		if !s.Header.TableID.isUnknown() {
			d, err := dmx.NextData()
			assert.NoError(t, err)
			ds = append(ds, d)
		}
	}
	assert.Equal(t, psi.toData(p, PIDPAT), ds)
	assert.Equal(t, map[uint16]uint16{0x3: 0x2, 0x5: 0x4}, dmx.programMap.p)

	// No more packets
	_, err = dmx.NextData()
	assert.EqualError(t, err, ErrNoMorePackets.Error())
}

func TestDemuxerRewind(t *testing.T) {
	r := bytes.NewReader([]byte("content"))
	dmx := NewDemuxer(context.Background(), r)
	dmx.packetPool.add(&Packet{Header: &PacketHeader{PID: 1}})
	dmx.dataBuffer = append(dmx.dataBuffer, &DemuxerData{})
	b := make([]byte, 2)
	_, err := r.Read(b)
	assert.NoError(t, err)
	n, err := dmx.Rewind()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, 7, r.Len())
	assert.Equal(t, 0, len(dmx.dataBuffer))
	assert.Equal(t, 0, len(dmx.packetPool.b))
	assert.Nil(t, dmx.packetBuffer)
}

func BenchmarkDemuxer_NextData(b *testing.B) {
	b.ReportAllocs()

	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	bs := psiBytes()
	b1, _ := packet(PacketHeader{ContinuityCounter: uint8(0), PayloadUnitStartIndicator: true, PID: PIDPAT}, PacketAdaptationField{}, bs[:147], true)
	w.Write(b1)
	b2, _ := packet(PacketHeader{ContinuityCounter: uint8(1), PID: PIDPAT}, PacketAdaptationField{}, bs[147:], true)
	w.Write(b2)

	r := bytes.NewReader(buf.Bytes())

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		dmx := NewDemuxer(context.Background(), r)

		for _, s := range psi.Sections {
			if !s.Header.TableID.isUnknown() {
				dmx.NextData()
			}
		}
	}
}
