package astits

import (
	"bytes"
	"fmt"

	"github.com/icza/bitio"
)

// PIDs.
const (
	// Program Association Table (PAT) contains a
	// directory listing of all Program Map Tables.
	PIDPAT uint16 = 0x0

	// Conditional Access Table (CAT) contains
	// a directory listing of all ITU-T Rec.
	// H.222 entitlement management message
	// streams used by Program Map Tables.
	PIDCAT uint16 = 0x1

	// Transport Stream Description Table (TSDT) contains
	// descriptors related to the overall transport stream.
	PIDTSDT uint16 = 0x2

	// Null Packet (used for fixed bandwidth padding).
	PIDNull uint16 = 0x1fff
)

// DemuxerData represents a data parsed by Demuxer.
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

// MuxerData represents a data to be written by Muxer.
type MuxerData struct {
	PID             uint16
	AdaptationField *PacketAdaptationField
	PES             *PESData
}

// parseData parses a payload spanning over
// multiple packets and returns a set of data.
func parseData(
	pkts []*Packet,
	prs PacketsParser,
	pm *programMap,
) ([]*DemuxerData, error) {
	// Use custom parser first
	var ds []*DemuxerData
	if prs != nil {
		data, skip, err := prs(pkts)
		if err != nil {
			return nil, fmt.Errorf("custom packets parsing failed: %w", err)
		}
		if skip {
			return data, nil
		}
		ds = data
	}

	var payloadLength int64
	for _, p := range pkts {
		payloadLength += int64(len(p.Payload) * 8)
	}
	payload := make([]byte, payloadLength/8)
	var n int
	for _, pkt := range pkts {
		n += copy(payload[n:], pkt.Payload)
	}

	pid := pkts[0].Header.PID

	// Parse payload
	if pid == PIDCAT {
		// Information in a CAT payload is private and dependent on the CA system.
		// Use the PacketsParser to parse this type of payload.
		return ds, nil
	}

	r := bitio.NewCountReader(bytes.NewReader(payload))

	if isPSIPayload(pid, pm) {
		psiData, err := parsePSIData(r)
		if err != nil {
			return nil, fmt.Errorf("parsing PSI data failed: %w", err)
		}
		ds = psiData.toData(pkts[0], pid)
		return ds, nil
	}

	if isPESPayload(payload) {
		pesData, err := parsePESData(r, payloadLength)
		if err != nil {
			return nil, fmt.Errorf("parsing PES data failed: %w", err)
		}
		ds = append(ds, &DemuxerData{
			FirstPacket: pkts[0],
			PES:         pesData,
			PID:         pid,
		})
	}

	return ds, nil
}

// isPSIPayload checks whether the payload is a PSI one.
func isPSIPayload(pid uint16, pm *programMap) bool {
	return pid == PIDPAT || // PAT
		pm.exists(pid) || // PMT
		((pid >= 0x10 && pid <= 0x14) || (pid >= 0x1e && pid <= 0x1f)) // DVB
}

// isPESPayload checks whether the payload is a PES one.
func isPESPayload(i []byte) bool {
	// Packet is not big enough.
	if len(i) < 3 {
		return false
	}

	// Check prefix
	return uint32(i[0])<<16|uint32(i[1])<<8|uint32(i[2]) == 1
}
