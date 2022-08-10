package astits

import (
	"errors"
	"fmt"

	"github.com/icza/bitio"
)

// Scrambling Controls.
const (
	ScramblingControlNotScrambled         = 0
	ScramblingControlReservedForFutureUse = 1
	ScramblingControlScrambledWithEvenKey = 2
	ScramblingControlScrambledWithOddKey  = 3
)

// Constants.
const (
	MpegTsPacketSize       = 188
	mpegTsPacketHeaderSize = 3
	pcrBytesSize           = 6
)

// Packet represents a packet.
// https://en.wikipedia.org/wiki/MPEG_transport_stream
type Packet struct {
	AdaptationField *PacketAdaptationField
	Header          *PacketHeader
	Payload         []byte // This is only the payload content.
}

// PacketHeader represents a packet header.
type PacketHeader struct {
	// TransportErrorIndicator set when a demodulator can't correct
	// errors from FEC data; indicating the packet is corrupt.
	TransportErrorIndicator bool

	// PayloadUnitStartIndicator set when a PES, PSI, or DVB-MIP
	// packet begins immediately following the header.
	PayloadUnitStartIndicator bool

	// TransportPriority set when the current packet has a higher
	// priority than other packets with the same PID.
	TransportPriority bool

	// PID Packet Identifier, describing the payload data.
	PID uint16 // 13 bits.

	TransportScramblingControl uint8 // 2 Bits.

	HasAdaptationField bool
	HasPayload         bool

	// ContinuityCounter Sequence number of payload packets
	// (0x00 to 0x0F) within each stream (except PID 8191)
	ContinuityCounter uint8
}

// PacketAdaptationField represents a packet adaptation field.
type PacketAdaptationField struct {
	AdaptationExtensionField *PacketAdaptationExtensionField

	// DiscontinuityIndicator set if current TS packet is
	// in a discontinuity state with respect to either the
	// continuity counter or the program clock reference.
	DiscontinuityIndicator bool

	// ElementaryStreamPriorityIndicator set when this
	// stream should be considered "high priority".
	ElementaryStreamPriorityIndicator bool
	HasAdaptationExtensionField       bool
	HasOPCR                           bool
	HasPCR                            bool
	HasTransportPrivateData           bool
	HasSplicingCountdown              bool
	Length                            int

	// IsOneByteStuffing only used for one byte
	// stuffing - if true, adaptation field will be
	// written as one uint8(0). Not part of TS format.
	IsOneByteStuffing bool

	// StuffingLength only used in writePacketAdaptationField
	// to request stuffing.
	StuffingLength int

	// OPCR Original Program clock reference.
	// Helps when one TS is copied into another.
	OPCR *ClockReference

	// PCR Program clock reference.
	PCR *ClockReference

	// RandomAccessIndicator set when the stream may
	// be decoded without errors from this point.
	RandomAccessIndicator bool

	// SpliceCountdown indicates how many TS packets
	// from this one a splicing point occurs
	// (Two's complement signed; may be negative).
	SpliceCountdown            uint8
	TransportPrivateDataLength uint8
	TransportPrivateData       []byte
}

// PacketAdaptationExtensionField represents a packet adaptation extension field.
type PacketAdaptationExtensionField struct {
	Length uint8

	HasLegalTimeWindow bool
	HasPiecewiseRate   bool
	HasSeamlessSplice  bool

	LegalTimeWindowIsValid bool

	// Extra information for rebroadcasters to determine
	// the state of buffers when packets may be missing.
	LegalTimeWindowOffset uint16 // 15 bits.

	// The rate of the stream, measured in 188-byte
	// packets, to define the end-time of the LTW.
	PiecewiseRate uint32 // 22 bits.

	// Indicates the parameters of the H.262 splice.
	SpliceType uint8 // 4 bits.

	// The PES DTS of the splice point. Split up as 3 bits,
	// 1 marker bit (0x1), 15 bits, 1 marker bit, 15 bits,
	// and 1 marker bit, for 33 data bits total.
	DTSNextAccessUnit *ClockReference
}

