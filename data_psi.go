package astits

import (
	"errors"
	"fmt"

	"github.com/icza/bitio"
)

// PSI table IDs.
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

// PSITableID .
type PSITableID uint16

// PSITableIDs.
const (
	PSITableIDPAT  PSITableID = 0x00
	PSITableIDPMT  PSITableID = 0x02
	PSITableIDBAT  PSITableID = 0x4a
	PSITableIDDIT  PSITableID = 0x7e
	PSITableIDRST  PSITableID = 0x71
	PSITableIDSIT  PSITableID = 0x7f
	PSITableIDST   PSITableID = 0x72
	PSITableIDTDT  PSITableID = 0x70
	PSITableIDTOT  PSITableID = 0x73
	PSITableIDNull PSITableID = 0xff

	PSITableIDEITStart    PSITableID = 0x4e
	PSITableIDEITEnd      PSITableID = 0x6f
	PSITableIDSDTVariant1 PSITableID = 0x42
	PSITableIDSDTVariant2 PSITableID = 0x46
	PSITableIDNITVariant1 PSITableID = 0x40
	PSITableIDNITVariant2 PSITableID = 0x41
)

// PSIData represents a PSI data.
// https://en.wikipedia.org/wiki/Program-specific_information
type PSIData struct {
	// PointerField it present at the start of the TS packet
	// payload signaled by the payload_unit_start_indicator
	// bit in the TS header. Used to set packet alignment
	// bytes or content before the start of tabled payload data.
	PointerField int
	Sections     []*PSISection
}

// PSISection represents a PSI section.
type PSISection struct {
	// CRC32 checksum of the entire table excluding the pointer
	// field, pointer filler bytes and the trailing CRC32.
	CRC32  uint32
	Header *PSISectionHeader
	Syntax *PSISectionSyntax
}

// PSISectionHeader represents a PSI section header.
type PSISectionHeader struct {
	// PrivateBit The PAT, PMT, and CAT all set this to -1.
	// Other tables set this to 1.
	PrivateBit bool

	// SectionLength The number of bytes that follow for the
	// syntax section (with CRC value) and/or table data.
	// These bytes must not exceed a value of 1021.
	SectionLength uint16

	// A flag that indicates if the syntax section
	// follows the section length. The PAT, PMT,
	// and CAT all set this to 1.
	SectionSyntaxIndicator bool

	// TableID that defines the structure of the syntax
	// section and other contained data. As an exception,
	// if this is the byte that immediately follow previous
	// table section and is set to 0xFF, then it indicates
	// that the repeat of table section end here and the rest of
	// TS data payload shall be stuffed with 0xFF. Consequently
	// the value 0xFF shall not be used for the Table Identifier.
	TableID PSITableID

	TableType string
}

// PSISectionSyntax represents a PSI section syntax.
type PSISectionSyntax struct {
	Data   *PSISectionSyntaxData
	Header *PSISectionSyntaxHeader
}

// PSISectionSyntaxHeader represents a PSI section syntax header.
type PSISectionSyntaxHeader struct {
	// TableIDExtension Informational only identifier.
	// The PAT uses this for the transport stream identifier
	// and the PMT uses this for the Program number.
	TableIDExtension uint16

	// VersionNumber Syntax version number.
	// Incremented when data is changed and wrapped
	// around on overflow for values greater than 32.
	VersionNumber uint8 // 5 bits.

	// CurrentNextIndicator Indicates if data is current in
	// effect or is for future use. If the bit is flagged on,
	// then the data is to be used at the present moment.
	CurrentNextIndicator bool

	// LastSectionNumber indicates which table is
	// the last table in the sequence of tables.
	LastSectionNumber uint8

	// SectionNumber is an index indicating which table
	// this is in a related sequence of tables.
	// The first table starts from 0.
	SectionNumber uint8
}

// PSISectionSyntaxData represents a PSI section syntax data.
type PSISectionSyntaxData struct {
	EIT *EITData
	NIT *NITData
	PAT *PATData
	PMT *PMTData
	SDT *SDTData
	TOT *TOTData
}

