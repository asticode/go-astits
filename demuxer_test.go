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

func TestDemuxerNextDataPMTComplex(t *testing.T) {
	// complex pmt with two tables (0xc0 and 0x2) split across two packets
	pmt := hexToBytes(`47403b1e00c0001500000100610000000000000100000000
0035e3e2d702b0b20001c50000eefdf01809044749e10b05
0441432d330504454143330504435545491beefdf0102a02
7e1f9700e9080c001f418507d04181eefef00f810706380f
ff1f003f0a04656e670081eefff00f8107061003ff1f003f
0a047370610086ef00f00f8a01009700e9080c001f418507
d041c0ef01f012050445545631a100e9080c001f418507d0
41c0ef02f013050445545631a20100e9080c001f47003b1f
418507d041c0ef03f008bf06496e76696469a5cff3afffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffffffffffffffffffff
ffffffffffffffffffffffffffffffff`)
	r := bytes.NewReader(pmt)
	assert.Equal(t, 188*2, r.Len())

	dmx := NewDemuxer(context.Background(), r, DemuxerOptPacketSize(188))
	dmx.programMap.set(59, 1)

	d, err := dmx.NextData()
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, uint16(59), d.FirstPacket.Header.PID)
	assert.NotNil(t, d.PMT)
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
