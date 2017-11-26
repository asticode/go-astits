package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasDiscontinuity(t *testing.T) {
	assert.False(t, hasDiscontinuity([]*Packet{{Header: &PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: &PacketHeader{ContinuityCounter: 0}}))
	assert.True(t, hasDiscontinuity([]*Packet{{Header: &PacketHeader{ContinuityCounter: 15}}}, &Packet{AdaptationField: &PacketAdaptationField{DiscontinuityIndicator: true}, Header: &PacketHeader{ContinuityCounter: 0, HasAdaptationField: true}}))
	assert.True(t, hasDiscontinuity([]*Packet{{Header: &PacketHeader{ContinuityCounter: 15}}}, &Packet{Header: &PacketHeader{ContinuityCounter: 1}}))
}

func TestPacketPool(t *testing.T) {
	b := newPacketPool()
	ps := b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 0, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 1, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 1, PayloadUnitStartIndicator: true, PID: 2}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 2, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 3, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 2)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 5, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 6, PayloadUnitStartIndicator: true, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.add(&Packet{Header: &PacketHeader{ContinuityCounter: 7, PID: 1}})
	assert.Len(t, ps, 0)
	ps = b.dump()
	assert.Len(t, ps, 2)
	assert.Equal(t, uint16(1), ps[0].Header.PID)
	ps = b.dump()
	assert.Len(t, ps, 1)
	assert.Equal(t, uint16(2), ps[0].Header.PID)
	ps = b.dump()
	assert.Len(t, ps, 0)
}
