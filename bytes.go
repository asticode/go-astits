package astits

import "fmt"

// This is copy-pasted from go-astikit package.
// This was done so I could add the `Reset` method to the BytesIterator, to
// avoid allocating a new one on every single TS packet.

// BytesIterator represents an object capable of iterating sequentially and safely
// through a slice of bytes. This is particularly useful when you need to iterate
// through a slice of bytes and don't want to check for "index out of range" errors
// manually.
type NoAllocBytesIterator struct {
	bs     []byte
	offset int
}

// NewBytesIterator creates a new BytesIterator
func NewNoAllocBytesIterator(bs []byte) *NoAllocBytesIterator {
	return &NoAllocBytesIterator{bs: bs}
}

func (i *NoAllocBytesIterator) Reset(bs []byte) {
	i.bs = bs
	i.offset = 0
}

// NextByte returns the next byte
func (i *NoAllocBytesIterator) NextByte() (b byte, err error) {
	if len(i.bs) < i.offset+1 {
		err = fmt.Errorf("astikit: slice length is %d, offset %d is invalid", len(i.bs), i.offset)
		return
	}
	b = i.bs[i.offset]
	i.offset++
	return
}

// NextBytes returns the n next bytes
func (i *NoAllocBytesIterator) NextBytes(n int) (bs []byte, err error) {
	if len(i.bs) < i.offset+n {
		err = fmt.Errorf("astikit: slice length is %d, offset %d is invalid", len(i.bs), i.offset+n)
		return
	}
	bs = make([]byte, n)
	copy(bs, i.bs[i.offset:i.offset+n])
	i.offset += n
	return
}

// NextBytesNoCopy returns the n next bytes
// Be careful with this function as it doesn't make a copy of returned data.
// bs will point to internal BytesIterator buffer.
// If you need to modify returned bytes or store it for some time, use NextBytes instead
func (i *NoAllocBytesIterator) NextBytesNoCopy(n int) (bs []byte, err error) {
	if len(i.bs) < i.offset+n {
		err = fmt.Errorf("astikit: slice length is %d, offset %d is invalid", len(i.bs), i.offset+n)
		return
	}
	bs = i.bs[i.offset : i.offset+n]
	i.offset += n
	return
}

// Seek seeks to the nth byte
func (i *NoAllocBytesIterator) Seek(n int) {
	i.offset = n
}

// Skip skips the n previous/next bytes
func (i *NoAllocBytesIterator) Skip(n int) {
	i.offset += n
}

// HasBytesLeft checks whether there are bytes left
func (i *NoAllocBytesIterator) HasBytesLeft() bool {
	return i.offset < len(i.bs)
}

// Offset returns the offset
func (i *NoAllocBytesIterator) Offset() int {
	return i.offset
}

// Dump dumps the rest of the slice
func (i *NoAllocBytesIterator) Dump() (bs []byte) {
	if !i.HasBytesLeft() {
		return
	}
	bs = make([]byte, len(i.bs)-i.offset)
	copy(bs, i.bs[i.offset:len(i.bs)])
	i.offset = len(i.bs)
	return
}

// Len returns the slice length
func (i *NoAllocBytesIterator) Len() int {
	return len(i.bs)
}
