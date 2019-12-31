package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func TestHasPESOptionalHeader(t *testing.T) {
	var a []int
	for i := 0; i <= 255; i++ {
		if !hasPESOptionalHeader(uint8(i)) {
			a = append(a, i)
		}
	}
	assert.Equal(t, []int{StreamIDPaddingStream, StreamIDPrivateStream2}, a)
}

var dsmTrickModeSlow = &DSMTrickMode{
	RepeatControl:    21,
	TrickModeControl: TrickModeControlSlowMotion,
}

func dsmTrickModeSlowBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("001")   // Control
	w.Write("10101") // Repeat control
	return buf.Bytes()
}

func TestParseDSMTrickMode(t *testing.T) {
	// Fast
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("011") // Control
	w.Write("10")  // Field ID
	w.Write("1")   // Intra slice refresh
	w.Write("11")  // Frequency truncation
	assert.Equal(t, parseDSMTrickMode(buf.Bytes()[0]), &DSMTrickMode{
		FieldID:             2,
		FrequencyTruncation: 3,
		IntraSliceRefresh:   1,
		TrickModeControl:    TrickModeControlFastReverse,
	})

	// Freeze
	buf.Reset()
	w.Write("010") // Control
	w.Write("10")  // Field ID
	w.Write("000") // Reserved
	assert.Equal(t, parseDSMTrickMode(buf.Bytes()[0]), &DSMTrickMode{
		FieldID:          2,
		TrickModeControl: TrickModeControlFreezeFrame,
	})

	// Slow
	assert.Equal(t, parseDSMTrickMode(dsmTrickModeSlowBytes()[0]), dsmTrickModeSlow)
}

var ptsClockReference = &ClockReference{Base: 5726623061}

func ptsBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("0010")            // Flag
	w.Write("101")             // 32...30
	w.Write("0")               // Dummy
	w.Write("010101010101010") // 29...15
	w.Write("0")               // Dummy
	w.Write("101010101010101") // 14...0
	w.Write("0")               // Dummy
	return buf.Bytes()
}

var dtsClockReference = &ClockReference{Base: 5726623060}

func dtsBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("0010")            // Flag
	w.Write("101")             // 32...30
	w.Write("0")               // Dummy
	w.Write("010101010101010") // 29...15
	w.Write("0")               // Dummy
	w.Write("101010101010100") // 14...0
	w.Write("0")               // Dummy
	return buf.Bytes()
}

func TestParsePTSOrDTS(t *testing.T) {
	v, err := parsePTSOrDTS(astikit.NewBytesIterator(ptsBytes()))
	assert.Equal(t, v, ptsClockReference)
	assert.NoError(t, err)
}

func escrBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("00")              // Dummy
	w.Write("011")             // 32...30
	w.Write("1")               // Dummy
	w.Write("000010111110000") // 29...15
	w.Write("1")               // Dummy
	w.Write("000010111001111") // 14...0
	w.Write("1")               // Dummy
	w.Write("000111010")       // Ext
	w.Write("1")               // Dummy
	return buf.Bytes()
}

func TestParseESCR(t *testing.T) {
	v, err := parseESCR(astikit.NewBytesIterator(escrBytes()))
	assert.Equal(t, v, clockReference)
	assert.NoError(t, err)
}

var pesWithoutHeader = &PESData{
	Data: []byte("data"),
	Header: &PESHeader{
		PacketLength: 4,
		StreamID:     StreamIDPaddingStream,
	},
}

func pesWithoutHeaderBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("000000000000000000000001")   // Prefix
	w.Write(uint8(StreamIDPaddingStream)) // Stream ID
	w.Write(uint16(4))                    // Packet length
	w.Write([]byte("datastuff"))          // Data
	return buf.Bytes()
}

