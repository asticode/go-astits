package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

var psi = &PSIData{
	PointerField: 4,
	Sections: []*PSISection{
		{
			CRC32: uint32(0x7ffc6102),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          30,
				SectionSyntaxIndicator: true,
				TableID:                78,
				TableType:              PSITableTypeEIT,
			},
			Syntax: &PSISectionSyntax{
				Data:   &PSISectionSyntaxData{EIT: eit},
				Header: psiSectionSyntaxHeader,
			},
		},
		{
			CRC32: uint32(0xfebaa941),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          25,
				SectionSyntaxIndicator: true,
				TableID:                64,
				TableType:              PSITableTypeNIT,
			},
			Syntax: &PSISectionSyntax{
				Data:   &PSISectionSyntaxData{NIT: nit},
				Header: psiSectionSyntaxHeader,
			},
		},
		{
			CRC32: uint32(0x60739f61),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          17,
				SectionSyntaxIndicator: true,
				TableID:                0,
				TableType:              PSITableTypePAT,
			},
			Syntax: &PSISectionSyntax{
				Data:   &PSISectionSyntaxData{PAT: pat},
				Header: psiSectionSyntaxHeader,
			},
		},
		{
			CRC32: uint32(0xc68442e8),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          24,
				SectionSyntaxIndicator: true,
				TableID:                2,
				TableType:              PSITableTypePMT,
			},
			Syntax: &PSISectionSyntax{
				Data:   &PSISectionSyntaxData{PMT: pmt},
				Header: psiSectionSyntaxHeader,
			},
		},
		{
			CRC32: uint32(0xef3751d6),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          20,
				SectionSyntaxIndicator: true,
				TableID:                66,
				TableType:              PSITableTypeSDT,
			},
			Syntax: &PSISectionSyntax{
				Data:   &PSISectionSyntaxData{SDT: sdt},
				Header: psiSectionSyntaxHeader,
			},
		},
		{
			CRC32: uint32(0x6969b13),
			Header: &PSISectionHeader{
				PrivateBit:             true,
				SectionLength:          14,
				SectionSyntaxIndicator: true,
				TableID:                115,
				TableType:              PSITableTypeTOT,
			},
			Syntax: &PSISectionSyntax{
				Data: &PSISectionSyntaxData{TOT: tot},
			},
		},
		{Header: &PSISectionHeader{TableID: 254, TableType: PSITableTypeUnknown}},
	},
}

func psiBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(4))                      // Pointer field
	w.Write([]byte("test"))                // Pointer field bytes
	w.Write(uint8(78))                     // EIT table ID
	w.Write("1")                           // EIT syntax section indicator
	w.Write("1")                           // EIT private bit
	w.Write("11")                          // EIT reserved
	w.Write("000000011110")                // EIT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // EIT syntax section header
	w.Write(eitBytes())                    // EIT data
	w.Write(uint32(0x7ffc6102))            // EIT CRC32
	w.Write(uint8(64))                     // NIT table ID
	w.Write("1")                           // NIT syntax section indicator
	w.Write("1")                           // NIT private bit
	w.Write("11")                          // NIT reserved
	w.Write("000000011001")                // NIT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // NIT syntax section header
	w.Write(nitBytes())                    // NIT data
	w.Write(uint32(0xfebaa941))            // NIT CRC32
	w.Write(uint8(0))                      // PAT table ID
	w.Write("1")                           // PAT syntax section indicator
	w.Write("1")                           // PAT private bit
	w.Write("11")                          // PAT reserved
	w.Write("000000010001")                // PAT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // PAT syntax section header
	w.Write(patBytes())                    // PAT data
	w.Write(uint32(0x60739f61))            // PAT CRC32
	w.Write(uint8(2))                      // PMT table ID
	w.Write("1")                           // PMT syntax section indicator
	w.Write("1")                           // PMT private bit
	w.Write("11")                          // PMT reserved
	w.Write("000000011000")                // PMT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // PMT syntax section header
	w.Write(pmtBytes())                    // PMT data
	w.Write(uint32(0xc68442e8))            // PMT CRC32
	w.Write(uint8(66))                     // SDT table ID
	w.Write("1")                           // SDT syntax section indicator
	w.Write("1")                           // SDT private bit
	w.Write("11")                          // SDT reserved
	w.Write("000000010100")                // SDT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // SDT syntax section header
	w.Write(sdtBytes())                    // SDT data
	w.Write(uint32(0xef3751d6))            // SDT CRC32
	w.Write(uint8(115))                    // TOT table ID
	w.Write("1")                           // TOT syntax section indicator
	w.Write("1")                           // TOT private bit
	w.Write("11")                          // TOT reserved
	w.Write("000000001110")                // TOT section length
	w.Write(totBytes())                    // TOT data
	w.Write(uint32(0x6969b13))             // TOT CRC32
	w.Write(uint8(254))                    // Unknown table ID
	w.Write(uint8(0))                      // PAT table ID
	return buf.Bytes()
}

