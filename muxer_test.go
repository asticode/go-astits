package astits

import (
	"bytes"
	"context"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func patExpectedBytes(versionNumber uint8, cc uint8) []byte {
	buf := bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
	w.Write(uint8(syncByte))
	w.Write("010") // no transport error, payload start, no priority
	w.WriteN(PIDPAT, 13)
	w.Write("0001") // no scrambling, no AF, payload present
	w.WriteN(cc, 4)

	w.Write(uint16(0))       // Table ID
	w.Write("1011")          // Syntax section indicator, private bit, reserved
	w.WriteN(uint16(13), 12) // Section length

	w.Write(uint16(PSITableIDPAT))
	w.Write("11")              // Reserved bits
	w.WriteN(versionNumber, 5) // Version number
	w.Write("1")               // Current/next indicator
	w.Write(uint8(0))          // Section number
	w.Write(uint8(0))          // Last section number

	w.Write(programNumberStart)
	w.Write("111") // reserved
	w.WriteN(pmtStartPID, 13)

	// CRC32
	if versionNumber == 0 {
		w.Write([]byte{0x71, 0x10, 0xd8, 0x78})
	} else {
		w.Write([]byte{0xef, 0xbe, 0x08, 0x5a})
	}

	w.Write(bytes.Repeat([]byte{0xff}, 167))

	return buf.Bytes()
}

func TestMuxer_generatePAT(t *testing.T) {
	muxer := NewMuxer(context.Background(), nil)

	err := muxer.generatePAT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.patBytes.Len())
	assert.Equal(t, patExpectedBytes(0, 0), muxer.patBytes.Bytes())

	// Version number shouldn't change
	err = muxer.generatePAT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.patBytes.Len())
	assert.Equal(t, patExpectedBytes(0, 1), muxer.patBytes.Bytes())

	// Version number should change
	muxer.pmUpdated = true
	err = muxer.generatePAT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.patBytes.Len())
	assert.Equal(t, patExpectedBytes(1, 2), muxer.patBytes.Bytes())
}

func pmtExpectedBytesVideoOnly(versionNumber, cc uint8) []byte {
	buf := bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
	w.Write(uint8(syncByte))
	w.Write("010") // no transport error, payload start, no priority
	w.WriteN(pmtStartPID, 13)
	w.Write("0001") // no scrambling, no AF, payload present
	w.WriteN(cc, 4)

	w.Write(uint16(PSITableIDPMT)) // Table ID
	w.Write("1011")                // Syntax section indicator, private bit, reserved
	w.WriteN(uint16(18), 12)       // Section length

	w.Write(programNumberStart)
	w.Write("11")              // Reserved bits
	w.WriteN(versionNumber, 5) // Version number
	w.Write("1")               // Current/next indicator
	w.Write(uint8(0))          // Section number
	w.Write(uint8(0))          // Last section number

	w.Write("111")               // reserved
	w.WriteN(uint16(0x1234), 13) // PCR PID

	w.Write("1111")         // reserved
	w.WriteN(uint16(0), 12) // program info length

	w.Write(uint8(StreamTypeH264Video))
	w.Write("111") // reserved
	w.WriteN(uint16(0x1234), 13)

	w.Write("1111")         // reserved
	w.WriteN(uint16(0), 12) // es info length

	w.Write([]byte{0x31, 0x48, 0x5b, 0xa2}) // CRC32

	w.Write(bytes.Repeat([]byte{0xff}, 162))

	return buf.Bytes()
}

func pmtExpectedBytesVideoAndAudio(versionNumber uint8, cc uint8) []byte {
	buf := bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
	w.Write(uint8(syncByte))
	w.Write("010") // no transport error, payload start, no priority
	w.WriteN(pmtStartPID, 13)
	w.Write("0001") // no scrambling, no AF, payload present
	w.WriteN(cc, 4)

	w.Write(uint16(PSITableIDPMT)) // Table ID
	w.Write("1011")                // Syntax section indicator, private bit, reserved
	w.WriteN(uint16(23), 12)       // Section length

	w.Write(programNumberStart)
	w.Write("11")              // Reserved bits
	w.WriteN(versionNumber, 5) // Version number
	w.Write("1")               // Current/next indicator
	w.Write(uint8(0))          // Section number
	w.Write(uint8(0))          // Last section number

	w.Write("111")               // reserved
	w.WriteN(uint16(0x1234), 13) // PCR PID

	w.Write("1111")         // reserved
	w.WriteN(uint16(0), 12) // program info length

	w.Write(uint8(StreamTypeH264Video))
	w.Write("111") // reserved
	w.WriteN(uint16(0x1234), 13)
	w.Write("1111")         // reserved
	w.WriteN(uint16(0), 12) // es info length

	w.Write(uint8(StreamTypeADTS))
	w.Write("111") // reserved
	w.WriteN(uint16(0x0234), 13)
	w.Write("1111")         // reserved
	w.WriteN(uint16(0), 12) // es info length

	// CRC32
	if versionNumber == 0 {
		w.Write([]byte{0x29, 0x52, 0xc4, 0x50})
	} else {
		w.Write([]byte{0x06, 0xf4, 0xa6, 0xea})
	}

	w.Write(bytes.Repeat([]byte{0xff}, 157))

	return buf.Bytes()
}

