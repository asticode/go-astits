package astits

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Sync byte
const syncByte = '\x47'

// Errors
var (
	ErrNoMorePackets                = errors.New("astits: no more packets")
	ErrPacketMustStartWithASyncByte = errors.New("astits: packet must start with a sync byte")
)

// Demuxer represents a demuxer
// https://en.wikipedia.org/wiki/MPEG_transport_stream
// http://seidl.cs.vsb.cz/download/dvb/DVB_Poster.pdf
// http://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.13.01_40/en_300468v011301o.pdf
type Demuxer struct {
	ctx              context.Context
	dataBuffer       []*Data
	optPacketSize    int
	optPacketsParser PacketsParser
	packetBuffer     *packetBuffer
	packetPool       *packetPool
	programMap       programMap
	r                io.Reader
}

// PacketsParser represents an object capable of parsing a set of packets containing a unique payload spanning over those packets
// Use the skip returned argument to indicate whether the default process should still be executed on the set of packets
type PacketsParser func(ps []*Packet) (ds []*Data, skip bool, err error)

// New creates a new transport stream based on a reader
func New(ctx context.Context, r io.Reader, opts ...func(*Demuxer)) (d *Demuxer) {
	// Init
	d = &Demuxer{
		ctx:        ctx,
		packetPool: newPacketPool(),
		programMap: newProgramMap(),
		r:          r,
	}

	// Apply options
	for _, opt := range opts {
		opt(d)
	}
	return
}

// OptPacketSize returns the option to set the packet size
func OptPacketSize(packetSize int) func(*Demuxer) {
	return func(d *Demuxer) {
		d.optPacketSize = packetSize
	}
}

// OptPacketsParser returns the option to set the packets parser
func OptPacketsParser(p PacketsParser) func(*Demuxer) {
	return func(d *Demuxer) {
		d.optPacketsParser = p
	}
}

// NextPacket retrieves the next packet
func (dmx *Demuxer) NextPacket() (p *Packet, err error) {
	// Check ctx error
	if err = dmx.ctx.Err(); err != nil {
		return
	}

	// Create packet buffer if not exists
	if dmx.packetBuffer == nil {
		if dmx.packetBuffer, err = newPacketBuffer(dmx.r, dmx.optPacketSize); err != nil {
			err = errors.Wrap(err, "astits: creating packet buffer failed")
			return
		}
	}

	// Fetch next packet from buffer
	if p, err = dmx.packetBuffer.next(); err != nil {
		if err != ErrNoMorePackets {
			err = errors.Wrap(err, "astits: fetching next packet from buffer failed")
		}
		return
	}
	return
}

// NextData retrieves the next data
func (dmx *Demuxer) NextData() (d *Data, err error) {
	// Check data buffer
	if len(dmx.dataBuffer) > 0 {
		d = dmx.dataBuffer[0]
		dmx.dataBuffer = dmx.dataBuffer[1:]
		return
	}

	// Loop through packets
	var p *Packet
	var ps []*Packet
	var ds []*Data
	for {
		// Get next packet
		if p, err = dmx.NextPacket(); err != nil {
			// If no more packets, we still need to dump the pool
			if ps = dmx.packetPool.dump(); err != ErrNoMorePackets || len(ps) == 0 {
				if err == ErrNoMorePackets {
					return
				}
				err = errors.Wrap(err, "astits: fetching next packet failed")
				return
			}
		} else {
			// Add packet to the pool
			if ps = dmx.packetPool.add(p); len(ps) == 0 {
				continue
			}
		}

		// Parse data
		if ds, err = parseData(ps, dmx.optPacketsParser, dmx.programMap); err != nil {
			err = errors.Wrap(err, "astits: building new data failed")
			return
		}

		// Check whether there is data to be processed
		if len(ds) > 0 {
			// Process data
			d = ds[0]
			dmx.dataBuffer = append(dmx.dataBuffer, ds[1:]...)

			// Update program map
			for _, v := range ds {
				if v.PAT != nil {
					for _, pgm := range v.PAT.Programs {
						// Program number 0 is reserved to NIT
						if pgm.ProgramNumber > 0 {
							dmx.programMap.set(pgm.ProgramMapID, pgm.ProgramNumber)
						}
					}
				}
			}
			return
		}
	}
}

// Rewind rewinds the demuxer reader
func (dmx *Demuxer) Rewind() (n int64, err error) {
	dmx.dataBuffer = []*Data{}
	dmx.packetBuffer = nil
	dmx.packetPool = newPacketPool()
	if n, err = rewind(dmx.r); err != nil {
		err = errors.Wrap(err, "astits: rewinding reader failed")
		return
	}
	return
}

// rewind rewinds the reader if possible, otherwise n = -1
func rewind(r io.Reader) (n int64, err error) {
	if s, ok := r.(io.Seeker); ok {
		if n, err = s.Seek(0, 0); err != nil {
			err = errors.Wrap(err, "astits: seeking to 0 failed")
			return
		}
		return
	}
	n = -1
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
		err = errors.Wrapf(err, "astits: reading first %d bytes failed", l)
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
				err = errors.Wrap(err, "astits: rewinding failed")
				return
			} else if n == -1 {
				var ls = packetSize - (l - packetSize)
				if _, err = r.Read(make([]byte, ls)); err != nil {
					err = errors.Wrapf(err, "astits: reading %d bytes to sync reader failed", ls)
					return
				}
			}
			return
		}
	}
	err = fmt.Errorf("astits: only one sync byte detected in first %d bytes", l)
	return
}