// parsePacket parses a packet.
func parsePacket(r *bitio.CountReader, pktLength int64) (*Packet, error) {
	// Packet must start with a sync byte.
	b := r.TryReadByte()
	if b != syncByte {
		return nil, ErrPacketStartSyncByte
	}

	p := &Packet{}

	// In case packet size is bigger than 188 bytes,
	// we don't care for the first bytes.
	var startOffset uint8
	if pktLength > 188*8 {
		startOffset = uint8(pktLength/8 - MpegTsPacketSize)

		skip := make([]byte, startOffset)
		TryReadFull(r, skip)
	}

	var err error
	if p.Header, err = parsePacketHeader(r); err != nil {
		return nil, fmt.Errorf("parsing packet header failed: %w", err)
	}

	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parsePacketAdaptationField(r); err != nil {
			return nil, fmt.Errorf("parsing packet adaptation field failed: %w", err)
		}
	}

	if p.Header.HasPayload {
		payloadOffset := int64(startOffset+4) * 8
		if p.Header.HasAdaptationField {
			payloadOffset += int64(1+p.AdaptationField.Length) * 8
		}

		skip := make([]byte, (payloadOffset-r.BitsCount)/8)
		TryReadFull(r, skip)

		if r.TryError != nil {
			return nil, fmt.Errorf("x %v : %w", (pktLength-r.BitsCount)/8, r.TryError)
		}
		// Read payload.
		p.Payload = make([]byte, (pktLength-r.BitsCount)/8)
		TryReadFull(r, p.Payload)
	}
	if r.TryError != nil {
		return nil, fmt.Errorf("y %v : %w", (pktLength-r.BitsCount)/8, r.TryError)
	}
	return p, r.TryError
}

// parsePacketHeader parses the packet header.
func parsePacketHeader(r *bitio.CountReader) (*PacketHeader, error) {
	h := &PacketHeader{
		TransportErrorIndicator:   r.TryReadBool(),
		PayloadUnitStartIndicator: r.TryReadBool(),
		TransportPriority:         r.TryReadBool(),
		PID:                       uint16(r.TryReadBits(13)),
	}

	h.TransportScramblingControl = uint8(r.TryReadBits(2))
	h.HasAdaptationField = r.TryReadBool()
	h.HasPayload = r.TryReadBool()
	h.ContinuityCounter = uint8(r.TryReadBits(4))

	return h, r.TryError
}

// parsePacketAdaptationField parses the packet adaptation field.
func parsePacketAdaptationField(r *bitio.CountReader) (*PacketAdaptationField, error) { //nolint:funlen
	a := &PacketAdaptationField{}

	a.Length = int(r.TryReadByte())

	afStartOffset := r.BitsCount

	// Invalid length.
	if a.Length <= 0 {
		a.StuffingLength = a.Length - int(r.BitsCount-afStartOffset/8)/8
		return a, nil
	}

	a.DiscontinuityIndicator = r.TryReadBool()
	a.RandomAccessIndicator = r.TryReadBool()
	a.ElementaryStreamPriorityIndicator = r.TryReadBool()
	a.HasPCR = r.TryReadBool()
	a.HasOPCR = r.TryReadBool()
	a.HasSplicingCountdown = r.TryReadBool()
	a.HasTransportPrivateData = r.TryReadBool()
	a.HasAdaptationExtensionField = r.TryReadBool()

	var err error
	if a.HasPCR {
		if a.PCR, err = parsePCR(r); err != nil {
			return nil, fmt.Errorf("parsing PCR failed: %w", err)
		}
	}

	if a.HasOPCR {
		if a.OPCR, err = parsePCR(r); err != nil {
			return nil, fmt.Errorf("parsing OPCR failed: %w", err)
		}
	}

	if a.HasSplicingCountdown {
		a.SpliceCountdown = r.TryReadByte()
	}

	if a.HasTransportPrivateData {
		a.TransportPrivateDataLength = r.TryReadByte()

		if a.TransportPrivateDataLength > 0 {
			a.TransportPrivateData = make([]byte, a.TransportPrivateDataLength)
			TryReadFull(r, a.TransportPrivateData)
		}
	}

	if !a.HasAdaptationExtensionField {
		a.StuffingLength = a.Length - int(r.BitsCount-afStartOffset)/8
		return a, nil
	}

	a.AdaptationExtensionField = &PacketAdaptationExtensionField{}

	a.AdaptationExtensionField.Length = r.TryReadByte()
	if a.AdaptationExtensionField.Length <= 0 {
		a.StuffingLength = a.Length - int(r.BitsCount-afStartOffset)/8
		return a, nil
	}

	a.AdaptationExtensionField.HasLegalTimeWindow = r.TryReadBool()
	a.AdaptationExtensionField.HasPiecewiseRate = r.TryReadBool()
	a.AdaptationExtensionField.HasSeamlessSplice = r.TryReadBool()
	_ = r.TryReadBits(5) // Reserved.

	if a.AdaptationExtensionField.HasLegalTimeWindow {
		a.AdaptationExtensionField.LegalTimeWindowIsValid = r.TryReadBool()
		a.AdaptationExtensionField.LegalTimeWindowOffset = uint16(r.TryReadBits(15))
	}

	if a.AdaptationExtensionField.HasPiecewiseRate {
		_ = r.TryReadBits(2) // Reserved.
		a.AdaptationExtensionField.PiecewiseRate = uint32(r.TryReadBits(22))
	}

	if a.AdaptationExtensionField.HasSeamlessSplice {
		a.AdaptationExtensionField.SpliceType = uint8(r.TryReadBits(4))

		a.AdaptationExtensionField.DTSNextAccessUnit, err = parsePTSOrDTS(r)
		if err != nil {
			return nil, fmt.Errorf("parsing DTSNextAccessUnit failed: %w", err)
		}
	}

	a.StuffingLength = a.Length - int(r.BitsCount-afStartOffset)/8

	return a, r.TryError
}