// parsePSIData parses a PSI data.
func parsePSIData(r *bitio.CountReader) (*PSIData, error) {
	d := &PSIData{}

	d.PointerField = int(r.TryReadByte())

	// Pointer filler bytes.
	skip := make([]byte, d.PointerField)
	TryReadFull(r, skip)

	var s *PSISection
	var stop bool
	var err error
	for !stop {
		if s, stop, err = parsePSISection(r); err != nil {
			err = fmt.Errorf("parsing PSI table failed: %w", err)
			return nil, err
		}
		d.Sections = append(d.Sections, s)
	}
	return d, r.TryError
}

// ErrPSIInvalidCRC32 .
var ErrPSIInvalidCRC32 = errors.New("computed CRC32 doesn't match table CRC32")

// parsePSISection parses a PSI section.
func parsePSISection(i *bitio.CountReader) (*PSISection, bool, error) {
	cr := NewCRC32Reader(i)
	r := bitio.NewCountReader(cr)
	r.BitsCount = i.BitsCount

	s := &PSISection{}

	header, offsetSectionsEnd, offsetEnd, err := parsePSISectionHeader(r)
	if err != nil {
		return nil, false, fmt.Errorf("parsing PSI section header failed: %w", err)
	}
	s.Header = header

	// Check whether we need to stop the parsing.
	if shouldStopPSIParsing(s.Header.TableID) {
		return s, true, nil
	}

	// Check whether there's a syntax section.
	if s.Header.SectionLength <= 0 {
		// Go to the end of the section.
		if offsetEnd > r.BitsCount {
			skip := make([]byte, (offsetEnd-r.BitsCount)/8)
			TryReadFull(r, skip)
		}
		return s, false, nil
	}

	if s.Syntax, err = parsePSISectionSyntax(r, s.Header, offsetSectionsEnd); err != nil {
		return nil, false, fmt.Errorf("parsing PSI section syntax failed: %w", err)
	}

	if s.Header.TableID.hasCRC32() {
		computedCRC32 := cr.CRC32()

		// Go to the end of the sections.
		if offsetSectionsEnd > r.BitsCount {
			skip := make([]byte, (offsetSectionsEnd-r.BitsCount)/8)
			TryReadFull(r, skip)
		}

		if s.CRC32, err = parseCRC32(r); err != nil {
			return nil, false, fmt.Errorf("parsing table CRC32 failed: %w", err)
		}

		if computedCRC32 != s.CRC32 {
			return nil, false, fmt.Errorf("%w computed=%v table=%v",
				ErrPSIInvalidCRC32, computedCRC32, s.CRC32)
		}
	}

	if offsetEnd > r.BitsCount {
		skip := make([]byte, (offsetEnd-r.BitsCount)/8)
		TryReadFull(r, skip)
	}

	return s, false, r.TryError
}

// parseCRC32 parses a CRC32.
func parseCRC32(r *bitio.CountReader) (uint32, error) {
	c := uint32(r.TryReadBits(32))
	return c, r.TryError
}

// shouldStopPSIParsing checks whether the PSI parsing should be stopped.
func shouldStopPSIParsing(tableID PSITableID) bool {
	return tableID == PSITableIDNull ||
		tableID.isUnknown()
}

// parsePSISectionHeader parses a PSI section header.
func parsePSISectionHeader(r *bitio.CountReader) (
	h *PSISectionHeader,
	offsetSectionsEnd,
	offsetEnd int64,
	err error,
) {
	h = &PSISectionHeader{}

	tableID := r.TryReadByte()
	h.TableID = PSITableID(tableID)

	h.TableType = h.TableID.Type()

	// Check whether we need to stop the parsing.
	if shouldStopPSIParsing(h.TableID) {
		return
	}

	h.SectionSyntaxIndicator = r.TryReadBool()
	h.PrivateBit = r.TryReadBool()
	_ = r.TryReadBits(2) // Reserved.
	h.SectionLength = uint16(r.TryReadBits(12))

	// Offsets
	offsetSectionsStart := r.BitsCount
	offsetEnd = offsetSectionsStart + int64(h.SectionLength*8)
	offsetSectionsEnd = offsetEnd
	if h.TableID.hasCRC32() {
		offsetSectionsEnd -= 4 * 8
	}

	return h, offsetSectionsEnd, offsetEnd, r.TryError
}

