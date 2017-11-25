package astits

import "time"

// EITData represents an EIT data
// Page: 36 | Chapter: 5.2.4 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type EITData struct {
	Events                   []*EITDataEvent
	LastTableID              uint8
	OriginalNetworkID        uint16
	SegmentLastSectionNumber uint8
	ServiceID                uint16
	TransportStreamID        uint16
}

// EITDataEvent represents an EIT data event
type EITDataEvent struct {
	Descriptors    []*Descriptor
	Duration       time.Duration
	EventID        uint16
	HasFreeCSAMode bool // When true indicates that access to one or more streams may be controlled by a CA system.
	RunningStatus  uint8
	StartTime      time.Time
}

// parseEITSection parses an EIT section
func parseEITSection(i []byte, offset *int, offsetSectionsEnd int, tableIDExtension uint16) (d *EITData) {
	// Init
	d = &EITData{ServiceID: tableIDExtension}

	// Transport stream ID
	d.TransportStreamID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	*offset += 2

	// Original network ID
	d.OriginalNetworkID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	*offset += 2

	// Segment last section number
	d.SegmentLastSectionNumber = uint8(i[*offset])
	*offset += 1

	// Last table ID
	d.LastTableID = uint8(i[*offset])
	*offset += 1

	// Loop until end of section data is reached
	for *offset < offsetSectionsEnd {
		// Event ID
		var e = &EITDataEvent{}
		e.EventID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
		*offset += 2

		// Start time
		e.StartTime = parseDVBTime(i, offset)

		// Duration
		e.Duration = parseDVBDurationSeconds(i, offset)

		// Running status
		e.RunningStatus = uint8(i[*offset]) >> 5

		// Free CA mode
		e.HasFreeCSAMode = uint8(i[*offset]&0x10) > 0

		// Descriptors
		e.Descriptors = parseDescriptors(i, offset)

		// Add event
		d.Events = append(d.Events, e)
	}
	return
}
