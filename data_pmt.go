package astits

import (
	"fmt"

	"github.com/asticode/go-astikit"
)

// Stream types
const (
	StreamTypeMPEG1Audio                 = 0x03 // ISO/IEC 11172-3
	StreamTypeMPEG2HalvedSampleRateAudio = 0x04 // ISO/IEC 13818-3
	StreamTypeMPEG2PacketizedData        = 0x06 // Rec. ITU-T H.222 | ISO/IEC 13818-1 i.e., DVB subtitles/VBI and AC-3
	StreamTypeADTS                       = 0x0F // ISO/IEC 13818-7 Audio with ADTS transport syntax
	StreamTypeH264Video                  = 0x1B // Rec. ITU-T H.264 | ISO/IEC 14496-10
	StreamTypeH265Video                  = 0x24 // Rec. ITU-T H.265 | ISO/IEC 23008-2
)

// PMTData represents a PMT data
// https://en.wikipedia.org/wiki/Program-specific_information
type PMTData struct {
	ElementaryStreams  []*PMTElementaryStream
	PCRPID             uint16        // The packet identifier that contains the program clock reference used to improve the random access accuracy of the stream's timing that is derived from the program timestamp. If this is unused. then it is set to 0x1FFF (all bits on).
	ProgramDescriptors []*Descriptor // Program descriptors
	ProgramNumber      uint16
}

// PMTElementaryStream represents a PMT elementary stream
type PMTElementaryStream struct {
	ElementaryPID               uint16        // The packet identifier that contains the stream type data.
	ElementaryStreamDescriptors []*Descriptor // Elementary stream descriptors
	StreamType                  uint8         // This defines the structure of the data contained within the elementary packet identifier.
}

// parsePMTSection parses a PMT section
func parsePMTSection(i *astikit.BytesIterator, offsetSectionsEnd int, tableIDExtension uint16) (d *PMTData, err error) {
	// Create data
	d = &PMTData{ProgramNumber: tableIDExtension}

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// PCR PID
	d.PCRPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

	// Program descriptors
	if d.ProgramDescriptors, err = parseDescriptors(i); err != nil {
		err = fmt.Errorf("astits: parsing descriptors failed: %w", err)
		return
	}

	// Loop until end of section data is reached
	for i.Offset() < offsetSectionsEnd {
		// Create stream
		e := &PMTElementaryStream{}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Stream type
		e.StreamType = uint8(b)

		// Get next bytes
		if bs, err = i.NextBytes(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Elementary PID
		e.ElementaryPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

		// Elementary descriptors
		if e.ElementaryStreamDescriptors, err = parseDescriptors(i); err != nil {
			err = fmt.Errorf("astits: parsing descriptors failed: %w", err)
			return
		}

		// Add elementary stream
		d.ElementaryStreams = append(d.ElementaryStreams, e)
	}
	return
}

func (p *PMTData) Serialise(b []byte) (int, error) {
	b[0] = 0x7<<5 | uint8(0x1f&(p.PCRPID>>8))
	b[1] = uint8(0xff & p.PCRPID)
	program_info_length := 0
	idx := 4
	for i := range p.ProgramDescriptors {
		n, err := p.ProgramDescriptors[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
		program_info_length += n
	}
	for i := range p.ElementaryStreams {
		n, err := p.ElementaryStreams[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
	}
	b[2] = 0xf0 | uint8(0x3&(uint8(program_info_length)>>8))
	b[3] = uint8(program_info_length)
	return idx, nil
}

func (pes *PMTElementaryStream) Serialise(b []byte) (int, error) {

	b[0] = pes.StreamType
	b[1] = 0x7<<5 | uint8(0x1f&(pes.ElementaryPID>>8))
	b[2] = uint8(0xff & pes.ElementaryPID)
	es_info_length := 0
	idx := 5
	for i := range pes.ElementaryStreamDescriptors {
		n, err := pes.ElementaryStreamDescriptors[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
		es_info_length += n
	}
	b[3] = 0xf0 | (uint8(0x3 & (es_info_length >> 8)))
	b[4] = uint8(es_info_length)
	return idx, nil
	// type PMTElementaryStream struct {
	// 	ElementaryPID               uint16        // The packet identifier that contains the stream type data.
	// 	ElementaryStreamDescriptors []*Descriptor // Elementary stream descriptors
	// 	StreamType                  uint8         // This defines the structure of the data contained within the elementary packet identifier.
	// }
}
