package astits

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
func parseNITSection(i []byte, offset *int, tableIDExtension uint16) (d *NITData) {
	// Init
	d = &NITData{NetworkID: tableIDExtension}

	// Network descriptors
	d.NetworkDescriptors = parseDescriptors(i, offset)

	// Transport stream loop length
	var transportStreamLoopLength = int(uint16(i[*offset]&0xf)<<8 | uint16(i[*offset+1]))
	*offset += 2

	// Transport stream loop
	transportStreamLoopLength += *offset
	for *offset < transportStreamLoopLength {
		// Transport stream ID
		var ts = &NITDataTransportStream{}
		ts.TransportStreamID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
		*offset += 2

		// Original network ID
		ts.OriginalNetworkID = uint16(i[*offset])<<8 | uint16(i[*offset+1])
		*offset += 2

		// Transport descriptors
		ts.TransportDescriptors = parseDescriptors(i, offset)

		// Append transport stream
		d.TransportStreams = append(d.TransportStreams, ts)
	}
	return
}
