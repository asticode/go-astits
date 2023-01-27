package astits

import "sync"

var poolOfPacketSlices = &packetSlicesPool{
	sp: sync.Pool{
		New: func() interface{} {
			// Prepare slice of somewhat sensible initial size to minimise calls to runtime.growslice
			ps := make([]*Packet, 0, 64)
			return &ps
		},
	},
}

type packetSlicesPool struct {
	sp sync.Pool
}

func (psp *packetSlicesPool) get() []*Packet {
	// Reset slice length to use with append
	return (*(psp.sp.Get().(*[]*Packet)))[:0]
}

func (psp *packetSlicesPool) put(ps []*Packet) {
	psp.sp.Put(&ps)
}

var poolOfData = &tempDataPool{
	sp: sync.Pool{
		New: func() interface{} {
			// Prepare slice of somewhat sensible initial size to minimise calls to runtime.growslice
			d := make([]byte, 0, 1024)
			return &d
		},
	},
}

type tempDataPool struct {
	sp sync.Pool
}

func (tdp *tempDataPool) get(size int) (payload []byte) {
	payload = *(tdp.sp.Get().(*[]byte))
	// Reset slice length or grow it to requested size to use with copy
	if cap(payload) >= size {
		payload = payload[:size]
	} else {
		n := size - cap(payload)
		payload = append(payload[:cap(payload)], make([]byte, n)...)[:size]
	}
	return
}

func (tdp *tempDataPool) put(payload []byte) {
	tdp.sp.Put(&payload)
}
