package astits

import (
	"bytes"
	"fmt"

	"github.com/icza/bitio"
)

// P-STD buffer scales.
const (
	PSTDBufferScale128Bytes  = 0
	PSTDBufferScale1024Bytes = 1
)

// PTS DTS indicator.
const (
	PTSDTSIndicatorBothPresent = 3
	PTSDTSIndicatorIsForbidden = 1
	PTSDTSIndicatorNoPTSOrDTS  = 0
	PTSDTSIndicatorOnlyPTS     = 2
)

// Stream IDs.
const (
	StreamIDPrivateStream1 = 189
	StreamIDPaddingStream  = 190
	StreamIDPrivateStream2 = 191
)

// Trick mode controls.
const (
	TrickModeControlFastForward = 0
	TrickModeControlFastReverse = 3
	TrickModeControlFreezeFrame = 2
	TrickModeControlSlowMotion  = 1
	TrickModeControlSlowReverse = 4
)

const (
	pesHeaderLength    = 6
	ptsOrDTSByteLength = 5
	escrLength         = 6
	dsmTrickModeLength = 1
)

// PESData represents a PES data.
// https://en.wikipedia.org/wiki/Packetized_elementary_stream
// http://dvd.sourceforge.net/dvdinfo/pes-hdr.html
// http://happy.emu.id.au/lab/tut/dttb/dtbtut4b.htm
type PESData struct {
	Data   []byte
	Header *PESHeader
}

// PESHeader represents a packet PES header.
type PESHeader struct {
	OptionalHeader *PESOptionalHeader

	// Specifies the number of bytes remaining in the packet
	// after this field. Can be zero. If the PES packet length
	// is set to zero, the PES packet can be of any length.
	// A value of zero for the PES packet length can be used
	// only when the PES packet payload is a video elementary stream.
	PacketLength uint16

	// Examples: Audio streams (0xC0-0xDF), Video streams (0xE0-0xEF)
	StreamID uint8
}

// PESOptionalHeader represents a PES optional header.
type PESOptionalHeader struct {
	MarkerBits             uint8 // 2 bits.
	ScramblingControl      uint8 // 2 bits.
	Priority               bool
	DataAlignmentIndicator bool
	IsCopyrighted          bool
	IsOriginal             bool

	PTSDTSIndicator       uint8 // 2 bits.
	HasESCR               bool
	HasESRate             bool
	HasDSMTrickMode       bool
	HasAdditionalCopyInfo bool
	HasCRC                bool
	HasExtension          bool

	HeaderLength uint8

	PTS                *ClockReference
	DTS                *ClockReference
	ESCR               *ClockReference
	ESRate             uint32 // 22 bits.
	DSMTrickMode       *DSMTrickMode
	AdditionalCopyInfo uint8 // 7 bits.
	CRC                uint16

	HasPrivateData                  bool
	HasPackHeaderField              bool
	HasProgramPacketSequenceCounter bool
	HasPSTDBuffer                   bool
	HasExtension2                   bool

	PrivateData []byte // 16 bytes.
	PackField   uint8

	PacketSequenceCounter  uint8 // 7 bits.
	MPEG1OrMPEG2ID         bool
	OriginalStuffingLength uint8 // 5 bits?

	PSTDBufferScale bool
	PSTDBufferSize  uint16 // 13 bits.

	Extension2Length uint8 // 7 bits.
	Extension2Data   []byte
}

// DSMTrickMode represents a DSM trick mode.
// https://patents.google.com/patent/US8213779B2/en
type DSMTrickMode struct {
	TrickModeControl    uint8 // 3 Bits.
	FieldID             uint8 // 2 Bits.
	IntraSliceRefresh   bool
	FrequencyTruncation uint8 // 2 Bits.
	RepeatControl       uint8 // 5 Bits.
}

// IsVideoStream .
func (h *PESHeader) IsVideoStream() bool {
	return h.StreamID == 0xe0 ||
		h.StreamID == 0xfd
}

