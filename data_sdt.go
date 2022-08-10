package astits

import (
	"fmt"

	"github.com/icza/bitio"
)

// Running statuses.
const (
	RunningStatusNotRunning          = 1
	RunningStatusPausing             = 3
	RunningStatusRunning             = 4
	RunningStatusServiceOffAir       = 5
	RunningStatusStartsInAFewSeconds = 2
	RunningStatusUndefined           = 0
)

// SDTData represents an SDT data.
// Page: 33 | Chapter: 5.2.3 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type SDTData struct {
	OriginalNetworkID uint16
	Services          []*SDTDataService
	TransportStreamID uint16
}

// SDTDataService represents an SDT data service.
type SDTDataService struct {
	Descriptors []*Descriptor

	// When true indicates that EIT present/following
	// information for the service is present in the current TS.
	HasEITPresentFollowing bool

	// When true indicates that EIT schedule information
	// for the service is present in the current TS.
	HasEITSchedule bool

	// When true indicates that access to one or
	// more streams may be controlled by a CA system.
	HasFreeCSAMode bool
	RunningStatus  uint8
	ServiceID      uint16
}

// parseSDTSection parses an SDT section.
func parseSDTSection(
	r *bitio.CountReader,
	offsetSectionsEnd int64,
	tableIDExtension uint16,
) (*SDTData, error) {
	d := &SDTData{TransportStreamID: tableIDExtension}

	d.OriginalNetworkID = uint16(r.TryReadBits(16))

	_ = r.TryReadByte() // Reserved.

	// Loop until end of section data is reached.
	for r.BitsCount < offsetSectionsEnd {
		s := &SDTDataService{}

		s.ServiceID = uint16(r.TryReadBits(16))

		_ = r.TryReadBits(6) // Reserved.
		s.HasEITSchedule = r.TryReadBool()
		s.HasEITPresentFollowing = r.TryReadBool()

		s.RunningStatus = uint8(r.TryReadBits(3))
		s.HasFreeCSAMode = r.TryReadBool()

		var err error
		if s.Descriptors, err = parseDescriptors(r); err != nil {
			return nil, fmt.Errorf("parsing descriptors failed: %w", err)
		}

		d.Services = append(d.Services, s)
	}
	return d, r.TryError
}