func TestParsePSIData(t *testing.T) {
	// Invalid CRC32
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(0))       // Pointer field
	w.Write(uint8(115))     // TOT table ID
	w.Write("1")            // TOT syntax section indicator
	w.Write("1")            // TOT private bit
	w.Write("11")           // TOT reserved
	w.Write("000000001110") // TOT section length
	w.Write(totBytes())     // TOT data
	w.Write(uint32(32))     // TOT CRC32
	_, err := parsePSIData(astikit.NewBytesIterator(buf.Bytes()))
	assert.EqualError(t, err, "astits: parsing PSI table failed: astits: Table CRC32 20 != computed CRC32 6969b13")

	// Valid
	d, err := parsePSIData(astikit.NewBytesIterator(psiBytes()))
	assert.NoError(t, err)
	assert.Equal(t, d, psi)
}

var psiSectionHeader = &PSISectionHeader{
	PrivateBit:             true,
	SectionLength:          2730,
	SectionSyntaxIndicator: true,
	TableID:                0,
	TableType:              PSITableTypePAT,
}

func psiSectionHeaderBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(0))       // Table ID
	w.Write("1")            // Syntax section indicator
	w.Write("1")            // Private bit
	w.Write("11")           // Reserved
	w.Write("101010101010") // Section length
	return buf.Bytes()
}

func TestParsePSISectionHeader(t *testing.T) {
	// Unknown table type
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint8(254)) // Table ID
	w.Write("1")        // Syntax section indicator
	w.Write("0000000")  // Finish the byte
	d, _, _, _, _, err := parsePSISectionHeader(astikit.NewBytesIterator(buf.Bytes()))
	assert.Equal(t, d, &PSISectionHeader{
		TableID:   254,
		TableType: PSITableTypeUnknown,
	})
	assert.NoError(t, err)

	// Valid table type
	d, offsetStart, offsetSectionsStart, offsetSectionsEnd, offsetEnd, err := parsePSISectionHeader(astikit.NewBytesIterator(psiSectionHeaderBytes()))
	assert.Equal(t, d, psiSectionHeader)
	assert.Equal(t, 0, offsetStart)
	assert.Equal(t, 3, offsetSectionsStart)
	assert.Equal(t, 2729, offsetSectionsEnd)
	assert.Equal(t, 2733, offsetEnd)
	assert.NoError(t, err)
}

func TestPSITableType(t *testing.T) {
	for i := PSITableIDEITStart; i <= PSITableIDEITEnd; i++ {
		assert.Equal(t, PSITableTypeEIT, i.Type())
	}
	assert.Equal(t, PSITableTypeDIT, PSITableIDDIT.Type())
	assert.Equal(t, PSITableTypeNIT, PSITableIDNITVariant1.Type())
	assert.Equal(t, PSITableTypeNIT, PSITableIDNITVariant2.Type())
	assert.Equal(t, PSITableTypeSDT, PSITableIDSDTVariant1.Type())
	assert.Equal(t, PSITableTypeSDT, PSITableIDSDTVariant2.Type())

	assert.Equal(t, PSITableTypeBAT, PSITableIDBAT.Type())
	assert.Equal(t, PSITableTypeNull, PSITableIDNull.Type())
	assert.Equal(t, PSITableTypePAT, PSITableIDPAT.Type())
	assert.Equal(t, PSITableTypePMT, PSITableIDPMT.Type())
	assert.Equal(t, PSITableTypeRST, PSITableIDRST.Type())
	assert.Equal(t, PSITableTypeSIT, PSITableIDSIT.Type())
	assert.Equal(t, PSITableTypeST, PSITableIDST.Type())
	assert.Equal(t, PSITableTypeTDT, PSITableIDTDT.Type())
	assert.Equal(t, PSITableTypeTOT, PSITableIDTOT.Type())
	assert.Equal(t, PSITableTypeUnknown, PSITableID(1).Type())
}

