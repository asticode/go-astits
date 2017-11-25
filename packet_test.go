package astits

import (
	"fmt"
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
)

func packet(h PacketHeader, a PacketAdaptationField, i []byte) ([]byte, *Packet) {
	w := astibinary.New()
	w.Write(uint8(syncByte))                             // Sync byte
	w.Write([]byte("test"))                              // Sometimes packets are 192 bytes
	w.Write(packetHeaderBytes(h))                        // Header
	w.Write(packetAdaptationFieldBytes(a))               // Adaptation field
	var payload = append(i, make([]byte, 147-len(i))...) // Payload
	w.Write(payload)
	return w.Bytes(), &Packet{
		AdaptationField: packetAdaptationField,
		Bytes:           w.Bytes(),
		Header:          packetHeader,
		Payload:         payload,
	}
}

func TestParsePacket(t *testing.T) {
	// Packet not starting with a sync
	w := astibinary.New()
	w.Write(uint16(1)) // Invalid sync byte
	_, err := parsePacket(w.Bytes())
	assert.EqualError(t, err, ErrPacketMustStartWithASyncByte.Error())

	// Valid
	b, ep := packet(*packetHeader, *packetAdaptationField, []byte("payload"))
	p, err := parsePacket(b)
	assert.NoError(t, err)
	assert.Equal(t, p, ep)
}

func TestPayloadOffset(t *testing.T) {
	assert.Equal(t, 3, payloadOffset(&PacketHeader{}, nil))
	assert.Equal(t, 6, payloadOffset(&PacketHeader{HasAdaptationField: true}, &PacketAdaptationField{Length: 2}))
}

var packetHeader = &PacketHeader{
	ContinuityCounter:         10,
	HasAdaptationField:        true,
	HasPayload:                true,
	PayloadUnitStartIndicator: true,
	PID: 5461,
	TransportErrorIndicator:    true,
	TransportPriority:          true,
	TransportScramblingControl: ScramblingControlScrambledWithEvenKey,
}

func packetHeaderBytes(h PacketHeader) []byte {
	w := astibinary.New()
	w.Write("1")                                      // Transport error indicator
	w.Write(h.PayloadUnitStartIndicator)              // Payload unit start indicator
	w.Write("1")                                      // Transport priority
	w.Write(fmt.Sprintf("%.13b", h.PID))              // PID
	w.Write("10")                                     // Scrambling control
	w.Write("11")                                     // Adaptation field control
	w.Write(fmt.Sprintf("%.4b", h.ContinuityCounter)) // Continuity counter
	return w.Bytes()
}

func TestParsePacketHeader(t *testing.T) {
	assert.Equal(t, packetHeader, parsePacketHeader(packetHeaderBytes(*packetHeader)))
}

var packetAdaptationField = &PacketAdaptationField{
	AdaptationExtensionField: &PacketAdaptationExtensionField{
		DTSNextAccessUnit:      dtsClockReference,
		HasLegalTimeWindow:     true,
		HasPiecewiseRate:       true,
		HasSeamlessSplice:      true,
		LegalTimeWindowIsValid: true,
		LegalTimeWindowOffset:  10922,
		Length:                 11,
		PiecewiseRate:          2796202,
		SpliceType:             2,
	},
	DiscontinuityIndicator:            true,
	ElementaryStreamPriorityIndicator: true,
	HasAdaptationExtensionField:       true,
	HasOPCR:                           true,
	HasPCR:                            true,
	HasTransportPrivateData:           true,
	HasSplicingCountdown:              true,
	Length:                            36,
	OPCR:                              pcr,
	PCR:                               pcr,
	RandomAccessIndicator:      true,
	SpliceCountdown:            2,
	TransportPrivateDataLength: 4,
	TransportPrivateData:       []byte("test"),
}

func packetAdaptationFieldBytes(a PacketAdaptationField) []byte {
	w := astibinary.New()
	w.Write(uint8(36))                // Length
	w.Write(a.DiscontinuityIndicator) // Discontinuity indicator
	w.Write("1")                      // Random access indicator
	w.Write("1")                      // Elementary stream priority indicator
	w.Write("1")                      // PCR flag
	w.Write("1")                      // OPCR flag
	w.Write("1")                      // Splicing point flag
	w.Write("1")                      // Transport data flag
	w.Write("1")                      // Adaptation field extension flag
	w.Write(pcrBytes())               // PCR
	w.Write(pcrBytes())               // OPCR
	w.Write(uint8(2))                 // Splice countdown
	w.Write(uint8(4))                 // Transport private data length
	w.Write([]byte("test"))           // Transport private data
	w.Write(uint8(11))                // Adaptation extension length
	w.Write("1")                      // LTW flag
	w.Write("1")                      // Piecewise rate flag
	w.Write("1")                      // Seamless splice flag
	w.Write("11111")                  // Reserved
	w.Write("1")                      // LTW valid flag
	w.Write("010101010101010")        // LTW offset
	w.Write("11")                     // Piecewise rate reserved
	w.Write("1010101010101010101010") // Piecewise rate
	w.Write(dtsBytes())               // Splice type + DTS next access unit
	w.Write([]byte("stuff"))          // Stuffing bytes
	return w.Bytes()
}

func TestParsePacketAdaptationField(t *testing.T) {
	assert.Equal(t, packetAdaptationField, parsePacketAdaptationField(packetAdaptationFieldBytes(*packetAdaptationField)))
}

var pcr = &ClockReference{
	Base:      5726623061,
	Extension: 341,
}

func pcrBytes() []byte {
	w := astibinary.New()
	w.Write("101010101010101010101010101010101") // Base
	w.Write("111111")                            // Reserved
	w.Write("101010101")                         // Extension
	return w.Bytes()
}

func TestParsePCR(t *testing.T) {
	assert.Equal(t, pcr, parsePCR(pcrBytes()))
}
