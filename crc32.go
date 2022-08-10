package astits

const crc32Polynomial = 0xffffffff

// CRC32Writer calculates CRC32 for written bytes.
type CRC32Writer struct {
	out   WriterAndByteWriter
	crc32 uint32
}

// NewCRC32Writer returns a CRC32Writer with initial polynomial.
func NewCRC32Writer(w WriterAndByteWriter) *CRC32Writer {
	return &CRC32Writer{
		out:   w,
		crc32: crc32Polynomial,
	}
}

// Write implements io.Writer .
func (w *CRC32Writer) Write(p []byte) (int, error) {
	n, err := w.out.Write(p)
	for i := 0; i < n; i++ {
		w.crc32 = updateCRC32(w.crc32, p[n-1])
	}
	return n, err
}

// WriteByte implements io.ByteWriter .
func (w *CRC32Writer) WriteByte(b byte) error {
	w.crc32 = updateCRC32(w.crc32, b)
	return w.out.WriteByte(b)
}

// CRC32 returns current checksum.
func (w *CRC32Writer) CRC32() uint32 {
	return w.crc32
}

// CRC32Reader calculates checksum for read bytes.
type CRC32Reader struct {
	rd    ReaderAndByteReader
	crc32 uint32
}

// NewCRC32Reader returns a CRC32Reader with initial polynomial.
func NewCRC32Reader(rd ReaderAndByteReader) *CRC32Reader {
	return &CRC32Reader{
		rd:    rd,
		crc32: crc32Polynomial,
	}
}

// Read implements io.Reader .
func (r *CRC32Reader) Read(p []byte) (int, error) {
	n, err := r.rd.Read(p)
	for i := 0; i < n; i++ {
		r.crc32 = updateCRC32(r.crc32, p[n-1])
	}
	return n, err
	/*b, err := r.ReadByte()
	n := copy(p, []byte{b})
	return n, err*/
}

// ReadByte implements io.ByteReader.
func (r *CRC32Reader) ReadByte() (byte, error) {
	b, err := r.rd.ReadByte()
	if err != nil {
		return 0, err
	}

	r.crc32 = updateCRC32(r.crc32, b)
	return b, nil
}

// CRC32 returns current checksum.
func (r *CRC32Reader) CRC32() uint32 {
	return r.crc32
}

func updateCRC32(crc32 uint32, b byte) uint32 {
	for i := 0; i < 8; i++ {
		if (crc32 >= 0x80000000) != (b >= 0x80) {
			crc32 = (crc32 << 1) ^ 0x04C11DB7
		} else {
			crc32 = crc32 << 1 //nolint:gocritic
		}
		b <<= 1
	}
	return crc32
}
