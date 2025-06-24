package astits

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/asticode/go-astikit"
)

// Scrambling Controls
const (
	ScramblingControlNotScrambled         = 0
	ScramblingControlReservedForFutureUse = 1
	ScramblingControlScrambledWithEvenKey = 2
	ScramblingControlScrambledWithOddKey  = 3
)

const (
	MpegTsPacketSize       = 188
	mpegTsPacketHeaderSize = 3
	pcrBytesSize           = 6
)

// Packet represents a packet
// https://en.wikipedia.org/wiki/MPEG_transport_stream
type Packet struct {
	AdaptationField *PacketAdaptationField
	Header          PacketHeader
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
	StuffingLength                    int  // Only used in writePacketAdaptationField to request stuffing
	IsOneByteStuffing                 bool // Only used for one byte stuffing - if true, adaptation field will be written as one uint8(0). Not part of TS format
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

var ErrNoRoomInBuffer = errors.New("No room to serialise into buffer")

// ParsePacket parses a packet into
func ParsePacket(b []byte) (p *Packet, err error) {
	return parsePacket(astikit.NewBytesIterator(b))
}

// ParsePacketWithoutPayload parses a packet without copying the packet payload
func ParsePacketWithoutPayload(b []byte) (p *Packet, err error) {
	return parsePacketWithoutPayload(astikit.NewBytesIterator(b))
}

// ParsePSIPacket parses a known PSI packet
func ParsePSIPacket(p *Packet) (*PSIData, error) {
	return parsePSIData(astikit.NewBytesIterator(p.Payload))
}

// ParsePESPacket parses a known PES packet
func ParsePESPacket(p *Packet) (d *PESData, err error) {
	//Need to protect against posibility of reading a header that doesn't have payload attached
	return parsePESData(astikit.NewBytesIterator(p.Payload))
}

// ParsePESPacket parses a known PES packet
func ParsePESPacketHeader(p *Packet) (d *PESData, err error) {
	//Need to protect against posibility of reading a header that doesn't have payload attached
	i := astikit.NewBytesIterator(p.Payload)
	// Create data
	d = &PESData{}

	// Skip first 3 bytes that are there to identify the PES payload
	i.Seek(3)

	// Parse header
	if d.Header, _, _, err = parsePESHeader(i); err != nil {
		err = fmt.Errorf("astits: parsing PES header failed: %w", err)
		return
	}
	return
}

func (p *Packet) Serialise(b []byte) (int, error) {
	if len(b) < 188 {
		return 0, errors.New("b not large enough to hold a packet")
	}
	b[0] = syncByte
	p.Header.Serialise(b)
	payloadStart := 4
	if p.Header.HasAdaptationField && p.AdaptationField == nil {
		return 0, errors.New("adaptation field not implemented")
	}
	if p.Header.HasAdaptationField {
		err := p.AdaptationField.Serialise(b[payloadStart:])
		if err != nil {
			return payloadStart, err
		}
		payloadStart += p.AdaptationField.Length + 1
	}
	copy(b[payloadStart:], p.Payload)

	return payloadStart, nil
}

func (h *PacketHeader) Serialise(b []byte) {
	teiBit, tpBit, pusiBit := uint8(0x0), uint8(0x0), uint8(0x0)
	if h.TransportErrorIndicator {
		teiBit = 0x80
	}
	if h.PayloadUnitStartIndicator {
		pusiBit = 0x40
	}
	if h.TransportPriority {
		tpBit = 0x20
	}
	pidBits := uint8((h.PID & uint16(0x1f00)) >> 8)
	b[1] = teiBit | tpBit | pusiBit | pidBits
	b[2] = uint8(h.PID & uint16(0xff))

	afBit, pBit, ccBits, tscBits := uint8(0x0), uint8(0x0), uint8(0x0), uint8(0x0)
	if h.HasAdaptationField {
		afBit = 0x20
	}
	if h.HasPayload {
		pBit = 0x10
	}
	ccBits = uint8(h.ContinuityCounter & 0xf)
	tscBits = h.TransportScramblingControl << 6
	b[3] = afBit | pBit | ccBits | tscBits
}

func (a *PacketAdaptationField) Serialise(b []byte) error {
	buf := new(bytes.Buffer)

	// Write adaptation_field_length
	buf.WriteByte(byte(a.Length))

	if a.Length > 0 {
		// Write Flags
		flags := byte(0)
		if a.DiscontinuityIndicator {
			flags |= 0x80
		}
		if a.RandomAccessIndicator {
			flags |= 0x40
		}
		if a.ElementaryStreamPriorityIndicator {
			flags |= 0x20
		}
		if a.HasPCR {
			flags |= 0x10
		}
		if a.HasOPCR {
			flags |= 0x08
		}
		if a.HasSplicingCountdown {
			flags |= 0x04
		}
		if a.HasTransportPrivateData {
			flags |= 0x02
		}
		if a.HasAdaptationExtensionField {
			flags |= 0x01
		}
		buf.WriteByte(flags)

		// Write PCR
		if a.HasPCR {
			pcr := uint64(a.PCR.Base<<15) | uint64(a.PCR.Extension&0x1FF) | (uint64(0x3f) << 9)
			buf.Write([]byte{
				byte(pcr >> 40),
				byte(pcr >> 32),
				byte(pcr >> 24),
				byte(pcr >> 16),
				byte(pcr >> 8),
				byte(pcr),
			})
		}

		// Write OPCR
		if a.HasOPCR {
			opcr := uint64(a.OPCR.Base<<15) | (uint64(a.OPCR.Extension & 0x1FF)) | (uint64(0x3f) << 9)
			buf.Write([]byte{
				byte(opcr >> 40),
				byte(opcr >> 32),
				byte(opcr >> 24),
				byte(opcr >> 16),
				byte(opcr >> 8),
				byte(opcr),
			})
		}

		// Write Splicing countdown
		if a.HasSplicingCountdown {
			buf.WriteByte(byte(a.SpliceCountdown))
		}

		// Write Transport private data
		if a.HasTransportPrivateData {
			buf.WriteByte(byte(a.TransportPrivateDataLength))
			buf.Write(a.TransportPrivateData)
		}

		// Write Adaptation extension field
		if a.HasAdaptationExtensionField {
			buf.WriteByte(byte(a.AdaptationExtensionField.Length))
			if a.AdaptationExtensionField.Length > 0 {
				extensionFlags := byte(0x1F)
				if a.AdaptationExtensionField.HasLegalTimeWindow {
					extensionFlags |= 0x80
				}
				if a.AdaptationExtensionField.HasPiecewiseRate {
					extensionFlags |= 0x40
				}
				if a.AdaptationExtensionField.HasSeamlessSplice {
					extensionFlags |= 0x20
				}
				buf.WriteByte(extensionFlags)

				if a.AdaptationExtensionField.HasLegalTimeWindow {
					ltw := (uint16(0) << 15) | a.AdaptationExtensionField.LegalTimeWindowOffset
					if a.AdaptationExtensionField.LegalTimeWindowIsValid {
						ltw |= 0x8000
					}
					binary.Write(buf, binary.BigEndian, ltw)
				}

				if a.AdaptationExtensionField.HasPiecewiseRate {
					piecewiseRate := 0xc00000 | 0x3fffff&a.AdaptationExtensionField.PiecewiseRate
					buf.Write([]byte{
						byte(piecewiseRate >> 16),
						byte(piecewiseRate >> 8),
						byte(piecewiseRate),
					})
				}

				if a.AdaptationExtensionField.HasSeamlessSplice {
					dtsNextAU := a.AdaptationExtensionField.DTSNextAccessUnit.Base
					buf.WriteByte(byte(a.AdaptationExtensionField.SpliceType<<4) | byte(dtsNextAU>>29)&0xe | 0x01)
					buf.Write([]byte{

						byte(dtsNextAU >> 22),
						byte(dtsNextAU>>14 | 0x01),
						byte(dtsNextAU >> 7),
						byte(dtsNextAU<<1 | 0x01),
					})
				}
			}
		}

		// Add stuffing bytes if needed
		for buf.Len() < a.Length+1 { //Include the length byte
			buf.WriteByte(0xFF)
		}
	}

	// Ensure the buffer length is correct
	if len(buf.Bytes()) > len(b) {
		return fmt.Errorf("astits: not enough room in buffer")
	}
	copy(b, buf.Bytes())

	return nil
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
	i.Seek(i.Len() - MpegTsPacketSize + 1)
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

func parsePacketWithoutPayload(i *astikit.BytesIterator) (p *Packet, err error) {
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

	return
}

func UnmarshalPacketWithoutPayload(i BytesIterator, p *Packet) error {
	var b byte
	var err error
	if b, err = i.NextByte(); err != nil {
		return fmt.Errorf("astits: getting next byte failed: %w", err)
	}

	// Packet must start with a sync byte
	if b != syncByte {
		return ErrPacketMustStartWithASyncByte
	}

	// In case packet size is bigger than 188 bytes, we don't care for the first bytes
	i.Seek(i.Len() - MpegTsPacketSize + 1)

	// Parse header
	if err := unmarshalPacketHeader(i, &p.Header); err != nil {
		return fmt.Errorf("astits: parsing packet header failed: %w", err)
	}

	// Parse adaptation field
	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parsePacketAdaptationField(i); err != nil {
			return fmt.Errorf("astits: parsing packet adaptation field failed: %w", err)
		}
	} else {
		p.AdaptationField = nil
	}

	return nil
}

// payloadOffset returns the payload offset
func payloadOffset(offsetStart int, h PacketHeader, a *PacketAdaptationField) (offset int) {
	offset = offsetStart + 3
	if h.HasAdaptationField {
		offset += 1 + a.Length
	}
	return
}

func unmarshalPacketHeader(i BytesIterator, ph *PacketHeader) (err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	ph.ContinuityCounter = uint8(bs[2] & 0xf)
	ph.HasAdaptationField = bs[2]&0x20 > 0
	ph.HasPayload = bs[2]&0x10 > 0
	ph.PayloadUnitStartIndicator = bs[0]&0x40 > 0
	ph.PID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])
	ph.TransportErrorIndicator = bs[0]&0x80 > 0
	ph.TransportPriority = bs[0]&0x20 > 0
	ph.TransportScramblingControl = uint8(bs[2]) >> 6 & 0x3
	return nil
}

