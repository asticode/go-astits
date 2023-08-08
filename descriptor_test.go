package astits

import (
	"bytes"
	"testing"

	"github.com/asticode/go-astikit"
	"github.com/stretchr/testify/assert"
)

var descriptors = []*Descriptor{{
	Length:           0x1,
	StreamIdentifier: &DescriptorStreamIdentifier{ComponentTag: 0x7},
	Tag:              DescriptorTagStreamIdentifier,
}}

func descriptorsBytes(w *astikit.BitsWriter) {
	w.Write("000000000011")                       // Overall length
	w.Write(uint8(DescriptorTagStreamIdentifier)) // Tag
	w.Write(uint8(1))                             // Length
	w.Write(uint8(7))                             // Component tag
}

type descriptorTest struct {
	name      string
	bytesFunc func(w *astikit.BitsWriter)
	desc      Descriptor
}

var descriptorTestTable = []descriptorTest{
	{
		"AC3",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagAC3)) // Tag
			w.Write(uint8(9))                // Length
			w.Write("1")                     // Component type flag
			w.Write("1")                     // BSID flag
			w.Write("1")                     // MainID flag
			w.Write("1")                     // ASVC flag
			w.Write("1111")                  // Reserved flags
			w.Write(uint8(1))                // Component type
			w.Write(uint8(2))                // BSID
			w.Write(uint8(3))                // MainID
			w.Write(uint8(4))                // ASVC
			w.Write([]byte("info"))          // Additional info
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
			}},
	},
	{
		"ISO639LanguageAndAudioType",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagISO639LanguageAndAudioType)) // Tag
			w.Write(uint8(4))                                       // Length
			w.Write([]byte("eng"))                                  // Language
			w.Write(uint8(AudioTypeCleanEffects))                   // Audio type
		},
		Descriptor{
			Tag:    DescriptorTagISO639LanguageAndAudioType,
			Length: 4,
			ISO639LanguageAndAudioType: &DescriptorISO639LanguageAndAudioType{
				Language: []byte("eng"),
				Type:     AudioTypeCleanEffects,
			}},
	},
	{
		"MaximumBitrate",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagMaximumBitrate)) // Tag
			w.Write(uint8(3))                           // Length
			w.Write("110000000000000000000001")         // Maximum bitrate
		},
		Descriptor{
			Tag:            DescriptorTagMaximumBitrate,
			Length:         3,
			MaximumBitrate: &DescriptorMaximumBitrate{Bitrate: uint32(50)}},
	},
	{
		"NetworkName",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagNetworkName)) // Tag
			w.Write(uint8(4))                        // Length
			w.Write([]byte("name"))                  // Name
		},
		Descriptor{
			Tag:         DescriptorTagNetworkName,
			Length:      4,
			NetworkName: &DescriptorNetworkName{Name: []byte("name")}},
	},
	{
		"Service",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagService))                // Tag
			w.Write(uint8(18))                                  // Length
			w.Write(uint8(ServiceTypeDigitalTelevisionService)) // Type
			w.Write(uint8(8))                                   // Provider name length
			w.Write([]byte("provider"))                         // Provider name
			w.Write(uint8(7))                                   // Service name length
			w.Write([]byte("service"))                          // Service name
		},
		Descriptor{
			Tag:    DescriptorTagService,
			Length: 18,
			Service: &DescriptorService{
				Name:     []byte("service"),
				Provider: []byte("provider"),
				Type:     ServiceTypeDigitalTelevisionService,
			}},
	},
	{
		"ShortEvent",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagShortEvent)) // Tag
			w.Write(uint8(14))                      // Length
			w.Write([]byte("eng"))                  // Language code
			w.Write(uint8(5))                       // Event name length
			w.Write([]byte("event"))                // Event name
			w.Write(uint8(4))                       // Text length
			w.Write([]byte("text"))
		},
		Descriptor{
			Tag:    DescriptorTagShortEvent,
			Length: 14,
			ShortEvent: &DescriptorShortEvent{
				EventName: []byte("event"),
				Language:  []byte("eng"),
				Text:      []byte("text"),
			}},
	},
	{
		"StreamIdentifier",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagStreamIdentifier)) // Tag
			w.Write(uint8(1))                             // Length
			w.Write(uint8(2))                             // Component tag
		},
		Descriptor{
			Tag:              DescriptorTagStreamIdentifier,
			Length:           1,
			StreamIdentifier: &DescriptorStreamIdentifier{ComponentTag: 0x2}},
	},
	{
		"Subtitling",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagSubtitling)) // Tag
			w.Write(uint8(16))                      // Length
			w.Write([]byte("lg1"))                  // Item #1 language
			w.Write(uint8(1))                       // Item #1 type
			w.Write(uint16(2))                      // Item #1 composition page
			w.Write(uint16(3))                      // Item #1 ancillary page
			w.Write([]byte("lg2"))                  // Item #2 language
			w.Write(uint8(4))                       // Item #2 type
			w.Write(uint16(5))                      // Item #2 composition page
			w.Write(uint16(6))                      // Item #2 ancillary page
		},
		Descriptor{
			Tag:    DescriptorTagSubtitling,
			Length: 16,
			Subtitling: &DescriptorSubtitling{Items: []*DescriptorSubtitlingItem{
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
			}}},
	},
	{
		"Teletext",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagTeletext)) // Tag
			w.Write(uint8(10))                    // Length
			w.Write([]byte("lg1"))                // Item #1 language
			w.Write("00001")                      // Item #1 type
			w.Write("010")                        // Item #1 magazine
			w.Write("00010010")                   // Item #1 page number
			w.Write([]byte("lg2"))                // Item #2 language
			w.Write("00011")                      // Item #2 type
			w.Write("100")                        // Item #2 magazine
			w.Write("00100011")                   // Item #2 page number
		},
		Descriptor{
			Tag:    DescriptorTagTeletext,
			Length: 10,
			Teletext: &DescriptorTeletext{Items: []*DescriptorTeletextItem{
				{
					Language: []byte("lg1"),
					Magazine: uint8(2),
					Page:     uint8(12),
					Type:     uint8(1),
				},
				{
					Language: []byte("lg2"),
					Magazine: uint8(4),
					Page:     uint8(23),
					Type:     uint8(3),
				},
			}}},
	},
	{
		"ExtendedEvent",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagExtendedEvent)) // Tag
			w.Write(uint8(30))                         // Length
			w.Write("0001")                            // Number
			w.Write("0010")                            // Last descriptor number
			w.Write([]byte("lan"))                     // ISO 639 language code
			w.Write(uint8(20))                         // Length of items
			w.Write(uint8(11))                         // Item #1 description length
			w.Write([]byte("description"))             // Item #1 description
			w.Write(uint8(7))                          // Item #1 content length
			w.Write([]byte("content"))                 // Item #1 content
			w.Write(uint8(4))                          // Text length
			w.Write([]byte("text"))                    // Text
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
			}},
	},
	{
		"EnhancedAC3",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagEnhancedAC3)) // Tag
			w.Write(uint8(12))                       // Length
			w.Write("1")                             // Component type flag
			w.Write("1")                             // BSID flag
			w.Write("1")                             // MainID flag
			w.Write("1")                             // ASVC flag
			w.Write("1")                             // Mix info exists
			w.Write("1")                             // SubStream1 flag
			w.Write("1")                             // SubStream2 flag
			w.Write("1")                             // SubStream3 flag
			w.Write(uint8(1))                        // Component type
			w.Write(uint8(2))                        // BSID
			w.Write(uint8(3))                        // MainID
			w.Write(uint8(4))                        // ASVC
			w.Write(uint8(5))                        // SubStream1
			w.Write(uint8(6))                        // SubStream2
			w.Write(uint8(7))                        // SubStream3
			w.Write([]byte("info"))                  // Additional info
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
			}},
	},
	{
		"Extension",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagExtension))                   // Tag
			w.Write(uint8(12))                                       // Length
			w.Write(uint8(DescriptorTagExtensionSupplementaryAudio)) // Extension tag
			w.Write("1")                                             // Mix type
			w.Write("10101")                                         // Editorial classification
			w.Write("1")                                             // Reserved
			w.Write("1")                                             // Language code flag
			w.Write([]byte("lan"))                                   // Language code
			w.Write([]byte("private"))                               // Private data
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
			}},
	},
	{
		"Component",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagComponent)) // Tag
			w.Write(uint8(10))                     // Length
			w.Write("1010")                        // Stream content ext
			w.Write("0101")                        // Stream content
			w.Write(uint8(1))                      // Component type
			w.Write(uint8(2))                      // Component tag
			w.Write([]byte("lan"))                 // ISO639 language code
			w.Write([]byte("text"))                // Text
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
			}},
	},
	{
		"Content",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagContent)) // Tag
			w.Write(uint8(2))                    // Length
			w.Write("0001")                      // Item #1 content nibble level 1
			w.Write("0010")                      // Item #1 content nibble level 2
			w.Write(uint8(3))                    // Item #1 user byte
		},
		Descriptor{
			Tag:    DescriptorTagContent,
			Length: 2,
			Content: &DescriptorContent{Items: []*DescriptorContentItem{{
				ContentNibbleLevel1: 1,
				ContentNibbleLevel2: 2,
				UserByte:            3,
			}}}},
	},
	{
		"ParentalRating",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagParentalRating)) // Tag
			w.Write(uint8(4))                           // Length
			w.Write([]byte("cou"))                      // Item #1 country code
			w.Write(uint8(2))                           // Item #1 rating
		},
		Descriptor{
			Tag:    DescriptorTagParentalRating,
			Length: 4,
			ParentalRating: &DescriptorParentalRating{Items: []*DescriptorParentalRatingItem{{
				CountryCode: []byte("cou"),
				Rating:      2,
			}}}},
	},
	{
		"LocalTimeOffset",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagLocalTimeOffset)) // Tag
			w.Write(uint8(13))                           // Length
			w.Write([]byte("cou"))                       // Country code
			w.Write("101010")                            // Country region ID
			w.Write("1")                                 // Reserved
			w.Write("1")                                 // Local time offset polarity
			w.Write(dvbDurationMinutesBytes)             // Local time offset
			w.Write(dvbTimeBytes)                        // Time of change
			w.Write(dvbDurationMinutesBytes)             // Next time offset
		},
		Descriptor{
			Tag:    DescriptorTagLocalTimeOffset,
			Length: 13,
			LocalTimeOffset: &DescriptorLocalTimeOffset{Items: []*DescriptorLocalTimeOffsetItem{{
				CountryCode:             []byte("cou"),
				CountryRegionID:         42,
				LocalTimeOffset:         dvbDurationMinutes,
				LocalTimeOffsetPolarity: true,
				NextTimeOffset:          dvbDurationMinutes,
				TimeOfChange:            dvbTime,
			}}}},
	},
	{
		"VBIData",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagVBIData))        // Tag
			w.Write(uint8(3))                           // Length
			w.Write(uint8(VBIDataServiceIDEBUTeletext)) // Service #1 id
			w.Write(uint8(1))                           // Service #1 descriptor length
			w.Write("11")                               // Service #1 descriptor reserved
			w.Write("1")                                // Service #1 descriptor field polarity
			w.Write("10101")                            // Service #1 descriptor line offset
		},
		Descriptor{
			Tag:    DescriptorTagVBIData,
			Length: 3,
			VBIData: &DescriptorVBIData{Services: []*DescriptorVBIDataService{{
				DataServiceID: VBIDataServiceIDEBUTeletext,
				Descriptors: []*DescriptorVBIDataDescriptor{{
					FieldParity: true,
					LineOffset:  21,
				}},
			}}}},
	},
	{
		"VBITeletext",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagVBITeletext)) // Tag
			w.Write(uint8(5))                        // Length
			w.Write([]byte("lan"))                   // Item #1 language
			w.Write("00001")                         // Item #1 type
			w.Write("010")                           // Item #1 magazine
			w.Write("00010010")                      // Item #1 page number
		},
		Descriptor{
			Tag:    DescriptorTagVBITeletext,
			Length: 5,
			VBITeletext: &DescriptorTeletext{Items: []*DescriptorTeletextItem{{
				Language: []byte("lan"),
				Magazine: uint8(2),
				Page:     uint8(12),
				Type:     uint8(1),
			}}}},
	},
	{
		"AVCVideo",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagAVCVideo)) // Tag
			w.Write(uint8(4))                     // Length
			w.Write(uint8(1))                     // Profile idc
			w.Write("1")                          // Constraint set0 flag
			w.Write("1")                          // Constraint set1 flag
			w.Write("1")                          // Constraint set1 flag
			w.Write("10101")                      // Compatible flags
			w.Write(uint8(2))                     // Level idc
			w.Write("1")                          // AVC still present
			w.Write("1")                          // AVC 24 hour picture flag
			w.Write("111111")                     // Reserved
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
			}},
	},
	{
		"PrivateDataSpecifier",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagPrivateDataSpecifier)) // Tag
			w.Write(uint8(4))                                 // Length
			w.Write(uint32(128))                              // Private data specifier
		},
		Descriptor{
			Tag:    DescriptorTagPrivateDataSpecifier,
			Length: 4,
			PrivateDataSpecifier: &DescriptorPrivateDataSpecifier{
				Specifier: 128,
			}},
	},
	{
		"DataStreamAlignment",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagDataStreamAlignment)) // Tag
			w.Write(uint8(1))                                // Length
			w.Write(uint8(2))                                // Type
		},
		Descriptor{
			Tag:    DescriptorTagDataStreamAlignment,
			Length: 1,
			DataStreamAlignment: &DescriptorDataStreamAlignment{
				Type: 2,
			}},
	},
	{
		"PrivateDataIndicator",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagPrivateDataIndicator)) // Tag
			w.Write(uint8(4))                                 // Length
			w.Write(uint32(127))                              // Private data indicator
		},
		Descriptor{
			Tag:    DescriptorTagPrivateDataIndicator,
			Length: 4,
			PrivateDataIndicator: &DescriptorPrivateDataIndicator{
				Indicator: 127,
			}},
	},
	{
		"UserDefined",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(0x80))    // Tag
			w.Write(uint8(4))       // Length
			w.Write([]byte("test")) // User defined
		},
		Descriptor{
			Tag:         0x80,
			Length:      4,
			UserDefined: []byte("test")},
	},
	{
		"Registration",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagRegistration)) // Tag
			w.Write(uint8(8))                         // Length
			w.Write(uint32(1))                        // Format identifier
			w.Write([]byte("test"))                   // Additional identification info
		},
		Descriptor{
			Tag:    DescriptorTagRegistration,
			Length: 8,
			Registration: &DescriptorRegistration{
				AdditionalIdentificationInfo: []byte("test"),
				FormatIdentifier:             uint32(1),
			}},
	},
	{
		"Unknown",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(0x1))     // Tag
			w.Write(uint8(4))       // Length
			w.Write([]byte("test")) // Content
		},
		Descriptor{
			Tag:    0x1,
			Length: 4,
			Unknown: &DescriptorUnknown{
				Content: []byte("test"),
				Tag:     0x1,
			}},
	},
	{
		"Extension",
		func(w *astikit.BitsWriter) {
			w.Write(uint8(DescriptorTagExtension)) // Tag
			w.Write(uint8(5))                      // Length
			w.Write(uint8(0))                      // Extension tag
			w.Write([]byte("test"))                // Content
		},
		Descriptor{
			Tag:    DescriptorTagExtension,
			Length: 5,
			Extension: &DescriptorExtension{
				Tag:     0,
				Unknown: &[]byte{'t', 'e', 's', 't'},
			}},
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
			w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
			tc.bytesFunc(w)
			descLen := uint16(buf.Len() - 2)
			descBytes := buf.Bytes()
			descBytes[0] = byte(descLen >> 8)
			descBytes[1] = byte(descLen & 0xff)

			ds, err := parseDescriptors(astikit.NewBytesIterator(descBytes))
			assert.NoError(t, err)
			assert.Equal(t, tc.desc, *ds[0])
		})
	}
}

