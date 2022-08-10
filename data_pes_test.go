package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
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
	w := bitio.NewWriter(buf)
	WriteBinary(w, "001")   // Control
	WriteBinary(w, "10101") // Repeat control
	return buf.Bytes()
}

type dsmTrickModeTestCase struct {
	name      string
	bytesFunc func(w *bitio.Writer)
	trickMode *DSMTrickMode
}

var dsmTrickModeTestCases = []dsmTrickModeTestCase{
	{
		"fast_forward",
		func(w *bitio.Writer) {
			WriteBinary(w, "000") // Control
			WriteBinary(w, "10")  // Field ID
			WriteBinary(w, "1")   // Intra slice refresh
			WriteBinary(w, "11")  // Frequency truncation
		},
		&DSMTrickMode{
			FieldID:             2,
			FrequencyTruncation: 3,
			IntraSliceRefresh:   true,
			TrickModeControl:    TrickModeControlFastForward,
		},
	},
	{
		"slow_motion",
		func(w *bitio.Writer) {
			WriteBinary(w, "001")
			WriteBinary(w, "10101")
		},
		&DSMTrickMode{
			RepeatControl:    0b10101,
			TrickModeControl: TrickModeControlSlowMotion,
		},
	},
	{
		"freeze_frame",
		func(w *bitio.Writer) {
			WriteBinary(w, "010") // Control
			WriteBinary(w, "10")  // Field ID
			WriteBinary(w, "111") // Reserved
		},
		&DSMTrickMode{
			FieldID:          2,
			TrickModeControl: TrickModeControlFreezeFrame,
		},
	},
	{
		"fast_reverse",
		func(w *bitio.Writer) {
			WriteBinary(w, "011") // Control
			WriteBinary(w, "10")  // Field ID
			WriteBinary(w, "1")   // Intra slice refresh
			WriteBinary(w, "11")  // Frequency truncation
		},
		&DSMTrickMode{
			FieldID:             2,
			FrequencyTruncation: 3,
			IntraSliceRefresh:   true,
			TrickModeControl:    TrickModeControlFastReverse,
		},
	},
	{
		"slow_reverse",
		func(w *bitio.Writer) {
			WriteBinary(w, "100")
			WriteBinary(w, "01010")
		},
		&DSMTrickMode{
			RepeatControl:    0b01010,
			TrickModeControl: TrickModeControlSlowReverse,
		},
	},
	{
		"reserved",
		func(w *bitio.Writer) {
			WriteBinary(w, "101")
			WriteBinary(w, "11111")
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
			w := bitio.NewWriter(buf)
			tc.bytesFunc(w)
			r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
			trickMode, err := parseDSMTrickMode(r)
			assert.NoError(t, err)
			assert.Equal(t, trickMode, tc.trickMode)
		})
	}
}

func TestWriteDSMTrickMode(t *testing.T) {
	for _, tc := range dsmTrickModeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := &bytes.Buffer{}
			wExpected := bitio.NewWriter(bufExpected)
			tc.bytesFunc(wExpected)

			bufActual := &bytes.Buffer{}
			wActual := bitio.NewWriter(bufActual)

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
	w := bitio.NewWriter(buf)
	WriteBinary(w, flag)              // Flag
	WriteBinary(w, "101")             // 32...30
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "010101010101010") // 29...15
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "101010101010101") // 14...0
	WriteBinary(w, "1")               // Dummy
	return buf.Bytes()
}

var dtsClockReference = &ClockReference{Base: 5726623060}

func dtsBytes(flag string) []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	WriteBinary(w, flag)              // Flag
	WriteBinary(w, "101")             // 32...30
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "010101010101010") // 29...15
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "101010101010100") // 14...0
	WriteBinary(w, "1")               // Dummy
	return buf.Bytes()
}

func TestParsePTSOrDTS(t *testing.T) {
	r := bitio.NewCountReader(bytes.NewReader(ptsBytes("0010")))
	_, err := r.ReadBits(4)
	assert.NoError(t, err)
	v, err := parsePTSOrDTS(r)
	assert.Equal(t, v, ptsClockReference)
	assert.NoError(t, err)
}

func TestWritePTSOrDTS(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	n, err := writePTSOrDTS(w, uint8(0b0010), dtsClockReference)
	assert.NoError(t, err)
	assert.Equal(t, n, 5)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, dtsBytes("0010"), buf.Bytes())
}

func escrBytes() []byte {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	WriteBinary(w, "11")              // Dummy
	WriteBinary(w, "011")             // 32...30
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "000010111110000") // 29...15
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "000010111001111") // 14...0
	WriteBinary(w, "1")               // Dummy
	WriteBinary(w, "000111010")       // Ext
	WriteBinary(w, "1")               // Dummy
	return buf.Bytes()
}

func TestParseESCR(t *testing.T) {
	r := bitio.NewCountReader(bytes.NewReader(escrBytes()))
	v, err := parseESCR(r)
	assert.Equal(t, v, clockReference)
	assert.NoError(t, err)
}

func TestWriteESCR(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	n, err := writeESCR(w, clockReference)
	assert.NoError(t, err)
	assert.Equal(t, n, 6)
	assert.Equal(t, n, buf.Len())
	assert.Equal(t, escrBytes(), buf.Bytes())
}

