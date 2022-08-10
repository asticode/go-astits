package astits

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/require"
)

func TestBitsWriter(t *testing.T) {
	// TODO Need to test LittleEndian
	bw := &bytes.Buffer{}
	w := bitio.NewWriter(bw)

	err := WriteBinary(w, "000000")
	require.NoError(t, err)
	require.Equal(t, 0, bw.Len())

	err = w.WriteBool(false)
	require.NoError(t, err)

	err = w.WriteBool(true)
	require.Equal(t, []byte{1}, bw.Bytes())

	_, err = w.Write([]byte{2, 3})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, bw.Bytes())

	err = w.WriteBits(uint64(4), 8)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, bw.Bytes())

	err = w.WriteBits(uint64(5), 16)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4, 0, 5}, bw.Bytes())

	err = w.WriteBits(uint64(6), 32)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4, 0, 5, 0, 0, 0, 6}, bw.Bytes())

	err = w.WriteBits(uint64(7), 64)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4, 0, 5, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 7}, bw.Bytes())

	bw.Reset()
	err = w.WriteBits(uint64(4), 3)
	require.NoError(t, err)

	err = w.WriteBits(uint64(4096), 13)
	require.NoError(t, err)
	require.Equal(t, []byte{144, 0}, bw.Bytes())
}

// testLimitedWriter is an implementation of io.Writer
// with max write size limit to test error handling
type testLimitedWriter struct {
	BytesLimit int
}

func (t *testLimitedWriter) Write(p []byte) (n int, err error) {
	t.BytesLimit -= len(p)
	if t.BytesLimit >= 0 {
		return len(p), nil
	}
	return len(p) + t.BytesLimit, io.EOF
}

func (t *testLimitedWriter) WriteByte(c byte) error {
	_, err := t.Write([]byte{c})
	return err
}

func BenchmarkBitsWriter_Write(b *testing.B) {
	benchmarks := []func(*bitio.Writer){
		func(w *bitio.Writer) { WriteBinary(w, "000000") },
		func(w *bitio.Writer) { w.WriteBool(false) },
		func(w *bitio.Writer) { w.WriteBool(true) },
		func(w *bitio.Writer) { w.Write([]byte{2, 3}) },
		func(w *bitio.Writer) { w.WriteByte(uint8(4)) },
		func(w *bitio.Writer) { w.WriteBits(uint64(5), 16) },
		func(w *bitio.Writer) { w.WriteBits(uint64(6), 32) },
		func(w *bitio.Writer) { w.WriteBits(uint64(7), 64) },
	}

	bw := &bytes.Buffer{}
	bw.Grow(1024)
	w := bitio.NewWriter(bw)

	for i, bm := range benchmarks {
		b.Run(fmt.Sprintf("%v", i), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bw.Reset()
				bm(w)
			}
		})
	}
}

func BenchmarkBitsWriter_WriteN(b *testing.B) {
	type benchData func(w *bitio.Writer)
	benchmarks := []benchData{}
	var i uint8
	for i = 1; i <= 8; i++ {
		benchmarks = append(benchmarks,
			func(w *bitio.Writer) { w.WriteBits(0xff, i) })
	}
	for i = 1; i <= 16; i++ {
		benchmarks = append(benchmarks,
			func(w *bitio.Writer) { w.WriteBits(0xffff, i) })
	}
	for i = 1; i <= 32; i++ {
		benchmarks = append(benchmarks,
			func(w *bitio.Writer) { w.WriteBits(0xffffffff, i) })
	}
	for i = 1; i <= 64; i++ {
		benchmarks = append(benchmarks,
			func(w *bitio.Writer) { w.WriteBits(0xffffffffffffffff, i) })
	}

	bw := &bytes.Buffer{}
	bw.Grow(1024)
	w := bitio.NewWriter(bw)

	for i, bm := range benchmarks {
		b.Run(fmt.Sprintf("%v", i), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bw.Reset()
				bm(w)
			}
		})
	}
}
