package astits

import (
	"io"

	"github.com/pkg/errors"
)

// packetBuffer represents a packet buffer
type packetBuffer struct {
	b          []*Packet
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
			err = errors.Wrap(err, "astits: auto detecting packet size failed")
			return
		}
	}
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
			err = errors.Wrapf(err, "astits: reading %d bytes failed", pb.packetSize)
		}
		return
	}

	// Parse packet
	if p, err = parsePacket(b); err != nil {
		err = errors.Wrap(err, "astits: building packet failed")
		return
	}
	return
}
