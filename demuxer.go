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
	ctx           context.Context
	dataBuffer    []*Data
	packetBuffer  *packetBuffer
	PacketSize    int
	PacketsParser PacketsParser
	programMap    programMap
	r             io.Reader
}

// PacketsParser represents an object capable of parsing a set of packets containing a unique payload spanning over those packets
// Use the skip returned argument to indicate whether the default process should still be executed on the set of packets
type PacketsParser func(ps []*Packet) (ds []*Data, skip bool, err error)

// New creates a new transport stream based on a reader
func New(ctx context.Context, r io.Reader) *Demuxer {
	return &Demuxer{
		ctx:          ctx,
		packetBuffer: newPacketBuffer(),
		programMap:   newProgramMap(),
		r:            r,
	}
}

// autoDetectPacketSize updates the packet size based on the first bytes
// Minimum packet size is 188 and is bounded by 2 sync bytes
// Assumption is made that the first byte of the reader is a sync byte
func (dmx *Demuxer) autoDetectPacketSize() (err error) {
	// Read first bytes
	const l = 193
	var b = make([]byte, l)
	if _, err = dmx.r.Read(b); err != nil {
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
			dmx.PacketSize = idx

			// Sync reader
			var ls = dmx.PacketSize - (l - dmx.PacketSize)
			if _, err = dmx.r.Read(make([]byte, ls)); err != nil {
				err = errors.Wrapf(err, "astits: reading %d bytes to sync reader failed", ls)
				return
			}
			return
		}
	}
	err = fmt.Errorf("astits: only one sync byte detected in first %d bytes", l)
	return
}

// NextPacket retrieves the next packet
func (dmx *Demuxer) NextPacket() (p *Packet, err error) {
	// Check ctx error
	if err = dmx.ctx.Err(); err != nil {
		return
	}

	// Auto detect packet size
	if dmx.PacketSize == 0 {
		// Auto detect packet size
		if err = dmx.autoDetectPacketSize(); err != nil {
			err = errors.Wrap(err, "astits: auto detecting packet size failed")
			return
		}

		// Rewind
		if _, err = dmx.Rewind(); err != nil {
			err = errors.Wrap(err, "astits: rewinding failed")
			return
		}
	}

	// Read
	var b = make([]byte, dmx.PacketSize)
	if _, err = io.ReadFull(dmx.r, b); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = ErrNoMorePackets
		} else {
			err = errors.Wrapf(err, "astits: reading %d bytes failed", dmx.PacketSize)
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
			// If no more packets, we still need to dump the buffer
			if ps = dmx.packetBuffer.dump(); err != ErrNoMorePackets || len(ps) == 0 {
				if err == ErrNoMorePackets {
					return
				}
				err = errors.Wrap(err, "astits: fetching next packet failed")
				return
			}
		} else {
			// Add packet to the buffer
			if ps = dmx.packetBuffer.add(p); len(ps) == 0 {
				continue
			}
		}

		// Parse data
		if ds, err = parseData(ps, dmx.PacketsParser, dmx.programMap); err != nil {
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
	if s, ok := dmx.r.(io.Seeker); ok {
		if n, err = s.Seek(0, 0); err != nil {
			err = errors.Wrap(err, "astits: seeking to 0 failed")
			return
		}
	}
	return
}