// parsePESData parses a PES data.
func parsePESData(r *bitio.CountReader, payloadLength int64) (*PESData, error) {
	d := &PESData{}

	// Skip first 3 bytes that are there to identify the PES payload
	skip := make([]byte, 3)
	TryReadFull(r, skip)

	header, dataStart, dataEnd, err := parsePESHeader(r, payloadLength)
	if err != nil {
		return nil, fmt.Errorf("parsing PES header failed: %w", err)
	}
	d.Header = header

	if dataStart > r.BitsCount {
		skip := make([]byte, (dataStart-r.BitsCount)/8)
		TryReadFull(r, skip)
	}

	d.Data = make([]byte, (dataEnd-dataStart)/8)
	TryReadFull(r, d.Data)

	return d, r.TryError
}

// hasPESOptionalHeader checks whether the data has a PES optional header.
func hasPESOptionalHeader(streamID uint8) bool {
	return streamID != StreamIDPaddingStream && streamID != StreamIDPrivateStream2
}

// parsePESHeader parses a PES header.
func parsePESHeader(r *bitio.CountReader, payloadLength int64) (h *PESHeader, dataStart, dataEnd int64, err error) {
	h = &PESHeader{}

	h.StreamID = r.TryReadByte()

	h.PacketLength = uint16(r.TryReadBits(16))

	// Update data end
	if h.PacketLength > 0 {
		dataEnd = r.BitsCount + int64(h.PacketLength*8)
	} else {
		dataEnd = payloadLength
	}

	if hasPESOptionalHeader(h.StreamID) {
		h.OptionalHeader, dataStart, err = parsePESOptionalHeader(r)
		if err != nil {
			err = fmt.Errorf("parsing PES optional header failed: %w", err)
			return
		}
	} else {
		dataStart = r.BitsCount
	}

	return h, dataStart, dataEnd, r.TryError
}

