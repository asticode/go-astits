package astits

import (
	"fmt"

	"github.com/asticode/go-astikit"
)

// P-STD buffer scales
const (
	PSTDBufferScale128Bytes  = 0
	PSTDBufferScale1024Bytes = 1
)

// PTS DTS indicator
const (
	PTSDTSIndicatorBothPresent = 3
	PTSDTSIndicatorIsForbidden = 1
	PTSDTSIndicatorNoPTSOrDTS  = 0
	PTSDTSIndicatorOnlyPTS     = 2
)

// Stream IDs
const (
	StreamIDPrivateStream1 = 189
	StreamIDPaddingStream  = 190
	StreamIDPrivateStream2 = 191
)

// Trick mode controls
const (
	TrickModeControlFastForward = 0
	TrickModeControlFastReverse = 3
	TrickModeControlFreezeFrame = 2
	TrickModeControlSlowMotion  = 1
	TrickModeControlSlowReverse = 4
)

// PESData represents a PES data
// https://en.wikipedia.org/wiki/Packetized_elementary_stream
// http://dvd.sourceforge.net/dvdinfo/pes-hdr.html
// http://happy.emu.id.au/lab/tut/dttb/dtbtut4b.htm
type PESData struct {
	Data   []byte
	Header *PESHeader
}

// PESHeader represents a packet PES header
type PESHeader struct {
	OptionalHeader *PESOptionalHeader
	PacketLength   uint16 // Specifies the number of bytes remaining in the packet after this field. Can be zero. If the PES packet length is set to zero, the PES packet can be of any length. A value of zero for the PES packet length can be used only when the PES packet payload is a video elementary stream.
	StreamID       uint8  // Examples: Audio streams (0xC0-0xDF), Video streams (0xE0-0xEF)
}

// PESOptionalHeader represents a PES optional header
type PESOptionalHeader struct {
	AdditionalCopyInfo              uint8
	CRC                             uint16
	DataAlignmentIndicator          bool // True indicates that the PES packet header is immediately followed by the video start code or audio syncword
	DSMTrickMode                    *DSMTrickMode
	DTS                             *ClockReference
	ESCR                            *ClockReference
	ESRate                          uint32
	Extension2Data                  []byte
	Extension2Length                uint8
	HasAdditionalCopyInfo           bool
	HasCRC                          bool
	HasDSMTrickMode                 bool
	HasESCR                         bool
	HasESRate                       bool
	HasExtension                    bool
	HasExtension2                   bool
	HasOptionalFields               bool
	HasPackHeaderField              bool
	HasPrivateData                  bool
	HasProgramPacketSequenceCounter bool
	HasPSTDBuffer                   bool
	HeaderLength                    uint8
	IsCopyrighted                   bool
	IsOriginal                      bool
	MarkerBits                      uint8
	MPEG1OrMPEG2ID                  uint8
	OriginalStuffingLength          uint8
	PacketSequenceCounter           uint8
	PackField                       uint8
	Priority                        bool
	PrivateData                     []byte
	PSTDBufferScale                 uint8
	PSTDBufferSize                  uint16
	PTS                             *ClockReference
	PTSDTSIndicator                 uint8
	ScramblingControl               uint8
}

// DSMTrickMode represents a DSM trick mode
// https://books.google.fr/books?id=vwUrAwAAQBAJ&pg=PT501&lpg=PT501&dq=dsm+trick+mode+control&source=bl&ots=fI-9IHXMRL&sig=PWnhxrsoMWNQcl1rMCPmJGNO9Ds&hl=fr&sa=X&ved=0ahUKEwjogafD8bjXAhVQ3KQKHeHKD5oQ6AEINDAB#v=onepage&q=dsm%20trick%20mode%20control&f=false
type DSMTrickMode struct {
	FieldID             uint8
	FrequencyTruncation uint8
	IntraSliceRefresh   uint8
	RepeatControl       uint8
	TrickModeControl    uint8
}

