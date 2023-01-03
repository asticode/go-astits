package astits

import (
	"encoding/binary"
	"fmt"

	"github.com/asticode/go-astikit"
)

// PIDs
const (
	PIDPAT  uint16 = 0x0    // Program Association Table (PAT) contains a directory listing of all Program Map Tables.
	PIDCAT  uint16 = 0x1    // Conditional Access Table (CAT) contains a directory listing of all ITU-T Rec. H.222 entitlement management message streams used by Program Map Tables.
	PIDTSDT uint16 = 0x2    // Transport Stream Description Table (TSDT) contains descriptors related to the overall transport stream
	PIDNull uint16 = 0x1fff // Null Packet (used for fixed bandwidth padding)
)

// DemuxerData represents a data parsed by Demuxer
type DemuxerData struct {
	EIT         *EITData
	FirstPacket *Packet
	NIT         *NITData
	PAT         *PATData
	PES         *PESData
	PID         uint16
	PMT         *PMTData
	SDT         *SDTData
	TOT         *TOTData
}

// MuxerData represents a data to be written by Muxer
type MuxerData struct {
	PID             uint16
	AdaptationField *PacketAdaptationField
	PES             *PESData
}

// parseData parses a payload spanning over multiple packets and returns a set of data
func parseData(ps []*Packet, prs PacketsParser, pm *programMap) (ds []*DemuxerData, err error) {
	// Use custom parser first
	if prs != nil {
		var skip bool
		if ds, skip, err = prs(ps); err != nil {
			err = fmt.Errorf("astits: custom packets parsing failed: %w", err)
			return
		} else if skip {
			return
		}
	}

	// Get payload length
	var l int
	for i := range ps {
		l += len(ps[i].Payload)
	}

	// Append payload
	var payload []byte

	if len(ps) > 1 {
		payload = make([]byte, l)
		o := copy(payload, ps[0].Payload)
		for i := range ps[1:] {
			o += copy(payload[o:], ps[i+1].Payload)
		}
	} else {
		payload = ps[0].Payload
	}

	// Create reader
	i := astikit.NewBytesIterator(payload)

	// Parse PID
	pid := ps[0].Header.PID

	// Parse payload
	if pid == PIDCAT {
		// Information in a CAT payload is private and dependent on the CA system. Use the PacketsParser
		// to parse this type of payload
	} else if isPSIPayload(pid, pm) {
		// Parse PSI data
		var psiData *PSIData
		if psiData, err = parsePSIData(i); err != nil {
			err = fmt.Errorf("astits: parsing PSI data failed: %w", err)
			return
		}

		// Append data
		ds = psiData.toData(ps[0], pid)
	} else if isPESPayload(payload) {
		// Parse PES data
		var pesData *PESData
		if pesData, err = parsePESData(i); err != nil {
			err = fmt.Errorf("astits: parsing PES data failed: %w", err)
			return
		}

		// Append data
		ds = append(ds, &DemuxerData{
			FirstPacket: ps[0],
			PES:         pesData,
			PID:         pid,
		})
	}
	return
}

// isPSIPayload checks whether the payload is a PSI one
func isPSIPayload(pid uint16, pm *programMap) bool {
	return pid == PIDPAT || // PAT
		pm.exists(pid) || // PMT
		((pid >= 0x10 && pid <= 0x14) || (pid >= 0x1e && pid <= 0x1f)) //DVB
}

// isPESPayload checks whether the payload is a PES one
func isPESPayload(i []byte) bool {
	// Packet is not big enough
	if len(i) < 3 {
		return false
	}

	// Check prefix
	return uint32(i[0])<<16|uint32(i[1])<<8|uint32(i[2]) == 1
}

// isPSIComplete checks whether we have sufficient amount of packets to parse PSI
func isPSIComplete(ps []*Packet, prs PacketsParser) bool {
	// Use custom parser first
	if prs != nil {
		if _, skip, err := prs(ps); err != nil {
			return false
		} else if skip {
			return true
		}
	}

	// Get payload length
	var l int
	for i := range ps {
		l += len(ps[i].Payload)
	}

	// Append payload
	var payload []byte

	if len(ps) > 1 {
		payload = make([]byte, l)
		o := copy(payload, ps[0].Payload)
		for i := range ps[1:] {
			o += copy(payload[o:], ps[i+1].Payload)
		}
	} else {
		payload = ps[0].Payload
	}

	// Create reader
	i := astikit.NewBytesIterator(payload)

	// Get next byte
	b, err := i.NextByte()
	if err != nil {
		return false
	}

	// Pointer filler bytes
	i.Skip(int(b))

	for i.HasBytesLeft() {

		// Get PSI table ID
		b, err = i.NextByte()
		if err != nil {
			return false
		}

		// Check whether we need to stop the parsing
		if shouldStopPSIParsing(PSITableID(b)) {
			break
		}

		// Get PSI section length
		var bs []byte
		bs, err = i.NextBytesNoCopy(2)
		if err != nil {
			return false
		}

		i.Skip(int(binary.BigEndian.Uint16(bs) & 0x0fff))
	}

	return i.Len() >= i.Offset()
}