var psiSectionSyntaxHeader = &PSISectionSyntaxHeader{
	CurrentNextIndicator: true,
	LastSectionNumber:    3,
	SectionNumber:        2,
	TableIDExtension:     1,
	VersionNumber:        21,
}

func psiSectionSyntaxHeaderBytes() []byte {
	buf := &bytes.Buffer{}
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: buf})
	w.Write(uint16(1)) // Table ID extension
	w.Write("11")      // Reserved bits
	w.Write("10101")   // Version number
	w.Write("1")       // Current/next indicator
	w.Write(uint8(2))  // Section number
	w.Write(uint8(3))  // Last section number
	return buf.Bytes()
}

func TestParsePSISectionSyntaxHeader(t *testing.T) {
	h, err := parsePSISectionSyntaxHeader(astikit.NewBytesIterator(psiSectionSyntaxHeaderBytes()))
	assert.Equal(t, psiSectionSyntaxHeader, h)
	assert.NoError(t, err)
}

func TestPSIToData(t *testing.T) {
	p := &Packet{}
	assert.Equal(t, []*DemuxerData{
		{EIT: eit, FirstPacket: p, PID: 2},
		{FirstPacket: p, NIT: nit, PID: 2},
		{FirstPacket: p, PAT: pat, PID: 2},
		{FirstPacket: p, PMT: pmt, PID: 2},
		{FirstPacket: p, SDT: sdt, PID: 2},
		{FirstPacket: p, TOT: tot, PID: 2},
	}, psi.toData(p, uint16(2)))
}

type psiDataTestCase struct {
	name      string
	bytesFunc func(*astikit.BitsWriter)
	data      *PSIData
}

var psiDataTestCases = []psiDataTestCase{
	{
		"PAT",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(4))                      // Pointer field
			w.Write([]byte{0, 0, 0, 0})            // Pointer field bytes
			w.Write(uint8(0))                      // PAT table ID
			w.Write("1")                           // PAT syntax section indicator
			w.Write("1")                           // PAT private bit
			w.Write("11")                          // PAT reserved
			w.Write("000000010001")                // PAT section length
			w.Write(psiSectionSyntaxHeaderBytes()) // PAT syntax section header
			w.Write(patBytes())                    // PAT data
			w.Write(uint32(0x60739f61))            // PAT CRC32
		},
		&PSIData{
			PointerField: 4,
			Sections: []*PSISection{
				{
					CRC32: uint32(0x60739f61),
					Header: &PSISectionHeader{
						PrivateBit:             true,
						SectionLength:          17,
						SectionSyntaxIndicator: true,
						TableID:                0,
						TableType:              PSITableTypePAT,
					},
					Syntax: &PSISectionSyntax{
						Data:   &PSISectionSyntaxData{PAT: pat},
						Header: psiSectionSyntaxHeader,
					},
				},
			},
		},
	},
	{
		"PMT",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(4))                      // Pointer field
			w.Write([]byte{0, 0, 0, 0})            // Pointer field bytes
			w.Write(uint8(2))                      // PMT table ID
			w.Write("1")                           // PMT syntax section indicator
			w.Write("1")                           // PMT private bit
			w.Write("11")                          // PMT reserved
			w.Write("000000011000")                // PMT section length
			w.Write(psiSectionSyntaxHeaderBytes()) // PMT syntax section header
			w.Write(pmtBytes())                    // PMT data
			w.Write(uint32(0xc68442e8))            // PMT CRC32
		},
		&PSIData{
			PointerField: 4,
			Sections: []*PSISection{
				{
					CRC32: uint32(0xc68442e8),
					Header: &PSISectionHeader{
						PrivateBit:             true,
						SectionLength:          24,
						SectionSyntaxIndicator: true,
						TableID:                2,
						TableType:              PSITableTypePMT,
					},
					Syntax: &PSISectionSyntax{
						Data:   &PSISectionSyntaxData{PMT: pmt},
						Header: psiSectionSyntaxHeader,
					},
				},
			},
		},
	},
}

func TestWritePSIData(t *testing.T) {
	for _, tc := range psiDataTestCases {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})
			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

			tc.bytesFunc(wExpected)

			n, err := writePSIData(wActual, tc.data)
			assert.NoError(t, err)
			assert.Equal(t, bufExpected.Len(), n)
			assert.Equal(t, n, bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

func BenchmarkParsePSIData(b *testing.B) {
	pb := psiBytes()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parsePSIData(astikit.NewBytesIterator(pb))
	}
}
