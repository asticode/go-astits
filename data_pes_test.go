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

type dsmTrickModeTestCase struct {
	name      string
	bytesFunc func(w *astikit.BitsWriter)
	trickMode *DSMTrickMode
}

var dsmTrickModeTestCases = []dsmTrickModeTestCase{
	{
		"fast_forward",
		func(w *astikit.BitsWriter) {
			w.Write("000") // Control
			w.Write("10")  // Field ID
			w.Write("1")   // Intra slice refresh
			w.Write("11")  // Frequency truncation
		},
		&DSMTrickMode{
			FieldID:             2,
			FrequencyTruncation: 3,
			IntraSliceRefresh:   1,
			TrickModeControl:    TrickModeControlFastForward,
		},
	},
	{
		"slow_motion",
		func(w *astikit.BitsWriter) {
			w.Write("001")
			w.Write("10101")
		},
		&DSMTrickMode{
			RepeatControl:    0b10101,
			TrickModeControl: TrickModeControlSlowMotion,
		},
	},
	{
		"freeze_frame",
		func(w *astikit.BitsWriter) {
			w.Write("010") // Control
			w.Write("10")  // Field ID
			w.Write("111") // Reserved
		},
		&DSMTrickMode{
			FieldID:          2,
			TrickModeControl: TrickModeControlFreezeFrame,
		},
	},
	{
		"fast_reverse",
		func(w *astikit.BitsWriter) {
			w.Write("011") // Control
			w.Write("10")  // Field ID
			w.Write("1")   // Intra slice refresh
			w.Write("11")  // Frequency truncation
		},
		&DSMTrickMode{
			FieldID:             2,
			FrequencyTruncation: 3,
			IntraSliceRefresh:   1,
			TrickModeControl:    TrickModeControlFastReverse,
		},
	},
	{
		"slow_reverse",
		func(w *astikit.BitsWriter) {
			w.Write("100")
			w.Write("01010")
		},
		&DSMTrickMode{
			RepeatControl:    0b01010,
			TrickModeControl: TrickModeControlSlowReverse,
		},
	},
	{
		"reserved",
		func(w *astikit.BitsWriter) {
			w.Write("101")
			w.Write("11111")
		},
		&DSMTrickMode{
			TrickModeControl: 5, // reserved
		},
	},
}

func TestParseDSMTrickMode(t *testing.T) {
	for _, tc := range dsmTrickModeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
			tc.bytesFunc(w)
			assert.Equal(t, parseDSMTrickMode(buf.Bytes()[0]), tc.trickMode)
		})
	}
}

func TestWriteDSMTrickMode(t *testing.T) {
	for _, tc := range dsmTrickModeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := &bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: bufExpected})
			tc.bytesFunc(wExpected)

			bufActual := &bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: bufActual})

			n, err := writeDSMTrickMode(wActual, tc.trickMode)
			assert.NoError(t, err)
			assert.Equal(t, 1, n)
			assert.Equal(t, n, bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

var ptsClockReference = &ClockReference{Base: 5726623061}

func ptsBytes(flag string) []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(flag)              // Flag
	w.Write("101")             // 32...30
	w.Write("1")               // Dummy
	w.Write("010101010101010") // 29...15
	w.Write("1")               // Dummy
	w.Write("101010101010101") // 14...0
	w.Write("1")               // Dummy
	return buf.Bytes()
}

var dtsClockReference = &ClockReference{Base: 5726623060}

func dtsBytes(flag string) []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(flag)              // Flag
	w.Write("101")             // 32...30
	w.Write("1")               // Dummy
	w.Write("010101010101010") // 29...15
	w.Write("1")               // Dummy
	w.Write("101010101010100") // 14...0
	w.Write("1")               // Dummy
	return buf.Bytes()
}

func TestParsePTSOrDTS(t *testing.T) {
	v, err := parsePTSOrDTS(astikit.NewBytesIterator(ptsBytes("0010")))
	assert.Equal(t, v, ptsClockReference)
	assert.NoError(t, err)
}

