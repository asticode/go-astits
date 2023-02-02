package astits

import "sync"

// poolOfTempPayload global variable is used to ease access to pool from any place of the code
var poolOfTempPayload = &poolTempPayload{
	sp: sync.Pool{
		New: func() interface{} {
			// Prepare the slice of somewhat sensible initial size to minimize calls to runtime.growslice
			return &tempPayload{
				s: make([]byte, 0, 1024),
			}
		},
	},
}

// tempPayload is an object containing payload slice
type tempPayload struct {
	s []byte
}

// poolTempPayload is a pool for temporary payload in parseData()
// Don't use it anywhere else to avoid pool pollution
type poolTempPayload struct {
	sp sync.Pool
}

// get returns the tempPayload object with byte slice of a 'size' length
func (ptp *poolTempPayload) get(size int) (payload *tempPayload) {
	payload = ptp.sp.Get().(*tempPayload)
	// Reset slice length or grow it to requested size for use with copy
	if cap(payload.s) >= size {
		payload.s = payload.s[:size]
	} else {
		n := size - cap(payload.s)
		payload.s = append(payload.s[:cap(payload.s)], make([]byte, n)...)[:size]
	}
	return
}

// put returns reference to the payload slice back to pool
// Don't use the payload after a call to put
func (ptp *poolTempPayload) put(payload *tempPayload) {
	ptp.sp.Put(payload)
}
