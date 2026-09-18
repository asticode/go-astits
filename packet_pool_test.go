package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasDiscontinuity(t *testing.T) {
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 0, HasPayload: true}}))
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 15}}))
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{AdaptationField: &PacketAdaptationField{DiscontinuityIndicator: true}, Header: PacketHeader{ContinuityCounter: 0, HasAdaptationField: true, HasPayload: true}}))
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{PID: PIDNull}}))
	assert.False(t, hasDiscontinuity([]*Packet{}, &Packet{Header: PacketHeader{ContinuityCounter: 5}}))
	assert.True(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 1, HasPayload: true}}))
	assert.True(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 0}}))
}

func TestIsSameAsPrevious(t *testing.T) {
	assert.False(t, isSameAsPrevious([]*Packet{{Header: PacketHeader{ContinuityCounter: 1}}}, &Packet{Header: PacketHeader{ContinuityCounter: 1}}))
	assert.False(t, isSameAsPrevious([]*Packet{{Header: PacketHeader{ContinuityCounter: 1}}}, &Packet{Header: PacketHeader{ContinuityCounter: 2, HasPayload: true}}))
	assert.True(t, isSameAsPrevious([]*Packet{{Header: PacketHeader{ContinuityCounter: 1}}}, &Packet{Header: PacketHeader{ContinuityCounter: 1, HasPayload: true}}))
}

func TestPacketPool(t *testing.T) {
	b := newPacketPool(nil)
	ps := b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 0, HasPayload: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 1, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 1, HasPayload: true, PayloadUnitStartIndicator: true, PID: 2}})
	assert.Len(t, ps, 0)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 2, HasPayload: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 3, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 2)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 5, HasPayload: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 6, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 7, HasPayload: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.dumpUnlocked()
	assert.Len(t, ps, 2)
	assert.Equal(t, uint16(1), ps[0].Header.PID)
	ps = b.dumpUnlocked()
	assert.Len(t, ps, 1)
	assert.Equal(t, uint16(2), ps[0].Header.PID)
	ps = b.dumpUnlocked()
	assert.Len(t, ps, 0)
}

func TestPacketPoolWithRarePackets(t *testing.T) {
	payloadDVBTeletext := hexToBytes(`000001bd00b2848024293c972af5ffffffffffffffffffffffffffffffff
	ffffffffffffffffffffffffffffff10032cf5e4a8a80b0ba80ba80b2692
	040404040404040404040404040404040404040404040404040404040404
	0404032cd5e4a8a85757a8a8a80b26a80404040404040404040404040404
	040404040404040404040404040404040404ff2cffffffffffffffffffff
	ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
	ffffffff`)
	b := newPacketPool(nil)
	ps := b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 0, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 1, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 2, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 3, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 3, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 4, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 6, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 7, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 7, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 9, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.addUnlocked(&Packet{Header: PacketHeader{ContinuityCounter: 10, HasPayload: true, PayloadUnitStartIndicator: true, PID: 1004}, Payload: payloadDVBTeletext})
	assert.Len(t, ps, 1)
	ps = b.dumpUnlocked()
	assert.Len(t, ps, 0)
}

// An unbounded PES (packet length 0, e.g. a large video access unit written by
// GStreamer's mpegtsmux) must not be dropped when the PES that starts right
// after it is complete within that single packet.
func TestPacketPoolUnboundedPESFollowedByOnePacketPES(t *testing.T) {
	// pesPacketPayload builds a PES payload: start code, stream id, packet length
	// (0 = unbounded), minimal optional header, then n data bytes.
	pesPacketPayload := func(packetLength uint16, n int) []byte {
		p := []byte{0, 0, 1, 0xe0, byte(packetLength >> 8), byte(packetLength), 0x80, 0x00, 0x00}
		return append(p, make([]byte, n)...)
	}

	packet := func(cc uint8, pusi bool, payload []byte) *Packet {
		return &Packet{
			Header: PacketHeader{
				ContinuityCounter:         cc,
				HasPayload:                true,
				PID:                       256,
				PayloadUnitStartIndicator: pusi,
			},
			Payload: payload,
		}
	}

	pp := newPacketPool(nil)

	// Unbounded PES spanning two packets.
	p1 := packet(0, true, pesPacketPayload(0, 175))
	p2 := packet(1, false, make([]byte, 184))
	// PES that is complete within one packet: length covers optional header + data.
	p3 := packet(2, true, pesPacketPayload(3+16, 16))
	// Next payload start, flushes whatever is still queued.
	p4 := packet(3, true, pesPacketPayload(0, 175))

	assert.Empty(t, pp.addUnlocked(p1))
	assert.Empty(t, pp.addUnlocked(p2))

	ps := pp.addUnlocked(p3)
	assert.Equal(t, []*Packet{p1, p2}, ps, "the unbounded PES must be returned, not dropped")

	ps = pp.addUnlocked(p4)
	assert.Equal(t, []*Packet{p3}, ps, "the one-packet PES must be returned by the next flush")
}
