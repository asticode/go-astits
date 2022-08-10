package astits

import (
	"fmt"

	"github.com/icza/bitio"
)

// NITData represents a NIT data.
// Page: 29 | Chapter: 5.2.1 | Link:
// https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type NITData struct {
	NetworkDescriptors []*Descriptor
	NetworkID          uint16
	TransportStreams   []*NITDataTransportStream
}

// NITDataTransportStream represents a NIT data transport stream.
type NITDataTransportStream struct {
	OriginalNetworkID    uint16
	TransportDescriptors []*Descriptor
	TransportStreamID    uint16
}

// parseNITSection parses a NIT section.
func parseNITSection(r *bitio.CountReader, tableIDExtension uint16) (*NITData, error) {
	d := &NITData{NetworkID: tableIDExtension}

	_ = r.TryReadBits(4)

	var err error
	if d.NetworkDescriptors, err = parseDescriptors(r); err != nil {
		return nil, fmt.Errorf("parsing descriptors failed: %w", err)
	}

	transportStreamLoopLength := int64(r.TryReadBits(16))

	offsetEnd := r.BitsCount/8 + transportStreamLoopLength
	for r.BitsCount/8 < offsetEnd {
		ts := &NITDataTransportStream{}

		ts.TransportStreamID = uint16(r.TryReadBits(16))

		ts.OriginalNetworkID = uint16(r.TryReadBits(16))

		_ = r.TryReadBits(4)
		if ts.TransportDescriptors, err = parseDescriptors(r); err != nil {
			return nil, fmt.Errorf("parsing descriptors failed: %w", err)
		}

		d.TransportStreams = append(d.TransportStreams, ts)
	}
	return d, r.TryError
}
