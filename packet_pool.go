package astits

import (
	"sort"
	"sync"
)

// packetAccumulator keeps track of packets for a single PID and decides when to flush them
type packetAccumulator struct {
	pid uint16
	q   []*Packet
}

// add adds a new packet for this PID to the queue
func (b *packetAccumulator) add(p *Packet) (ps []*Packet) {
	mps := b.q

	// Empty buffer if we detect a discontinuity
	if hasDiscontinuity(mps, p) {
		mps = []*Packet{}
	}

	// Throw away packet if it's the same as the previous one
	if isSameAsPrevious(mps, p) {
		return
	}

	// Flush buffer if new payload starts here
	if p.Header.PayloadUnitStartIndicator {
		ps = mps
		mps = []*Packet{p}
	} else {
		mps = append(mps, p)
	}

	b.q = mps
	return
}

// newPacketAccumulator creates a new packet queue for a single PID
func newPacketAccumulator(pid uint16) *packetAccumulator {
	return &packetAccumulator{
		pid: pid,
	}
}

// packetPool represents a queue of packets for each PID in the stream
type packetPool struct {
	b map[uint16]*packetAccumulator // Indexed by PID
	m *sync.Mutex
}

// newPacketPool creates a new packet pool
func newPacketPool() *packetPool {
	return &packetPool{
		b: make(map[uint16]*packetAccumulator),
		m: &sync.Mutex{},
	}
}

// add adds a new packet to the pool
func (b *packetPool) add(p *Packet) (ps []*Packet) {
	// Throw away packet if error indicator
	if p.Header.TransportErrorIndicator {
		return
	}

	// Throw away packets that don't have a payload until we figure out what we're going to do with them
	// TODO figure out what we're going to do with them :D
	if !p.Header.HasPayload {
		return
	}

	// Lock
	b.m.Lock()
	defer b.m.Unlock()

	// Init buffer
	var acc *packetAccumulator
	var ok bool
	if acc, ok = b.b[p.Header.PID]; !ok {
		acc = newPacketAccumulator(p.Header.PID)
	}

	// Add to the accumulator
	ps = acc.add(p)

	// Assign
	b.b[p.Header.PID] = acc
	return
}

// dump dumps the packet pool by looking for the first item with packets inside
func (b *packetPool) dump() (ps []*Packet) {
	b.m.Lock()
	defer b.m.Unlock()
	var keys []int
	for k := range b.b {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		ps = b.b[uint16(k)].q
		delete(b.b, uint16(k))
		if len(ps) > 0 {
			return
		}
	}
	return
}

// hasDiscontinuity checks whether a packet is discontinuous with a set of packets
func hasDiscontinuity(ps []*Packet, p *Packet) bool {
	return (p.Header.HasAdaptationField && p.AdaptationField.DiscontinuityIndicator) ||
		(len(ps) > 0 && p.Header.HasPayload && p.Header.ContinuityCounter != (ps[len(ps)-1].Header.ContinuityCounter+1)%16) ||
		(len(ps) > 0 && !p.Header.HasPayload && p.Header.ContinuityCounter != ps[len(ps)-1].Header.ContinuityCounter)
}

// isSameAsPrevious checks whether a packet is the same as the last packet of a set of packets
func isSameAsPrevious(ps []*Packet, p *Packet) bool {
	return len(ps) > 0 && p.Header.HasPayload && p.Header.ContinuityCounter == ps[len(ps)-1].Header.ContinuityCounter
}