type pesTestCase struct {
	name                    string
	headerBytesFunc         func(w *bitio.Writer, withStuffing bool, withCRC bool)
	optionalHeaderBytesFunc func(w *bitio.Writer, withStuffing bool, withCRC bool)
	bytesFunc               func(w *bitio.Writer, withStuffing bool, withCRC bool)
	pesData                 *PESData
}

var pesTestCases = []pesTestCase{
	{
		"without_header",
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
			WriteBinary(w, "000000000000000000000001") // Prefix
			w.WriteByte(StreamIDPaddingStream)         // Stream ID
			w.WriteBits(4, 16)                         // Packet length
		},
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
			// do nothing here
		},
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
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
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
			packetLength := 67
			stuffing := []byte("stuff")

			if !withStuffing {
				packetLength -= len(stuffing)
			}

			if !withCRC {
				packetLength -= 2
			}

			WriteBinary(w, "000000000000000000000001") // Prefix
			w.WriteByte(1)                             // Stream ID
			w.WriteBits(uint64(packetLength), 16)      // Packet length
		},
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
			optionalHeaderLength := 60
			stuffing := []byte("stuff")

			if !withStuffing {
				optionalHeaderLength -= len(stuffing)
			}

			if !withCRC {
				optionalHeaderLength -= 2
			}

			WriteBinary(w, "10")                       // Marker bits
			WriteBinary(w, "01")                       // Scrambling control
			WriteBinary(w, "1")                        // Priority
			WriteBinary(w, "1")                        // Data alignment indicator
			WriteBinary(w, "1")                        // Copyright
			WriteBinary(w, "1")                        // Original or copy
			WriteBinary(w, "11")                       // PTS/DTS indicator
			WriteBinary(w, "1")                        // ESCR flag
			WriteBinary(w, "1")                        // ES rate flag
			WriteBinary(w, "1")                        // DSM trick mode flag
			WriteBinary(w, "1")                        // Additional copy flag
			w.WriteBool(withCRC)                       // CRC flag
			WriteBinary(w, "1")                        // Extension flag
			w.WriteByte(uint8(optionalHeaderLength))   // Header length
			w.Write(ptsBytes("0011"))                  // PTS
			w.Write(dtsBytes("0001"))                  // DTS
			w.Write(escrBytes())                       // ESCR
			WriteBinary(w, "101010101010101010101011") // ES rate
			w.Write(dsmTrickModeSlowBytes())           // DSM trick mode
			WriteBinary(w, "11111111")                 // Additional copy info
			if withCRC {
				w.WriteBits(4, 16) // CRC
			}
			// Extension starts here
			WriteBinary(w, "1")                 // Private data flag
			WriteBinary(w, "0")                 // Pack header field flag
			WriteBinary(w, "1")                 // Program packet sequence counter flag
			WriteBinary(w, "1")                 // PSTD buffer flag
			WriteBinary(w, "111")               // Dummy
			WriteBinary(w, "1")                 // Extension 2 flag
			w.Write([]byte("1234567890123456")) // Private data
			// w.WriteByte(uint8(5))                   // Pack field
			WriteBinary(w, "1101010111010101") // Packet sequence counter
			WriteBinary(w, "0111010101010101") // PSTD buffer
			WriteBinary(w, "10001010")         // Extension 2 header
			w.Write([]byte("extension2"))      // Extension 2 data
			if withStuffing {
				w.Write(stuffing) // Optional header stuffing bytes
			}
		},
		func(w *bitio.Writer, withStuffing bool, withCRC bool) {
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
					MPEG1OrMPEG2ID:                  true,
					OriginalStuffingLength:          21,
					PacketSequenceCounter:           85,
					// PackField:                       5,
					Priority:          true,
					PrivateData:       []byte("1234567890123456"),
					PSTDBufferScale:   true,
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
	w := bitio.NewWriter(&buf)
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
			w := bitio.NewWriter(&buf)
			tc.headerBytesFunc(w, true, true)
			tc.optionalHeaderBytesFunc(w, true, true)
			tc.bytesFunc(w, true, true)
			r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
			d, err := parsePESData(r, int64(len(buf.Bytes())*8))
			assert.NoError(t, err)
			assert.Equal(t, tc.pesData, d)
		})
	}
}

func TestWritePESData(t *testing.T) {
	for _, tc := range pesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := bitio.NewWriter(&bufExpected)
			tc.headerBytesFunc(wExpected, false, false)
			tc.optionalHeaderBytesFunc(wExpected, false, false)
			tc.bytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := bitio.NewWriter(&bufActual)

			start := true
			var totalBytes int
			var payloadPos int

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
			wExpected := bitio.NewWriter(&bufExpected)
			tc.headerBytesFunc(wExpected, false, false)
			tc.optionalHeaderBytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := bitio.NewWriter(&bufActual)

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
			wExpected := bitio.NewWriter(&bufExpected)
			tc.optionalHeaderBytesFunc(wExpected, false, false)

			bufActual := bytes.Buffer{}
			wActual := bitio.NewWriter(&bufActual)

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
		w := bitio.NewWriter(&buf)
		tc.headerBytesFunc(w, true, true)
		tc.optionalHeaderBytesFunc(w, true, true)
		tc.bytesFunc(w, true, true)
		bss[ti] = buf.Bytes()
	}

	for ti, tc := range pesTestCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r := bitio.NewCountReader(bytes.NewReader(bss[ti]))
				parsePESData(r, int64(len(bss[ti])*8))
			}
		})
	}
}