// parsePacketHeader parses the packet header
func parsePacketHeader(i *astikit.BytesIterator) (h PacketHeader, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create header
	return PacketHeader{
		ContinuityCounter:          uint8(bs[2] & 0xf),
		HasAdaptationField:         bs[2]&0x20 > 0,
		HasPayload:                 bs[2]&0x10 > 0,
		PayloadUnitStartIndicator:  bs[0]&0x40 > 0,
		PID:                        uint16(bs[0]&0x1f)<<8 | uint16(bs[1]),
		TransportErrorIndicator:    bs[0]&0x80 > 0,
		TransportPriority:          bs[0]&0x20 > 0,
		TransportScramblingControl: uint8(bs[2]) >> 6 & 0x3,
	}, nil
}

type BytesIterator interface {
	NextByte() (b byte, err error)
	Offset() int
	NextBytes(n int) (bs []byte, err error)
	NextBytesNoCopy(n int) (bs []byte, err error)
	Skip(n int)
	Seek(n int)
	Len() int
}

// parsePacketAdaptationField parses the packet adaptation field
func parsePacketAdaptationField(i BytesIterator) (a *PacketAdaptationField, err error) {
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

	afStartOffset := i.Offset()

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
					if bs, err = i.NextBytesNoCopy(2); err != nil {
						err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
						return
					}
					a.AdaptationExtensionField.LegalTimeWindowIsValid = bs[0]&0x80 > 0
					a.AdaptationExtensionField.LegalTimeWindowOffset = uint16(bs[0]&0x7f)<<8 | uint16(bs[1])
				}

				// Piecewise rate
				if a.AdaptationExtensionField.HasPiecewiseRate {
					var bs []byte
					if bs, err = i.NextBytesNoCopy(3); err != nil {
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

	a.StuffingLength = a.Length - (i.Offset() - afStartOffset)

	return
}

// parsePCR parses a Program Clock Reference
// Program clock reference, stored as 33 bits base, 6 bits reserved, 9 bits extension.
func parsePCR(i BytesIterator) (cr *ClockReference, err error) {
	var bs []byte
	if bs, err = i.NextBytesNoCopy(6); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	pcr := uint64(bs[0])<<40 | uint64(bs[1])<<32 | uint64(bs[2])<<24 | uint64(bs[3])<<16 | uint64(bs[4])<<8 | uint64(bs[5])
	cr = newClockReference(int64(pcr>>15), int64(pcr&0x1ff))
	return
}