// parsePCR parses a Program Clock Reference
// Program clock reference, stored as 33 bits base,
// 6 bits reserved, 9 bits extension.
func parsePCR(r *bitio.CountReader) (*ClockReference, error) {
	base := int64(r.TryReadBits(33))
	_ = r.TryReadBits(6) // Reserved.
	ext := int64(r.TryReadBits(9))

	return newClockReference(base, ext), r.TryError
}

// ErrShortPayload .
var ErrShortPayload = errors.New("short payload")

func writePacket(w *bitio.Writer, p *Packet, targetPacketSize int) (written int, err error) {
	if err = w.WriteByte(uint8(syncByte)); err != nil {
		return
	}
	written++

	n, err := writePacketHeader(w, p.Header)
	if err != nil {
		return
	}
	written += n

	if p.Header.HasAdaptationField {
		n, err = writePacketAdaptationField(w, p.AdaptationField)
		if err != nil {
			return
		}
		written += n
	}

	if targetPacketSize-written < len(p.Payload) {
		return 0, fmt.Errorf(
			"%w: payload=%d available=%d",
			ErrShortPayload,
			len(p.Payload),
			targetPacketSize-written,
		)
	}

	if p.Header.HasPayload {
		_, err = w.Write(p.Payload)
		if err != nil {
			return
		}
		written += len(p.Payload)
	}

	for written < targetPacketSize {
		if err = w.WriteByte(uint8(0xff)); err != nil {
			return
		}
		written++
	}

	return written, nil
}

func writePacketHeader(w *bitio.Writer, h *PacketHeader) (written int, retErr error) {
	w.TryWriteBool(h.TransportErrorIndicator)
	w.TryWriteBool(h.PayloadUnitStartIndicator)
	w.TryWriteBool(h.TransportPriority)
	w.TryWriteBits(uint64(h.PID), 13)
	w.TryWriteBits(uint64(h.TransportScramblingControl), 2)
	w.TryWriteBool(h.HasAdaptationField) // adaptation_field_control higher bit.
	w.TryWriteBool(h.HasPayload)         // adaptation_field_control lower bit.
	w.TryWriteBits(uint64(h.ContinuityCounter), 4)

	return mpegTsPacketHeaderSize, w.TryError
}

func writePCR(w *bitio.Writer, cr *ClockReference) (int, error) {
	w.TryWriteBits(uint64(cr.Base), 33)
	w.TryWriteBits(0xff, 6)
	w.TryWriteBits(uint64(cr.Extension), 9)
	return pcrBytesSize, w.TryError
}

