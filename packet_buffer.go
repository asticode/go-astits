package astits

import (
	"fmt"
	"io"

	"github.com/asticode/go-astikit"
)

// packetBuffer represents a packet buffer
type packetBuffer struct {
	packetSize int
	r          io.Reader
}

// newPacketBuffer creates a new packet buffer
func newPacketBuffer(r io.Reader, packetSize int) (pb *packetBuffer, err error) {
	// Init
	pb = &packetBuffer{
		packetSize: packetSize,
		r:          r,
	}

	// Packet size is not set
	if pb.packetSize == 0 {
		// Auto detect packet size
		if pb.packetSize, err = autoDetectPacketSize(r); err != nil {
			err = fmt.Errorf("astits: auto detecting packet size failed: %w", err)
			return
		}
	}
	return
}

// autoDetectPacketSize updates the packet size based on the first bytes
// Minimum packet size is 188 and is bounded by 2 sync bytes
// Assumption is made that the first byte of the reader is a sync byte
func autoDetectPacketSize(r io.Reader) (packetSize int, err error) {
	// Read first bytes
	const l = 193
	var b = make([]byte, l)
	if _, err = r.Read(b); err != nil {
		err = fmt.Errorf("astits: reading first %d bytes failed: %w", l, err)
		return
	}

	// Packet must start with a sync byte
	if b[0] != syncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	// Look for sync bytes
	for idx, b := range b {
		if b == syncByte && idx >= 188 {
			// Update packet size
			packetSize = idx

			// Rewind or sync reader
			var n int64
			if n, err = rewind(r); err != nil {
				err = fmt.Errorf("astits: rewinding failed: %w", err)
				return
			} else if n == -1 {
				var ls = packetSize - (l - packetSize)
				if _, err = r.Read(make([]byte, ls)); err != nil {
					err = fmt.Errorf("astits: reading %d bytes to sync reader failed: %w", ls, err)
					return
				}
			}
			return
		}
	}
	err = fmt.Errorf("astits: only one sync byte detected in first %d bytes", l)
	return
}

// rewind rewinds the reader if possible, otherwise n = -1
func rewind(r io.Reader) (n int64, err error) {
	if s, ok := r.(io.Seeker); ok {
		if n, err = s.Seek(0, 0); err != nil {
			err = fmt.Errorf("astits: seeking to 0 failed: %w", err)
			return
		}
		return
	}
	n = -1
	return
}

// next fetches the next packet from the buffer
func (pb *packetBuffer) next() (p *Packet, err error) {
	// Read
	var b = make([]byte, pb.packetSize)
	if _, err = io.ReadFull(pb.r, b); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = ErrNoMorePackets
		} else {
			err = fmt.Errorf("astits: reading %d bytes failed: %w", pb.packetSize, err)
		}
		return
	}

	// Parse packet
	if p, err = parsePacket(astikit.NewBytesIterator(b)); err != nil {
		err = fmt.Errorf("astits: building packet failed: %w", err)
		return
	}
	return
}
