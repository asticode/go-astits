package astits

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

func TestAutoDetectPacketSize(t *testing.T) {
	// Packet should start with a sync byte
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(2))
	w.Write(byte(syncByte))
	_, err := autoDetectPacketSize(bytes.NewReader(buf.Bytes()))
	assert.EqualError(t, err, ErrPacketMustStartWithASyncByte.Error())

	// Valid packet size
	buf.Reset()
	w.Write(byte(syncByte))
	w.Write(make([]byte, 20))
	w.Write(byte(syncByte))
	w.Write(make([]byte, 166))
	w.Write(byte(syncByte))
	w.Write(make([]byte, 187))
	w.Write([]byte("test"))
	r := bytes.NewReader(buf.Bytes())
	p, err := autoDetectPacketSize(r)
	assert.NoError(t, err)
	assert.Equal(t, MpegTsPacketSize, p)
	assert.Equal(t, 380, r.Len())
}

func TestFindSync(t *testing.T) {
	sizes := []int{MpegTsPacketSize, 192, 204}
	for _, c := range []struct {
		name   string
		b      []byte
		offset int
		size   int
		ok     bool
	}{
		{name: "188 at nonzero offset", b: syncedBytes(5, 188, 3), offset: 5, size: 188, ok: true},
		{name: "192 with leading junk", b: syncedBytes(17, 192, 3), offset: 17, size: 192, ok: true},
		{name: "204 with leading junk", b: syncedBytes(100, 204, 3), offset: 100, size: 204, ok: true},
		{name: "lone sync byte in junk is skipped", b: func() []byte {
			b := syncedBytes(50, 188, 3)
			b[10] = syncByte
			return b
		}(), offset: 50, size: 188, ok: true},
		{name: "no match", b: append([]byte{syncByte}, make([]byte, 400)...)},
	} {
		t.Run(c.name, func(t *testing.T) {
			offset, size, ok := findSync(c.b, sizes)
			assert.Equal(t, c.ok, ok)
			assert.Equal(t, c.offset, offset)
			assert.Equal(t, c.size, size)
		})
	}
}

// syncedBytes returns junk bytes followed by count packets of the given size
func syncedBytes(junk, size, count int) []byte {
	b := make([]byte, junk+size*count)
	for i := range b {
		b[i] = 0xff
	}
	for i := 0; i < count; i++ {
		b[junk+i*size] = syncByte
	}
	return b
}

type packetID struct {
	pid  uint16
	cc   uint8
	pusi bool
}

func demuxPacketIDs(t *testing.T, r io.Reader, opts ...func(*Demuxer)) (ids []packetID) {
	dmx := NewDemuxer(context.Background(), r, opts...)
	for {
		p, err := dmx.NextPacket()
		if errors.Is(err, ErrNoMorePackets) {
			return
		}
		if !assert.NoError(t, err) {
			return
		}
		ids = append(ids, packetID{pid: p.Header.PID, cc: p.Header.ContinuityCounter, pusi: p.Header.PayloadUnitStartIndicator})
	}
}

func TestDemuxerInitialSync(t *testing.T) {
	b, err := os.ReadFile("testdata/ts/silent_audio.ts")
	if !assert.NoError(t, err) {
		return
	}
	ref := demuxPacketIDs(t, bytes.NewReader(b))
	if !assert.Len(t, ref, 4) {
		return
	}

	readers := map[string]func([]byte) io.Reader{
		"plain":  func(b []byte) io.Reader { return struct{ io.Reader }{bytes.NewReader(b)} },
		"bufio":  func(b []byte) io.Reader { return bufio.NewReader(bytes.NewReader(b)) },
		"seeker": func(b []byte) io.Reader { return bytes.NewReader(b) },
	}
	for _, cut := range []int{0, 1, 100, 187} {
		for _, packetSize := range []int{0, MpegTsPacketSize} {
			for name, newReader := range readers {
				t.Run(fmt.Sprintf("cut=%d/size=%d/%s", cut, packetSize, name), func(t *testing.T) {
					var opts []func(*Demuxer)
					if packetSize > 0 {
						opts = append(opts, DemuxerOptPacketSize(packetSize))
					}

					first := 1
					if cut == 0 {
						first = 0
						if packetSize == 0 && name == "plain" {
							// Existing behavior: the reader is synced past the detection bytes
							first = 2
						}
					} else if packetSize == 0 && name == "plain" {
						// Packets within the auto detection window can't be recovered
						n := syncWindow
						if l := len(b) - cut; l < n {
							n = l
						}
						offset := MpegTsPacketSize - cut
						boundary := offset + (n-offset+MpegTsPacketSize-1)/MpegTsPacketSize*MpegTsPacketSize
						first = (cut + boundary) / MpegTsPacketSize
					}

					ids := demuxPacketIDs(t, newReader(b[cut:]), opts...)
					assert.Equal(t, len(ref)-first, len(ids))
					if len(ids) > 0 {
						assert.Equal(t, ref[first:], ids)
					}
				})
			}
		}
	}
}
