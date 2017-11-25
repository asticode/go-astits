package astits

// Running statuses
const (
	RunningStatusNotRunning          = 1
	RunningStatusPausing             = 3
	RunningStatusRunning             = 4
	RunningStatusServiceOffAir       = 5
	RunningStatusStartsInAFewSeconds = 2
	RunningStatusUndefined           = 0
)

// SDTData represents an SDT data
// Page: 33 | Chapter: 5.2.3 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type SDTData struct {
	OriginalNetworkID uint16
	Services          []*SDTDataService
	TransportStreamID uint16
}

// SDTDataService represents an SDT data service
type SDTDataService struct {
	Descriptors            []*Descriptor
	HasEITPresentFollowing bool // When true indicates that EIT present/following information for the service is present in the current TS
	HasEITSchedule         bool // When true indicates that EIT schedule information for the service is present in the current TS
	HasFreeCSAMode         bool // When true indicates that access to one or more streams may be controlled by a CA system.
	RunningStatus          uint8
	ServiceID              uint16
}

// parseSDTSection parses an SDT section
func parseSDTSection(i []byte, offset *int, offsetSectionsEnd int, tableIDExtension uint16) (d *SDTData) {
	// Init
	d = &SDTData{TransportStreamID: tableIDExtension}

	// Original network ID
	d.OriginalNetworkID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	*offset += 2

	// Reserved for future use
	*offset += 1

	// Loop until end of section data is reached
	for *offset < offsetSectionsEnd {
		// Service ID
		var s = &SDTDataService{}
		s.ServiceID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
		*offset += 2

		// EIT schedule flag
		s.HasEITSchedule = uint8(i[*offset]&0x2) > 0

		// EIT present/following flag
		s.HasEITPresentFollowing = uint8(i[*offset]&0x1) > 0
		*offset += 1

		// Running status
		s.RunningStatus = uint8(i[*offset]) >> 5

		// Free CA mode
		s.HasFreeCSAMode = uint8(i[*offset]&0x10) > 0

		// Descriptors
		s.Descriptors = parseDescriptors(i, offset)

		// Append service
		d.Services = append(d.Services, s)
	}
	return
}
