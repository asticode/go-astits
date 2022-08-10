package astits

import (
	"io"
	"log"

	"github.com/icza/bitio"
)

// WriterAndByteWriter An io.Writer and io.ByteWriter at the same time.
type WriterAndByteWriter interface {
	io.Writer
	io.ByteWriter
}

// ReaderAndByteReader An io.Reader and io.ByteReader at the same time.
type ReaderAndByteReader interface {
	io.Reader
	io.ByteReader
}

// WriteBinary .
func WriteBinary(w *bitio.Writer, str string) error {
	for _, r := range str {
		var err error

		switch r {
		case '1':
			err = w.WriteBool(true)
		case '0':
			err = w.WriteBool(false)
		default:
			log.Fatalf("invalid rune: %v", r)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// TryReadFull .
func TryReadFull(r *bitio.CountReader, p []byte) {
	if r.TryError == nil {
		_, r.TryError = io.ReadFull(r, p)
	}
}