func TestWritePTSOrDTS(t *testing.T) {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	n, err := writePTSOrDTS(w, uint8(0b0010), dtsClockReference)
	assert.NoError(t, err)
	assert.Equal(t, n, 5)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, dtsBytes("0010"), buf.Bytes())
}

func escrBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write("11")              // Dummy
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

func TestWriteESCR(t *testing.T) {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	n, err := writeESCR(w, clockReference)
	assert.NoError(t, err)
	assert.Equal(t, n, 6)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, escrBytes(), buf.Bytes())
}

type pesTestCase struct {
	name                    string
	headerBytesFunc         func(w *astikit.BitsWriter, withStuffing bool, withCRC bool)
	optionalHeaderBytesFunc func(w *astikit.BitsWriter, withStuffing bool, withCRC bool)
	bytesFunc               func(w *astikit.BitsWriter, withStuffing bool, withCRC bool)
	pesData                 *PESData
}

var pesTestCases = []pesTestCase{
	{
		"without_header",
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			w.Write("000000000000000000000001")   // Prefix
			w.Write(uint8(StreamIDPaddingStream)) // Stream ID
			w.Write(uint16(4))                    // Packet length
		},
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			// do nothing here
		},
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			w.Write([]byte("data")) // Data
		},
		&PESData{
			Data: []byte("data"),
			Header: &PESHeader{
				PacketLength: 4,
				StreamID:     StreamIDPaddingStream,
			},
		},
	},
	{
		"with_header",
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			packetLength := 67
			stuffing := []byte("stuff")

			if !withStuffing {
				packetLength -= len(stuffing)
			}

			if !withCRC {
				packetLength -= 2
			}

			w.Write("000000000000000000000001") // Prefix
			w.Write(uint8(1))                   // Stream ID
			w.Write(uint16(packetLength))       // Packet length

		},
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			optionalHeaderLength := 60
			stuffing := []byte("stuff")

			if !withStuffing {
				optionalHeaderLength -= len(stuffing)
			}

			if !withCRC {
				optionalHeaderLength -= 2
			}

			w.Write("10")                        // Marker bits
			w.Write("01")                        // Scrambling control
			w.Write("1")                         // Priority
			w.Write("1")                         // Data alignment indicator
			w.Write("1")                         // Copyright
			w.Write("1")                         // Original or copy
			w.Write("11")                        // PTS/DTS indicator
			w.Write("1")                         // ESCR flag
			w.Write("1")                         // ES rate flag
			w.Write("1")                         // DSM trick mode flag
			w.Write("1")                         // Additional copy flag
			w.Write(withCRC)                     // CRC flag
			w.Write("1")                         // Extension flag
			w.Write(uint8(optionalHeaderLength)) // Header length
			w.Write(ptsBytes("0011"))            // PTS
			w.Write(dtsBytes("0001"))            // DTS
			w.Write(escrBytes())                 // ESCR
			w.Write("101010101010101010101011")  // ES rate
			w.Write(dsmTrickModeSlowBytes())     // DSM trick mode
			w.Write("11111111")                  // Additional copy info
			if withCRC {
				w.Write(uint16(4)) // CRC
			}
			// Extension starts here
			w.Write("1")                        // Private data flag
			w.Write("0")                        // Pack header field flag
			w.Write("1")                        // Program packet sequence counter flag
			w.Write("1")                        // PSTD buffer flag
			w.Write("111")                      // Dummy
			w.Write("1")                        // Extension 2 flag
			w.Write([]byte("1234567890123456")) // Private data
			//w.Write(uint8(5))                   // Pack field
			w.Write("1101010111010101")   // Packet sequence counter
			w.Write("0111010101010101")   // PSTD buffer
			w.Write("10001010")           // Extension 2 header
			w.Write([]byte("extension2")) // Extension 2 data
			if withStuffing {
				w.Write(stuffing) // Optional header stuffing bytes
			}
		},
		func(w *astikit.BitsWriter, withStuffing bool, withCRC bool) {
			stuffing := []byte("stuff")
			w.Write([]byte("data")) // Data
			if withStuffing {
				w.Write(stuffing) // Stuffing
			}
		},
		&PESData{
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
					HasPackHeaderField:              false,
					HasPrivateData:                  true,
					HasProgramPacketSequenceCounter: true,
					HasPSTDBuffer:                   true,
					HeaderLength:                    60,
					IsCopyrighted:                   true,
					IsOriginal:                      true,
					MarkerBits:                      2,
					MPEG1OrMPEG2ID:                  1,
					OriginalStuffingLength:          21,
					PacketSequenceCounter:           85,
					//PackField:                       5,
					Priority:          true,
					PrivateData:       []byte("1234567890123456"),
					PSTDBufferScale:   1,
					PSTDBufferSize:    5461,
					PTSDTSIndicator:   3,
					PTS:               ptsClockReference,
					ScramblingControl: 1,
				},
				PacketLength: 67,
				StreamID:     1,
			},
		},
	},
}

