package astits

import (
	"fmt"

	"github.com/asticode/go-astikit"
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
const (
	PSITableTypeIdPAT  = 0x00
	PSITableTypeIdPMT  = 0x02
	PSITableTypeIdBAT  = 0x4a
	PSITableTypeIdDIT  = 0x7e
	PSITableTypeIdRST  = 0x71
	PSITableTypeIdSIT  = 0x7f
	PSITableTypeIdST   = 0x72
	PSITableTypeIdTDT  = 0x70
	PSITableTypeIdTOT  = 0x73
	PSITableTypeIdNull = 0xff

	PSITableTypeIdEITStart    = 0x4e
	PSITableTypeIdEITEnd      = 0x6f
	PSITableTypeIdSDTVariant1 = 0x42
	PSITableTypeIdSDTVariant2 = 0x46
	PSITableTypeIdNITVariant1 = 0x40
	PSITableTypeIdNITVariant2 = 0x41
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
func parsePSIData(i *astikit.BytesIterator) (d *PSIData, err error) {
	// Init data
	d = &PSIData{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Pointer field
	d.PointerField = int(b)

	// Pointer filler bytes
	i.Skip(d.PointerField)

	// Parse sections
	var s *PSISection
	var stop bool
	for i.HasBytesLeft() && !stop {
		if s, stop, err = parsePSISection(i); err != nil {
			err = fmt.Errorf("astits: parsing PSI table failed: %w", err)
			return
		}
		d.Sections = append(d.Sections, s)
	}
	return
}

// parsePSISection parses a PSI section
func parsePSISection(i *astikit.BytesIterator) (s *PSISection, stop bool, err error) {
	// Init section
	s = &PSISection{}

	// Parse header
	var offsetStart, offsetSectionsEnd, offsetEnd int
	if s.Header, offsetStart, _, offsetSectionsEnd, offsetEnd, err = parsePSISectionHeader(i); err != nil {
		err = fmt.Errorf("astits: parsing PSI section header failed: %w", err)
		return
	}

	// Check whether we need to stop the parsing
	if shouldStopPSIParsing(s.Header.TableType) {
		stop = true
		return
	}

	// Check whether there's a syntax section
	if s.Header.SectionLength > 0 {
		// Parse syntax
		if s.Syntax, err = parsePSISectionSyntax(i, s.Header, offsetSectionsEnd); err != nil {
			err = fmt.Errorf("astits: parsing PSI section syntax failed: %w", err)
			return
		}

		// Process CRC32
		if hasCRC32(s.Header.TableType) {
			// Seek to the end of the sections
			i.Seek(offsetSectionsEnd)

			// Parse CRC32
			if s.CRC32, err = parseCRC32(i); err != nil {
				err = fmt.Errorf("astits: parsing CRC32 failed: %w", err)
				return
			}

			// Get CRC32 data
			i.Seek(offsetStart)
			var crc32Data []byte
			if crc32Data, err = i.NextBytes(offsetSectionsEnd - offsetStart); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}

			// Compute CRC32
			crc32 := computeCRC32(crc32Data)

			// Check CRC32
			if crc32 != s.CRC32 {
				err = fmt.Errorf("astits: Table CRC32 %x != computed CRC32 %x", s.CRC32, crc32)
				return
			}
		}
	}

	// Seek to the end of the section
	i.Seek(offsetEnd)
	return
}

// parseCRC32 parses a CRC32
func parseCRC32(i *astikit.BytesIterator) (c uint32, err error) {
	var bs []byte
	if bs, err = i.NextBytes(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	c = uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])
	return
}

// shouldStopPSIParsing checks whether the PSI parsing should be stopped
func shouldStopPSIParsing(tableType string) bool {
	return tableType == PSITableTypeNull || tableType == PSITableTypeUnknown
}

// parsePSISectionHeader parses a PSI section header
func parsePSISectionHeader(i *astikit.BytesIterator) (h *PSISectionHeader, offsetStart, offsetSectionsStart, offsetSectionsEnd, offsetEnd int, err error) {
	// Init
	h = &PSISectionHeader{}
	offsetStart = i.Offset()

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Table ID
	h.TableID = int(b)

	// Table type
	h.TableType = psiTableType(h.TableID)

	// Check whether we need to stop the parsing
	if shouldStopPSIParsing(h.TableType) {
		return
	}

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Section syntax indicator
	h.SectionSyntaxIndicator = bs[0]&0x80 > 0

	// Private bit
	h.PrivateBit = bs[0]&0x40 > 0

	// Section length
	h.SectionLength = uint16(bs[0]&0xf)<<8 | uint16(bs[1])

	// Offsets
	offsetSectionsStart = i.Offset()
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
// Page: 28 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
// (barbashov) the link above can be broken, alternative: https://dvb.org/wp-content/uploads/2019/12/a038_tm1217r37_en300468v1_17_1_-_rev-134_-_si_specification.pdf
func psiTableType(tableID int) string {
	switch {
	case tableID == PSITableTypeIdBAT:
		return PSITableTypeBAT
	case tableID >= PSITableTypeIdEITStart && tableID <= PSITableTypeIdEITEnd:
		return PSITableTypeEIT
	case tableID == PSITableTypeIdDIT:
		return PSITableTypeDIT
	case tableID == PSITableTypeIdNITVariant1, tableID == PSITableTypeIdNITVariant2:
		return PSITableTypeNIT
	case tableID == PSITableTypeIdNull:
		return PSITableTypeNull
	case tableID == PSITableTypeIdPAT:
		return PSITableTypePAT
	case tableID == PSITableTypeIdPMT:
		return PSITableTypePMT
	case tableID == PSITableTypeIdRST:
		return PSITableTypeRST
	case tableID == PSITableTypeIdSDTVariant1, tableID == PSITableTypeIdSDTVariant2:
		return PSITableTypeSDT
	case tableID == PSITableTypeIdSIT:
		return PSITableTypeSIT
	case tableID == PSITableTypeIdST:
		return PSITableTypeST
	case tableID == PSITableTypeIdTDT:
		return PSITableTypeTDT
	case tableID == PSITableTypeIdTOT:
		return PSITableTypeTOT
	default:
		return PSITableTypeUnknown
	}
}

// parsePSISectionSyntax parses a PSI section syntax
func parsePSISectionSyntax(i *astikit.BytesIterator, h *PSISectionHeader, offsetSectionsEnd int) (s *PSISectionSyntax, err error) {
	// Init
	s = &PSISectionSyntax{}

	// Header
	if hasPSISyntaxHeader(h.TableType) {
		if s.Header, err = parsePSISectionSyntaxHeader(i); err != nil {
			err = fmt.Errorf("astits: parsing PSI section syntax header failed: %w", err)
			return
		}
	}

	// Parse data
	if s.Data, err = parsePSISectionSyntaxData(i, h, s.Header, offsetSectionsEnd); err != nil {
		err = fmt.Errorf("astits: parsing PSI section syntax data failed: %w", err)
		return
	}
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
func parsePSISectionSyntaxHeader(i *astikit.BytesIterator) (h *PSISectionSyntaxHeader, err error) {
	// Init
	h = &PSISectionSyntaxHeader{}

	// Get next 2 bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Table ID extension
	h.TableIDExtension = uint16(bs[0])<<8 | uint16(bs[1])

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Version number
	h.VersionNumber = uint8(b&0x3f) >> 1

	// Current/Next indicator
	h.CurrentNextIndicator = b&0x1 > 0

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Section number
	h.SectionNumber = uint8(b)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Last section number
	h.LastSectionNumber = uint8(b)
	return
}

// parsePSISectionSyntaxData parses a PSI section data
func parsePSISectionSyntaxData(i *astikit.BytesIterator, h *PSISectionHeader, sh *PSISectionSyntaxHeader, offsetSectionsEnd int) (d *PSISectionSyntaxData, err error) {
	// Init
	d = &PSISectionSyntaxData{}

	// Switch on table type
	switch h.TableType {
	case PSITableTypeBAT:
		// TODO Parse BAT
	case PSITableTypeDIT:
		// TODO Parse DIT
	case PSITableTypeEIT:
		if d.EIT, err = parseEITSection(i, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			err = fmt.Errorf("astits: parsing EIT section failed: %w", err)
			return
		}
	case PSITableTypeNIT:
		if d.NIT, err = parseNITSection(i, sh.TableIDExtension); err != nil {
			err = fmt.Errorf("astits: parsing NIT section failed: %w", err)
			return
		}
	case PSITableTypePAT:
		if d.PAT, err = parsePATSection(i, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			err = fmt.Errorf("astits: parsing PAT section failed: %w", err)
			return
		}
	case PSITableTypePMT:
		if d.PMT, err = parsePMTSection(i, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			err = fmt.Errorf("astits: parsing PMT section failed: %w", err)
			return
		}
	case PSITableTypeRST:
		// TODO Parse RST
	case PSITableTypeSDT:
		if d.SDT, err = parseSDTSection(i, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			err = fmt.Errorf("astits: parsing PMT section failed: %w", err)
			return
		}
	case PSITableTypeSIT:
		// TODO Parse SIT
	case PSITableTypeST:
		// TODO Parse ST
	case PSITableTypeTOT:
		if d.TOT, err = parseTOTSection(i); err != nil {
			err = fmt.Errorf("astits: parsing TOT section failed: %w", err)
			return
		}
	case PSITableTypeTDT:
		// TODO Parse TDT
	}
	return
}

// toData parses the PSI tables and returns a set of Data
func (d *PSIData) toData(firstPacket *Packet, pid uint16) (ds []*Data) {
	// Loop through sections
	for _, s := range d.Sections {
		// Switch on table type
		switch s.Header.TableType {
		case PSITableTypeEIT:
			ds = append(ds, &Data{EIT: s.Syntax.Data.EIT, FirstPacket: firstPacket, PID: pid})
		case PSITableTypeNIT:
			ds = append(ds, &Data{FirstPacket: firstPacket, NIT: s.Syntax.Data.NIT, PID: pid})
		case PSITableTypePAT:
			ds = append(ds, &Data{FirstPacket: firstPacket, PAT: s.Syntax.Data.PAT, PID: pid})
		case PSITableTypePMT:
			ds = append(ds, &Data{FirstPacket: firstPacket, PID: pid, PMT: s.Syntax.Data.PMT})
		case PSITableTypeSDT:
			ds = append(ds, &Data{FirstPacket: firstPacket, PID: pid, SDT: s.Syntax.Data.SDT})
		case PSITableTypeTOT:
			ds = append(ds, &Data{FirstPacket: firstPacket, PID: pid, TOT: s.Syntax.Data.TOT})
		}
	}
	return
}

func writePSIData(w *astikit.BitsWriter, d *PSIData) (int, error) {
	b := astikit.NewBitsWriterBatch(w)
	b.Write(uint8(d.PointerField))
	for i := 0; i < d.PointerField; i++ {
		b.Write(uint8(0x00))
	}

	bytesWritten := 1 + d.PointerField

	if err := b.Err(); err != nil {
		return 0, err
	}

	for _, s := range d.Sections {
		s.Header.TableType = psiTableType(s.Header.TableID)
		n, err := writePSISection(w, s)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

func calcPSISectionLength(s *PSISection) uint16 {
	ret := uint16(0)
	if hasPSISyntaxHeader(s.Header.TableType) {
		ret += 5 // PSI syntax header length
	}

	switch s.Header.TableType {
	case PSITableTypePAT:
		ret += calcPATSectionLength(s.Syntax.Data.PAT)
	case PSITableTypePMT:
		ret += calcPMTSectionLength(s.Syntax.Data.PMT)
	}

	if hasCRC32(s.Header.TableType) {
		ret += 4
	}

	return ret
}

func writePSISection(w *astikit.BitsWriter, s *PSISection) (int, error) {
	if s.Header.TableType != PSITableTypePAT && s.Header.TableType != PSITableTypePMT {
		return 0, fmt.Errorf("writePSISection: table %s is not implemented", s.Header.TableType)
	}

	b := astikit.NewBitsWriterBatch(w)

	sectionLength := calcPSISectionLength(s)
	sectionCRC32 := CRC32Polynomial

	if hasCRC32(s.Header.TableType) {
		w.SetWriteCallback(func(bs []byte) {
			sectionCRC32 = updateCRC32(sectionCRC32, bs)
		})
		defer w.SetWriteCallback(nil)
	}

	b.Write(uint8(s.Header.TableID))
	b.Write(s.Header.SectionSyntaxIndicator)
	b.Write(s.Header.PrivateBit)
	b.WriteN(uint8(0xff), 2)
	b.WriteN(sectionLength, 12)
	bytesWritten := 3

	if s.Header.SectionLength > 0 {
		n, err := writePSISectionSyntax(w, s)
		if err != nil {
			return 0, err
		}
		bytesWritten += n

		if hasCRC32(s.Header.TableType) {
			b.Write(sectionCRC32)
			bytesWritten += 4
		}
	}

	return bytesWritten, b.Err()
}

func writePSISectionSyntax(w *astikit.BitsWriter, s *PSISection) (int, error) {
	bytesWritten := 0
	if hasPSISyntaxHeader(s.Header.TableType) {
		n, err := writePSISectionSyntaxHeader(w, s.Syntax.Header)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	n, err := writePSISectionSyntaxData(w, s.Syntax.Data, s.Header.TableType)
	if err != nil {
		return 0, err
	}
	bytesWritten += n

	return bytesWritten, nil
}

func writePSISectionSyntaxHeader(w *astikit.BitsWriter, h *PSISectionSyntaxHeader) (int, error) {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(h.TableIDExtension)
	b.WriteN(uint8(0xff), 2)
	b.WriteN(h.VersionNumber, 5)
	b.Write(h.CurrentNextIndicator)
	b.Write(h.SectionNumber)
	b.Write(h.LastSectionNumber)

	return 5, b.Err()
}

func writePSISectionSyntaxData(w *astikit.BitsWriter, d *PSISectionSyntaxData, tableType string) (int, error) {
	switch tableType {
	// TODO write other table types
	case PSITableTypePAT:
		return writePATSection(w, d.PAT)
	case PSITableTypePMT:
		return writePMTSection(w, d.PMT)
	}

	return 0, nil
}
