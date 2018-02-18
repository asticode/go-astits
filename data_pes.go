package astits

import (
	"fmt"

	"github.com/pkg/errors"
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
func parsePESData(i []byte) (d *PESData, err error) {
	// Init
	d = &PESData{}

	// Parse header
	var offset, dataStart, dataEnd = 3, 0, 0
	if d.Header, dataStart, dataEnd, err = parsePESHeader(i, &offset); err != nil {
		err = errors.Wrap(err, "astits: parsing PES header failed")
		return
	}

	// Parse data
	d.Data = i[dataStart:dataEnd]
	return
}

// hasPESOptionalHeader checks whether the data has a PES optional header
func hasPESOptionalHeader(streamID uint8) bool {
	return streamID != StreamIDPaddingStream && streamID != StreamIDPrivateStream2
}

// parsePESData parses a PES header
func parsePESHeader(i []byte, offset *int) (h *PESHeader, dataStart, dataEnd int, err error) {
	// Init
	h = &PESHeader{}

	// Stream ID
	h.StreamID = uint8(i[*offset])
	*offset += 1

	// Length
	h.PacketLength = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	*offset += 2

	// Data end
	if h.PacketLength > 0 {
		dataEnd = *offset + int(h.PacketLength)
	} else {
		dataEnd = len(i)
	}

	// Check for incomplete data
	if dataEnd > len(i) {
		err = fmt.Errorf("astits: pes dataEnd (%d) > len(i) (%d)", dataEnd, len(i))
		return
	}

	// Optional header
	if hasPESOptionalHeader(h.StreamID) {
		h.OptionalHeader, dataStart = parsePESOptionalHeader(i, offset)
	} else {
		dataStart = *offset
	}
	return
}

// parsePESOptionalHeader parses a PES optional header
func parsePESOptionalHeader(i []byte, offset *int) (h *PESOptionalHeader, dataStart int) {
	// Init
	h = &PESOptionalHeader{}

	// Marker bits
	h.MarkerBits = uint8(i[*offset]) >> 6

	// Scrambling control
	h.ScramblingControl = uint8(i[*offset]) >> 4 & 0x3

	// Priority
	h.Priority = uint8(i[*offset])&0x8 > 0

	// Data alignment indicator
	h.DataAlignmentIndicator = uint8(i[*offset])&0x4 > 0

	// Copyrighted
	h.IsCopyrighted = uint(i[*offset])&0x2 > 0

	// Original or copy
	h.IsOriginal = uint8(i[*offset])&0x1 > 0
	*offset += 1

	// PTS DST indicator
	h.PTSDTSIndicator = uint8(i[*offset]) >> 6 & 0x3

	// Flags
	h.HasESCR = uint8(i[*offset])&0x20 > 0
	h.HasESRate = uint8(i[*offset])&0x10 > 0
	h.HasDSMTrickMode = uint8(i[*offset])&0x8 > 0
	h.HasAdditionalCopyInfo = uint8(i[*offset])&0x4 > 0
	h.HasCRC = uint8(i[*offset])&0x2 > 0
	h.HasExtension = uint8(i[*offset])&0x1 > 0
	*offset += 1

	// Header length
	h.HeaderLength = uint8(i[*offset])
	*offset += 1

	// Data start
	dataStart = *offset + int(h.HeaderLength)

	// PTS/DTS
	if h.PTSDTSIndicator == PTSDTSIndicatorOnlyPTS {
		h.PTS = parsePTSOrDTS(i[*offset:])
		*offset += 5
	} else if h.PTSDTSIndicator == PTSDTSIndicatorBothPresent {
		h.PTS = parsePTSOrDTS(i[*offset:])
		*offset += 5
		h.DTS = parsePTSOrDTS(i[*offset:])
		*offset += 5
	}

	// ESCR
	if h.HasESCR {
		h.ESCR = parseESCR(i[*offset:])
		*offset += 6
	}

	// ES rate
	if h.HasESRate {
		h.ESRate = uint32(i[*offset])&0x7f<<15 | uint32(i[*offset+1])<<7 | uint32(i[*offset+2])>>1
		*offset += 3
	}

	// Trick mode
	if h.HasDSMTrickMode {
		h.DSMTrickMode = parseDSMTrickMode(i[*offset])
		*offset += 1
	}

	// Additional copy info
	if h.HasAdditionalCopyInfo {
		h.AdditionalCopyInfo = i[*offset] & 0x7f
		*offset += 1
	}

	// CRC
	if h.HasCRC {
		h.CRC = uint16(i[*offset])>>8 | uint16(i[*offset+1])
		*offset += 2
	}

	// Extension
	if h.HasExtension {
		// Flags
		h.HasPrivateData = i[*offset]&0x80 > 0
		h.HasPackHeaderField = i[*offset]&0x40 > 0
		h.HasProgramPacketSequenceCounter = i[*offset]&0x20 > 0
		h.HasPSTDBuffer = i[*offset]&0x10 > 0
		h.HasExtension2 = i[*offset]&0x1 > 0
		*offset += 1

		// Private data
		if h.HasPrivateData {
			h.PrivateData = i[*offset : *offset+16]
			*offset += 16
		}

		// Pack field length
		if h.HasPackHeaderField {
			h.PackField = uint8(i[*offset])
			*offset += 1
		}

		// Program packet sequence counter
		if h.HasProgramPacketSequenceCounter {
			h.PacketSequenceCounter = uint8(i[*offset]) & 0x7f
			h.MPEG1OrMPEG2ID = uint8(i[*offset+1]) >> 6 & 0x1
			h.OriginalStuffingLength = uint8(i[*offset+1]) & 0x3f
			*offset += 2
		}

		// P-STD buffer
		if h.HasPSTDBuffer {
			h.PSTDBufferScale = i[*offset] >> 5 & 0x1
			h.PSTDBufferSize = uint16(i[*offset])&0x1f<<8 | uint16(i[*offset+1])
			*offset += 2
		}

		// Extension 2
		if h.HasExtension2 {
			// Length
			h.Extension2Length = uint8(i[*offset]) & 0x7f
			*offset += 2

			// Data
			h.Extension2Data = i[*offset : *offset+int(h.Extension2Length)]
			*offset += int(h.Extension2Length)
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
func parsePTSOrDTS(i []byte) *ClockReference {
	return newClockReference(int(uint64(i[0])>>1&0x7<<30|uint64(i[1])<<22|uint64(i[2])>>1&0x7f<<15|uint64(i[3])<<7|uint64(i[4])>>1&0x7f), 0)
}

// parseESCR parses an ESCR
func parseESCR(i []byte) *ClockReference {
	var escr = uint64(i[0])>>3&0x7<<39 | uint64(i[0])&0x3<<37 | uint64(i[1])<<29 | uint64(i[2])>>3<<24 | uint64(i[2])&0x3<<22 | uint64(i[3])<<14 | uint64(i[4])>>3<<9 | uint64(i[4])&0x3<<7 | uint64(i[5])>>1
	return newClockReference(int(escr>>9), int(escr&0x1ff))
}