// Type returns the psi table type based on the table id.
// Page: 28 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
func (t PSITableID) Type() string {
	switch {
	case t == PSITableIDBAT:
		return PSITableTypeBAT
	case t >= PSITableIDEITStart && t <= PSITableIDEITEnd:
		return PSITableTypeEIT
	case t == PSITableIDDIT:
		return PSITableTypeDIT
	case t == PSITableIDNITVariant1, t == PSITableIDNITVariant2:
		return PSITableTypeNIT
	case t == PSITableIDNull:
		return PSITableTypeNull
	case t == PSITableIDPAT:
		return PSITableTypePAT
	case t == PSITableIDPMT:
		return PSITableTypePMT
	case t == PSITableIDRST:
		return PSITableTypeRST
	case t == PSITableIDSDTVariant1, t == PSITableIDSDTVariant2:
		return PSITableTypeSDT
	case t == PSITableIDSIT:
		return PSITableTypeSIT
	case t == PSITableIDST:
		return PSITableTypeST
	case t == PSITableIDTDT:
		return PSITableTypeTDT
	case t == PSITableIDTOT:
		return PSITableTypeTOT
	default:
		return PSITableTypeUnknown
	}
}

// hasPSISyntaxHeader checks whether the section has a syntax header.
func (t PSITableID) hasPSISyntaxHeader() bool {
	return t == PSITableIDPAT ||
		t == PSITableIDPMT ||
		t == PSITableIDNITVariant1 || t == PSITableIDNITVariant2 ||
		t == PSITableIDSDTVariant1 || t == PSITableIDSDTVariant2 ||
		(t >= PSITableIDEITStart && t <= PSITableIDEITEnd)
}

// hasCRC32 checks whether the table has a CRC32.
func (t PSITableID) hasCRC32() bool {
	return t == PSITableIDPAT ||
		t == PSITableIDPMT ||
		t == PSITableIDTOT ||
		t == PSITableIDNITVariant1 || t == PSITableIDNITVariant2 ||
		t == PSITableIDSDTVariant1 || t == PSITableIDSDTVariant2 ||
		(t >= PSITableIDEITStart && t <= PSITableIDEITEnd)
}

func (t PSITableID) isUnknown() bool {
	switch t {
	case PSITableIDBAT,
		PSITableIDDIT,
		PSITableIDNITVariant1, PSITableIDNITVariant2,
		PSITableIDNull,
		PSITableIDPAT,
		PSITableIDPMT,
		PSITableIDRST,
		PSITableIDSDTVariant1, PSITableIDSDTVariant2,
		PSITableIDSIT,
		PSITableIDST,
		PSITableIDTDT,
		PSITableIDTOT:
		return false
	}
	if t >= PSITableIDEITStart && t <= PSITableIDEITEnd {
		return false
	}
	return true
}

// parsePSISectionSyntax parses a PSI section syntax.
func parsePSISectionSyntax(
	r *bitio.CountReader,
	h *PSISectionHeader,
	offsetSectionsEnd int64,
) (*PSISectionSyntax, error) {
	s := &PSISectionSyntax{}
	var err error

	if h.TableID.hasPSISyntaxHeader() {
		s.Header, err = parsePSISectionSyntaxHeader(r)
		if err != nil {
			return nil, fmt.Errorf("parsing PSI section syntax header failed: %w", err)
		}
	}

	s.Data, err = parsePSISectionSyntaxData(r, h, s.Header, offsetSectionsEnd)
	if err != nil {
		return nil, fmt.Errorf("parsing PSI section syntax data failed: %w", err)
	}

	return s, nil
}

