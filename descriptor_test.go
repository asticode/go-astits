package astits

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
	"github.com/stretchr/testify/assert"
)

var descriptors = []*Descriptor{{
	Length:           0x1,
	StreamIdentifier: DescriptorStreamIdentifier{ComponentTag: 0x7},
	Tag:              DescriptorTagStreamIdentifier,
}}

func descriptorsBytes(w *bitio.Writer) {
	WriteBinary(w, "000000000011")             // Overall length
	w.WriteByte(DescriptorTagStreamIdentifier) // Tag
	w.WriteByte(1)                             // Length
	w.WriteByte(7)                             // Component tag
}

type descriptorTest struct {
	name      string
	bytesFunc func(w *bitio.Writer)
	desc      Descriptor
}

var descriptorTestTable = []descriptorTest{
	{
		"AC3",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagAC3) // Tag
			w.WriteByte(9)                // Length
			WriteBinary(w, "1")           // Component type flag
			WriteBinary(w, "1")           // BSID flag
			WriteBinary(w, "1")           // MainID flag
			WriteBinary(w, "1")           // ASVC flag
			WriteBinary(w, "1111")        // Reserved flags
			w.WriteByte(1)                // Component type
			w.WriteByte(2)                // BSID
			w.WriteByte(3)                // MainID
			w.WriteByte(4)                // ASVC
			w.Write([]byte("info"))       // Additional info
		},
		Descriptor{
			Tag:    DescriptorTagAC3,
			Length: 9,
			AC3: &DescriptorAC3{
				AdditionalInfo:   []byte("info"),
				ASVC:             uint8(4),
				BSID:             uint8(2),
				ComponentType:    uint8(1),
				HasASVC:          true,
				HasBSID:          true,
				HasComponentType: true,
				HasMainID:        true,
				MainID:           uint8(3),
			},
		},
	},
	{
		"ISO639LanguageAndAudioType",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagISO639LanguageAndAudioType) // Tag
			w.WriteByte(4)                                       // Length
			w.Write([]byte("eng"))                               // Language
			w.WriteByte(AudioTypeCleanEffects)                   // Audio type
		},
		Descriptor{
			Tag:    DescriptorTagISO639LanguageAndAudioType,
			Length: 4,
			ISO639LanguageAndAudioType: &DescriptorISO639LanguageAndAudioType{
				Language: []byte("eng"),
				Type:     AudioTypeCleanEffects,
			},
		},
	},
	{
		"MaximumBitrate",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagMaximumBitrate)   // Tag
			w.WriteByte(3)                             // Length
			WriteBinary(w, "110000000000000000000001") // Maximum bitrate
		},
		Descriptor{
			Tag:            DescriptorTagMaximumBitrate,
			Length:         3,
			MaximumBitrate: DescriptorMaximumBitrate{Bitrate: 1},
		},
	},
	{
		"NetworkName",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagNetworkName) // Tag
			w.WriteByte(4)                        // Length
			w.Write([]byte("name"))               // Name
		},
		Descriptor{
			Tag:         DescriptorTagNetworkName,
			Length:      4,
			NetworkName: DescriptorNetworkName{Name: []byte("name")},
		},
	},
	{
		"Service",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagService)                // Tag
			w.WriteByte(18)                                  // Length
			w.WriteByte(ServiceTypeDigitalTelevisionService) // Type
			w.WriteByte(8)                                   // Provider name length
			w.Write([]byte("provider"))                      // Provider name
			w.WriteByte(7)                                   // Service name length
			w.Write([]byte("service"))                       // Service name
		},
		Descriptor{
			Tag:    DescriptorTagService,
			Length: 18,
			Service: &DescriptorService{
				Name:     []byte("service"),
				Provider: []byte("provider"),
				Type:     ServiceTypeDigitalTelevisionService,
			},
		},
	},
	{
		"ShortEvent",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagShortEvent) // Tag
			w.WriteByte(14)                      // Length
			w.Write([]byte("eng"))               // Language code
			w.WriteByte(5)                       // Event name length
			w.Write([]byte("event"))             // Event name
			w.WriteByte(4)                       // Text length
			w.Write([]byte("text"))
		},
		Descriptor{
			Tag:    DescriptorTagShortEvent,
			Length: 14,
			ShortEvent: &DescriptorShortEvent{
				EventName: []byte("event"),
				Language:  []byte("eng"),
				Text:      []byte("text"),
			},
		},
	},
	{
		"StreamIdentifier",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagStreamIdentifier) // Tag
			w.WriteByte(1)                             // Length
			w.WriteByte(2)                             // Component tag
		},
		Descriptor{
			Tag:              DescriptorTagStreamIdentifier,
			Length:           1,
			StreamIdentifier: DescriptorStreamIdentifier{ComponentTag: 0x2},
		},
	},
	{
		"Subtitling",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagSubtitling) // Tag
			w.WriteByte(16)                      // Length
			w.Write([]byte("lg1"))               // Item #1 language
			w.WriteByte(1)                       // Item #1 type
			w.WriteBits(2, 16)                   // Item #1 composition page
			w.WriteBits(3, 16)                   // Item #1 ancillary page
			w.Write([]byte("lg2"))               // Item #2 language
			w.WriteByte(4)                       // Item #2 type
			w.WriteBits(5, 16)                   // Item #2 composition page
			w.WriteBits(6, 16)                   // Item #2 ancillary page
		},
		Descriptor{
			Tag:    DescriptorTagSubtitling,
			Length: 16,
			Subtitling: DescriptorSubtitling{
				Items: []*DescriptorSubtitlingItem{
					{
						AncillaryPageID:   3,
						CompositionPageID: 2,
						Language:          []byte("lg1"),
						Type:              1,
					},
					{
						AncillaryPageID:   6,
						CompositionPageID: 5,
						Language:          []byte("lg2"),
						Type:              4,
					},
				},
			},
		},
	},
	{
		"Teletext",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagTeletext) // Tag
			w.WriteByte(10)                    // Length
			w.Write([]byte("lg1"))             // Item #1 language
			WriteBinary(w, "00001")            // Item #1 type
			WriteBinary(w, "010")              // Item #1 magazine
			WriteBinary(w, "00010010")         // Item #1 page number
			w.Write([]byte("lg2"))             // Item #2 language
			WriteBinary(w, "00011")            // Item #2 type
			WriteBinary(w, "100")              // Item #2 magazine
			WriteBinary(w, "00100011")         // Item #2 page number
		},
		Descriptor{
			Tag:    DescriptorTagTeletext,
			Length: 10,
			Teletext: DescriptorTeletext{
				Items: []*DescriptorTeletextItem{
					{
						Language: []byte("lg1"),
						Magazine: 2,
						Page:     12,
						Type:     1,
					},
					{
						Language: []byte("lg2"),
						Magazine: 4,
						Page:     23,
						Type:     3,
					},
				},
			},
		},
	},
	{
		"ExtendedEvent",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagExtendedEvent) // Tag
			w.WriteByte(30)                         // Length
			WriteBinary(w, "0001")                  // Number
			WriteBinary(w, "0010")                  // Last descriptor number
			w.Write([]byte("lan"))                  // ISO 639 language code
			w.WriteByte(20)                         // Length of items
			w.WriteByte(11)                         // Item #1 description length
			w.Write([]byte("description"))          // Item #1 description
			w.WriteByte(7)                          // Item #1 content length
			w.Write([]byte("content"))              // Item #1 content
			w.WriteByte(4)                          // Text length
			w.Write([]byte("text"))                 // Text
		},
		Descriptor{
			Tag:    DescriptorTagExtendedEvent,
			Length: 30,
			ExtendedEvent: &DescriptorExtendedEvent{
				ISO639LanguageCode: []byte("lan"),
				Items: []*DescriptorExtendedEventItem{{
					Content:     []byte("content"),
					Description: []byte("description"),
				}},
				LastDescriptorNumber: 0x2,
				Number:               0x1,
				Text:                 []byte("text"),
			},
		},
	},
	{
		"EnhancedAC3",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagEnhancedAC3) // Tag
			w.WriteByte(12)                       // Length
			WriteBinary(w, "1")                   // Component type flag
			WriteBinary(w, "1")                   // BSID flag
			WriteBinary(w, "1")                   // MainID flag
			WriteBinary(w, "1")                   // ASVC flag
			WriteBinary(w, "1")                   // Mix info exists
			WriteBinary(w, "1")                   // SubStream1 flag
			WriteBinary(w, "1")                   // SubStream2 flag
			WriteBinary(w, "1")                   // SubStream3 flag
			w.WriteByte(1)                        // Component type
			w.WriteByte(2)                        // BSID
			w.WriteByte(3)                        // MainID
			w.WriteByte(4)                        // ASVC
			w.WriteByte(5)                        // SubStream1
			w.WriteByte(6)                        // SubStream2
			w.WriteByte(7)                        // SubStream3
			w.Write([]byte("info"))               // Additional info
		},
		Descriptor{
			Tag:    DescriptorTagEnhancedAC3,
			Length: 12,
			EnhancedAC3: &DescriptorEnhancedAC3{
				AdditionalInfo:   []byte("info"),
				ASVC:             uint8(4),
				BSID:             uint8(2),
				ComponentType:    uint8(1),
				HasASVC:          true,
				HasBSID:          true,
				HasComponentType: true,
				HasMainID:        true,
				HasSubStream1:    true,
				HasSubStream2:    true,
				HasSubStream3:    true,
				MainID:           uint8(3),
				MixInfoExists:    true,
				SubStream1:       5,
				SubStream2:       6,
				SubStream3:       7,
			},
		},
	},
	{
		"Extension",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagExtension)                   // Tag
			w.WriteByte(12)                                       // Length
			w.WriteByte(DescriptorTagExtensionSupplementaryAudio) // Extension tag
			WriteBinary(w, "1")                                   // Mix type
			WriteBinary(w, "10101")                               // Editorial classification
			WriteBinary(w, "1")                                   // Reserved
			WriteBinary(w, "1")                                   // Language code flag
			w.Write([]byte("lan"))                                // Language code
			w.Write([]byte("private"))                            // Private data
		},
		Descriptor{
			Tag:    DescriptorTagExtension,
			Length: 12,
			Extension: &DescriptorExtension{
				SupplementaryAudio: &DescriptorExtensionSupplementaryAudio{
					EditorialClassification: 21,
					HasLanguageCode:         true,
					LanguageCode:            []byte("lan"),
					MixType:                 true,
					PrivateData:             []byte("private"),
				},
				Tag:     DescriptorTagExtensionSupplementaryAudio,
				Unknown: nil,
			},
		},
	},
	{
		"Component",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagComponent) // Tag
			w.WriteByte(10)                     // Length
			WriteBinary(w, "1010")              // Stream content ext
			WriteBinary(w, "0101")              // Stream content
			w.WriteByte(1)                      // Component type
			w.WriteByte(2)                      // Component tag
			w.Write([]byte("lan"))              // ISO639 language code
			w.Write([]byte("text"))             // Text
		},
		Descriptor{
			Tag:    DescriptorTagComponent,
			Length: 10,
			Component: &DescriptorComponent{
				ComponentTag:       2,
				ComponentType:      1,
				ISO639LanguageCode: []byte("lan"),
				StreamContentExt:   10,
				StreamContent:      5,
				Text:               []byte("text"),
			},
		},
	},
	{
		"Content",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagContent) // Tag
			w.WriteByte(2)                    // Length
			WriteBinary(w, "0001")            // Item #1 content nibble level 1
			WriteBinary(w, "0010")            // Item #1 content nibble level 2
			w.WriteByte(3)                    // Item #1 user byte
		},
		Descriptor{
			Tag:    DescriptorTagContent,
			Length: 2,
			Content: DescriptorContent{
				Items: []*DescriptorContentItem{{
					ContentNibbleLevel1: 1,
					ContentNibbleLevel2: 2,
					UserByte:            3,
				}},
			},
		},
	},
	{
		"ParentalRating",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagParentalRating) // Tag
			w.WriteByte(4)                           // Length
			w.Write([]byte("cou"))                   // Item #1 country code
			w.WriteByte(2)                           // Item #1 rating
		},
		Descriptor{
			Tag:    DescriptorTagParentalRating,
			Length: 4,
			ParentalRating: DescriptorParentalRating{
				Items: []*DescriptorParentalRatingItem{{
					CountryCode: []byte("cou"),
					Rating:      2,
				}},
			},
		},
	},
	{
		"LocalTimeOffset",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagLocalTimeOffset) // Tag
			w.WriteByte(13)                           // Length
			w.Write([]byte("cou"))                    // Country code
			WriteBinary(w, "101010")                  // Country region ID
			WriteBinary(w, "1")                       // Reserved
			WriteBinary(w, "1")                       // Local time offset polarity
			w.Write(dvbDurationMinutesBytes)          // Local time offset
			w.Write(dvbTimeBytes)                     // Time of change
			w.Write(dvbDurationMinutesBytes)          // Next time offset
		},
		Descriptor{
			Tag:    DescriptorTagLocalTimeOffset,
			Length: 13,
			LocalTimeOffset: []*DescriptorLocalTimeOffsetItem{{
				CountryCode:             []byte("cou"),
				CountryRegionID:         42,
				LocalTimeOffset:         dvbDurationMinutes,
				LocalTimeOffsetPolarity: true,
				NextTimeOffset:          dvbDurationMinutes,
				TimeOfChange:            dvbTime,
			}},
		},
	},
	{
		"VBIData",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagVBIData)        // Tag
			w.WriteByte(3)                           // Length
			w.WriteByte(VBIDataServiceIDEBUTeletext) // Service #1 id
			w.WriteByte(1)                           // Service #1 descriptor length
			WriteBinary(w, "11")                     // Service #1 descriptor reserved
			WriteBinary(w, "1")                      // Service #1 descriptor field polarity
			WriteBinary(w, "10101")                  // Service #1 descriptor line offset
		},
		Descriptor{
			Tag:    DescriptorTagVBIData,
			Length: 3,
			VBIData: []*DescriptorVBIDataService{{
				DataServiceID: VBIDataServiceIDEBUTeletext,
				Descriptors: []*DescriptorVBIDataDescriptor{{
					FieldParity: true,
					LineOffset:  21,
				}},
			}},
		},
	},
	{
		"VBITeletext",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagVBITeletext) // Tag
			w.WriteByte(5)                        // Length
			w.Write([]byte("lan"))                // Item #1 language
			WriteBinary(w, "00001")               // Item #1 type
			WriteBinary(w, "010")                 // Item #1 magazine
			WriteBinary(w, "00010010")            // Item #1 page number
		},
		Descriptor{
			Tag:    DescriptorTagVBITeletext,
			Length: 5,
			VBITeletext: DescriptorTeletext{
				[]*DescriptorTeletextItem{{
					// Language: 7102830, // "lan"
					Language: []byte("lan"),
					Magazine: 2,
					Page:     12,
					Type:     1,
				}},
			},
		},
	},
	{
		"AVCVideo",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagAVCVideo) // Tag
			w.WriteByte(4)                     // Length
			w.WriteByte(1)                     // Profile idc
			WriteBinary(w, "1")                // Constraint set0 flag
			WriteBinary(w, "1")                // Constraint set1 flag
			WriteBinary(w, "1")                // Constraint set1 flag
			WriteBinary(w, "10101")            // Compatible flags
			w.WriteByte(2)                     // Level idc
			WriteBinary(w, "1")                // AVC still present
			WriteBinary(w, "1")                // AVC 24 hour picture flag
			WriteBinary(w, "111111")           // Reserved
		},
		Descriptor{
			Tag:    DescriptorTagAVCVideo,
			Length: 4,
			AVCVideo: &DescriptorAVCVideo{
				AVC24HourPictureFlag: true,
				AVCStillPresent:      true,
				CompatibleFlags:      21,
				ConstraintSet0Flag:   true,
				ConstraintSet1Flag:   true,
				ConstraintSet2Flag:   true,
				LevelIDC:             2,
				ProfileIDC:           1,
			},
		},
	},
	{
		"PrivateDataSpecifier",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagPrivateDataSpecifier) // Tag
			w.WriteByte(4)                                 // Length
			w.WriteBits(128, 32)                           // Private data specifier
		},
		Descriptor{
			Tag:                  DescriptorTagPrivateDataSpecifier,
			Length:               4,
			PrivateDataSpecifier: DescriptorPrivateDataSpecifier{Specifier: 128},
		},
	},
	{
		"DataStreamAlignment",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagDataStreamAlignment) // Tag
			w.WriteByte(1)                                // Length
			w.WriteByte(2)                                // Type
		},
		Descriptor{
			Tag:    DescriptorTagDataStreamAlignment,
			Length: 1,

			DataStreamAlignment: 2,
		},
	},
	{
		"PrivateDataIndicator",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagPrivateDataIndicator) // Tag
			w.WriteByte(4)                                 // Length
			w.WriteBits(127, 32)                           // Private data indicator
		},
		Descriptor{
			Tag:    DescriptorTagPrivateDataIndicator,
			Length: 4,

			PrivateDataIndicator: 127,
		},
	},
	{
		"UserDefined",
		func(w *bitio.Writer) {
			w.WriteByte(0x80)       // Tag
			w.WriteByte(4)          // Length
			w.Write([]byte("test")) // User defined
		},
		Descriptor{
			Tag:         0x80,
			Length:      4,
			UserDefined: []byte("test"),
		},
	},
	{
		"Registration",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagRegistration) // Tag
			w.WriteByte(8)                         // Length
			w.WriteBits(1, 32)                     // Format identifier
			w.Write([]byte("test"))                // Additional identification info
		},
		Descriptor{
			Tag:    DescriptorTagRegistration,
			Length: 8,
			Registration: &DescriptorRegistration{
				AdditionalIdentificationInfo: []byte("test"),
				FormatIdentifier:             1,
			},
		},
	},
	{
		"Unknown",
		func(w *bitio.Writer) {
			w.WriteByte(0x1)        // Tag
			w.WriteByte(4)          // Length
			w.Write([]byte("test")) // Content
		},
		Descriptor{
			Tag:    0x1,
			Length: 4,
			Unknown: &DescriptorUnknown{
				Content: []byte("test"),
				Tag:     0x1,
			},
		},
	},
	{
		"Extension",
		func(w *bitio.Writer) {
			w.WriteByte(DescriptorTagExtension) // Tag
			w.WriteByte(5)                      // Length
			w.WriteByte(0)                      // Extension tag
			w.Write([]byte("test"))             // Content
		},
		Descriptor{
			Tag:    DescriptorTagExtension,
			Length: 5,
			Extension: &DescriptorExtension{
				Tag:     0,
				Unknown: &[]byte{'t', 'e', 's', 't'},
			},
		},
	},
}

