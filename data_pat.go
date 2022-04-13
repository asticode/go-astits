package astits

import (
	"fmt"

	"github.com/asticode/go-astikit"
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
func parsePATSection(i *astikit.BytesIterator, offsetSectionsEnd int, tableIDExtension uint16) (d *PATData, err error) {
	// Create data
	d = &PATData{TransportStreamID: tableIDExtension}

	// Loop until end of section data is reached
	for i.Offset() < offsetSectionsEnd {
		// Get next bytes
		var bs []byte
		if bs, err = i.NextBytes(4); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
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

func (p *PATData) Serialise(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrNoRoomInBuffer
	}
	currentIdx := 0
	for i := range p.Programs {
		n, err := p.Programs[i].Serialise(b[currentIdx:])
		if err != nil {
			return currentIdx, err
		}
		currentIdx += n
	}

	return currentIdx, nil
}

func (p *PATProgram) Serialise(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, ErrNoRoomInBuffer
	}
	b[0], b[1] = U16toU8s(p.ProgramNumber)
	// if p.ProgramNumber == 0 {
	// 	//TODO figure out Network PID
	// 	return 2, errors.New("Network PID not implemented")
	// }
	b[2] = uint8(0x1f&(p.ProgramMapID>>8)) | 7<<5
	b[3] = uint8(0xff & p.ProgramMapID)
	return 4, nil
}