// parsePESOptionalHeader parses a PES optional header.
func parsePESOptionalHeader(r *bitio.CountReader) (*PESOptionalHeader, int64, error) { //nolint:funlen
	// Create header
	h := &PESOptionalHeader{}

	h.MarkerBits = uint8(r.TryReadBits(2))
	h.ScramblingControl = uint8(r.TryReadBits(2))
	h.Priority = r.TryReadBool()
	h.DataAlignmentIndicator = r.TryReadBool()
	h.IsCopyrighted = r.TryReadBool()
	h.IsOriginal = r.TryReadBool()

	h.PTSDTSIndicator = uint8(r.TryReadBits(2))
	h.HasESCR = r.TryReadBool()
	h.HasESRate = r.TryReadBool()
	h.HasDSMTrickMode = r.TryReadBool()
	h.HasAdditionalCopyInfo = r.TryReadBool()
	h.HasCRC = r.TryReadBool()
	h.HasExtension = r.TryReadBool()

	h.HeaderLength = r.TryReadByte()

	// Update data start
	dataStart := r.BitsCount + int64(h.HeaderLength)*8
	var err error

	// PTS/DTS
	if h.PTSDTSIndicator == PTSDTSIndicatorOnlyPTS {
		_ = r.TryReadBits(4) // Reserved.
		if h.PTS, err = parsePTSOrDTS(r); err != nil {
			return nil, 0, fmt.Errorf("parsing PTS failed: %w", err)
		}
	} else if h.PTSDTSIndicator == PTSDTSIndicatorBothPresent {
		_ = r.TryReadBits(4) // Reserved.
		if h.PTS, err = parsePTSOrDTS(r); err != nil {
			return nil, 0, fmt.Errorf("parsing PTS failed: %w", err)
		}
		_ = r.TryReadBits(4) // Reserved.
		if h.DTS, err = parsePTSOrDTS(r); err != nil {
			return nil, 0, fmt.Errorf("parsing PTS failed: %w", err)
		}
	}

	if h.HasESCR {
		if h.ESCR, err = parseESCR(r); err != nil {
			return nil, 0, fmt.Errorf("parsing ESCR failed: %w", err)
		}
	}

	if h.HasESRate {
		_ = r.TryReadBool() // Reserved.
		h.ESRate = uint32(r.TryReadBits(22))
		_ = r.TryReadBool() // Reserved.
	}

	if h.HasDSMTrickMode {
		h.DSMTrickMode, err = parseDSMTrickMode(r)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing DSM trick mode failed: %w", err)
		}
	}

	if h.HasAdditionalCopyInfo {
		_ = r.TryReadBool() // Reserved.
		h.AdditionalCopyInfo = uint8(r.TryReadBits(7))
	}

	if h.HasCRC {
		h.CRC = uint16(r.TryReadBits(16))
	}

	if !h.HasExtension {
		return h, dataStart, nil
	}

	h.HasPrivateData = r.TryReadBool()
	h.HasPackHeaderField = r.TryReadBool()
	h.HasProgramPacketSequenceCounter = r.TryReadBool()
	h.HasPSTDBuffer = r.TryReadBool()
	_ = r.TryReadBits(3) // Reserved.
	h.HasExtension2 = r.TryReadBool()

	if h.HasPrivateData {
		h.PrivateData = make([]byte, 16)
		TryReadFull(r, h.PrivateData)
	}

	if h.HasPackHeaderField {
		// TODO it's only a length of pack_header,
		// should read it all. now it's wrong.
		h.PackField = r.TryReadByte()
	}

	if h.HasProgramPacketSequenceCounter {
		_ = r.TryReadBool() // Reserved.
		h.PacketSequenceCounter = uint8(r.TryReadBits(7))

		h.MPEG1OrMPEG2ID = r.TryReadBool()
		h.OriginalStuffingLength = uint8(r.TryReadBits(5)) // 5 bits?
		_ = r.TryReadBits(2)                               // Reserved.
	}

	if h.HasPSTDBuffer {
		_ = r.TryReadBits(2) // Reserved.
		h.PSTDBufferScale = r.TryReadBool()
		h.PSTDBufferSize = uint16(r.TryReadBits(13))
	}

	if h.HasExtension2 {
		_ = r.TryReadBool() // Reserved.
		h.Extension2Length = uint8(r.TryReadBits(7))

		h.Extension2Data = make([]byte, h.Extension2Length)
		TryReadFull(r, h.Extension2Data)
	}
	return h, dataStart, r.TryError
}

// parseDSMTrickMode parses a DSM trick mode.
func parseDSMTrickMode(r *bitio.CountReader) (*DSMTrickMode, error) {
	m := &DSMTrickMode{}
	m.TrickModeControl = uint8(r.TryReadBits(3))

	switch m.TrickModeControl {
	case TrickModeControlFastForward, TrickModeControlFastReverse:
		m.FieldID = uint8(r.TryReadBits(2))
		m.IntraSliceRefresh = r.TryReadBool()
		m.FrequencyTruncation = uint8(r.TryReadBits(2))

	case TrickModeControlFreezeFrame:
		m.FieldID = uint8(r.TryReadBits(2))
		_ = r.TryReadBits(3)

	case TrickModeControlSlowMotion, TrickModeControlSlowReverse:
		m.RepeatControl = uint8(r.TryReadBits(5))

	default:
		_ = uint8(r.TryReadBits(5))
	}
	return m, r.TryError
}

