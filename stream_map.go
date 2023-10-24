package astits

// elementaryStreamMap represents an elementary stream ids map
type elementaryStreamMap struct {
	// We use map[uint32] instead map[uint16] as go runtime provide optimized hash functions for (u)int32/64 keys
	es map[uint32]uint16 // map[StreamID]ProgramNumber
}

// newElementaryStreamMap creates a new elementary stream ids map
func newElementaryStreamMap() *elementaryStreamMap {
	return &elementaryStreamMap{
		es: make(map[uint32]uint16),
	}
}

// setLocked sets a new stream id to the elementary stream
func (m elementaryStreamMap) setLocked(pid, number uint16) {
	m.es[uint32(pid)] = number
}

// existsLocked checks whether the stream with this pid exists
func (m elementaryStreamMap) existsLocked(pid uint16) (ok bool) {
	_, ok = m.es[uint32(pid)]
	return
}

func (m elementaryStreamMap) unsetLocked(pid uint16) {
	delete(m.es, uint32(pid))
}
