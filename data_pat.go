package astits

import (
	astibyte "github.com/asticode/go-astitools/byte"
	"github.com/pkg/errors"
)

// PATData represents a PAT data
// https://en.wikipedia.org/wiki/Program-specific_information
type PATData struct {
	Programs          []*PATProgram
	TransportStreamID uint16
}

// PATProgram represents a PAT program
type PATProgram struct {
	ProgramMapID  uint16 // The packet identifier that contains the associated PMT
	ProgramNumber uint16 // Relates to the Table ID extension in the associated PMT. A value of 0 is reserved for a NIT packet identifier.
}

// parsePATSection parses a PAT section
func parsePATSection(i *astibyte.Iterator, offsetSectionsEnd int, tableIDExtension uint16) (d *PATData, err error) {
	// Create data
	d = &PATData{TransportStreamID: tableIDExtension}

	// Loop until end of section data is reached
	for i.Offset() < offsetSectionsEnd {
		// Get next bytes
		var bs []byte
		if bs, err = i.NextBytes(4); err != nil {
			err = errors.Wrap(err, "astits: fetching next bytes failed")
			return
		}

		// Append program
		d.Programs = append(d.Programs, &PATProgram{
			ProgramMapID:  uint16(bs[2]&0x1f)<<8 | uint16(bs[3]),
			ProgramNumber: uint16(bs[0])<<8 | uint16(bs[1]),
		})
	}
	return
}
