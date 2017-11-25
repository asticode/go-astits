package astits

import "time"

// TOTData represents a TOT data
// Page: 39 | Chapter: 5.2.6 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type TOTData struct {
	Descriptors []*Descriptor
	UTCTime     time.Time
}

// parseTOTSection parses a TOT section
func parseTOTSection(i []byte, offset *int) (d *TOTData) {
	// Init
	d = &TOTData{}

	// UTC time
	d.UTCTime = parseDVBTime(i, offset)

	// Descriptors
	d.Descriptors = parseDescriptors(i, offset)
	return
}
