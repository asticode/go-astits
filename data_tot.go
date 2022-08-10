package astits

import (
	"fmt"
	"time"

	"github.com/icza/bitio"
)

// TOTData represents a TOT data.
// Page: 39 | Chapter: 5.2.6 | Link:
// https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type TOTData struct {
	Descriptors []*Descriptor
	UTCTime     time.Time
}

// parseTOTSection parses a TOT section.
func parseTOTSection(r *bitio.CountReader) (*TOTData, error) {
	d := &TOTData{}

	var err error
	if d.UTCTime, err = parseDVBTime(r); err != nil {
		return nil, fmt.Errorf("parsing DVB time failed: %w", err)
	}

	if _, err = r.ReadBits(4); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	// Descriptors
	if d.Descriptors, err = parseDescriptors(r); err != nil {
		return nil, fmt.Errorf("parsing descriptors failed: %w", err)
	}
	return d, nil
}