func TestParseDescriptorOneByOne(t *testing.T) {
	for _, tc := range descriptorTestTable {
		t.Run(tc.name, func(t *testing.T) {
			// idea is following:
			// 1. get descriptor bytes and update its length
			// 2. parse bytes and get a Descriptor instance
			// 3. compare expected descriptor value and actual
			buf := bytes.Buffer{}
			buf.Write([]byte{0x00, 0x00}) // reserve two bytes for length
			w := bitio.NewWriter(&buf)
			tc.bytesFunc(w)
			descLen := uint16(buf.Len() - 2)
			descBytes := buf.Bytes()
			descBytes[0] = byte(descLen >> 8)
			descBytes[1] = byte(descLen & 0xff)

			r := bitio.NewCountReader(bytes.NewReader(descBytes))
			_, err := r.ReadBits(4)
			assert.NoError(t, err)

			ds, err := parseDescriptors(r)
			assert.NoError(t, err)
			assert.Equal(t, tc.desc, *ds[0])
		})
	}
}

func TestParseDescriptorAll(t *testing.T) {
	buf := bytes.Buffer{}
	buf.Write([]byte{0x00, 0x00}) // reserve two bytes for length
	w := bitio.NewWriter(&buf)

	for _, tc := range descriptorTestTable {
		tc.bytesFunc(w)
	}

	descLen := uint16(buf.Len() - 2)
	descBytes := buf.Bytes()
	descBytes[0] = byte(descLen >> 8)
	descBytes[1] = byte(descLen & 0xff)

	r := bitio.NewCountReader(bytes.NewReader(descBytes))
	_, err := r.ReadBits(4)
	assert.NoError(t, err)

	ds, err := parseDescriptors(r)
	assert.NoError(t, err)

	for i, tc := range descriptorTestTable {
		assert.Equal(t, tc.desc, *ds[i])
	}
}

