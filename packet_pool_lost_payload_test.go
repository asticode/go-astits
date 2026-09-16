package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// pesPacketPayload builds a PES payload: start code, stream id, packet length
// (0 = unbounded), minimal optional header, then n data bytes.
func pesPacketPayload(packetLength uint16, n int) []byte {
	p := []byte{0, 0, 1, 0xe0, byte(packetLength >> 8), byte(packetLength), 0x80, 0x00, 0x00}
	return append(p, make([]byte, n)...)
}

func tsPacket(cc uint8, pusi bool, payload []byte) *Packet {
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

// An unbounded PES (packet length 0, e.g. a large video access unit written by
// GStreamer's mpegtsmux) must not be dropped when the PES that starts right
// after it is complete within that single packet.
func TestPacketPoolUnboundedPESFollowedByOnePacketPES(t *testing.T) {
	pp := newPacketPool(nil)

	// Unbounded PES spanning two packets.
	p1 := tsPacket(0, true, pesPacketPayload(0, 175))
	p2 := tsPacket(1, false, make([]byte, 184))
	// PES that is complete within one packet: length covers optional header + data.
	p3 := tsPacket(2, true, pesPacketPayload(3+16, 16))
	// Next payload start, flushes whatever is still queued.
	p4 := tsPacket(3, true, pesPacketPayload(0, 175))

	assert.Empty(t, pp.addUnlocked(p1))
	assert.Empty(t, pp.addUnlocked(p2))

	ps := pp.addUnlocked(p3)
	assert.Equal(t, []*Packet{p1, p2}, ps, "the unbounded PES must be returned, not dropped")

	ps = pp.addUnlocked(p4)
	assert.Equal(t, []*Packet{p3}, ps, "the one-packet PES must be returned by the next flush")
}
