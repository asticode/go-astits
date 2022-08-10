package astits

import (
	"fmt"
	"time"

	"github.com/icza/bitio"
)

// EITData represents an EIT data
// Page: 36 | Chapter: 5.2.4 | Link:
// https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type EITData struct {
	Events                   []*EITDataEvent
	LastTableID              uint8
	OriginalNetworkID        uint16
	SegmentLastSectionNumber uint8
	ServiceID                uint16
	TransportStreamID        uint16
}

// EITDataEvent represents an EIT data event.
type EITDataEvent struct {
	Duration      time.Duration
	EventID       uint16
	StartTime     time.Time
	RunningStatus uint8

	// When true indicates that access to one or
	// more streams may be controlled by a CA system.
	HasFreeCSAMode bool
	Descriptors    []*Descriptor
}

// parseEITSection parses an EIT section.
func parseEITSection(
	r *bitio.CountReader,
	offsetSectionsEnd int64,
	tableIDExtension uint16,
) (*EITData, error) {
	d := &EITData{ServiceID: tableIDExtension}

	d.TransportStreamID = uint16(r.TryReadBits(16))

	d.OriginalNetworkID = uint16(r.TryReadBits(16))

	d.SegmentLastSectionNumber = r.TryReadByte()

	d.LastTableID = r.TryReadByte()

	// Loop until end of section data is reached.
	for r.BitsCount < offsetSectionsEnd {
		e := &EITDataEvent{}

		e.EventID = uint16(r.TryReadBits(16))

		var err error
		if e.StartTime, err = parseDVBTime(r); err != nil {
			return nil, fmt.Errorf("parsing DVB time: %w", err)
		}

		if e.Duration, err = parseDVBDurationSeconds(r); err != nil {
			return nil, fmt.Errorf("parsing DVB duration seconds failed: %w", err)
		}

		e.RunningStatus = uint8(r.TryReadBits(3))

		e.HasFreeCSAMode = r.TryReadBool()

		if e.Descriptors, err = parseDescriptors(r); err != nil {
			return nil, fmt.Errorf("parsing descriptors failed: %w", err)
		}

		d.Events = append(d.Events, e)
	}

	return d, r.TryError
}