func TestWriteDescriptorOneByOne(t *testing.T) {
	for _, tc := range descriptorTestTable {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := bitio.NewWriter(&bufExpected)
			tc.bytesFunc(wExpected)

			bufActual := bytes.Buffer{}
			wActual := bitio.NewWriter(&bufActual)
			n, err := writeDescriptor(wActual, &tc.desc)
			assert.NoError(t, err)
			assert.Equal(t, n, bufActual.Len())
			assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
		})
	}
}

func TestWriteDescriptorAll(t *testing.T) {
	bufExpected := bytes.Buffer{}
	bufExpected.Write([]byte{0x00, 0x00}) // reserve two bytes for length
	wExpected := bitio.NewWriter(&bufExpected)

	dss := []*Descriptor{}

	for _, tc := range descriptorTestTable {
		tc.bytesFunc(wExpected)
		tcc := tc
		dss = append(dss, &tcc.desc)
	}

	descLen := uint16(bufExpected.Len() - 2)
	descBytes := bufExpected.Bytes()
	descBytes[0] = byte(descLen>>8) | 0b11110000 // program_info_length is preceded by 4 reserved bits
	descBytes[1] = byte(descLen & 0xff)

	bufActual := bytes.Buffer{}
	wActual := bitio.NewWriter(&bufActual)

	n, err := writeDescriptorsWithLength(wActual, dss)
	assert.NoError(t, err)
	assert.Equal(t, n, bufActual.Len())
	assert.Equal(t, bufExpected.Len(), bufActual.Len())
	assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
}

func BenchmarkWriteDescriptor(b *testing.B) {
	buf := bytes.Buffer{}
	buf.Grow(1024)
	w := bitio.NewWriter(&buf)

	for _, bm := range descriptorTestTable {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				writeDescriptor(w, &bm.desc)
			}
		})
	}
}

func BenchmarkParseDescriptor(b *testing.B) {
	bss := make([][]byte, len(descriptorTestTable))

	for ti, tc := range descriptorTestTable {
		buf := bytes.Buffer{}
		buf.Write([]byte{0x00, 0x00}) // reserve two bytes for length
		w := bitio.NewWriter(&buf)
		tc.bytesFunc(w)
		descLen := uint16(buf.Len() - 2)
		descBytes := buf.Bytes()
		descBytes[0] = byte(descLen >> 8)
		descBytes[1] = byte(descLen & 0xff)
		bss[ti] = descBytes
	}

	for ti, tc := range descriptorTestTable {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r := bitio.NewCountReader(bytes.NewReader(bss[ti]))

				_, err := r.ReadBits(4)
				assert.NoError(b, err)
				parseDescriptors(r)
			}
		})
	}
}
