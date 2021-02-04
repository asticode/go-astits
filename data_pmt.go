package astits

import (
	"errors"
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

func calcPMTProgramInfoLength(d *PMTData) int {
	return 0
}

func writePMTSection(w *astikit.BitsWriter, d *PMTData) (int, error) {
	b := astikit.NewBitsWriterBatch(w)

	b.WriteN(uint8(0xff), 3)
	b.WriteN(d.PCRPID, 13)
	b.WriteN(uint8(0xff), 4)
	b.WriteN(uint16(calcPMTProgramInfoLength(d)), 12)
	//bytesWritten := 4
	//
	//for _, desc := range d.ProgramDescriptors {
	//	desc.
	//}

	return 0, errors.New("not implemented")
}
