package astits

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

func packet(h PacketHeader, a PacketAdaptationField, i []byte, packet192bytes bool) ([]byte, *Packet) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(uint8(syncByte)) // Sync byte
	if packet192bytes {
		w.Write([]byte("test")) // Sometimes packets are 192 bytes
	}
	w.Write(packetHeaderBytes(h, "11"))                          // Header
	w.Write(packetAdaptationFieldBytes(a))                       // Adaptation field
	payload := append(i, bytes.Repeat([]byte{0}, 147-len(i))...) // Payload
	w.Write(payload)
	return buf.Bytes(), &Packet{
		AdaptationField: packetAdaptationField,
		Header:          packetHeader,
		Payload:         payload,
	}
}

func packetShort(h PacketHeader, payload []byte) ([]byte, *Packet) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(uint8(syncByte))        // Sync byte
	w.Write(packetHeaderBytes(h, "01")) // Header
	p := append(payload, bytes.Repeat([]byte{0}, MpegTsPacketSize-buf.Len())...)
	w.Write(p)
	return buf.Bytes(), &Packet{
		Header:  &h,
		Payload: payload,
	}
}

func TestParsePacket(t *testing.T) {
	// Packet not starting with a sync
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteBits(1, 16) // Invalid sync byte
	r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
	_, err := parsePacket(r, int64(len(buf.Bytes())*8))
	assert.ErrorIs(t, err, ErrPacketStartSyncByte)

	// Valid
	b, ep := packet(*packetHeader, *packetAdaptationField, []byte("payload"), true)
	r = bitio.NewCountReader(bytes.NewReader(b))
	p, err := parsePacket(r, int64(len(b)*8))
	assert.NoError(t, err)
	assert.Equal(t, p, ep)
}

func TestWritePacket(t *testing.T) {
	eb, ep := packet(*packetHeader, *packetAdaptationField, []byte("payload"), false)
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	n, err := writePacket(w, ep, MpegTsPacketSize)
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, n)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, len(eb), buf.Len())
	assert.Equal(t, eb, buf.Bytes())
}

func TestWritePacket_HeaderOnly(t *testing.T) {
	shortPacketHeader := *packetHeader
	shortPacketHeader.HasPayload = false
	shortPacketHeader.HasAdaptationField = false
	_, ep := packetShort(shortPacketHeader, nil)

	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)

	n, err := writePacket(w, ep, MpegTsPacketSize)
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, n)
	assert.Equal(t, n, buf.Len())

	// we can't just compare bytes returned by packetShort since they're not completely correct,
	//  so we just cross-check writePacket with parsePacket
	r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
	p, err := parsePacket(r, int64(len(buf.Bytes())*8))
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
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

func packetHeaderBytes(h PacketHeader, afControl string) []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteBool(h.TransportErrorIndicator)                   // Transport error indicator
	w.WriteBool(h.PayloadUnitStartIndicator)                 // Payload unit start indicator
	WriteBinary(w, "1")                                      // Transport priority
	WriteBinary(w, fmt.Sprintf("%.13b", h.PID))              // PID
	WriteBinary(w, "10")                                     // Scrambling control
	WriteBinary(w, afControl)                                // Adaptation field control
	WriteBinary(w, fmt.Sprintf("%.4b", h.ContinuityCounter)) // Continuity counter
	return buf.Bytes()
}

func TestParsePacketHeader(t *testing.T) {
	bs := packetHeaderBytes(*packetHeader, "11")
	r := bitio.NewCountReader(bytes.NewReader(bs))
	v, err := parsePacketHeader(r)
	assert.Equal(t, packetHeader, v)
	assert.NoError(t, err)
}

func TestWritePacketHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	bytesWritten, err := writePacketHeader(w, packetHeader)
	assert.NoError(t, err)
	assert.Equal(t, bytesWritten, 3)
	assert.Equal(t, bytesWritten, buf.Len())
	assert.Equal(t, packetHeaderBytes(*packetHeader, "11"), buf.Bytes())
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
	StuffingLength:                    5,
}

