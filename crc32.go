package astits

const (
	crc32Polynomial = uint32(0xffffffff)
)

// this solution based on vlc implementation + static crc table (1kb additional memory on start, without reallocations)
// you can find generator in internal/cmd/crc32_table
// https://github.com/videolan/vlc/blob/master/modules/mux/mpeg/ps.c

func computeCRC32(bs []byte) uint32 {
	return updateCRC32(crc32Polynomial, bs)
}

func updateCRC32(iCrc uint32, bs []byte) uint32 {
	for _, b := range bs {
		iCrc = (iCrc << 8) ^ tableCRC32[((iCrc>>24)^uint32(b))&0xff]
	}
	return iCrc
}
