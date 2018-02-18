package astits

import (
	"github.com/pkg/errors"
)

// PIDs
const (
	PIDPAT  = 0x0    // Program Association Table (PAT) contains a directory listing of all Program Map Tables.
	PIDCAT  = 0x1    // Conditional Access Table (CAT) contains a directory listing of all ITU-T Rec. H.222 entitlement management message streams used by Program Map Tables.
	PIDTSDT = 0x2    // Transport Stream Description Table (TSDT) contains descriptors related to the overall transport stream
	PIDNull = 0x1fff // Null Packet (used for fixed bandwidth padding)
)

// Data represents a data
type Data struct {
	EIT *EITData
	NIT *NITData
	PAT *PATData
	PES *PESData
	PID uint16
	PMT *PMTData
	SDT *SDTData
	TOT *TOTData
}

// parseData parses a payload spanning over multiple packets and returns a set of data
func parseData(ps []*Packet, prs PacketsParser, pm programMap) (ds []*Data, err error) {
	// Use custom parser first
	if prs != nil {
		var skip bool
		if ds, skip, err = prs(ps); err != nil {
			err = errors.Wrap(err, "astits: custom packets parsing failed")
			return
		} else if skip {
			return
		}
	}

	// Reconstruct payload
	var l int
	for _, p := range ps {
		l += len(p.Payload)
	}
	var payload = make([]byte, l)
	var c int
	for _, p := range ps {
		c += copy(payload[c:], p.Payload)
	}

	// Parse PID
	var pid = ps[0].Header.PID

	// Parse payload
	if pid == PIDCAT {
		// Information in a CAT payload is private and dependent on the CA system. Use the PacketsParser
		// to parse this type of payload
	} else if isPSIPayload(pid, pm) {
		var psiData *PSIData
		if psiData, err = parsePSIData(payload); err != nil {
			err = errors.Wrap(err, "astits: parsing PSI data failed")
			return
		}
		ds = psiData.toData(pid)
	} else if isPESPayload(payload) {
		d, err := parsePESData(payload)
		if err == nil {
			ds = append(ds, &Data{PES: d, PID: pid})
		}
	}
	return
}

// isPSIPayload checks whether the payload is a PSI one
func isPSIPayload(pid uint16, pm programMap) bool {
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
