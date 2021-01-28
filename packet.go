package astits

import (
	"fmt"
	"github.com/asticode/go-astikit"
)

// Scrambling Controls
const (
	ScramblingControlNotScrambled         = 0
	ScramblingControlReservedForFutureUse = 1
	ScramblingControlScrambledWithEvenKey = 2
	ScramblingControlScrambledWithOddKey  = 3
	MpegTsPacketSize                      = 188
)

// Packet represents a packet
// https://en.wikipedia.org/wiki/MPEG_transport_stream
type Packet struct {
	AdaptationField *PacketAdaptationField
	Header          *PacketHeader
	Payload         []byte // This is only the payload content
}

// PacketHeader represents a packet header
type PacketHeader struct {
	ContinuityCounter          uint8 // Sequence number of payload packets (0x00 to 0x0F) within each stream (except PID 8191)
	HasAdaptationField         bool
	HasPayload                 bool
	PayloadUnitStartIndicator  bool   // Set when a PES, PSI, or DVB-MIP packet begins immediately following the header.
	PID                        uint16 // Packet Identifier, describing the payload data.
	TransportErrorIndicator    bool   // Set when a demodulator can't correct errors from FEC data; indicating the packet is corrupt.
	TransportPriority          bool   // Set when the current packet has a higher priority than other packets with the same PID.
	TransportScramblingControl uint8
}

// PacketAdaptationField represents a packet adaptation field
type PacketAdaptationField struct {
	AdaptationExtensionField          *PacketAdaptationExtensionField
	DiscontinuityIndicator            bool // Set if current TS packet is in a discontinuity state with respect to either the continuity counter or the program clock reference
	ElementaryStreamPriorityIndicator bool // Set when this stream should be considered "high priority"
	HasAdaptationExtensionField       bool
	HasOPCR                           bool
	HasPCR                            bool
	HasTransportPrivateData           bool
	HasSplicingCountdown              bool
	Length                            int
	OPCR                              *ClockReference // Original Program clock reference. Helps when one TS is copied into another
	PCR                               *ClockReference // Program clock reference
	RandomAccessIndicator             bool            // Set when the stream may be decoded without errors from this point
	SpliceCountdown                   int             // Indicates how many TS packets from this one a splicing point occurs (Two's complement signed; may be negative)
	TransportPrivateDataLength        int
	TransportPrivateData              []byte
}

// PacketAdaptationExtensionField represents a packet adaptation extension field
type PacketAdaptationExtensionField struct {
	DTSNextAccessUnit      *ClockReference // The PES DTS of the splice point. Split up as 3 bits, 1 marker bit (0x1), 15 bits, 1 marker bit, 15 bits, and 1 marker bit, for 33 data bits total.
	HasLegalTimeWindow     bool
	HasPiecewiseRate       bool
	HasSeamlessSplice      bool
	LegalTimeWindowIsValid bool
	LegalTimeWindowOffset  uint16 // Extra information for rebroadcasters to determine the state of buffers when packets may be missing.
	Length                 int
	PiecewiseRate          uint32 // The rate of the stream, measured in 188-byte packets, to define the end-time of the LTW.
	SpliceType             uint8  // Indicates the parameters of the H.262 splice.
}

// parsePacket parses a packet
func parsePacket(i *astikit.BytesIterator) (p *Packet, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: getting next byte failed: %w", err)
		return
	}

	// Packet must start with a sync byte
	if b != syncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	// Create packet
	p = &Packet{}

	// In case packet size is bigger than 188 bytes, we don't care for the first bytes
	i.Seek(i.Len() - 188 + 1)
	offsetStart := i.Offset()

	// Parse header
	if p.Header, err = parsePacketHeader(i); err != nil {
		err = fmt.Errorf("astits: parsing packet header failed: %w", err)
		return
	}

	// Parse adaptation field
	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parsePacketAdaptationField(i); err != nil {
			err = fmt.Errorf("astits: parsing packet adaptation field failed: %w", err)
			return
		}
	}

	// Build payload
	if p.Header.HasPayload {
		i.Seek(payloadOffset(offsetStart, p.Header, p.AdaptationField))
		p.Payload = i.Dump()
	}
	return
}

// payloadOffset returns the payload offset
func payloadOffset(offsetStart int, h *PacketHeader, a *PacketAdaptationField) (offset int) {
	offset = offsetStart + 3
	if h.HasAdaptationField {
		offset += 1 + a.Length
	}
	return
}

// parsePacketHeader parses the packet header
func parsePacketHeader(i *astikit.BytesIterator) (h *PacketHeader, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create header
	h = &PacketHeader{
		ContinuityCounter:          uint8(bs[2] & 0xf),
		HasAdaptationField:         bs[2]&0x20 > 0,
		HasPayload:                 bs[2]&0x10 > 0,
		PayloadUnitStartIndicator:  bs[0]&0x40 > 0,
		PID:                        uint16(bs[0]&0x1f)<<8 | uint16(bs[1]),
		TransportErrorIndicator:    bs[0]&0x80 > 0,
		TransportPriority:          bs[0]&0x20 > 0,
		TransportScramblingControl: uint8(bs[2]) >> 6 & 0x3,
	}
	return
}

