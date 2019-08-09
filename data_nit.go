package astits

import (
	astibyte "github.com/asticode/go-astitools/byte"
	"github.com/pkg/errors"
)

// NITData represents a NIT data
// Page: 29 | Chapter: 5.2.1 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type NITData struct {
	NetworkDescriptors []*Descriptor
	NetworkID          uint16
	TransportStreams   []*NITDataTransportStream
}

// NITDataTransportStream represents a NIT data transport stream
type NITDataTransportStream struct {
	OriginalNetworkID    uint16
	TransportDescriptors []*Descriptor
	TransportStreamID    uint16
}

// parseNITSection parses a NIT section
func parseNITSection(i *astibyte.Iterator, tableIDExtension uint16) (d *NITData, err error) {
	// Create data
	d = &NITData{NetworkID: tableIDExtension}

	// Network descriptors
	if d.NetworkDescriptors, err = parseDescriptors(i); err != nil {
		err = errors.Wrap(err, "astits: parsing descriptors failed")
		return
	}

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = errors.Wrap(err, "astits: fetching next bytes failed")
		return
	}

	// Transport stream loop length
	transportStreamLoopLength := int(uint16(bs[0]&0xf)<<8 | uint16(bs[1]))

	// Transport stream loop
	offsetEnd := i.Offset() + transportStreamLoopLength
	for i.Offset() < offsetEnd {
		// Create transport stream
		ts := &NITDataTransportStream{}

		// Get next bytes
		if bs, err = i.NextBytes(2); err != nil {
			err = errors.Wrap(err, "astits: fetching next bytes failed")
			return
		}

		// Transport stream ID
		ts.TransportStreamID = uint16(bs[0])<<8 | uint16(bs[1])

		// Get next bytes
		if bs, err = i.NextBytes(2); err != nil {
			err = errors.Wrap(err, "astits: fetching next bytes failed")
			return
		}

		// Original network ID
		ts.OriginalNetworkID = uint16(bs[0])<<8 | uint16(bs[1])

		// Transport descriptors
		if ts.TransportDescriptors, err = parseDescriptors(i); err != nil {
			err = errors.Wrap(err, "astits: parsing descriptors failed")
			return
		}

		// Append transport stream
		d.TransportStreams = append(d.TransportStreams, ts)
	}
	return
}
