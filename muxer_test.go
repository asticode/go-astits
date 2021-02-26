package astits

import (
	"bytes"
	"context"
	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMuxer_generatePAT(t *testing.T) {
	patExpectedBytes := func() []byte {
		buf := bytes.Buffer{}
		w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
		w.Write(uint8(syncByte))
		w.Write("010") // no transport error, payload start, no priority
		w.WriteN(PIDPAT, 13)
		w.Write("0001") // no scrambling, no AF, payload present
		w.Write("0000") // CC

		w.Write(uint16(0))       // Table ID
		w.Write("1011")          // Syntax section indicator, private bit, reserved
		w.WriteN(uint16(13), 12) // Section length

		w.Write(uint16(PSITableTypeIdPAT))
		w.Write("11")     // Reserved bits
		w.Write("00000")  // Version number
		w.Write("1")      // Current/next indicator
		w.Write(uint8(0)) // Section number
		w.Write(uint8(0)) // Last section number

		w.Write(ProgramNumberStart)
		w.Write("111") // reserved
		w.WriteN(PMTStartPID, 13)

		w.Write([]byte{0x71, 0x10, 0xd8, 0x78}) // CRC32

		w.Write(bytes.Repeat([]byte{0xff}, 167))

		return buf.Bytes()
	}

	muxer := NewMuxer(context.Background(), nil)
	err := muxer.generatePAT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.patBytes.Len())
	assert.Equal(t, patExpectedBytes(), muxer.patBytes.Bytes())
}

func TestMuxer_generatePMT(t *testing.T) {
	pmtExpectedBytes := func() []byte {
		buf := bytes.Buffer{}
		w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
		w.Write(uint8(syncByte))
		w.Write("010") // no transport error, payload start, no priority
		w.WriteN(PMTStartPID, 13)
		w.Write("0001") // no scrambling, no AF, payload present
		w.Write("0000") // CC

		w.Write(uint16(PSITableTypeIdPMT)) // Table ID
		w.Write("1011")                    // Syntax section indicator, private bit, reserved
		w.WriteN(uint16(18), 12)           // Section length

		w.Write(ProgramNumberStart)
		w.Write("11")     // Reserved bits
		w.Write("00000")  // Version number
		w.Write("1")      // Current/next indicator
		w.Write(uint8(0)) // Section number
		w.Write(uint8(0)) // Last section number

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

	muxer := NewMuxer(context.Background(), nil)
	err := muxer.AddElementaryStream(PMTElementaryStream{
		ElementaryPID: 0x1234,
		StreamType:    StreamTypeH264Video,
	}, true)
	assert.NoError(t, err)

	err = muxer.generatePMT()
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, muxer.pmtBytes.Len())
	assert.Equal(t, pmtExpectedBytes(), muxer.pmtBytes.Bytes())
}