// parsePacketAdaptationField parses the packet adaptation field
func parsePacketAdaptationField(i *astikit.BytesIterator) (a *PacketAdaptationField, err error) {
	// Create adaptation field
	a = &PacketAdaptationField{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Length
	a.Length = int(b)

	// Valid length
	if a.Length > 0 {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Flags
		a.DiscontinuityIndicator = b&0x80 > 0
		a.RandomAccessIndicator = b&0x40 > 0
		a.ElementaryStreamPriorityIndicator = b&0x20 > 0
		a.HasPCR = b&0x10 > 0
		a.HasOPCR = b&0x08 > 0
		a.HasSplicingCountdown = b&0x04 > 0
		a.HasTransportPrivateData = b&0x02 > 0
		a.HasAdaptationExtensionField = b&0x01 > 0

		// PCR
		if a.HasPCR {
			if a.PCR, err = parsePCR(i); err != nil {
				err = fmt.Errorf("astits: parsing PCR failed: %w", err)
				return
			}
		}

		// OPCR
		if a.HasOPCR {
			if a.OPCR, err = parsePCR(i); err != nil {
				err = fmt.Errorf("astits: parsing PCR failed: %w", err)
				return
			}
		}

		// Splicing countdown
		if a.HasSplicingCountdown {
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}
			a.SpliceCountdown = int(b)
		}

		// Transport private data
		if a.HasTransportPrivateData {
			// Length
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}
			a.TransportPrivateDataLength = int(b)

			// Data
			if a.TransportPrivateDataLength > 0 {
				if a.TransportPrivateData, err = i.NextBytes(a.TransportPrivateDataLength); err != nil {
					err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
					return
				}
			}
		}

		// Adaptation extension
		if a.HasAdaptationExtensionField {
			// Create extension field
			a.AdaptationExtensionField = &PacketAdaptationExtensionField{}

			// Get next byte
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}

			// Length
			a.AdaptationExtensionField.Length = int(b)
			if a.AdaptationExtensionField.Length > 0 {
				// Get next byte
				if b, err = i.NextByte(); err != nil {
					err = fmt.Errorf("astits: fetching next byte failed: %w", err)
					return
				}

				// Basic
				a.AdaptationExtensionField.HasLegalTimeWindow = b&0x80 > 0
				a.AdaptationExtensionField.HasPiecewiseRate = b&0x40 > 0
				a.AdaptationExtensionField.HasSeamlessSplice = b&0x20 > 0

				// Legal time window
				if a.AdaptationExtensionField.HasLegalTimeWindow {
					var bs []byte
					if bs, err = i.NextBytes(2); err != nil {
						err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
						return
					}
					a.AdaptationExtensionField.LegalTimeWindowIsValid = bs[0]&0x80 > 0
					a.AdaptationExtensionField.LegalTimeWindowOffset = uint16(bs[0]&0x7f)<<8 | uint16(bs[1])
				}

				// Piecewise rate
				if a.AdaptationExtensionField.HasPiecewiseRate {
					var bs []byte
					if bs, err = i.NextBytes(3); err != nil {
						err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
						return
					}
					a.AdaptationExtensionField.PiecewiseRate = uint32(bs[0]&0x3f)<<16 | uint32(bs[1])<<8 | uint32(bs[2])
				}

				// Seamless splice
				if a.AdaptationExtensionField.HasSeamlessSplice {
					// Get next byte
					if b, err = i.NextByte(); err != nil {
						err = fmt.Errorf("astits: fetching next byte failed: %w", err)
						return
					}

					// Splice type
					a.AdaptationExtensionField.SpliceType = uint8(b&0xf0) >> 4

					// We need to rewind since the current byte is used by the DTS next access unit as well
					i.Skip(-1)

					// DTS Next access unit
					if a.AdaptationExtensionField.DTSNextAccessUnit, err = parsePTSOrDTS(i); err != nil {
						err = fmt.Errorf("astits: parsing DTS failed: %w", err)
						return
					}
				}
			}
		}
	}
	return
}

// parsePCR parses a Program Clock Reference
// Program clock reference, stored as 33 bits base, 6 bits reserved, 9 bits extension.
func parsePCR(i *astikit.BytesIterator) (cr *ClockReference, err error) {
	var bs []byte
	if bs, err = i.NextBytes(6); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	pcr := uint64(bs[0])<<40 | uint64(bs[1])<<32 | uint64(bs[2])<<24 | uint64(bs[3])<<16 | uint64(bs[4])<<8 | uint64(bs[5])
	cr = newClockReference(int64(pcr>>15), int64(pcr&0x1ff))
	return
}

func recoverAndSaveError(retErr *error) {
	if err := recover(); err != nil {
		*retErr = err.(error)
	}
}

func tryWrite(w *astikit.BitsWriter, i interface{}) {
	if err := w.Write(i); err != nil {
		panic(err)
	}
}

func tryWriteN(w *astikit.BitsWriter, i interface{}, n int) {
	if err := w.WriteN(i, n); err != nil {
		panic(err)
	}
}