// readPTSOrDTS reads a PTS or a DTS.
func readPTSOrDTS(r *bitio.CountReader) (int64, error) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)

	w.TryWriteBits(r.TryReadBits(3), 3)   // bits 32-30.
	_ = r.TryReadBool()                   // Reserved.
	w.TryWriteBits(r.TryReadBits(15), 15) // bits 27-15.
	_ = r.TryReadBool()                   // Reserved.
	w.TryWriteBits(r.TryReadBits(15), 15) // bits 27-15.
	_ = r.TryReadBool()                   // Reserved.

	if r.TryError != nil {
		return 0, fmt.Errorf("read: %w", r.TryError)
	}
	if w.TryError != nil {
		return 0, fmt.Errorf("write: %w", w.TryError)
	}

	if _, err := w.Align(); err != nil {
		return 0, fmt.Errorf("align: %w", err)
	}

	base, err := bitio.NewReader(buf).ReadBits(33)
	if err != nil {
		return 0, fmt.Errorf("base: %w", w.TryError)
	}

	return int64(base), nil
}

// parsePTSOrDTS parses a PTS or a DTS.
func parsePTSOrDTS(r *bitio.CountReader) (*ClockReference, error) {
	base, err := readPTSOrDTS(r)
	return newClockReference(base, 0), err
}

// parseESCR parses an ESCR.
func parseESCR(r *bitio.CountReader) (*ClockReference, error) {
	r.TryReadBits(2) // Reserved
	base, err := readPTSOrDTS(r)
	if err != nil {
		return nil, err
	}

	ext := int64(r.TryReadBits(9))
	_ = r.TryReadBool() // Reserved.

	return newClockReference(base, ext), r.TryError
}

// writePESData first packet will contain PES header with
// optional PES header and payload, if possible all consequential
// packets will contain just payload for the last packet caller
// must add  AF with stuffing, see calcPESDataLength.
func writePESData(
	w *bitio.Writer,
	h *PESHeader,
	payloadLeft []byte,
	isPayloadStart bool,
	bytesAvailable int,
) (totalBytesWritten, payloadBytesWritten int, err error) {
	if isPayloadStart {
		var n int
		n, err = writePESHeader(w, h, len(payloadLeft))
		if err != nil {
			err = fmt.Errorf("writing PES header failed: %w", err)
			return
		}
		totalBytesWritten += n
	}

	payloadBytesWritten = bytesAvailable - totalBytesWritten
	if payloadBytesWritten > len(payloadLeft) {
		payloadBytesWritten = len(payloadLeft)
	}

	_, err = w.Write(payloadLeft[:payloadBytesWritten])
	if err != nil {
		err = fmt.Errorf("writing payload failed: %w", err)
		return
	}

	totalBytesWritten += payloadBytesWritten
	return
}

func writePESHeader(w *bitio.Writer, h *PESHeader, payloadSize int) (int, error) {
	w.TryWriteBits(0x000001, 24) // packet_start_code_prefix
	w.TryWriteByte(h.StreamID)

	pesPacketLength := 0

	if !h.IsVideoStream() {
		pesPacketLength = payloadSize
		if hasPESOptionalHeader(h.StreamID) {
			pesPacketLength += int(calcPESOptionalHeaderLength(h.OptionalHeader))
		}
		if pesPacketLength > 0xffff {
			pesPacketLength = 0
		}
	}

	w.TryWriteBits(uint64(pesPacketLength), 16)

	bytesWritten := pesHeaderLength

	if hasPESOptionalHeader(h.StreamID) {
		n, err := writePESOptionalHeader(bitio.NewWriter(w), h.OptionalHeader)
		if err != nil {
			return 0, fmt.Errorf("writing optional header failed: %w", err)
		}
		bytesWritten += n
	}

	return bytesWritten, w.TryError
}

func calcPESOptionalHeaderLength(h *PESOptionalHeader) uint8 {
	if h == nil {
		return 0
	}
	return 3 + calcPESOptionalHeaderDataLength(h)
}

