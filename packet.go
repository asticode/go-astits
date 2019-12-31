package astits

import (
	"github.com/asticode/go-astikit"
	"github.com/pkg/errors"
)

// Scrambling Controls
const (
	ScramblingControlNotScrambled         = 0
	ScramblingControlReservedForFutureUse = 1
	ScramblingControlScrambledWithEvenKey = 2
	ScramblingControlScrambledWithOddKey  = 3
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
		err = errors.Wrap(err, "astits: getting next byte failed")
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
		err = errors.Wrap(err, "astits: parsing packet header failed")
		return
	}

	// Parse adaptation field
	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parsePacketAdaptationField(i); err != nil {
			err = errors.Wrap(err, "astits: parsing packet adaptation field failed")
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
		err = errors.Wrap(err, "astits: fetching next bytes failed")
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
		err = errors.Wrap(err, "astits: fetching next byte failed")
		return
	}

	// Length
	a.Length = int(b)

	// Valid length
	if a.Length > 0 {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = errors.Wrap(err, "astits: fetching next byte failed")
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
				err = errors.Wrap(err, "astits: parsing PCR failed")
				return
			}
		}

		// OPCR
		if a.HasOPCR {
			if a.OPCR, err = parsePCR(i); err != nil {
				err = errors.Wrap(err, "astits: parsing PCR failed")
				return
			}
		}

		// Splicing countdown
		if a.HasSplicingCountdown {
			if b, err = i.NextByte(); err != nil {
				err = errors.Wrap(err, "astits: fetching next byte failed")
				return
			}
			a.SpliceCountdown = int(b)
		}

		// Transport private data
		if a.HasTransportPrivateData {
			// Length
			if b, err = i.NextByte(); err != nil {
				err = errors.Wrap(err, "astits: fetching next byte failed")
				return
			}
			a.TransportPrivateDataLength = int(b)

			// Data
			if a.TransportPrivateDataLength > 0 {
				if a.TransportPrivateData, err = i.NextBytes(a.TransportPrivateDataLength); err != nil {
					err = errors.Wrap(err, "astits: fetching next bytes failed")
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
				err = errors.Wrap(err, "astits: fetching next byte failed")
				return
			}

			// Length
			a.AdaptationExtensionField.Length = int(b)
			if a.AdaptationExtensionField.Length > 0 {
				// Get next byte
				if b, err = i.NextByte(); err != nil {
					err = errors.Wrap(err, "astits: fetching next byte failed")
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
						err = errors.Wrap(err, "astits: fetching next bytes failed")
						return
					}
					a.AdaptationExtensionField.LegalTimeWindowIsValid = bs[0]&0x80 > 0
					a.AdaptationExtensionField.LegalTimeWindowOffset = uint16(bs[0]&0x7f)<<8 | uint16(bs[1])
				}

				// Piecewise rate
				if a.AdaptationExtensionField.HasPiecewiseRate {
					var bs []byte
					if bs, err = i.NextBytes(3); err != nil {
						err = errors.Wrap(err, "astits: fetching next bytes failed")
						return
					}
					a.AdaptationExtensionField.PiecewiseRate = uint32(bs[0]&0x3f)<<16 | uint32(bs[1])<<8 | uint32(bs[2])
				}

				// Seamless splice
				if a.AdaptationExtensionField.HasSeamlessSplice {
					// Get next byte
					if b, err = i.NextByte(); err != nil {
						err = errors.Wrap(err, "astits: fetching next byte failed")
						return
					}

					// Splice type
					a.AdaptationExtensionField.SpliceType = uint8(b&0xf0) >> 4

					// We need to rewind since the current byte is used by the DTS next access unit as well
					i.Skip(-1)

					// DTS Next access unit
					if a.AdaptationExtensionField.DTSNextAccessUnit, err = parsePTSOrDTS(i); err != nil {
						err = errors.Wrap(err, "astits: parsing DTS failed")
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
		err = errors.Wrap(err, "astits: fetching next bytes failed")
		return
	}
	pcr := uint64(bs[0])<<40 | uint64(bs[1])<<32 | uint64(bs[2])<<24 | uint64(bs[3])<<16 | uint64(bs[4])<<8 | uint64(bs[5])
	cr = newClockReference(int(pcr>>15), int(pcr&0x1ff))
	return
}