var pesWithHeader = &PESData{
	Data: []byte("data"),
	Header: &PESHeader{
		OptionalHeader: &PESOptionalHeader{
			AdditionalCopyInfo:              127,
			CRC:                             4,
			DataAlignmentIndicator:          true,
			DSMTrickMode:                    dsmTrickModeSlow,
			DTS:                             dtsClockReference,
			ESCR:                            clockReference,
			ESRate:                          1398101,
			Extension2Data:                  []byte("extension2"),
			Extension2Length:                10,
			HasAdditionalCopyInfo:           true,
			HasCRC:                          true,
			HasDSMTrickMode:                 true,
			HasESCR:                         true,
			HasESRate:                       true,
			HasExtension:                    true,
			HasExtension2:                   true,
			HasPackHeaderField:              true,
			HasPrivateData:                  true,
			HasProgramPacketSequenceCounter: true,
			HasPSTDBuffer:                   true,
			HeaderLength:                    62,
			IsCopyrighted:                   true,
			IsOriginal:                      true,
			MarkerBits:                      2,
			MPEG1OrMPEG2ID:                  1,
			OriginalStuffingLength:          21,
			PacketSequenceCounter:           85,
			PackField:                       5,
			Priority:                        true,
			PrivateData:                     []byte("1234567890123456"),
			PSTDBufferScale:                 1,
			PSTDBufferSize:                  5461,
			PTSDTSIndicator:                 3,
			PTS:                             ptsClockReference,
			ScramblingControl:               1,
		},
		PacketLength: 69,
		StreamID:     1,
	},
}

func pesWithHeaderBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("000000000000000000000001") // Prefix
	w.Write(uint8(1))                   // Stream ID
	w.Write(uint16(69))                 // Packet length
	w.Write("10")                       // Marker bits
	w.Write("01")                       // Scrambling control
	w.Write("1")                        // Priority
	w.Write("1")                        // Data alignment indicator
	w.Write("1")                        // Copyright
	w.Write("1")                        // Original or copy
	w.Write("11")                       // PTS/DTS indicator
	w.Write("1")                        // ESCR flag
	w.Write("1")                        // ES rate flag
	w.Write("1")                        // DSM trick mode flag
	w.Write("1")                        // Additional copy flag
	w.Write("1")                        // CRC flag
	w.Write("1")                        // Extension flag
	w.Write(uint8(62))                  // Header length
	w.Write(ptsBytes())                 // PTS
	w.Write(dtsBytes())                 // DTS
	w.Write(escrBytes())                // ESCR
	w.Write("101010101010101010101010") // ES rate
	w.Write(dsmTrickModeSlowBytes())    // DSM trick mode
	w.Write("11111111")                 // Additional copy info
	w.Write(uint16(4))                  // CRC
	w.Write("1")                        // Private data flag
	w.Write("1")                        // Pack header field flag
	w.Write("1")                        // Program packet sequence counter flag
	w.Write("1")                        // PSTD buffer flag
	w.Write("000")                      // Dummy
	w.Write("1")                        // Extension 2 flag
	w.Write([]byte("1234567890123456")) // Private data
	w.Write(uint8(5))                   // Pack field
	w.Write("0101010101010101")         // Packet sequence counter
	w.Write("0111010101010101")         // PSTD buffer
	w.Write("0000101000000000")         // Extension 2 header
	w.Write([]byte("extension2"))       // Extension 2 data
	w.Write([]byte("stuff"))            // Optional header stuffing bytes
	w.Write([]byte("datastuff"))        // Data
	return buf.Bytes()
}

func TestParsePESData(t *testing.T) {
	// No optional header and specific packet length
	d, err := parsePESData(astikit.NewBytesIterator(pesWithoutHeaderBytes()))
	assert.NoError(t, err)
	assert.Equal(t, pesWithoutHeader, d)

	// Optional header and no specific header length
	d, err = parsePESData(astikit.NewBytesIterator(pesWithHeaderBytes()))
	assert.NoError(t, err)
	assert.Equal(t, pesWithHeader, d)
}