func calcPacketAdaptationFieldLength(af *PacketAdaptationField) (length uint8) {
	length++
	if af.HasPCR {
		length += pcrBytesSize
	}
	if af.HasOPCR {
		length += pcrBytesSize
	}
	if af.HasSplicingCountdown {
		length++
	}
	if af.HasTransportPrivateData {
		length += 1 + uint8(len(af.TransportPrivateData))
	}
	if af.HasAdaptationExtensionField {
		length += 1 + calcPacketAdaptationFieldExtensionLength(af.AdaptationExtensionField)
	}
	length += uint8(af.StuffingLength)
	return
}

func writePacketAdaptationField(w *bitio.Writer, af *PacketAdaptationField) (int, error) { //nolint:funlen
	var bytesWritten int

	if af.IsOneByteStuffing {
		w.TryWriteByte(0)
		return 1, nil
	}

	length := calcPacketAdaptationFieldLength(af)
	w.TryWriteByte(length)
	bytesWritten++

	w.TryWriteBool(af.DiscontinuityIndicator)
	w.TryWriteBool(af.RandomAccessIndicator)
	w.TryWriteBool(af.ElementaryStreamPriorityIndicator)
	w.TryWriteBool(af.HasPCR)
	w.TryWriteBool(af.HasOPCR)
	w.TryWriteBool(af.HasSplicingCountdown)
	w.TryWriteBool(af.HasTransportPrivateData)
	w.TryWriteBool(af.HasAdaptationExtensionField)

	bytesWritten++

	if af.HasPCR {
		n, err := writePCR(w, af.PCR)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	if af.HasOPCR {
		n, err := writePCR(w, af.OPCR)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	if af.HasSplicingCountdown {
		w.TryWriteByte(af.SpliceCountdown)
		bytesWritten++
	}

	if af.HasTransportPrivateData {
		// we can get length from TransportPrivateData itself, why do we need separate field?
		w.TryWriteByte(af.TransportPrivateDataLength)
		bytesWritten++
		if af.TransportPrivateDataLength > 0 {
			w.TryWrite(af.TransportPrivateData)
		}
		bytesWritten += len(af.TransportPrivateData)
	}

	if af.HasAdaptationExtensionField {
		n, err := writePacketAdaptationFieldExtension(w, af.AdaptationExtensionField)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	// stuffing
	for i := 0; i < af.StuffingLength; i++ {
		w.TryWriteByte(0xff)
		bytesWritten++
	}
	return bytesWritten, w.TryError
}

func calcPacketAdaptationFieldExtensionLength(
	afe *PacketAdaptationExtensionField,
) (length uint8) {
	length++
	if afe.HasLegalTimeWindow {
		length += 2
	}
	if afe.HasPiecewiseRate {
		length += 3
	}
	if afe.HasSeamlessSplice {
		length += ptsOrDTSByteLength
	}
	return length
}

func writePacketAdaptationFieldExtension(
	w *bitio.Writer, afe *PacketAdaptationExtensionField,
) (int, error) {
	var bytesWritten int

	length := calcPacketAdaptationFieldExtensionLength(afe)
	w.TryWriteByte(length)
	bytesWritten++

	w.TryWriteBool(afe.HasLegalTimeWindow)
	w.TryWriteBool(afe.HasPiecewiseRate)
	w.TryWriteBool(afe.HasSeamlessSplice)
	w.TryWriteBits(0xff, 5) // reserved
	bytesWritten++

	if afe.HasLegalTimeWindow {
		w.TryWriteBool(afe.LegalTimeWindowIsValid)
		w.TryWriteBits(uint64(afe.LegalTimeWindowOffset), 15)
		bytesWritten += 2
	}

	if afe.HasPiecewiseRate {
		w.TryWriteBits(0xff, 2)
		w.TryWriteBits(uint64(afe.PiecewiseRate), 22)
		bytesWritten += 3
	}

	if afe.HasSeamlessSplice {
		n, err := writePTSOrDTS(w, afe.SpliceType, afe.DTSNextAccessUnit)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	return bytesWritten, w.TryError
}

func newStuffingAdaptationField(bytesToStuff int) *PacketAdaptationField {
	if bytesToStuff == 1 {
		return &PacketAdaptationField{
			IsOneByteStuffing: true,
		}
	}

	return &PacketAdaptationField{
		// one byte for length and one for flags
		StuffingLength: bytesToStuff - 2,
	}
}
