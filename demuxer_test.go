package astits

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func TestDemuxerNew(t *testing.T) {
	ps := 1
	pp := func(ps []*Packet) (ds []*Data, skip bool, err error) { return }
	dmx := New(context.Background(), nil, OptPacketSize(ps), OptPacketsParser(pp))
	assert.Equal(t, ps, dmx.optPacketSize)
	assert.Equal(t, fmt.Sprintf("%p", pp), fmt.Sprintf("%p", dmx.optPacketsParser))
}

func TestDemuxerNextPacket(t *testing.T) {
	// Ctx error
	ctx, cancel := context.WithCancel(context.Background())
	dmx := New(ctx, bytes.NewReader([]byte{}))
	cancel()
	_, err := dmx.NextPacket()
	assert.Error(t, err)

	// Valid
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	b1, p1 := packet(*packetHeader, *packetAdaptationField, []byte("1"))
	w.Write(b1)
	b2, p2 := packet(*packetHeader, *packetAdaptationField, []byte("2"))
	w.Write(b2)
	dmx = New(context.Background(), bytes.NewReader(buf.Bytes()))

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
	b1, _ := packet(PacketHeader{ContinuityCounter: uint8(0), PayloadUnitStartIndicator: true, PID: PIDPAT}, PacketAdaptationField{}, b[:147])
	w.Write(b1)
	b2, _ := packet(PacketHeader{ContinuityCounter: uint8(1), PID: PIDPAT}, PacketAdaptationField{}, b[147:])
	w.Write(b2)
	dmx := New(context.Background(), bytes.NewReader(buf.Bytes()))
	p, err := dmx.NextPacket()
	assert.NoError(t, err)
	_, err = dmx.Rewind()
	assert.NoError(t, err)

	// Next data
	var ds []*Data
	for _, s := range psi.Sections {
		if s.Header.TableType != PSITableTypeUnknown {
			d, err := dmx.NextData()
			assert.NoError(t, err)
			ds = append(ds, d)
		}
	}
	//Remove originalBytes field from all descriptors
	for i := range ds {
		removeOriginalBytesFromData(ds[i])
	}
	assert.Equal(t, psi.toData(p, PIDPAT), ds)
	assert.Equal(t, map[uint16]uint16{0x3: 0x2, 0x5: 0x4}, dmx.programMap.p)

	// No more packets
	_, err = dmx.NextData()
	assert.EqualError(t, err, ErrNoMorePackets.Error())
}

func TestDemuxerRewind(t *testing.T) {
	r := bytes.NewReader([]byte("content"))
	dmx := New(context.Background(), r)
	dmx.packetPool.Add(&Packet{Header: &PacketHeader{PID: 1}})
	dmx.dataBuffer = append(dmx.dataBuffer, &Data{})
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

func removeOriginalBytesFromData(d *Data) {
	if d.PMT != nil {
		for j := range d.PMT.ProgramDescriptors {
			d.PMT.ProgramDescriptors[j].originalBytes = nil
		}
		for k := range d.PMT.ElementaryStreams {
			for l := range d.PMT.ElementaryStreams[k].ElementaryStreamDescriptors {
				d.PMT.ElementaryStreams[k].ElementaryStreamDescriptors[l].originalBytes = nil
			}
		}
	}
	if d.EIT != nil {
		for j := range d.EIT.Events {
			for k := range d.EIT.Events[j].Descriptors {
				d.EIT.Events[j].Descriptors[k].originalBytes = nil
			}
		}
	}
	if d.NIT != nil {
		for j := range d.NIT.TransportStreams {
			for k := range d.NIT.TransportStreams[j].TransportDescriptors {
				d.NIT.TransportStreams[j].TransportDescriptors[k].originalBytes = nil
			}
		}
		for l := range d.NIT.NetworkDescriptors {
			d.NIT.NetworkDescriptors[l].originalBytes = nil
		}
	}
	if d.SDT != nil {
		for j := range d.SDT.Services {
			for k := range d.SDT.Services[j].Descriptors {
				d.SDT.Services[j].Descriptors[k].originalBytes = nil
			}
		}
	}
	if d.TOT != nil {
		for k := range d.TOT.Descriptors {
			d.TOT.Descriptors[k].originalBytes = nil
		}
	}
}
