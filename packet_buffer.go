package astits

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/icza/bitio"
)

// packetBuffer represents a packet buffer.
type packetBuffer struct {
	packetSize       int
	r                io.Reader
	packetReadBuffer []byte
}

// newPacketBuffer creates a new packet buffer.
func newPacketBuffer(r io.Reader, packetSize int) (pb *packetBuffer, err error) {
	// Init
	pb = &packetBuffer{
		packetSize: packetSize,
		r:          r,
	}

	// Packet size is not set.
	if pb.packetSize == 0 {
		// Auto detect packet size.
		if pb.packetSize, err = autoDetectPacketSize(r); err != nil {
			err = fmt.Errorf("auto detecting packet size failed: %w", err)
			return
		}
	}
	return
}

// ErrSingleSyncByte .
var ErrSingleSyncByte = errors.New("only one sync byte detected")

// autoDetectPacketSize updates the packet size based on the first bytes
// Minimum packet size is 188 and is bounded by 2 sync bytes
// Assumption is made that the first byte of the reader is a sync byte.
func autoDetectPacketSize(r io.Reader) (int, error) {
	// Read first bytes
	const l = 193
	b := make([]byte, l)
	shouldRewind, err := peek(r, b)
	if err != nil {
		return 0, fmt.Errorf("reading first %d bytes failed: %w", l, err)
	}

	// Packet must start with a sync byte.
	if b[0] != syncByte {
		return 0, ErrPacketStartSyncByte
	}

	var packetSize int
	// Look for sync bytes.
	for idx, b := range b {
		if b != syncByte || idx < MpegTsPacketSize {
			continue
		}

		// Update packet size.
		packetSize = idx

		if !shouldRewind {
			return packetSize, nil
		}

		// Rewind or sync reader.
		var n int64
		if n, err = rewind(r); err != nil {
			return 0, fmt.Errorf("rewinding failed: %w", err)
		} else if n == -1 {
			ls := packetSize - (l - packetSize)
			_, err := r.Read(make([]byte, ls))
			if err != nil {
				return 0, fmt.Errorf("reading %d bytes to sync reader failed: %w", ls, err)
			}
		}
		return packetSize, nil
	}
	return 0, fmt.Errorf("%w in first %d bytes", ErrSingleSyncByte, l)
}

// peek bufio.Reader can't be rewinded, which leads to packet
// loss on packet size autodetection but it has handy Peek()
// method so what we do here is peeking bytes for bufio.Reader
// and falling back to rewinding/syncing for all other readers.
func peek(r io.Reader, b []byte) (shouldRewind bool, err error) {
	if br, ok := r.(*bufio.Reader); ok {
		var bs []byte
		bs, err = br.Peek(len(b))
		if err != nil {
			return
		}
		copy(b, bs)
		return false, nil
	}

	_, err = r.Read(b)
	shouldRewind = true
	return
}

// rewind rewinds the reader if possible, otherwise n = -1 .
func rewind(r io.Reader) (n int64, err error) {
	if s, ok := r.(io.Seeker); ok {
		if n, err = s.Seek(0, 0); err != nil {
			err = fmt.Errorf("seeking to 0 failed: %w", err)
			return
		}
		return
	}
	n = -1
	return
}

// next fetches the next packet from the buffer.
func (pb *packetBuffer) next() (*Packet, error) {
	// Read
	if pb.packetReadBuffer == nil || len(pb.packetReadBuffer) != pb.packetSize {
		pb.packetReadBuffer = make([]byte, pb.packetSize)
	}

	_, err := io.ReadFull(pb.r, pb.packetReadBuffer)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("reading %d bytes failed: %w", pb.packetSize, err)
	}

	r := bitio.NewCountReader(bytes.NewReader(pb.packetReadBuffer))
	pktBufferLength := int64(len(pb.packetReadBuffer) * 8)

	// Parse packet.
	p, err := parsePacket(r, pktBufferLength)
	if err != nil {
		return nil, fmt.Errorf("building packet failed: %w", err)
	}

	return p, nil
}
