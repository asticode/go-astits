package astits

import (
	"fmt"

	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// PSI table IDs
const (
	PSITableTypeBAT     = "BAT"
	PSITableTypeDIT     = "DIT"
	PSITableTypeEIT     = "EIT"
	PSITableTypeNIT     = "NIT"
	PSITableTypeNull    = "Null"
	PSITableTypePAT     = "PAT"
	PSITableTypePMT     = "PMT"
	PSITableTypeRST     = "RST"
	PSITableTypeSDT     = "SDT"
	PSITableTypeSIT     = "SIT"
	PSITableTypeST      = "ST"
	PSITableTypeTDT     = "TDT"
	PSITableTypeTOT     = "TOT"
	PSITableTypeUnknown = "Unknown"
)

// PSIData represents a PSI data
// https://en.wikipedia.org/wiki/Program-specific_information
type PSIData struct {
	PointerField int // Present at the start of the TS packet payload signaled by the payload_unit_start_indicator bit in the TS header. Used to set packet alignment bytes or content before the start of tabled payload data.
	Sections     []*PSISection
}

// PSISection represents a PSI section
type PSISection struct {
	CRC32  uint32 // A checksum of the entire table excluding the pointer field, pointer filler bytes and the trailing CRC32.
	Header *PSISectionHeader
	Syntax *PSISectionSyntax
}

// PSISectionHeader represents a PSI section header
type PSISectionHeader struct {
	PrivateBit             bool   // The PAT, PMT, and CAT all set this to 0. Other tables set this to 1.
	SectionLength          uint16 // The number of bytes that follow for the syntax section (with CRC value) and/or table data. These bytes must not exceed a value of 1021.
	SectionSyntaxIndicator bool   // A flag that indicates if the syntax section follows the section length. The PAT, PMT, and CAT all set this to 1.
	TableID                int    // Table Identifier, that defines the structure of the syntax section and other contained data. As an exception, if this is the byte that immediately follow previous table section and is set to 0xFF, then it indicates that the repeat of table section end here and the rest of TS data payload shall be stuffed with 0xFF. Consequently the value 0xFF shall not be used for the Table Identifier.
	TableType              string
}

// PSISectionSyntax represents a PSI section syntax
type PSISectionSyntax struct {
	Data   *PSISectionSyntaxData
	Header *PSISectionSyntaxHeader
}

// PSISectionSyntaxHeader represents a PSI section syntax header
type PSISectionSyntaxHeader struct {
	CurrentNextIndicator bool   // Indicates if data is current in effect or is for future use. If the bit is flagged on, then the data is to be used at the present moment.
	LastSectionNumber    uint8  // This indicates which table is the last table in the sequence of tables.
	SectionNumber        uint8  // This is an index indicating which table this is in a related sequence of tables. The first table starts from 0.
	TableIDExtension     uint16 // Informational only identifier. The PAT uses this for the transport stream identifier and the PMT uses this for the Program number.
	VersionNumber        uint8  // Syntax version number. Incremented when data is changed and wrapped around on overflow for values greater than 32.
}

// PSISectionSyntaxData represents a PSI section syntax data
type PSISectionSyntaxData struct {
	EIT *EITData
	NIT *NITData
	PAT *PATData
	PMT *PMTData
	SDT *SDTData
	TOT *TOTData
}

// parsePSIData parses a PSI data
func parsePSIData(i []byte) (d *PSIData, err error) {
	// Init data
	d = &PSIData{}
	var offset int

	// Pointer field
	d.PointerField = int(i[offset])
	offset += 1

	// Pointer filler bytes
	offset += d.PointerField

	// Parse sections
	var s *PSISection
	var stop bool
	for offset < len(i) && !stop {
		if s, stop, err = parsePSISection(i, &offset); err != nil {
			err = errors.Wrap(err, "astits: parsing PSI table failed")
			return
		}
		d.Sections = append(d.Sections, s)
	}
	return
}

// parsePSISection parses a PSI section
func parsePSISection(i []byte, offset *int) (s *PSISection, stop bool, err error) {
	// Init section
	s = &PSISection{}

	// Parse header
	var offsetStart, offsetSectionsEnd, offsetEnd int
	s.Header, offsetStart, _, offsetSectionsEnd, offsetEnd = parsePSISectionHeader(i, offset)

	// Check whether we need to stop the parsing
	if shouldStopPSIParsing(s.Header.TableType) {
		stop = true
		return
	}

	// Check whether there's a syntax section
	if s.Header.SectionLength > 0 && s.Header.SectionSyntaxIndicator {
		// Parse syntax
		s.Syntax = parsePSISectionSyntax(i, offset, s.Header, offsetSectionsEnd)

		// Process CRC32
		if hasCRC32(s.Header.TableType) {
			// Parse CRC32
			s.CRC32 = parseCRC32(i[offsetSectionsEnd:offsetEnd])
			*offset += 4

			// Check CRC32
			var c = computeCRC32(i[offsetStart:offsetSectionsEnd])
			if c != s.CRC32 {
				err = fmt.Errorf("astits: Table CRC32 %x != computed CRC32 %x", s.CRC32, c)
				return
			}
		}
	}
	return
}

// parseCRC32 parses a CRC32
func parseCRC32(i []byte) uint32 {
	return uint32(i[len(i)-4])<<24 | uint32(i[len(i)-3])<<16 | uint32(i[len(i)-2])<<8 | uint32(i[len(i)-1])
}

// computeCRC32 computes a CRC32
// https://stackoverflow.com/questions/35034042/how-to-calculate-crc32-in-psi-si-packet
func computeCRC32(i []byte) (o uint32) {
	o = uint32(0xffffffff)
	for _, b := range i {
		for i := 0; i < 8; i++ {
			if (o >= uint32(0x80000000)) != (b >= uint8(0x80)) {
				o = (o << 1) ^ 0x04C11DB7
			} else {
				o = o << 1
			}
			b <<= 1
		}
	}
	return
}

// shouldStopPSIParsing checks whether the PSI parsing should be stopped
func shouldStopPSIParsing(tableType string) bool {
	return tableType == PSITableTypeNull || tableType == PSITableTypeUnknown
}

// parsePSISectionHeader parses a PSI section header
func parsePSISectionHeader(i []byte, offset *int) (h *PSISectionHeader, offsetStart, offsetSectionsStart, offsetSectionsEnd, offsetEnd int) {
	// Init
	h = &PSISectionHeader{}
	offsetStart = *offset

	// Table ID
	h.TableID = int(i[*offset])
	*offset += 1

	// Table type
	h.TableType = psiTableType(h.TableID)

	// Check whether we need to stop the parsing
	if shouldStopPSIParsing(h.TableType) {
		return
	}

	// Section syntax indicator
	h.SectionSyntaxIndicator = i[*offset]&0x80 > 0

	// Private bit
	h.PrivateBit = i[*offset]&0x40 > 0

	// Section length
	h.SectionLength = uint16(i[*offset]&0xf)<<8 | uint16(i[*offset+1])
	*offset += 2

	// Offsets
	offsetSectionsStart = *offset
	offsetEnd = offsetSectionsStart + int(h.SectionLength)
	offsetSectionsEnd = offsetEnd
	if hasCRC32(h.TableType) {
		offsetSectionsEnd -= 4
	}
	return
}

// hasCRC32 checks whether the table has a CRC32
func hasCRC32(tableType string) bool {
	return tableType == PSITableTypePAT ||
		tableType == PSITableTypePMT ||
		tableType == PSITableTypeEIT ||
		tableType == PSITableTypeNIT ||
		tableType == PSITableTypeTOT ||
		tableType == PSITableTypeSDT
}

// psiTableType returns the psi table type based on the table id
func psiTableType(tableID int) string {
	switch {
	case tableID == 0x4a:
		return PSITableTypeBAT
	case tableID >= 0x4e && tableID <= 0x6f:
		return PSITableTypeEIT
	case tableID == 0x7e:
		return PSITableTypeDIT
	case tableID == 0x40, tableID == 0x41:
		return PSITableTypeNIT
	case tableID == 0xff:
		return PSITableTypeNull
	case tableID == 0:
		return PSITableTypePAT
	case tableID == 2:
		return PSITableTypePMT
	case tableID == 0x71:
		return PSITableTypeRST
	case tableID == 0x42, tableID == 0x46:
		return PSITableTypeSDT
	case tableID == 0x7f:
		return PSITableTypeSIT
	case tableID == 0x72:
		return PSITableTypeST
	case tableID == 0x70:
		return PSITableTypeTDT
	case tableID == 0x73:
		return PSITableTypeTOT
	}
	// TODO Remove this log
	astilog.Debugf("unlisted PSI table ID %d", tableID)
	return PSITableTypeUnknown
}

// parsePSISectionSyntax parses a PSI section syntax
func parsePSISectionSyntax(i []byte, offset *int, h *PSISectionHeader, offsetSectionsEnd int) (s *PSISectionSyntax) {
	// Init
	s = &PSISectionSyntax{}

	// Header
	if hasPSISyntaxHeader(h.TableType) {
		s.Header = parsePSISectionSyntaxHeader(i, offset)
	}

	// Parse data
	s.Data = parsePSISectionSyntaxData(i, offset, h, s.Header, offsetSectionsEnd)
	return
}

// hasPSISyntaxHeader checks whether the section has a syntax header
func hasPSISyntaxHeader(tableType string) bool {
	return tableType == PSITableTypeEIT ||
		tableType == PSITableTypeNIT ||
		tableType == PSITableTypePAT ||
		tableType == PSITableTypePMT ||
		tableType == PSITableTypeSDT
}

// parsePSISectionSyntaxHeader parses a PSI section syntax header
func parsePSISectionSyntaxHeader(i []byte, offset *int) (h *PSISectionSyntaxHeader) {
	// Init
	h = &PSISectionSyntaxHeader{}

	// Table ID extension
	h.TableIDExtension = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	*offset += 2

	// Version number
	h.VersionNumber = uint8(i[*offset]&0x3f) >> 1

	// Current/Next indicator
	h.CurrentNextIndicator = i[*offset]&0x1 > 0
	*offset += 1

	// Section number
	h.SectionNumber = uint8(i[*offset])
	*offset += 1

	// Last section number
	h.LastSectionNumber = uint8(i[*offset])
	*offset += 1
	return
}

// parsePSISectionSyntaxData parses a PSI section data
func parsePSISectionSyntaxData(i []byte, offset *int, h *PSISectionHeader, sh *PSISectionSyntaxHeader, offsetSectionsEnd int) (d *PSISectionSyntaxData) {
	// Init
	d = &PSISectionSyntaxData{}

	// Switch on table type
	switch h.TableType {
	case PSITableTypeBAT:
		// TODO Parse BAT
	case PSITableTypeDIT:
		// TODO Parse DIT
	case PSITableTypeEIT:
		d.EIT = parseEITSection(i, offset, offsetSectionsEnd, sh.TableIDExtension)
	case PSITableTypeNIT:
		d.NIT = parseNITSection(i, offset, sh.TableIDExtension)
	case PSITableTypePAT:
		d.PAT = parsePATSection(i, offset, offsetSectionsEnd, sh.TableIDExtension)
	case PSITableTypePMT:
		d.PMT = parsePMTSection(i, offset, offsetSectionsEnd, sh.TableIDExtension)
	case PSITableTypeRST:
		// TODO Parse RST
	case PSITableTypeSDT:
		d.SDT = parseSDTSection(i, offset, offsetSectionsEnd, sh.TableIDExtension)
	case PSITableTypeSIT:
		// TODO Parse SIT
	case PSITableTypeST:
		// TODO Parse ST
	case PSITableTypeTOT:
		d.TOT = parseTOTSection(i, offset)
	case PSITableTypeTDT:
		// TODO Parse TDT
	}
	return
}

// toData parses the PSI tables and returns a set of Data
func (d *PSIData) toData(pid uint16) (ds []*Data) {
	// Loop through sections
	for _, s := range d.Sections {
		// Switch on table type
		switch s.Header.TableType {
		case PSITableTypeEIT:
			ds = append(ds, &Data{EIT: s.Syntax.Data.EIT, PID: pid})
		case PSITableTypeNIT:
			ds = append(ds, &Data{NIT: s.Syntax.Data.NIT, PID: pid})
		case PSITableTypePAT:
			ds = append(ds, &Data{PAT: s.Syntax.Data.PAT, PID: pid})
		case PSITableTypePMT:
			ds = append(ds, &Data{PID: pid, PMT: s.Syntax.Data.PMT})
		case PSITableTypeSDT:
			ds = append(ds, &Data{PID: pid, SDT: s.Syntax.Data.SDT})
		case PSITableTypeTOT:
			ds = append(ds, &Data{PID: pid, TOT: s.Syntax.Data.TOT})
		}
	}
	return
}
