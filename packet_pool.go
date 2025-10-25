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

	// Empty buffer if we detect a discontinuity
	if hasDiscontinuity(mps, p) {
		// Reset current slice or make new
		if cap(mps) > 0 {
			mps = mps[:0]
		} else {
			mps = make([]*Packet, 0, 10)
		}
	}

	// Throw away packet if it's the same as the previous one
	if isSameAsPrevious(mps, p) {
		return
	}

	// Flush buffer if new payload starts here
	if p.Header.PayloadUnitStartIndicator {
		ps = mps
		mps = make([]*Packet, 0, cap(mps))
	}

	mps = append(mps, p)

	// Check if PSI payload is complete
	if b.programMap != nil &&
		(b.pid == PIDPAT || b.programMap.existsUnlocked(b.pid)) &&
		isPSIComplete(mps) {
		ps = mps
		mps = nil
	} else if len(mps) > 0 && isPESComplete(mps) {
		// Check if PES payload is complete based on PES header length
		ps = mps
		mps = nil
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
			return
		}
	}
	return
}

// hasDiscontinuity checks whether a packet is discontinuous with a set of packets
func hasDiscontinuity(ps []*Packet, p *Packet) bool {
	l := len(ps)
	return (p.Header.HasAdaptationField && p.AdaptationField.DiscontinuityIndicator) || (l > 0 && ((p.Header.HasPayload && p.Header.ContinuityCounter != (ps[l-1].Header.ContinuityCounter+1)%16) ||
		(!p.Header.HasPayload && p.Header.ContinuityCounter != ps[l-1].Header.ContinuityCounter)))
}

// isSameAsPrevious checks whether a packet is the same as the last packet of a set of packets
func isSameAsPrevious(ps []*Packet, p *Packet) bool {
	l := len(ps)
	return l > 0 && p.Header.HasPayload && p.Header.ContinuityCounter == ps[l-1].Header.ContinuityCounter
}

// isPESComplete checks whether we have sufficient packets to form a complete PES packet
func isPESComplete(ps []*Packet) bool {
	// Must have at least one packet with PUSI to start the PES
	if len(ps) == 0 || !ps[0].Header.PayloadUnitStartIndicator {
		return false
	}

	// Calculate total payload length
	var totalPayload int
	for _, p := range ps {
		totalPayload += len(p.Payload)
	}

	// Need at least 6 bytes for PES header
	if totalPayload < 6 {
		return false
	}

	// Check if payload looks like PES (starts with 0x00 0x00 0x01)
	payload := ps[0].Payload
	if len(payload) < 3 || payload[0] != 0x00 || payload[1] != 0x00 || payload[2] != 0x01 {
		return false
	}

	// Try to extract PES packet length from header
	if totalPayload < 6 {
		return false
	}

	// Get first 6 bytes to read PES header
	pesHeader := make([]byte, 0, 6)
	remaining := 6
	for _, p := range ps {
		if remaining <= 0 {
			break
		}
		take := len(p.Payload)
		if take > remaining {
			take = remaining
		}
		pesHeader = append(pesHeader, p.Payload[:take]...)
		remaining -= take
	}

	if len(pesHeader) < 6 {
		return false
	}

	// Extract packet length (bytes 4-5, big endian)
	packetLength := (uint16(pesHeader[4]) << 8) | uint16(pesHeader[5])

	// If packet length is 0, this indicates variable length (video streams)
	// For now, we can't determine completeness, so return false
	if packetLength == 0 {
		return false
	}

	// Calculate expected total PES size (6 bytes header + packet length)
	expectedTotalSize := 6 + int(packetLength)

	// Check if we have enough payload
	return totalPayload >= expectedTotalSize
}
