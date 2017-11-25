package astits

import (
	"testing"

	"github.com/asticode/go-astitools/binary"
	"github.com/stretchr/testify/assert"
)

var descriptors = []*Descriptor{{
	Length:           0x1,
	StreamIdentifier: &DescriptorStreamIdentifier{ComponentTag: 0x7},
	Tag:              DescriptorTagStreamIdentifier,
}}

func descriptorsBytes(w *astibinary.Writer) {
	w.Write("000000000011")                       // Overall length
	w.Write(uint8(DescriptorTagStreamIdentifier)) // Tag
	w.Write(uint8(1))                             // Length
	w.Write(uint8(7))                             // Component tag
}

func TestParseDescriptor(t *testing.T) {
	// Init
	w := astibinary.New()
	w.Write(uint16(155)) // Descriptors length
	// AC3
	w.Write(uint8(DescriptorTagAC3)) // Tag
	w.Write(uint8(9))                // Length
	w.Write("1")                     // Component type flag
	w.Write("1")                     // BSID flag
	w.Write("1")                     // MainID flag
	w.Write("1")                     // ASVC flag
	w.Write("0000")                  // Reserved flags
	w.Write(uint8(1))                // Component type
	w.Write(uint8(2))                // BSID
	w.Write(uint8(3))                // MainID
	w.Write(uint8(4))                // ASVC
	w.Write([]byte("info"))          // Additional info
	// ISO639 language and audio type
	w.Write(uint8(DescriptorTagISO639LanguageAndAudioType)) // Tag
	w.Write(uint8(4))                                       // Length
	w.Write([]byte("eng"))                                  // Language
	w.Write(uint8(AudioTypeCleanEffects))                   // Audio type
	// Maximum bitrate
	w.Write(uint8(DescriptorTagMaximumBitrate)) // Tag
	w.Write(uint8(3))                           // Length
	w.Write("000000000000000000000001")         // Maximum bitrate
	// Network name
	w.Write(uint8(DescriptorTagNetworkName)) // Tag
	w.Write(uint8(4))                        // Length
	w.Write([]byte("name"))                  // Name
	// Service
	w.Write(uint8(DescriptorTagService))                // Tag
	w.Write(uint8(18))                                  // Length
	w.Write(uint8(ServiceTypeDigitalTelevisionService)) // Type
	w.Write(uint8(8))                                   // Provider name length
	w.Write([]byte("provider"))                         // Provider name
	w.Write(uint8(7))                                   // Service name length
	w.Write([]byte("service"))                          // Service name
	// Short event
	w.Write(uint8(DescriptorTagShortEvent)) // Tag
	w.Write(uint8(14))                      // Length
	w.Write([]byte("eng"))                  // Language code
	w.Write(uint8(5))                       // Event name length
	w.Write([]byte("event"))                // Event name
	w.Write(uint8(4))                       // Text length
	w.Write([]byte("text"))
	// Stream identifier
	w.Write(uint8(DescriptorTagStreamIdentifier)) // Tag
	w.Write(uint8(1))                             // Length
	w.Write(uint8(2))                             // Component tag
	// Subtitling
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
	// Teletext
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
	// Extended event
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
	// Enhanced AC3
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
	// Extension supplementary audio
	w.Write(uint8(DescriptorTagExtension))                   // Tag
	w.Write(uint8(12))                                       // Length
	w.Write(uint8(DescriptorTagExtensionSupplementaryAudio)) // Extension tag
	w.Write("1")                                             // Mix type
	w.Write("10101")                                         // Editorial classification
	w.Write("1")                                             // Reserved
	w.Write("1")                                             // Language code flag
	w.Write([]byte("lan"))                                   // Language code
	w.Write([]byte("private"))                               // Private data

	// Assert
	var offset int
	ds := parseDescriptors(w.Bytes(), &offset)
	assert.Equal(t, *ds[0].AC3, DescriptorAC3{
		AdditionalInfo:   []byte("info"),
		ASVC:             uint8(4),
		BSID:             uint8(2),
		ComponentType:    uint8(1),
		HasASVC:          true,
		HasBSID:          true,
		HasComponentType: true,
		HasMainID:        true,
		MainID:           uint8(3),
	})
	assert.Equal(t, *ds[1].ISO639LanguageAndAudioType, DescriptorISO639LanguageAndAudioType{
		Language: []byte("eng"),
		Type:     AudioTypeCleanEffects,
	})
	assert.Equal(t, *ds[2].MaximumBitrate, DescriptorMaximumBitrate{Bitrate: uint32(50)})
	assert.Equal(t, *ds[3].NetworkName, DescriptorNetworkName{Name: []byte("name")})
	assert.Equal(t, *ds[4].Service, DescriptorService{
		Name:     []byte("service"),
		Provider: []byte("provider"),
		Type:     ServiceTypeDigitalTelevisionService,
	})
	assert.Equal(t, *ds[5].ShortEvent, DescriptorShortEvent{
		EventName: []byte("event"),
		Language:  []byte("eng"),
		Text:      []byte("text"),
	})
	assert.Equal(t, *ds[6].StreamIdentifier, DescriptorStreamIdentifier{ComponentTag: 0x2})
	assert.Equal(t, *ds[7].Subtitling, DescriptorSubtitling{Items: []*DescriptorSubtitlingItem{
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
	}})
	assert.Equal(t, *ds[8].Teletext, DescriptorTeletext{Items: []*DescriptorTeletextItem{
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
	}})
	assert.Equal(t, *ds[9].ExtendedEvent, DescriptorExtendedEvent{
		ISO639LanguageCode: []byte("lan"),
		Items: []*DescriptorExtendedEventItem{{
			Content:     []byte("content"),
			Description: []byte("description"),
		}},
		LastDescriptorNumber: 0x2,
		Number:               0x1,
		Text:                 []byte("text"),
	})
	assert.Equal(t, *ds[10].EnhancedAC3, DescriptorEnhancedAC3{
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
	})
	assert.Equal(t, *ds[11].Extension.SupplementaryAudio, DescriptorExtensionSupplementaryAudio{
		EditorialClassification: 21,
		HasLanguageCode:         true,
		LanguageCode:            []byte("lan"),
		MixType:                 true,
		PrivateData:             []byte("private"),
	})
}