func writePacketHeader(w *astikit.BitsWriter, h *PacketHeader) (written int, retErr error) {
	defer recoverAndSaveError(&retErr)

	tryWrite(w, h.TransportErrorIndicator)
	tryWrite(w, h.PayloadUnitStartIndicator)
	tryWrite(w, h.TransportPriority)
	tryWriteN(w, h.PID, 13)
	tryWriteN(w, h.TransportScramblingControl, 2)
	tryWrite(w, h.HasAdaptationField) // adaptation_field_control higher bit
	tryWrite(w, h.HasPayload)         // adaptation_field_control lower bit
	tryWriteN(w, h.ContinuityCounter, 4)

	return 3, nil
}

func writePCR(w *astikit.BitsWriter, cr *ClockReference) (int, error) {
	var bs [6]byte
	base := cr.Base << 15
	bs[0] = byte((base >> 40) & 0xff)
	bs[1] = byte((base >> 32) & 0xff)
	bs[2] = byte((base >> 24) & 0xff)
	bs[3] = byte((base >> 16) & 0xff)
	bs[4] = byte((base>>8)&0x80) | byte((cr.Extension>>8)&0x7f) | byte(0b1111110) // last 6 are reserved bits
	bs[5] = byte(cr.Extension & 0xff)

	if err := w.Write(bs[:]); err != nil {
		return 0, err
	}
	return len(bs), nil
}

func writePacketAdaptationField(w *astikit.BitsWriter, af *PacketAdaptationField) (writtenBytes int, retErr error) {
	defer recoverAndSaveError(&retErr)

	tryWrite(w, uint8(af.Length))
	writtenBytes++

	if af.Length == 0 {
		return
	}

	tryWrite(w, af.DiscontinuityIndicator)
	tryWrite(w, af.RandomAccessIndicator)
	tryWrite(w, af.ElementaryStreamPriorityIndicator)
	tryWrite(w, af.HasPCR)
	tryWrite(w, af.HasOPCR)
	tryWrite(w, af.HasSplicingCountdown)
	tryWrite(w, af.HasTransportPrivateData)
	tryWrite(w, af.HasAdaptationExtensionField)

	writtenBytes++

	if af.HasPCR {
		n, err := writePCR(w, af.PCR)
		if err != nil {
			return 0, err
		}
		writtenBytes += n
	}

	if af.HasOPCR {
		n, err := writePCR(w, af.OPCR)
		if err != nil {
			return 0, err
		}
		writtenBytes += n
	}

	if af.HasSplicingCountdown {
		tryWrite(w, uint8(af.SpliceCountdown))
		writtenBytes++
	}

	if af.HasTransportPrivateData {
		tryWrite(w, uint8(af.TransportPrivateDataLength))
		writtenBytes++
		if af.TransportPrivateDataLength > 0 {
			tryWrite(w, af.TransportPrivateData)
		}
		writtenBytes += len(af.TransportPrivateData)
	}

	if af.HasAdaptationExtensionField {
		n, err := writePacketAdaptationFieldExtension(w, af.AdaptationExtensionField)
		if err != nil {
			return 0, err
		}
		writtenBytes += n
	}

	if writtenBytes-1 > af.Length {
		return writtenBytes, fmt.Errorf(
			"PacketAdaptationField provided Length %d is less than actually written %d",
			af.Length, writtenBytes,
		)
	}

	// stuffing
	for writtenBytes-1 < af.Length {
		tryWrite(w, uint8(0))
		writtenBytes++
	}

	return
}

func writePacketAdaptationFieldExtension(w *astikit.BitsWriter, afe *PacketAdaptationExtensionField) (writtenBytes int, retErr error) {
	defer recoverAndSaveError(&retErr)

	tryWrite(w, uint8(afe.Length))
	writtenBytes++

	if afe.Length == 0 {
		return writtenBytes, nil
	}

	tryWrite(w, afe.HasLegalTimeWindow)
	tryWrite(w, afe.HasPiecewiseRate)
	tryWrite(w, afe.HasSeamlessSplice)
	tryWriteN(w, uint8(0xff), 5) // reserved
	writtenBytes++

	if afe.HasLegalTimeWindow {
		tryWrite(w, afe.LegalTimeWindowIsValid)
		tryWriteN(w, afe.LegalTimeWindowOffset, 15)
		writtenBytes += 2
	}

	if afe.HasPiecewiseRate {
		tryWriteN(w, uint8(0xff), 2)
		tryWriteN(w, afe.PiecewiseRate, 22)
		writtenBytes += 3
	}

	if afe.HasSeamlessSplice {
		n, err := writePTSOrDTS(w, afe.SpliceType, afe.DTSNextAccessUnit)
		if err != nil {
			return 0, err
		}
		writtenBytes += n
	}

	if writtenBytes-1 > afe.Length {
		return writtenBytes, fmt.Errorf(
			"PacketAdaptationFieldExtension provided Length %d is less than actually written %d",
			afe.Length, writtenBytes,
		)
	}

	// reserved bytes
	for writtenBytes-1 < afe.Length {
		tryWrite(w, uint8(1))
		writtenBytes++
	}

	return
}
