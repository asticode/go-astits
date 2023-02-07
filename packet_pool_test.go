package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasDiscontinuity(t *testing.T) {
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 0, HasPayload: true}}))
	assert.False(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: PacketHeader{ContinuityCounter: 15}}))
	assert.True(t, hasDiscontinuity([]*Packet{{Header: PacketHeader{ContinuityCounter: 15}}}, &Packet{AdaptationField: &PacketAdaptationField{DiscontinuityIndicator: true}, Header: PacketHeader{ContinuityCounter: 0, HasAdaptationField: true, HasPayload: true}}))
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