func TestParseDescriptorAll(t *testing.T) {
	buf := bytes.Buffer{}
	buf.Write([]byte{0x00, 0x00}) // reserve two bytes for length
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})

	for _, tc := range descriptorTestTable {
		tc.bytesFunc(w)
	}

	descLen := uint16(buf.Len() - 2)
	descBytes := buf.Bytes()
	descBytes[0] = byte(descLen >> 8)
	descBytes[1] = byte(descLen & 0xff)

	ds, err := parseDescriptors(astikit.NewBytesIterator(descBytes))
	assert.NoError(t, err)

	for i, tc := range descriptorTestTable {
		assert.Equal(t, tc.desc, *ds[i])
	}
}

func TestWriteDescriptorOneByOne(t *testing.T) {
	for _, tc := range descriptorTestTable {
		t.Run(tc.name, func(t *testing.T) {
			bufExpected := bytes.Buffer{}
			wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})
			tc.bytesFunc(wExpected)

			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})
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
	wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})

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
	wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

	n, err := writeDescriptorsWithLength(wActual, dss)
	assert.NoError(t, err)
	assert.Equal(t, n, bufActual.Len())
	assert.Equal(t, bufExpected.Len(), bufActual.Len())
	assert.Equal(t, bufExpected.Bytes(), bufActual.Bytes())
}

func BenchmarkWriteDescriptor(b *testing.B) {
	buf := bytes.Buffer{}
	buf.Grow(1024)
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})

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
		w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &buf})
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
				parseDescriptors(astikit.NewBytesIterator(bss[ti]))
			}
		})
	}
}

func FuzzDescriptor(f *testing.F) {
	bufExpected := bytes.Buffer{}
	bufExpected.Write([]byte{0x00, 0x00}) // reserve two bytes for length
	wExpected := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufExpected})

	for _, tc := range descriptorTestTable {
		tc.bytesFunc(wExpected)
	}

	descLen := uint16(bufExpected.Len() - 2)
	descBytes := bufExpected.Bytes()
	descBytes[0] = byte(descLen>>8) | 0b11110000 // program_info_length is preceded by 4 reserved bits
	descBytes[1] = byte(descLen & 0xff)

	f.Add(descBytes)

	f.Fuzz(func(t *testing.T, b []byte) {
		ds, err := parseDescriptors(astikit.NewBytesIterator(b))

		if err == nil {
			bufActual := bytes.Buffer{}
			wActual := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &bufActual})

			writeDescriptorsWithLength(wActual, ds)
		}
	})
}
