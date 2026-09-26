package astits

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/asticode/go-astikit"
)

// packetBuffer represents a packet buffer
type packetBuffer struct {
	packetSize       int
	s                PacketSkipper
	r                io.Reader
	packetReadBuffer []byte
	aligned          bool
}

// syncWindow is the number of bytes scanned to find the first packet when the
// stream doesn't start with a sync byte: it must hold a sync byte followed by
// another one exactly one packet later, for the biggest supported packet size
const syncWindow = 2*204 + 1

// findSync returns the first offset in b holding a sync byte followed by another
// sync byte exactly one packet later, trying each packet size in order
func findSync(b []byte, sizes []int) (offset, size int, ok bool) {
	for off := range b {
		if b[off] != syncByte {
			continue
		}
		for _, s := range sizes {
			if off+s < len(b) && b[off+s] == syncByte {
				return off, s, true
			}
		}
	}
	return
}

// newPacketBuffer creates a new packet buffer
func newPacketBuffer(r io.Reader, packetSize int, s PacketSkipper) (pb *packetBuffer, err error) {
	// Init
	pb = &packetBuffer{
		packetSize: packetSize,
		s:          s,
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
// If the first byte of the reader is not a sync byte (e.g. when joining a live
// stream), the reader is synced to the first packet found in the first bytes
func autoDetectPacketSize(r io.Reader) (packetSize int, err error) {
	// Read first bytes
	const l = 193
	var b = make([]byte, l)
	read, shouldRewind, rerr := peek(r, b)
	if rerr != nil {
		err = fmt.Errorf("astits: reading first %d bytes failed: %w", l, rerr)
		return
	}

	// Stream doesn't start with a sync byte
	if b[0] != syncByte {
		return syncToFirstPacket(r, b[:read], shouldRewind)
	}

	// Look for sync bytes
	for idx, b := range b {
		if b == syncByte && idx >= MpegTsPacketSize {
			// Update packet size
			packetSize = idx

			if !shouldRewind {
				return
			}

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

// bufio.Reader can't be rewinded, which leads to packet loss on packet size autodetection
// but it has handy Peek() method
// so what we do here is peeking bytes for bufio.Reader and falling back to rewinding/syncing for all other readers
func peek(r io.Reader, b []byte) (n int, shouldRewind bool, err error) {
	if br, ok := r.(*bufio.Reader); ok {
		var bs []byte
		bs, err = br.Peek(len(b))
		if err != nil {
			return
		}
		n = copy(b, bs)
		return n, false, nil
	}

	n, err = r.Read(b)
	shouldRewind = true
	return
}

// syncToFirstPacket looks for the first packet in the first bytes of a stream
// that doesn't start with a sync byte, and positions the reader at a packet boundary
// b holds the bytes already read (or peeked) from the reader
func syncToFirstPacket(r io.Reader, b []byte, shouldRewind bool) (packetSize int, err error) {
	// Widen the window
	var n = len(b)
	if br, ok := r.(*bufio.Reader); ok {
		var bs []byte
		if bs, err = br.Peek(syncWindow); err != nil && err != io.EOF {
			err = fmt.Errorf("astits: peeking first %d bytes failed: %w", syncWindow, err)
			return
		}
		err = nil
		b = bs
		n = len(b)
	} else if n < syncWindow {
		b = append(b, make([]byte, syncWindow-n)...)
		var m int
		if m, err = io.ReadFull(r, b[n:]); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			err = fmt.Errorf("astits: reading first %d bytes failed: %w", syncWindow, err)
			return
		}
		err = nil
		n += m
		b = b[:n]
	}

	// Look for the first packet
	offset, packetSize, ok := findSync(b, []int{MpegTsPacketSize, 192, 204})
	if !ok {
		packetSize = 0
		err = ErrPacketMustStartWithASyncByte
		return
	}

	// Position the reader at the first packet
	if !shouldRewind {
		if _, err = r.(*bufio.Reader).Discard(offset); err != nil {
			err = fmt.Errorf("astits: discarding %d bytes failed: %w", offset, err)
		}
		return
	}
	if s, ok := r.(io.Seeker); ok {
		if _, err = s.Seek(int64(offset), io.SeekStart); err != nil {
			err = fmt.Errorf("astits: seeking to %d failed: %w", offset, err)
		}
		return
	}

	// Reader can't be rewinded: sync it to the next packet boundary
	var boundary = offset + (n-offset+packetSize-1)/packetSize*packetSize
	if _, err = io.CopyN(io.Discard, r, int64(boundary-n)); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = fmt.Errorf("astits: discarding %d bytes to sync reader failed: %w", boundary-n, err)
		}
	}
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
	if pb.packetReadBuffer == nil || len(pb.packetReadBuffer) != pb.packetSize {
		pb.packetReadBuffer = make([]byte, pb.packetSize)
	}

	// Loop to make sure we return a packet even if first packets are skipped
	for p == nil {
		if _, err = io.ReadFull(pb.r, pb.packetReadBuffer); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				err = ErrNoMorePackets
			} else {
				err = fmt.Errorf("astits: reading %d bytes failed: %w", pb.packetSize, err)
			}
			return
		}

		// Sync to the first packet if the stream doesn't start with a sync byte
		if !pb.aligned {
			pb.aligned = true
			if pb.packetReadBuffer[0] != syncByte {
				if err = pb.syncFirstPacket(); err != nil {
					return
				}
			}
		}

		// Parse packet
		if p, err = parsePacket(astikit.NewBytesIterator(pb.packetReadBuffer), pb.s); err != nil {
			if !errors.Is(err, errSkippedPacket) {
				err = fmt.Errorf("astits: building packet failed: %w", err)
				return
			}
		}
	}

	return
}

// syncFirstPacket looks for the first packet in the first two packet-size chunks
// and puts the bytes read past it back in front of the reader
func (pb *packetBuffer) syncFirstPacket() (err error) {
	// Read another chunk
	var b = make([]byte, 2*pb.packetSize)
	copy(b, pb.packetReadBuffer)
	var n int
	if n, err = io.ReadFull(pb.r, b[pb.packetSize:]); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			err = fmt.Errorf("astits: reading %d bytes failed: %w", pb.packetSize, err)
			return
		}
		err = nil
	}
	b = b[:pb.packetSize+n]

	// No packet found: parsing will fail with the proper error
	offset, _, ok := findSync(b, []int{pb.packetSize})
	if !ok {
		return
	}

	// Use the first packet and put the rest back in front of the reader
	copy(pb.packetReadBuffer, b[offset:offset+pb.packetSize])
	pb.r = io.MultiReader(bytes.NewReader(b[offset+pb.packetSize:]), pb.r)
	return
}