func calcPESOptionalHeaderDataLength(h *PESOptionalHeader) uint8 {
	var length uint8
	if h.PTSDTSIndicator == PTSDTSIndicatorOnlyPTS {
		length += ptsOrDTSByteLength
	} else if h.PTSDTSIndicator == PTSDTSIndicatorBothPresent {
		length += 2 * ptsOrDTSByteLength
	}

	if h.HasESCR {
		length += escrLength
	}

	if h.HasESRate {
		length += 3
	}

	if h.HasDSMTrickMode {
		length += dsmTrickModeLength
	}

	if h.HasAdditionalCopyInfo {
		length++
	}

	/*if h.HasCRC {
		// length += 4 // TODO
	}*/

	if h.HasExtension {
		length++

		if h.HasPrivateData {
			length += 16
		}

		/*if h.HasPackHeaderField {
			// TODO
		}*/

		if h.HasProgramPacketSequenceCounter {
			length += 2
		}

		if h.HasPSTDBuffer {
			length += 2
		}

		if h.HasExtension2 {
			length += 1 + uint8(len(h.Extension2Data))
		}
	}

	return length
}

func writePESOptionalHeader(w *bitio.Writer, h *PESOptionalHeader) (int, error) { //nolint:funlen
	if h == nil {
		return 0, nil
	}

	w.TryWriteBits(0b10, 2) // Marker bits.
	w.TryWriteBits(uint64(h.ScramblingControl), 2)
	w.TryWriteBool(h.Priority)
	w.TryWriteBool(h.DataAlignmentIndicator)
	w.TryWriteBool(h.IsCopyrighted)
	w.TryWriteBool(h.IsOriginal)

	w.TryWriteBits(uint64(h.PTSDTSIndicator), 2)
	w.TryWriteBool(h.HasESCR)
	w.TryWriteBool(h.HasESRate)
	w.TryWriteBool(h.HasDSMTrickMode)
	w.TryWriteBool(h.HasAdditionalCopyInfo)
	w.TryWriteBool(false) // CRC of previous PES packet. not supported yet
	// b.Write(h.HasCRC)
	w.TryWriteBool(h.HasExtension)

	pesOptionalHeaderDataLength := calcPESOptionalHeaderDataLength(h)
	w.TryWriteByte(pesOptionalHeaderDataLength)

	bytesWritten := 3

	if h.PTSDTSIndicator == PTSDTSIndicatorOnlyPTS {
		n, err := writePTSOrDTS(w, 0b0010, h.PTS)
		if err != nil {
			return 0, fmt.Errorf("PTS only: %w", err)
		}
		bytesWritten += n
	}

	if h.PTSDTSIndicator == PTSDTSIndicatorBothPresent {
		n, err := writePTSOrDTS(w, 0b0011, h.PTS)
		if err != nil {
			return 0, fmt.Errorf("PTS: %w", err)
		}
		bytesWritten += n

		n, err = writePTSOrDTS(w, 0b0001, h.DTS)
		if err != nil {
			return 0, fmt.Errorf("DTS: %w", err)
		}
		bytesWritten += n
	}

	if h.HasESCR {
		n, err := writeESCR(w, h.ESCR)
		if err != nil {
			return 0, fmt.Errorf("ESCR: %w", err)
		}
		bytesWritten += n
	}

	if h.HasESRate {
		w.TryWriteBool(true)
		w.TryWriteBits(uint64(h.ESRate), 22)
		w.TryWriteBool(true)
		bytesWritten += 3
	}

	if h.HasDSMTrickMode {
		n, err := writeDSMTrickMode(w, h.DSMTrickMode)
		if err != nil {
			return 0, fmt.Errorf("DMS trick mode: %w", err)
		}
		bytesWritten += n
	}

	if h.HasAdditionalCopyInfo {
		w.TryWriteBool(true) // marker_bit
		w.TryWriteBits(uint64(h.AdditionalCopyInfo), 7)
		bytesWritten++
	}

	/*if h.HasCRC {
		// TODO, not supported
	}*/

	if h.HasExtension {
		writePESExtension(w, h, &bytesWritten)
	}

	return bytesWritten, w.TryError
}

