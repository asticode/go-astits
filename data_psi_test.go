package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
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
	w := bitio.NewWriter(buf)
	w.WriteByte(4)                         // Pointer field
	w.Write([]byte("test"))                // Pointer field bytes
	w.WriteByte(78)                        // EIT table ID
	WriteBinary(w, "1")                    // EIT syntax section indicator
	WriteBinary(w, "1")                    // EIT private bit
	WriteBinary(w, "11")                   // EIT reserved
	WriteBinary(w, "000000011110")         // EIT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // EIT syntax section header
	w.Write(eitBytes())                    // EIT data
	w.WriteBits(0x7ffc6102, 32)            // EIT CRC32
	w.WriteByte(64)                        // NIT table ID
	WriteBinary(w, "1")                    // NIT syntax section indicator
	WriteBinary(w, "1")                    // NIT private bit
	WriteBinary(w, "11")                   // NIT reserved
	WriteBinary(w, "000000011001")         // NIT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // NIT syntax section header
	w.Write(nitBytes())                    // NIT data
	w.WriteBits(0xfebaa941, 32)            // NIT CRC32
	w.WriteByte(0)                         // PAT table ID
	WriteBinary(w, "1")                    // PAT syntax section indicator
	WriteBinary(w, "1")                    // PAT private bit
	WriteBinary(w, "11")                   // PAT reserved
	WriteBinary(w, "000000010001")         // PAT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // PAT syntax section header
	w.Write(patBytes())                    // PAT data
	w.WriteBits(0x60739f61, 32)            // PAT CRC32
	w.WriteByte(2)                         // PMT table ID
	WriteBinary(w, "1")                    // PMT syntax section indicator
	WriteBinary(w, "1")                    // PMT private bit
	WriteBinary(w, "11")                   // PMT reserved
	WriteBinary(w, "000000011000")         // PMT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // PMT syntax section header
	w.Write(pmtBytes())                    // PMT data
	w.WriteBits(0xc68442e8, 32)            // PMT CRC32
	w.WriteByte(66)                        // SDT table ID
	WriteBinary(w, "1")                    // SDT syntax section indicator
	WriteBinary(w, "1")                    // SDT private bit
	WriteBinary(w, "11")                   // SDT reserved
	WriteBinary(w, "000000010100")         // SDT section length
	w.Write(psiSectionSyntaxHeaderBytes()) // SDT syntax section header
	w.Write(sdtBytes())                    // SDT data
	w.WriteBits(0xef3751d6, 32)            // SDT CRC32
	w.WriteByte(115)                       // TOT table ID
	WriteBinary(w, "1")                    // TOT syntax section indicator
	WriteBinary(w, "1")                    // TOT private bit
	WriteBinary(w, "11")                   // TOT reserved
	WriteBinary(w, "000000001110")         // TOT section length
	w.Write(totBytes())                    // TOT data
	w.WriteBits(0x6969b13, 32)             // TOT CRC32
	w.WriteByte(254)                       // Unknown table ID
	w.WriteByte(0)                         // PAT table ID
	return buf.Bytes()
}

func TestParsePSIData(t *testing.T) {
	// Invalid CRC32
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(0)                 // Pointer field
	w.WriteByte(115)               // TOT table ID
	WriteBinary(w, "1")            // TOT syntax section indicator
	WriteBinary(w, "1")            // TOT private bit
	WriteBinary(w, "11")           // TOT reserved
	WriteBinary(w, "000000001110") // TOT section length
	w.Write(totBytes())            // TOT data
	w.WriteBits(32, 32)            // TOT CRC32

	r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
	_, err := parsePSIData(r)
	assert.ErrorIs(t, err, ErrPSIInvalidCRC32)

	// Valid
	r = bitio.NewCountReader(bytes.NewReader(psiBytes()))
	d, err := parsePSIData(r)
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
	w := bitio.NewWriter(buf)
	w.WriteByte(0)                 // Table ID
	WriteBinary(w, "1")            // Syntax section indicator
	WriteBinary(w, "1")            // Private bit
	WriteBinary(w, "11")           // Reserved
	WriteBinary(w, "101010101010") // Section length
	return buf.Bytes()
}