// used by TestParseData
func pesWithHeaderBytes() []byte {
	buf := bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
	pesTestCases[1].headerBytesFunc(w, true, true)
	pesTestCases[1].optionalHeaderBytesFunc(w, true, true)
	pesTestCases[1].bytesFunc(w, true, true)
	return buf.Bytes()
}

// used by TestParseData
func pesWithHeader() *PESData {
	return pesTestCases[1].pesData
}

func TestParsePESData(t *testing.T) {
	for _, tc := range pesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
			tc.headerBytesFunc(w, true, true)
			tc.optionalHeaderBytesFunc(w, true, true)
			tc.bytesFunc(w, true, true)
			d, err := parsePESData(astikit.NewBytesIterator(buf.Bytes()))
			assert.NoError(t, err)
			assert.Equal(t, tc.pesData, d)
		})
	}
}

func TestWritePESData(t *testing.T) {
	for _, tc := range pesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})
			tc.headerBytesFunc(wExpected, false, false)
			tc.optionalHeaderBytesFunc(wExpected, false, false)
			tc.bytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

			start := true
			totalBytes := 0
			payloadPos := 0

			for payloadPos+1 < len(tc.pesData.Data) {
				n, payloadN, err := writePESData(
					wActual,
					tc.pesData.Header,
					tc.pesData.Data[payloadPos:],
					start,
					MpegTsPacketSize-mpegTsPacketHeaderSize,
				)
				assert.NoError(t, err)
				start = false

				totalBytes += n
				payloadPos += payloadN
			}

			assert.Equal(t, totalBytes, bufActual.Len())
			assert.Equal(t, bufExpected.Len(), bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

func TestWritePESHeader(t *testing.T) {
	for _, tc := range pesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})
			tc.headerBytesFunc(wExpected, false, false)
			tc.optionalHeaderBytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

			n, err := writePESHeader(wActual, tc.pesData.Header, len(tc.pesData.Data))
			assert.NoError(t, err)
			assert.Equal(t, n, bufActual.Len())
			assert.Equal(t, bufExpected.Len(), bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

func TestWritePESOptionalHeader(t *testing.T) {
	for _, tc := range pesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})
			tc.optionalHeaderBytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

			n, err := writePESOptionalHeader(wActual, tc.pesData.Header.OptionalHeader)
			assert.NoError(t, err)
			assert.Equal(t, n, bufActual.Len())
			assert.Equal(t, bufExpected.Len(), bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

func BenchmarkParsePESData(b *testing.B) {
	bss := make([][]byte, len(pesTestCases))

	for ti, tc := range pesTestCases {
		buf := bytes.Buffer{}
		w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
		tc.headerBytesFunc(w, true, true)
		tc.optionalHeaderBytesFunc(w, true, true)
		tc.bytesFunc(w, true, true)
		bss[ti] = buf.Bytes()
	}

	for ti, tc := range pesTestCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				parsePESData(astikit.NewBytesIterator(bss[ti]))
			}
		})
	}
}
