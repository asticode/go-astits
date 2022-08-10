package astits

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// Sync byte.
const syncByte = '\x47'

// ErrPacketStartSyncByte packet must start with a sync byte.
var ErrPacketStartSyncByte = errors.New("packet must start with a sync byte")

// Demuxer represents a demuxer.
// https://en.wikipedia.org/wiki/MPEG_transport_stream
// http://seidl.cs.vsb.cz/download/dvb/DVB_Poster.pdf
// http://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.13.01_40/en_300468v011301o.pdf
type Demuxer struct {
	ctx              context.Context
	dataBuffer       []*DemuxerData
	optPacketSize    int
	optPacketsParser PacketsParser
	packetBuffer     *packetBuffer
	packetPool       *packetPool
	programMap       *programMap
	r                io.Reader
}

// PacketsParser represents an object capable of parsing
// a set of packets containing a unique payload spanning
// over those packets. Use the skip returned argument
// to indicate whether the default process should
// still be executed on the set of packets.
type PacketsParser func(ps []*Packet) (ds []*DemuxerData, skip bool, err error)

// NewDemuxer creates a new transport stream based on a reader.
func NewDemuxer(ctx context.Context, r io.Reader, opts ...func(*Demuxer)) (d *Demuxer) {
	// Init
	d = &Demuxer{
		ctx:        ctx,
		programMap: newProgramMap(),
		r:          r,
	}
	d.packetPool = newPacketPool(d.optPacketsParser, d.programMap)

	// Apply options
	for _, opt := range opts {
		opt(d)
	}

	return
}

// DemuxerOptPacketSize returns the option to set the packet size.
func DemuxerOptPacketSize(packetSize int) func(*Demuxer) {
	return func(d *Demuxer) {
		d.optPacketSize = packetSize
	}
}

// DemuxerOptPacketsParser returns the option to set the packets parser.
func DemuxerOptPacketsParser(p PacketsParser) func(*Demuxer) {
	return func(d *Demuxer) {
		d.optPacketsParser = p
	}
}

// NextPacket retrieves the next packet.
func (dmx *Demuxer) NextPacket() (*Packet, error) {
	// Check ctx error
	// TODO Handle ctx error another way since if the read blocks,
	// everything blocks Maybe execute everything in a goroutine
	// and listen the ctx channel in the same for loop.
	var err error
	if err = dmx.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// Create packet buffer if not exists.
	if dmx.packetBuffer == nil {
		dmx.packetBuffer, err = newPacketBuffer(dmx.r, dmx.optPacketSize)
		if err != nil {
			return nil, fmt.Errorf("creating packet buffer failed: %w", err)
		}
	}

	// Fetch next packet from buffer.
	p, err := dmx.packetBuffer.next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("fetching next packet from buffer failed: %w", err)
	}

	return p, nil
}

// NextData retrieves the next data.
func (dmx *Demuxer) NextData() (*DemuxerData, error) {
	// Check data buffer.
	if len(dmx.dataBuffer) > 0 {
		d := dmx.dataBuffer[0]
		dmx.dataBuffer = dmx.dataBuffer[1:]
		return d, nil
	}

	// Loop through packets.
	var p *Packet
	var err error
	var ps []*Packet
	var ds []*DemuxerData
	for {
		// Get next packet.
		if p, err = dmx.NextPacket(); err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("fetching next packet failed: %w", err)
			}
			// If the end of the stream has been reached, we dump the packet pool.
			for {
				if ps = dmx.packetPool.dump(); len(ps) == 0 {
					break
				}

				var errParseData error
				if ds, errParseData = parseData(ps, dmx.optPacketsParser, dmx.programMap); errParseData != nil {
					// We need to silence this error as there may be some
					// incomplete data here  We still want to try to
					// parse all packets, in case final data is complete.
					continue
				}

				if d := dmx.updateData(ds); d != nil {
					return d, nil
				}
			}
			return nil, err
		}

		if ps = dmx.packetPool.add(p); len(ps) == 0 {
			continue
		}

		if ds, err = parseData(ps, dmx.optPacketsParser, dmx.programMap); err != nil {
			return nil, fmt.Errorf("building new data failed: %w", err)
		}

		if d := dmx.updateData(ds); d != nil {
			return d, nil
		}
	}
}

func (dmx *Demuxer) updateData(ds []*DemuxerData) (d *DemuxerData) {
	// Check whether there is data to be processed.
	if len(ds) > 0 {
		// Process data.
		d = ds[0]
		dmx.dataBuffer = append(dmx.dataBuffer, ds[1:]...)

		// Update program map.
		for _, v := range ds {
			if v.PAT != nil {
				for _, pgm := range v.PAT.Programs {
					// Program number 0 is reserved to NIT.
					if pgm.ProgramNumber > 0 {
						dmx.programMap.set(pgm.ProgramMapID, pgm.ProgramNumber)
					}
				}
			}
		}
	}
	return
}

// Rewind rewinds the demuxer reader.
func (dmx *Demuxer) Rewind() (n int64, err error) {
	dmx.dataBuffer = []*DemuxerData{}
	dmx.packetBuffer = nil
	dmx.packetPool = newPacketPool(dmx.optPacketsParser, dmx.programMap)
	if n, err = rewind(dmx.r); err != nil {
		err = fmt.Errorf("rewinding reader failed: %w", err)
		return
	}
	return
}
