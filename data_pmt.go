package astits

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
func parsePMTSection(i []byte, offset *int, offsetSectionsEnd int, tableIDExtension uint16) (d *PMTData) {
	// Init
	d = &PMTData{ProgramNumber: tableIDExtension}

	// PCR PID
	d.PCRPID = uint16(i[*offset]&0x1f)<<8 | uint16(i[*offset+1])
	*offset += 2

	// Program descriptors
	d.ProgramDescriptors = parseDescriptors(i, offset)

	// Loop until end of section data is reached
	for *offset < offsetSectionsEnd {
		// Stream type
		var e = &PMTElementaryStream{}
		e.StreamType = uint8(i[*offset])
		*offset += 1

		// Elementary PID
		e.ElementaryPID = uint16(i[*offset]&0x1f)<<8 | uint16(i[*offset+1])
		*offset += 2

		// Elementary descriptors
		e.ElementaryStreamDescriptors = parseDescriptors(i, offset)

		// Add elementary stream
		d.ElementaryStreams = append(d.ElementaryStreams, e)
	}
	return
}
