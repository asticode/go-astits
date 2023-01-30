package astits

import "sync"

// poolOfPacketSlice global variable is used to ease access to pool from any place of the code
var poolOfPacketSlice = &poolPacketSlice{
	sp: sync.Pool{
		New: func() interface{} {
			// Prepare the slice of somewhat sensible initial size to minimise calls to runtime.growslice
			ps := make([]*Packet, 0, 64)
			return &ps
		},
	},
}

// poolOfTempPayload global variable is used to ease access to pool from any place of the code
var poolOfTempPayload = &poolTempPayload{
	sp: sync.Pool{
		New: func() interface{} {
			// Prepare the slice of somewhat sensible initial size to minimize calls to runtime.growslice
			d := make([]byte, 0, 1024)
			return &d
		},
	},
}

// poolPacketSlice is a pool of packet references slices
// You should use it whenever this kind of object created or destroyed
type poolPacketSlice struct {
	sp sync.Pool
}

// get returns the slice of packet references of a zero length and some capacity
func (pps *poolPacketSlice) get() []*Packet {
	// Reset slice length to use with append
	return (*(pps.sp.Get().(*[]*Packet)))[:0]
}

// put returns reference to packet slice back to pool
// Don't use packet slice after a call to put
func (pps *poolPacketSlice) put(ps []*Packet) {
	pps.sp.Put(&ps)
}

// poolTempPayload is a pool for temporary payload in parseData()
// Don't use it anywhere else to avoid pool pollution
type poolTempPayload struct {
	sp sync.Pool
}

// get returns the byte slice of a 'size' length
func (ptp *poolTempPayload) get(size int) (payload []byte) {
	payload = *(ptp.sp.Get().(*[]byte))
	// Reset slice length or grow it to requested size for use with copy
	if cap(payload) >= size {
		payload = payload[:size]
	} else {
		n := size - cap(payload)
		payload = append(payload[:cap(payload)], make([]byte, n)...)[:size]
	}
	return
}

// put returns reference to the payload slice back to pool
// Don't use the payload after a call to put
func (ptp *poolTempPayload) put(payload []byte) {
	ptp.sp.Put(&payload)
}