func TestMuxer_generatePMT(t *testing.T) {
	muxer := NewMuxer(context.Background(), nil)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	muxer.SetPCRPID(0x1234)
	assert.NoError(t, err)

	err = muxer.generatePMT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.pmtBytes.Len())
	assert.Equal(t, pmtExpectedBytesVideoOnly(0, 0), muxer.pmtBytes.Bytes())

	// Version number shouldn't change
	err = muxer.generatePMT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.pmtBytes.Len())
	assert.Equal(t, pmtExpectedBytesVideoOnly(0, 1), muxer.pmtBytes.Bytes())

	err = muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x0234,
		StreamType:    StreamTypeAACAudio,
	})
	assert.NoError(t, err)

	// Version number should change
	err = muxer.generatePMT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.pmtBytes.Len())
	assert.Equal(t, pmtExpectedBytesVideoAndAudio(1, 2), muxer.pmtBytes.Bytes())
}

func TestMuxer_WriteTables(t *testing.T) {
	buf := bytes.Buffer{}
	muxer := NewMuxer(context.Background(), &buf)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	muxer.SetPCRPID(0x1234)
	assert.NoError(t, err)

	n, err := muxer.WriteTables()
	assert.NoError(t, err)
	assert.Equal(t, 2*MpegTsPacketSize, n)
	assert.Equal(t, n, buf.Len())

	expectedBytes := append(patExpectedBytes(0, 0), pmtExpectedBytesVideoOnly(0, 0)...)
	assert.Equal(t, expectedBytes, buf.Bytes())
}

func TestMuxer_WriteTables_Error(t *testing.T) {
	muxer := NewMuxer(context.Background(), nil)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	assert.NoError(t, err)

	_, err = muxer.WriteTables()
	assert.Equal(t, ErrPCRPIDInvalid, err)
}

func TestMuxer_AddElementaryStream(t *testing.T) {
	muxer := NewMuxer(context.Background(), nil)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	assert.NoError(t, err)

	err = muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	assert.Equal(t, ErrPIDAlreadyExists, err)
}

func TestMuxer_RemoveElementaryStream(t *testing.T) {
	muxer := NewMuxer(context.Background(), nil)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	assert.NoError(t, err)

	err = muxer.RemoveElementaryStream(0x1234)
	assert.NoError(t, err)

	err = muxer.RemoveElementaryStream(0x1234)
	assert.Equal(t, ErrPIDNotFound, err)
}

func testPayload() []byte {
	ret := make([]byte, 0xff+1)
	for i := 0; i <= 0xff; i++ {
		ret[i] = byte(i)
	}
	return ret
}

func TestMuxer_WritePayload(t *testing.T) {
	buf := bytes.Buffer{}
	muxer := NewMuxer(context.Background(), &buf)

	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	})
	muxer.SetPCRPID(0x1234)
	assert.NoError(t, err)

	err = muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x0234,
		StreamType:    StreamTypeAACAudio,
	})
	assert.NoError(t, err)

	payload := testPayload()
	pcr := ClockReference{
		Base:      5726623061,
		Extension: 341,
	}
	pts := ClockReference{Base: 5726623060}

	n, err := muxer.WriteData(&MuxerData{
		PID: 0x1234,
		AdaptationField: &PacketAdaptationField{
			HasPCR:                true,
			PCR:                   &pcr,
			RandomAccessIndicator: true,
		},
		PES: &PESData{
			Data: payload,
			Header: &PESHeader{
				OptionalHeader: &PESOptionalHeader{
					DTS:             &pts,
					PTS:             &pts,
					PTSDTSIndicator: PTSDTSIndicatorBothPresent,
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, buf.Len(), n)

	bytesTotal := n

	n, err = muxer.WriteData(&MuxerData{
		PID: 0x0234,
		AdaptationField: &PacketAdaptationField{
			HasPCR:                true,
			PCR:                   &pcr,
			RandomAccessIndicator: true,
		},
		PES: &PESData{
			Data: payload,
			Header: &PESHeader{
				OptionalHeader: &PESOptionalHeader{
					DTS:             &pts,
					PTS:             &pts,
					PTSDTSIndicator: PTSDTSIndicatorBothPresent,
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, buf.Len(), bytesTotal+n)
	assert.Equal(t, 0, buf.Len()%MpegTsPacketSize)

	bs := buf.Bytes()
	assert.Equal(t, patExpectedBytes(0, 0), bs[:MpegTsPacketSize])
	assert.Equal(t, pmtExpectedBytesVideoAndAudio(0, 0), bs[MpegTsPacketSize:MpegTsPacketSize*2])
}
