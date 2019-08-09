package astits

import (
	astibyte "github.com/asticode/go-astitools/byte"
	"github.com/pkg/errors"
)

// Stream types
const (
	StreamTypeLowerBitrateVideo          = 27 // ITU-T Rec. H.264 and ISO/IEC 14496-10
	StreamTypeMPEG1Audio                 = 3  // ISO/IEC 11172-3
	StreamTypeMPEG2HalvedSampleRateAudio = 4  // ISO/IEC 13818-3
	StreamTypeMPEG2PacketizedData        = 6  // ITU-T Rec. H.222 and ISO/IEC 13818-1 i.e., DVB subtitles/VBI and AC-3
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
func parsePMTSection(i *astibyte.Iterator, offsetSectionsEnd int, tableIDExtension uint16) (d *PMTData, err error) {
	// Create data
	d = &PMTData{ProgramNumber: tableIDExtension}

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = errors.Wrap(err, "astits: fetching next bytes failed")
		return
	}

	// PCR PID
	d.PCRPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

	// Program descriptors
	if d.ProgramDescriptors, err = parseDescriptors(i); err != nil {
		err = errors.Wrap(err, "astits: parsing descriptors failed")
		return
	}

	// Loop until end of section data is reached
	for i.Offset() < offsetSectionsEnd {
		// Create stream
		e := &PMTElementaryStream{}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = errors.Wrap(err, "astits: fetching next byte failed")
			return
		}

		// Stream type
		e.StreamType = uint8(b)

		// Get next bytes
		if bs, err = i.NextBytes(2); err != nil {
			err = errors.Wrap(err, "astits: fetching next bytes failed")
			return
		}

		// Elementary PID
		e.ElementaryPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

		// Elementary descriptors
		if e.ElementaryStreamDescriptors, err = parseDescriptors(i); err != nil {
			err = errors.Wrap(err, "astits: parsing descriptors failed")
			return
		}

		// Add elementary stream
		d.ElementaryStreams = append(d.ElementaryStreams, e)
	}
	return
}