func TestParsePSISectionHeader(t *testing.T) {
	// Unknown table type
	buf := &bytes.Buffer{}
	w := bitio.NewWriter(buf)
	w.WriteByte(254)          // Table ID
	WriteBinary(w, "1")       // Syntax section indicator
	WriteBinary(w, "0000000") // Finish the byte
	r := bitio.NewCountReader(bytes.NewReader(buf.Bytes()))
	d, _, _, err := parsePSISectionHeader(r)
	assert.Equal(t, d, &PSISectionHeader{
		TableID:   254,
		TableType: PSITableTypeUnknown,
	})
	assert.NoError(t, err)

	// Valid table type
	r = bitio.NewCountReader(bytes.NewReader(psiSectionHeaderBytes()))
	d, offsetSectionsEnd, offsetEnd, err := parsePSISectionHeader(r)
	assert.Equal(t, d, psiSectionHeader)
	assert.Equal(t, int64(2729*8), offsetSectionsEnd)
	assert.Equal(t, int64(2733*8), offsetEnd)
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
	w := bitio.NewWriter(buf)
	w.WriteBits(1, 16)      // Table ID extension
	WriteBinary(w, "11")    // Reserved bits
	WriteBinary(w, "10101") // Version number
	WriteBinary(w, "1")     // Current/next indicator
	w.WriteByte(2)          // Section number
	w.WriteByte(3)          // Last section number
	return buf.Bytes()
}

func TestParsePSISectionSyntaxHeader(t *testing.T) {
	r := bitio.NewCountReader(bytes.NewReader(psiSectionSyntaxHeaderBytes()))
	h, err := parsePSISectionSyntaxHeader(r)
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
	bytesFunc func(*bitio.Writer)
	data      *PSIData
}

var psiDataTestCases = []psiDataTestCase{
	{
		"PAT",
		func(w *bitio.Writer) {
			w.WriteByte(4)                         // Pointer field
			w.Write([]byte{0, 0, 0, 0})            // Pointer field bytes
			w.WriteByte(0)                         // PAT table ID
			WriteBinary(w, "1")                    // PAT syntax section indicator
			WriteBinary(w, "1")                    // PAT private bit
			WriteBinary(w, "11")                   // PAT reserved
			WriteBinary(w, "000000010001")         // PAT section length
			w.Write(psiSectionSyntaxHeaderBytes()) // PAT syntax section header
			w.Write(patBytes())                    // PAT data
			w.WriteBits(0x60739f61, 32)            // PAT CRC32
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
		func(w *bitio.Writer) {
			w.WriteByte(4)                         // Pointer field
			w.Write([]byte{0, 0, 0, 0})            // Pointer field bytes
			w.WriteByte(2)                         // PMT table ID
			WriteBinary(w, "1")                    // PMT syntax section indicator
			WriteBinary(w, "1")                    // PMT private bit
			WriteBinary(w, "11")                   // PMT reserved
			WriteBinary(w, "000000011000")         // PMT section length
			w.Write(psiSectionSyntaxHeaderBytes()) // PMT syntax section header
			w.Write(pmtBytes())                    // PMT data
			w.WriteBits(0xc68442e8, 32)            // PMT CRC32
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
			wExpected := bitio.NewWriter(&bufExpected)
			bufActual := bytes.Buffer{}
			wActual := bitio.NewWriter(&bufActual)

			tc.bytesFunc(wExpected)

			err := writePSIData(wActual, tc.data)
			assert.NoError(t, err)
			assert.Equal(t, bufActual.Len(), bufExpected.Len())
			assert.Equal(t, bufActual.Bytes(), bufExpected.Bytes())
		})
	}
}

func BenchmarkParsePSIData(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := bitio.NewCountReader(bytes.NewReader(psiBytes()))
		parsePSIData(r)
	}
}