func writePESExtension(w *bitio.Writer, h *PESOptionalHeader, bytesWritten *int) {
	w.TryWriteBool(h.HasPrivateData)
	w.TryWriteBool(false) // TODO pack_header_field_flag, not implemented
	// b.Write(h.HasPackHeaderField)
	w.TryWriteBool(h.HasProgramPacketSequenceCounter)
	w.TryWriteBool(h.HasPSTDBuffer)
	w.TryWriteBits(0xff, 3) // reserved
	w.TryWriteBool(h.HasExtension2)
	*bytesWritten++

	if h.HasPrivateData {
		w.TryWrite(h.PrivateData)
		*bytesWritten += 16
	}

	/*if h.HasPackHeaderField {
		// TODO (see parsePESOptionalHeader)
	}*/

	if h.HasProgramPacketSequenceCounter {
		w.TryWriteBool(true) // marker_bit
		w.TryWriteBits(uint64(h.PacketSequenceCounter), 7)
		w.TryWriteBool(true) // marker_bit
		w.TryWriteBool(h.MPEG1OrMPEG2ID)
		w.TryWriteBits(uint64(h.OriginalStuffingLength), 6)
		*bytesWritten += 2
	}

	if h.HasPSTDBuffer {
		w.TryWriteBits(0b01, 2)
		w.TryWriteBool(h.PSTDBufferScale)
		w.TryWriteBits(uint64(h.PSTDBufferSize), 13)
		*bytesWritten += 2
	}

	if h.HasExtension2 {
		w.TryWriteBool(true) // marker_bit
		w.TryWriteBits(uint64(len(h.Extension2Data)), 7)
		w.TryWrite(h.Extension2Data)
		*bytesWritten += 1 + len(h.Extension2Data)
	}
}

func writeDSMTrickMode(w *bitio.Writer, m *DSMTrickMode) (int, error) {
	w.TryWriteBits(uint64(m.TrickModeControl), 3)

	switch m.TrickModeControl {
	case TrickModeControlFastForward, TrickModeControlFastReverse:
		w.TryWriteBits(uint64(m.FieldID), 2)
		w.TryWriteBool(m.IntraSliceRefresh)
		w.TryWriteBits(uint64(m.FrequencyTruncation), 2)

	case TrickModeControlFreezeFrame:
		w.TryWriteBits(uint64(m.FieldID), 2)
		w.TryWriteBits(0xff, 3)

	case TrickModeControlSlowMotion, TrickModeControlSlowReverse:
		w.TryWriteBits(uint64(m.RepeatControl), 5)

	default:
		w.TryWriteBits(0xff, 5)
	}

	return dsmTrickModeLength, w.TryError
}

func writeESCR(w *bitio.Writer, cr *ClockReference) (int, error) {
	w.TryWriteBits(0xff, 2)
	w.TryWriteBits(uint64(cr.Base>>30), 3)
	w.TryWriteBool(true)
	w.TryWriteBits(uint64(cr.Base>>15), 15)
	w.TryWriteBool(true)
	w.TryWriteBits(uint64(cr.Base), 15)
	w.TryWriteBool(true)
	w.TryWriteBits(uint64(cr.Extension), 9)
	w.TryWriteBool(true)

	return escrLength, w.TryError
}

func writePTSOrDTS(w *bitio.Writer, flag uint8, cr *ClockReference) (bytesWritten int, retErr error) {
	w.TryWriteBits(uint64(flag), 4)
	w.TryWriteBits(uint64(cr.Base>>30), 3)
	w.TryWriteBool(true)
	w.TryWriteBits(uint64(cr.Base>>15), 15)
	w.TryWriteBool(true)
	w.TryWriteBits(uint64(cr.Base), 15)
	w.TryWriteBool(true)

	return ptsOrDTSByteLength, w.TryError
}
