package astits

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
	Bytes           []byte // This is the whole packet content
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
func parsePacket(i []byte) (p *Packet, err error) {
	// Packet must start with a sync byte
	if i[0] != syncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	// Init
	p = &Packet{Bytes: i}

	// In case packet size is bigger than 188 bytes, we don't care for the first bytes
	i = i[len(i)-188+1:]

	// Parse header
	p.Header = parsePacketHeader(i)

	// Parse adaptation field
	if p.Header.HasAdaptationField {
		p.AdaptationField = parsePacketAdaptationField(i[3:])
	}

	// Build payload
	if p.Header.HasPayload {
		p.Payload = i[payloadOffset(p.Header, p.AdaptationField):]
	}
	return
}

// payloadOffset returns the payload offset
func payloadOffset(h *PacketHeader, a *PacketAdaptationField) (offset int) {
	offset = 3
	if h.HasAdaptationField {
		offset += 1 + a.Length
	}
	return
}

// parsePacketHeader parses the packet header
func parsePacketHeader(i []byte) *PacketHeader {
	return &PacketHeader{
		ContinuityCounter:         uint8(i[2] & 0xf),
		HasAdaptationField:        i[2]&0x20 > 0,
		HasPayload:                i[2]&0x10 > 0,
		PayloadUnitStartIndicator: i[0]&0x40 > 0,
		PID: uint16(i[0]&0x1f)<<8 | uint16(i[1]),
		TransportErrorIndicator:    i[0]&0x80 > 0,
		TransportPriority:          i[0]&0x20 > 0,
		TransportScramblingControl: uint8(i[2]) >> 6 & 0x3,
	}
}

// parsePacketAdaptationField parses the packet adaptation field
func parsePacketAdaptationField(i []byte) (a *PacketAdaptationField) {
	// Init
	a = &PacketAdaptationField{}
	var offset int

	// Length
	a.Length = int(i[offset])
	offset += 1

	// Valid length
	if a.Length > 0 {
		// Flags
		a.DiscontinuityIndicator = i[offset]&0x80 > 0
		a.RandomAccessIndicator = i[offset]&0x40 > 0
		a.ElementaryStreamPriorityIndicator = i[offset]&0x20 > 0
		a.HasPCR = i[offset]&0x10 > 0
		a.HasOPCR = i[offset]&0x08 > 0
		a.HasSplicingCountdown = i[offset]&0x04 > 0
		a.HasTransportPrivateData = i[offset]&0x02 > 0
		a.HasAdaptationExtensionField = i[offset]&0x01 > 0
		offset += 1

		// PCR
		if a.HasPCR {
			a.PCR = parsePCR(i[offset:])
			offset += 6
		}

		// OPCR
		if a.HasOPCR {
			a.OPCR = parsePCR(i[offset:])
			offset += 6
		}

		// Splicing countdown
		if a.HasSplicingCountdown {
			a.SpliceCountdown = int(i[offset])
			offset += 1
		}

		// Transport private data
		if a.HasTransportPrivateData {
			a.TransportPrivateDataLength = int(i[offset])
			offset += 1
			if a.TransportPrivateDataLength > 0 {
				a.TransportPrivateData = i[offset : offset+a.TransportPrivateDataLength]
				offset += a.TransportPrivateDataLength
			}
		}

		// Adaptation extension
		if a.HasAdaptationExtensionField {
			a.AdaptationExtensionField = &PacketAdaptationExtensionField{Length: int(i[offset])}
			offset += 1
			if a.AdaptationExtensionField.Length > 0 {
				// Basic
				a.AdaptationExtensionField.HasLegalTimeWindow = i[offset]&0x80 > 0
				a.AdaptationExtensionField.HasPiecewiseRate = i[offset]&0x40 > 0
				a.AdaptationExtensionField.HasSeamlessSplice = i[offset]&0x20 > 0
				offset += 1

				// Legal time window
				if a.AdaptationExtensionField.HasLegalTimeWindow {
					a.AdaptationExtensionField.LegalTimeWindowIsValid = i[offset]&0x80 > 0
					a.AdaptationExtensionField.LegalTimeWindowOffset = uint16(i[offset]&0x7f)<<8 | uint16(i[offset+1])
					offset += 2
				}

				// Piecewise rate
				if a.AdaptationExtensionField.HasPiecewiseRate {
					a.AdaptationExtensionField.PiecewiseRate = uint32(i[offset]&0x3f)<<16 | uint32(i[offset+1])<<8 | uint32(i[offset+2])
					offset += 3
				}

				// Seamless splice
				if a.AdaptationExtensionField.HasSeamlessSplice {
					a.AdaptationExtensionField.SpliceType = uint8(i[offset]&0xf0) >> 4
					a.AdaptationExtensionField.DTSNextAccessUnit = parsePTSOrDTS(i[offset:])
				}
			}
		}
	}
	return
}

// parsePCR parses a Program Clock Reference
// Program clock reference, stored as 33 bits base, 6 bits reserved, 9 bits extension.
func parsePCR(i []byte) *ClockReference {
	var pcr = uint64(i[0])<<40 | uint64(i[1])<<32 | uint64(i[2])<<24 | uint64(i[3])<<16 | uint64(i[4])<<8 | uint64(i[5])
	return newClockReference(int(pcr>>15), int(pcr&0x1ff))
}