// parsePSISectionSyntaxHeader parses a PSI section syntax header.
func parsePSISectionSyntaxHeader(r *bitio.CountReader) (*PSISectionSyntaxHeader, error) {
	h := &PSISectionSyntaxHeader{}

	h.TableIDExtension = uint16(r.TryReadBits(16))

	_ = r.TryReadBits(2) // Reserved.
	h.VersionNumber = uint8(r.TryReadBits(5))
	h.CurrentNextIndicator = r.TryReadBool()

	h.SectionNumber = r.TryReadByte()

	h.LastSectionNumber = r.TryReadByte()
	return h, r.TryError
}

// parsePSISectionSyntaxData parses a PSI section data.
func parsePSISectionSyntaxData(
	r *bitio.CountReader,
	h *PSISectionHeader,
	sh *PSISectionSyntaxHeader,
	offsetSectionsEnd int64,
) (*PSISectionSyntaxData, error) {
	d := &PSISectionSyntaxData{}
	var err error

	// Switch on table type.
	switch h.TableID {
	case PSITableIDBAT:
		// TODO Parse BAT.
	case PSITableIDDIT:
		// TODO Parse DIT.
	case PSITableIDNITVariant1, PSITableIDNITVariant2:
		if d.NIT, err = parseNITSection(r, sh.TableIDExtension); err != nil {
			return nil, fmt.Errorf("parsing NIT section failed: %w", err)
		}
	case PSITableIDPAT:
		if d.PAT, err = parsePATSection(r, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			return nil, fmt.Errorf("parsing PAT section failed: %w", err)
		}
	case PSITableIDPMT:
		if d.PMT, err = parsePMTSection(r, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			return nil, fmt.Errorf("parsing PMT section failed: %w", err)
		}
	case PSITableIDRST:
		// TODO Parse RST.
	case PSITableIDSDTVariant1, PSITableIDSDTVariant2:
		if d.SDT, err = parseSDTSection(r, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			return nil, fmt.Errorf("parsing PMT section failed: %w", err)
		}
	case PSITableIDSIT:
		// TODO Parse SIT.
	case PSITableIDST:
		// TODO Parse ST.
	case PSITableIDTOT:
		if d.TOT, err = parseTOTSection(r); err != nil {
			return nil, fmt.Errorf("parsing TOT section failed: %w", err)
		}
	case PSITableIDTDT:
		// TODO Parse TDT.
	}

	if h.TableID >= PSITableIDEITStart && h.TableID <= PSITableIDEITEnd {
		if d.EIT, err = parseEITSection(r, offsetSectionsEnd, sh.TableIDExtension); err != nil {
			return nil, fmt.Errorf("parsing EIT section failed: %w", err)
		}
	}

	return d, nil
}

// toData parses the PSI tables and returns a set of DemuxerData.
func (d *PSIData) toData(firstPacket *Packet, pid uint16) (ds []*DemuxerData) {
	// Loop through sections.
	for _, s := range d.Sections {
		// Switch on table type.
		switch s.Header.TableID {
		case PSITableIDNITVariant1, PSITableIDNITVariant2:
			ds = append(ds, &DemuxerData{FirstPacket: firstPacket, NIT: s.Syntax.Data.NIT, PID: pid})
		case PSITableIDPAT:
			ds = append(ds, &DemuxerData{FirstPacket: firstPacket, PAT: s.Syntax.Data.PAT, PID: pid})
		case PSITableIDPMT:
			ds = append(ds, &DemuxerData{FirstPacket: firstPacket, PID: pid, PMT: s.Syntax.Data.PMT})
		case PSITableIDSDTVariant1, PSITableIDSDTVariant2:
			ds = append(ds, &DemuxerData{FirstPacket: firstPacket, PID: pid, SDT: s.Syntax.Data.SDT})
		case PSITableIDTOT:
			ds = append(ds, &DemuxerData{FirstPacket: firstPacket, PID: pid, TOT: s.Syntax.Data.TOT})
		}
		if s.Header.TableID >= PSITableIDEITStart && s.Header.TableID <= PSITableIDEITEnd {
			data := &DemuxerData{
				EIT:         s.Syntax.Data.EIT,
				FirstPacket: firstPacket,
				PID:         pid,
			}
			ds = append(ds, data)
		}
	}
	return
}

