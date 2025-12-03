package astits

import (
	"sort"
)

// packetAccumulator keeps track of packets for a single PID and decides when to flush them
type packetAccumulator struct {
	pid        uint16
	programMap *programMap
	q          []*Packet
}

// newPacketAccumulator creates a new packet queue for a single PID
func newPacketAccumulator(pid uint16, programMap *programMap) *packetAccumulator {
	return &packetAccumulator{
		pid:        pid,
		programMap: programMap,
	}
}

// add adds a new packet for this PID to the queue
func (b *packetAccumulator) add(p *Packet) (ps []*Packet) {
	mps := b.q

	var needSkipBufferMPS bool
	// Empty buffer if we detect a discontinuity
	if hasDiscontinuity(mps, p) {
		// Reset current slice or make new
		if cap(mps) > 0 {
			mps = mps[:0]
		} else {
			mps = make([]*Packet, 0, 10)
		}
	} else {
		needSkipBufferMPS = len(mps) == 0 || isPacketsAlreadySent(mps)
	}

	// Throw away packet if it's the same as the previous one
	if isSameAsPrevious(mps, p) {
		return
	}

	// Flush buffer if new payload starts here
	if p.Header.PayloadUnitStartIndicator {
		if !needSkipBufferMPS {
			ps = mps
		}
		mps = make([]*Packet, 0, cap(mps))
	}

	mps = append(mps, p)

	// Check if PSI payload is complete
	if b.programMap != nil &&
		(b.pid == PIDPAT || b.programMap.existsUnlocked(b.pid)) &&
		isPSIComplete(mps) {
		ps = mps
		mps = nil
	} else if needSkipBufferMPS && isPayloadCompletePES(p) {
		ps = mps
	}

	b.q = mps
	return
}

// packetPool represents a queue of packets for each PID in the stream
type packetPool struct {
	// We use map[uint32] instead map[uint16] as go runtime provide optimized hash functions for (u)int32/64 keys
	b map[uint32]*packetAccumulator // Indexed by PID

	programMap *programMap
}

// newPacketPool creates a new packet pool with an optional parser and programMap
func newPacketPool(programMap *programMap) *packetPool {
	return &packetPool{
		b: make(map[uint32]*packetAccumulator),

		programMap: programMap,
	}
}

// addUnlocked adds a new packet to the pool
func (b *packetPool) addUnlocked(p *Packet) (ps []*Packet) {
	// Throw away packet if error indicator
	if p.Header.TransportErrorIndicator {
		return
	}

	// Throw away packets that don't have a payload until we figure out what we're going to do with them
	// TODO figure out what we're going to do with them :D
	if !p.Header.HasPayload {
		return
	}

	// Make sure accumulator exists
	acc, ok := b.b[uint32(p.Header.PID)]
	if !ok {
		acc = newPacketAccumulator(p.Header.PID, b.programMap)
		b.b[uint32(p.Header.PID)] = acc
	}

	// Add to the accumulator
	return acc.add(p)
}

// dumpUnlocked dumps the packet pool by looking for the first item with packets inside
func (b *packetPool) dumpUnlocked() (ps []*Packet) {
	var keys []int
	for k := range b.b {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		ps = b.b[uint32(k)].q
		delete(b.b, uint32(k))
		if len(ps) > 0 {
			if isPacketsAlreadySent(ps) {
				ps = nil
				continue
			}
			return
		}
	}
	return
}

// hasDiscontinuity checks whether a packet is discontinuous with a set of packets
func hasDiscontinuity(ps []*Packet, p *Packet) bool {
	if p.Header.PID == PIDNull {
		return false
	}

	isDiscontinuityIndicator := p.Header.HasAdaptationField && p.AdaptationField.DiscontinuityIndicator
	if isDiscontinuityIndicator {
		return false
	}

	l := len(ps)
	if l == 0 {
		return false
	}

	var expectedContinuityCounter uint8
	if p.Header.HasPayload {
		expectedContinuityCounter = (ps[l-1].Header.ContinuityCounter + 1) & 0x0f
	} else {
		expectedContinuityCounter = ps[l-1].Header.ContinuityCounter
	}
	if expectedContinuityCounter == p.Header.ContinuityCounter {
		return false
	}

	return true
}

// isSameAsPrevious checks whether a packet is the same as the last packet of a set of packets
func isSameAsPrevious(ps []*Packet, p *Packet) bool {
	l := len(ps)
	return l > 0 && p.Header.HasPayload && p.Header.ContinuityCounter == ps[l-1].Header.ContinuityCounter
}

func isPayloadCompletePES(packet *Packet) bool {
	return packet.Header.PayloadUnitStartIndicator && isPESPayload(packet.Payload) && isPESComplete(packet.Payload)
}

func isPacketsAlreadySent(ps []*Packet) bool {
	return len(ps) == 1 && isPayloadCompletePES(ps[0])
}