func packetAdaptationFieldBytes(a PacketAdaptationField) []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(uint8(36))                   // Length
	w.WriteBool(a.DiscontinuityIndicator)    // Discontinuity indicator
	WriteBinary(w, "1")                      // Random access indicator
	WriteBinary(w, "1")                      // Elementary stream priority indicator
	WriteBinary(w, "1")                      // PCR flag
	WriteBinary(w, "1")                      // OPCR flag
	WriteBinary(w, "1")                      // Splicing point flag
	WriteBinary(w, "1")                      // Transport data flag
	WriteBinary(w, "1")                      // Adaptation field extension flag
	w.Write(pcrBytes())                      // PCR
	w.Write(pcrBytes())                      // OPCR
	w.WriteByte(uint8(2))                    // Splice countdown
	w.WriteByte(uint8(4))                    // Transport private data length
	w.Write([]byte("test"))                  // Transport private data
	w.WriteByte(uint8(11))                   // Adaptation extension length
	WriteBinary(w, "1")                      // LTW flag
	WriteBinary(w, "1")                      // Piecewise rate flag
	WriteBinary(w, "1")                      // Seamless splice flag
	WriteBinary(w, "11111")                  // Reserved
	WriteBinary(w, "1")                      // LTW valid flag
	WriteBinary(w, "010101010101010")        // LTW offset
	WriteBinary(w, "11")                     // Piecewise rate reserved
	WriteBinary(w, "1010101010101010101010") // Piecewise rate
	w.Write(dtsBytes("0010"))                // Splice type + DTS next access unit
	w.WriteBits(^uint64(0), 40)              // Stuffing bytes
	return buf.Bytes()
}

func TestParsePacketAdaptationField(t *testing.T) {
	bs := packetAdaptationFieldBytes(*packetAdaptationField)
	r := bitio.NewCountReader(bytes.NewReader(bs))
	v, err := parsePacketAdaptationField(r)
	assert.Equal(t, packetAdaptationField, v)
	assert.NoError(t, err)
}

func TestWritePacketAdaptationField(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	eb := packetAdaptationFieldBytes(*packetAdaptationField)
	bytesWritten, err := writePacketAdaptationField(w, packetAdaptationField)
	assert.NoError(t, err)
	assert.Equal(t, bytesWritten, buf.Len())
	assert.Equal(t, len(eb), buf.Len())
	assert.Equal(t, eb, buf.Bytes())
}

var pcr = &ClockReference{
	Base:      5726623061,
	Extension: 341,
}

func pcrBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	WriteBinary(w, "101010101010101010101010101010101") // Base
	WriteBinary(w, "111111")                            // Reserved
	WriteBinary(w, "101010101")                         // Extension
	return buf.Bytes()
}

func TestParsePCR(t *testing.T) {
	r := bitio.NewCountReader(bytes.NewReader(pcrBytes()))
	v, err := parsePCR(r)
	assert.Equal(t, pcr, v)
	assert.NoError(t, err)
}

func TestWritePCR(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	bytesWritten, err := writePCR(w, pcr)
	assert.NoError(t, err)
	assert.Equal(t, bytesWritten, 6)
	assert.Equal(t, bytesWritten, buf.Len())
	assert.Equal(t, pcrBytes(), buf.Bytes())
}

func BenchmarkWritePCR(b *testing.B) {
	buf := &bytes.Buffer{}
	buf.Grow(6)
	w := bitio.NewWriter(buf)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		writePCR(w, pcr)
	}
}

func BenchmarkParsePacket(b *testing.B) {
	bs, _ := packet(*packetHeader, *packetAdaptationField, []byte("payload"), true)

	for i := 0; i < b.N; i++ {
		b.ReportAllocs()
		r := bitio.NewCountReader(bytes.NewReader(bs))
		parsePacket(r, int64(len(bs)*8))
	}
}