func writePSIData(w *bitio.Writer, d *PSIData) error {
	w.TryWriteByte(uint8(d.PointerField))
	for i := 0; i < d.PointerField; i++ {
		w.TryWriteByte(0x00)
	}

	bytesWritten := 1 + d.PointerField

	if err := w.TryError; err != nil {
		return fmt.Errorf("write: %w", w.TryError)
	}

	for _, s := range d.Sections {
		n, err := writePSISection(w, s)
		if err != nil {
			return fmt.Errorf("writing PSI sections failed: %w", err)
		}
		bytesWritten += n
	}

	return nil
}

func calcPSISectionLength(s *PSISection) uint16 {
	ret := uint16(0)
	if s.Header.TableID.hasPSISyntaxHeader() {
		ret += 5 // PSI syntax header length.
	}

	switch s.Header.TableID {
	case PSITableIDPAT:
		ret += calcPATSectionLength(s.Syntax.Data.PAT)
	case PSITableIDPMT:
		ret += calcPMTSectionLength(s.Syntax.Data.PMT)
	}

	if s.Header.TableID.hasCRC32() {
		ret += 4
	}

	return ret
}

// ErrPSIUnsupportedTable .
var ErrPSIUnsupportedTable = errors.New("unsupported table")

func writePSISection(w *bitio.Writer, s *PSISection) (int, error) {
	if s.Header.TableID != PSITableIDPAT && s.Header.TableID != PSITableIDPMT {
		return 0, fmt.Errorf("%w: %s", ErrPSIUnsupportedTable, s.Header.TableID.Type())
	}

	sectionLength := calcPSISectionLength(s)

	var cw *CRC32Writer

	if s.Header.TableID.hasCRC32() {
		cw = NewCRC32Writer(w)
		w = bitio.NewWriter(cw)
	}

	w.TryWriteByte(uint8(s.Header.TableID))

	w.TryWriteBool(s.Header.SectionSyntaxIndicator)
	w.TryWriteBool(s.Header.PrivateBit)
	w.TryWriteBits(0xff, 2)
	w.TryWriteBits(uint64(sectionLength), 12)
	bytesWritten := 3

	if s.Header.SectionLength > 0 {
		n, err := writePSISectionSyntax(w, s)
		if err != nil {
			return 0, fmt.Errorf("writing PSI section syntax failed: %w", err)
		}
		bytesWritten += n

		if s.Header.TableID.hasCRC32() {
			w.TryWriteBits(uint64(cw.CRC32()), 32)
			bytesWritten += 4
		}
	}

	return bytesWritten, w.TryError
}

func writePSISectionSyntax(w *bitio.Writer, s *PSISection) (int, error) {
	bytesWritten := 0
	if s.Header.TableID.hasPSISyntaxHeader() {
		n, err := writePSISectionSyntaxHeader(w, s.Syntax.Header)
		if err != nil {
			return 0, fmt.Errorf("header: %w", err)
		}
		bytesWritten += n
	}

	n, err := writePSISectionSyntaxData(w, s.Syntax.Data, s.Header.TableID)
	if err != nil {
		return 0, fmt.Errorf("data: %w", err)
	}
	bytesWritten += n

	return bytesWritten, nil
}

func writePSISectionSyntaxHeader(w *bitio.Writer, h *PSISectionSyntaxHeader) (int, error) {
	w.TryWriteBits(uint64(h.TableIDExtension), 16)

	w.TryWriteBits(0xff, 2) // Reserved.
	w.TryWriteBits(uint64(h.VersionNumber), 5)
	w.TryWriteBool(h.CurrentNextIndicator)

	w.TryWriteByte(h.SectionNumber)

	w.TryWriteByte(h.LastSectionNumber)

	return 5, w.TryError
}

func writePSISectionSyntaxData(w *bitio.Writer, d *PSISectionSyntaxData, tableID PSITableID) (int, error) {
	switch tableID {
	// TODO write other table types.
	case PSITableIDPAT:
		return writePATSection(w, d.PAT)
	case PSITableIDPMT:
		return writePMTSection(w, d.PMT)
	}

	return 0, nil
}