// parsePESData parses a PES data
func parsePESData(i *astikit.BytesIterator) (d *PESData, err error) {
	// Create data
	d = &PESData{}

	// Skip first 3 bytes that are there to identify the PES payload
	i.Seek(3)

	// Parse header
	var dataStart, dataEnd int
	if d.Header, dataStart, dataEnd, err = parsePESHeader(i); err != nil {
		err = fmt.Errorf("astits: parsing PES header failed: %w", err)
		return
	}

	// Seek to data
	i.Seek(dataStart)

	// Extract data
	if d.Data, err = i.NextBytes(dataEnd - dataStart); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// hasPESOptionalHeader checks whether the data has a PES optional header
func hasPESOptionalHeader(streamID uint8) bool {
	return streamID != StreamIDPaddingStream && streamID != StreamIDPrivateStream2
}

// parsePESData parses a PES header
func parsePESHeader(i *astikit.BytesIterator) (h *PESHeader, dataStart, dataEnd int, err error) {
	// Create header
	h = &PESHeader{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Stream ID
	h.StreamID = uint8(b)

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Length
	h.PacketLength = uint16(bs[0])<<8 | uint16(bs[1])

	// Update data end
	if h.PacketLength > 0 {
		dataEnd = i.Offset() + int(h.PacketLength)
	} else {
		dataEnd = i.Len()
	}

	// Optional header
	if hasPESOptionalHeader(h.StreamID) {
		if h.OptionalHeader, dataStart, err = parsePESOptionalHeader(i); err != nil {
			err = fmt.Errorf("astits: parsing PES optional header failed: %w", err)
			return
		}
	} else {
		dataStart = i.Offset()
	}
	return
}

// parsePESOptionalHeader parses a PES optional header
func parsePESOptionalHeader(i *astikit.BytesIterator) (h *PESOptionalHeader, dataStart int, err error) {
	// Create header
	h = &PESOptionalHeader{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Marker bits
	h.MarkerBits = uint8(b) >> 6

	// Scrambling control
	h.ScramblingControl = uint8(b) >> 4 & 0x3

	// Priority
	h.Priority = uint8(b)&0x8 > 0

	// Data alignment indicator
	h.DataAlignmentIndicator = uint8(b)&0x4 > 0

	// Copyrighted
	h.IsCopyrighted = uint(b)&0x2 > 0

	// Original or copy
	h.IsOriginal = uint8(b)&0x1 > 0

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// PTS DST indicator
	h.PTSDTSIndicator = uint8(b) >> 6 & 0x3

	// Flags
	h.HasESCR = uint8(b)&0x20 > 0
	h.HasESRate = uint8(b)&0x10 > 0
	h.HasDSMTrickMode = uint8(b)&0x8 > 0
	h.HasAdditionalCopyInfo = uint8(b)&0x4 > 0
	h.HasCRC = uint8(b)&0x2 > 0
	h.HasExtension = uint8(b)&0x1 > 0

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Header length
	h.HeaderLength = uint8(b)

	// Update data start
	dataStart = i.Offset() + int(h.HeaderLength)

	// PTS/DTS
	if h.PTSDTSIndicator == PTSDTSIndicatorOnlyPTS {
		if h.PTS, err = parsePTSOrDTS(i); err != nil {
			err = fmt.Errorf("astits: parsing PTS failed: %w", err)
			return
		}
	} else if h.PTSDTSIndicator == PTSDTSIndicatorBothPresent {
		if h.PTS, err = parsePTSOrDTS(i); err != nil {
			err = fmt.Errorf("astits: parsing PTS failed: %w", err)
			return
		}
		if h.DTS, err = parsePTSOrDTS(i); err != nil {
			err = fmt.Errorf("astits: parsing PTS failed: %w", err)
			return
		}
	}

	// ESCR
	if h.HasESCR {
		if h.ESCR, err = parseESCR(i); err != nil {
			err = fmt.Errorf("astits: parsing ESCR failed: %w", err)
			return
		}
	}

	// ES rate
	if h.HasESRate {
		var bs []byte
		if bs, err = i.NextBytes(3); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
		h.ESRate = uint32(bs[0])&0x7f<<15 | uint32(bs[1])<<7 | uint32(bs[2])>>1
	}

	// Trick mode
	if h.HasDSMTrickMode {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		h.DSMTrickMode = parseDSMTrickMode(b)
	}

	// Additional copy info
	if h.HasAdditionalCopyInfo {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		h.AdditionalCopyInfo = b & 0x7f
	}

	// CRC
	if h.HasCRC {
		var bs []byte
		if bs, err = i.NextBytes(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
		h.CRC = uint16(bs[0])>>8 | uint16(bs[1])
	}

	// Extension
	if h.HasExtension {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Flags
		h.HasPrivateData = b&0x80 > 0
		h.HasPackHeaderField = b&0x40 > 0
		h.HasProgramPacketSequenceCounter = b&0x20 > 0
		h.HasPSTDBuffer = b&0x10 > 0
		h.HasExtension2 = b&0x1 > 0

		// Private data
		if h.HasPrivateData {
			if h.PrivateData, err = i.NextBytes(16); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
		}

		// Pack field length
		if h.HasPackHeaderField {
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}
			h.PackField = uint8(b)
		}

		// Program packet sequence counter
		if h.HasProgramPacketSequenceCounter {
			var bs []byte
			if bs, err = i.NextBytes(2); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
			h.PacketSequenceCounter = uint8(bs[0]) & 0x7f
			h.MPEG1OrMPEG2ID = uint8(bs[1]) >> 6 & 0x1
			h.OriginalStuffingLength = uint8(bs[1]) & 0x3f
		}

		// P-STD buffer
		if h.HasPSTDBuffer {
			var bs []byte
			if bs, err = i.NextBytes(2); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
			h.PSTDBufferScale = bs[0] >> 5 & 0x1
			h.PSTDBufferSize = uint16(bs[0])&0x1f<<8 | uint16(bs[1])
		}

		// Extension 2
		if h.HasExtension2 {
			// Length
			var bs []byte
			if bs, err = i.NextBytes(2); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
			h.Extension2Length = uint8(bs[0]) & 0x7f

			// Data
			if h.Extension2Data, err = i.NextBytes(int(h.Extension2Length)); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
		}
	}
	return
}

// parseDSMTrickMode parses a DSM trick mode
func parseDSMTrickMode(i byte) (m *DSMTrickMode) {
	m = &DSMTrickMode{}
	m.TrickModeControl = i >> 5
	if m.TrickModeControl == TrickModeControlFastForward || m.TrickModeControl == TrickModeControlFastReverse {
		m.FieldID = i >> 3 & 0x3
		m.IntraSliceRefresh = i >> 2 & 0x1
		m.FrequencyTruncation = i & 0x3
	} else if m.TrickModeControl == TrickModeControlFreezeFrame {
		m.FieldID = i >> 3 & 0x3
	} else if m.TrickModeControl == TrickModeControlSlowMotion || m.TrickModeControl == TrickModeControlSlowReverse {
		m.RepeatControl = i & 0x1f
	}
	return
}

// parsePTSOrDTS parses a PTS or a DTS
func parsePTSOrDTS(i BytesIterator) (cr *ClockReference, err error) {
	var bs []byte
	if bs, err = i.NextBytes(5); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	cr = newClockReference(int64(uint64(bs[0])>>1&0x7<<30|uint64(bs[1])<<22|uint64(bs[2])>>1&0x7f<<15|uint64(bs[3])<<7|uint64(bs[4])>>1&0x7f), 0)
	return
}

// parseESCR parses an ESCR
func parseESCR(i BytesIterator) (cr *ClockReference, err error) {
	var bs []byte
	if bs, err = i.NextBytes(6); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	escr := uint64(bs[0])>>3&0x7<<39 | uint64(bs[0])&0x3<<37 | uint64(bs[1])<<29 | uint64(bs[2])>>3<<24 | uint64(bs[2])&0x3<<22 | uint64(bs[3])<<14 | uint64(bs[4])>>3<<9 | uint64(bs[4])&0x3<<7 | uint64(bs[5])>>1
	cr = newClockReference(int64(escr>>9), int64(escr&0x1ff))
	return
}
