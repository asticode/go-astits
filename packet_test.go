package astits

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func packet(h PacketHeader, a PacketAdaptationField, i []byte) ([]byte, *Packet) {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(syncByte))                             // Sync byte
	w.Write([]byte("test"))                              // Sometimes packets are 192 bytes
	w.Write(packetHeaderBytes(h))                        // Header
	w.Write(packetAdaptationFieldBytes(a))               // Adaptation field
	var payload = append(i, make([]byte, 147-len(i))...) // Payload
	w.Write(payload)
	return buf.Bytes(), &Packet{
		AdaptationField: packetAdaptationField,
		Header:          packetHeader,
		Payload:         payload,
	}
}

func TestParsePacket(t *testing.T) {
	// Packet not starting with a sync
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint16(1)) // Invalid sync byte
	_, err := parsePacket(astikit.NewBytesIterator(buf.Bytes()))
	assert.EqualError(t, err, ErrPacketMustStartWithASyncByte.Error())

	// Valid
	b, ep := packet(*packetHeader, *packetAdaptationField, []byte("payload"))
	p, err := parsePacket(astikit.NewBytesIterator(b))
	assert.NoError(t, err)
	assert.Equal(t, p, ep)
}

func TestPayloadOffset(t *testing.T) {
	assert.Equal(t, 3, payloadOffset(0, &PacketHeader{}, nil))
	assert.Equal(t, 7, payloadOffset(1, &PacketHeader{HasAdaptationField: true}, &PacketAdaptationField{Length: 2}))
}

var packetHeader = &PacketHeader{
	ContinuityCounter:          10,
	HasAdaptationField:         true,
	HasPayload:                 true,
	PayloadUnitStartIndicator:  true,
	PID:                        5461,
	TransportErrorIndicator:    true,
	TransportPriority:          true,
	TransportScramblingControl: ScramblingControlScrambledWithEvenKey,
}

func packetHeaderBytes(h PacketHeader) []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(h.TransportErrorIndicator)                // Transport error indicator
	w.Write(h.PayloadUnitStartIndicator)              // Payload unit start indicator
	w.Write("1")                                      // Transport priority
	w.Write(fmt.Sprintf("%.13b", h.PID))              // PID
	w.Write("10")                                     // Scrambling control
	w.Write("11")                                     // Adaptation field control
	w.Write(fmt.Sprintf("%.4b", h.ContinuityCounter)) // Continuity counter
	return buf.Bytes()
}

func TestParsePacketHeader(t *testing.T) {
	v, err := parsePacketHeader(astikit.NewBytesIterator(packetHeaderBytes(*packetHeader)))
	assert.Equal(t, packetHeader, v)
	assert.NoError(t, err)
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
	RandomAccessIndicator:             true,
	SpliceCountdown:                   2,
	TransportPrivateDataLength:        4,
	TransportPrivateData:              []byte("test"),
}

func packetAdaptationFieldBytes(a PacketAdaptationField) []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
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
	return buf.Bytes()
}

func TestParsePacketAdaptationField(t *testing.T) {
	v, err := parsePacketAdaptationField(astikit.NewBytesIterator(packetAdaptationFieldBytes(*packetAdaptationField)))
	assert.Equal(t, packetAdaptationField, v)
	assert.NoError(t, err)
}

var pcr = &ClockReference{
	Base:      5726623061,
	Extension: 341,
}

func pcrBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("101010101010101010101010101010101") // Base
	w.Write("111111")                            // Reserved
	w.Write("101010101")                         // Extension
	return buf.Bytes()
}

func TestParsePCR(t *testing.T) {
	v, err := parsePCR(astikit.NewBytesIterator(pcrBytes()))
	assert.Equal(t, pcr, v)
	assert.NoError(t, err)
}
