package astits

import "github.com/icza/bitio"

const (
	patSectionEntryBytesSize = 4 // 16 bits + 3 reserved + 13 bits = 32 bits
)

// PATData represents a PAT data
// https://en.wikipedia.org/wiki/Program-specific_information
type PATData struct {
	Programs          []*PATProgram
	TransportStreamID uint16
}

// PATProgram represents a PAT program.
type PATProgram struct {
	// ProgramNumber Relates to the Table ID extension in the associated PMT.
	// A value of 0 is reserved for a NIT packet identifier.
	ProgramNumber uint16

	// ProgramMapID 13 bits. The packet identifier that contains the associated PMT
	ProgramMapID uint16
}

// parsePATSection parses a PAT section.
func parsePATSection(
	r *bitio.CountReader,
	offsetSectionsEnd int64,
	tableIDExtension uint16,
) (*PATData, error) {
	d := &PATData{TransportStreamID: tableIDExtension}

	for r.BitsCount < offsetSectionsEnd {
		p := &PATProgram{}

		p.ProgramNumber = uint16(r.TryReadBits(16))
		r.TryReadBits(3)
		p.ProgramMapID = uint16(r.TryReadBits(13))

		d.Programs = append(d.Programs, p)
	}
	return d, r.TryError
}

func calcPATSectionLength(d *PATData) uint16 {
	return uint16(4 * len(d.Programs))
}

func writePATSection(w *bitio.Writer, d *PATData) (int, error) {
	for _, p := range d.Programs {
		w.TryWriteBits(uint64(p.ProgramNumber), 16)
		w.TryWriteBits(0xff, 3)
		w.TryWriteBits(uint64(p.ProgramMapID), 13)
	}

	return len(d.Programs) * patSectionEntryBytesSize, w.TryError
}
