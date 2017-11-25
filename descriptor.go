package astits

import (
	"time"

	"github.com/asticode/go-astilog"
)

// Audio types
// Page: 683 | https://books.google.fr/books?id=6dgWB3-rChYC&printsec=frontcover&hl=fr
const (
	AudioTypeCleanEffects             = 0x1
	AudioTypeHearingImpaired          = 0x2
	AudioTypeVisualImpairedCommentary = 0x3
)

// Descriptor tags
// Page: 42 | Chapter: 6.1 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
const (
	DescriptorTagAC3                        = 0x6a
	DescriptorTagComponent                  = 0x50
	DescriptorTagContent                    = 0x54
	DescriptorTagEnhancedAC3                = 0x7a
	DescriptorTagExtendedEvent              = 0x4e
	DescriptorTagExtension                  = 0x7f
	DescriptorTagISO639LanguageAndAudioType = 0xa
	DescriptorTagLocalTimeOffset            = 0x58
	DescriptorTagMaximumBitrate             = 0xe
	DescriptorTagNetworkName                = 0x40
	DescriptorTagParentalRating             = 0x55
	DescriptorTagService                    = 0x48
	DescriptorTagShortEvent                 = 0x4d
	DescriptorTagStreamIdentifier           = 0x52
	DescriptorTagSubtitling                 = 0x59
	DescriptorTagTeletext                   = 0x56
)

// Descriptor extension tags
// Page: 111 | Chapter: 6.1 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
const (
	DescriptorTagExtensionSupplementaryAudio = 0x6
)

// Service types
// Page: 97 | Chapter: 6.2.33 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
// https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf / page 97
const (
	ServiceTypeDigitalTelevisionService = 0x1
)

// Teletext types
// Page: 106 | Chapter: 6.2.43 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
const (
	TeletextTypeAdditionalInformationPage                    = 0x3
	TeletextTypeInitialTeletextPage                          = 0x1
	TeletextTypeProgramSchedulePage                          = 0x4
	TeletextTypeTeletextSubtitlePage                         = 0x2
	TeletextTypeTeletextSubtitlePageForHearingImpairedPeople = 0x5
)

// Descriptor represents a descriptor
type Descriptor struct {
	AC3                        *DescriptorAC3
	Component                  *DescriptorComponent
	Content                    *DescriptorContent
	EnhancedAC3                *DescriptorEnhancedAC3
	ExtendedEvent              *DescriptorExtendedEvent
	Extension                  *DescriptorExtension
	ISO639LanguageAndAudioType *DescriptorISO639LanguageAndAudioType
	Length                     uint8
	LocalTimeOffset            *DescriptorLocalTimeOffset
	MaximumBitrate             *DescriptorMaximumBitrate
	NetworkName                *DescriptorNetworkName
	ParentalRating             *DescriptorParentalRating
	Service                    *DescriptorService
	ShortEvent                 *DescriptorShortEvent
	StreamIdentifier           *DescriptorStreamIdentifier
	Subtitling                 *DescriptorSubtitling
	Tag                        uint8 // the tag defines the structure of the contained data following the descriptor length.
	Teletext                   *DescriptorTeletext
}

// DescriptorAC3 represents an AC3 descriptor
// Page: 165 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorAC3 struct {
	AdditionalInfo   []byte
	ASVC             uint8
	BSID             uint8
	ComponentType    uint8
	HasASVC          bool
	HasBSID          bool
	HasComponentType bool
	HasMainID        bool
	MainID           uint8
}

func newDescriptorAC3(i []byte) (d *DescriptorAC3) {
	var offset int
	d = &DescriptorAC3{}
	d.HasComponentType = uint8(i[offset]&0x80) > 0
	d.HasBSID = uint8(i[offset]&0x40) > 0
	d.HasMainID = uint8(i[offset]&0x20) > 0
	d.HasASVC = uint8(i[offset]&0x10) > 0
	offset += 1
	if d.HasComponentType {
		d.ComponentType = uint8(i[offset])
		offset += 1
	}
	if d.HasBSID {
		d.BSID = uint8(i[offset])
		offset += 1
	}
	if d.HasMainID {
		d.MainID = uint8(i[offset])
		offset += 1
	}
	if d.HasASVC {
		d.ASVC = uint8(i[offset])
		offset += 1
	}
	for offset < len(i) {
		d.AdditionalInfo = append(d.AdditionalInfo, i[offset])
		offset += 1
	}
	return
}

// DescriptorComponent represents a component descriptor
// Page: 51 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorComponent struct {
	ComponentTag       uint8
	ComponentType      uint8
	ISO639LanguageCode []byte
	StreamContent      uint8
	StreamContentExt   uint8
	Text               []byte
}

func newDescriptorComponent(i []byte) (d *DescriptorComponent) {
	// Init
	d = &DescriptorComponent{}
	var offset int

	// Stream content ext
	d.StreamContentExt = uint8(i[offset] >> 4)

	// Stream content
	d.StreamContent = uint8(i[offset] & 0xf)
	offset += 1

	// Component type
	d.ComponentType = uint8(i[offset])
	offset += 1

	// Component tag
	d.ComponentTag = uint8(i[offset])
	offset += 1

	// ISO639 language code
	d.ISO639LanguageCode = i[offset : offset+3]
	offset += 3

	// Text
	for offset < len(i) {
		d.Text = append(d.Text, i[offset])
		offset += 1
	}
	return
}

// DescriptorContent represents a content descriptor
// Page: 58 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorContent struct {
	Items []*DescriptorContentItem
}

// DescriptorContentItem represents a content item descriptor
// Check page 59 of https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf for content nibble
// levels associations
type DescriptorContentItem struct {
	ContentNibbleLevel1 uint8
	ContentNibbleLevel2 uint8
	UserByte            uint8
}

func newDescriptorContent(i []byte) (d *DescriptorContent) {
	// Init
	d = &DescriptorContent{}
	var offset int

	// Add items
	for offset < len(i) {
		d.Items = append(d.Items, &DescriptorContentItem{
			ContentNibbleLevel1: uint8(i[offset] >> 4),
			ContentNibbleLevel2: uint8(i[offset] & 0xf),
			UserByte:            uint8(i[offset+1]),
		})
		offset += 2
	}
	return
}

// DescriptorEnhancedAC3 represents an enhanced AC3 descriptor
// Page: 166 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorEnhancedAC3 struct {
	AdditionalInfo   []byte
	ASVC             uint8
	BSID             uint8
	ComponentType    uint8
	HasASVC          bool
	HasBSID          bool
	HasComponentType bool
	HasMainID        bool
	HasSubStream1    bool
	HasSubStream2    bool
	HasSubStream3    bool
	MainID           uint8
	MixInfoExists    bool
	SubStream1       uint8
	SubStream2       uint8
	SubStream3       uint8
}

func newDescriptorEnhancedAC3(i []byte) (d *DescriptorEnhancedAC3) {
	var offset int
	d = &DescriptorEnhancedAC3{}
	d.HasComponentType = uint8(i[offset]&0x80) > 0
	d.HasBSID = uint8(i[offset]&0x40) > 0
	d.HasMainID = uint8(i[offset]&0x20) > 0
	d.HasASVC = uint8(i[offset]&0x10) > 0
	d.MixInfoExists = uint8(i[offset]&0x8) > 0
	d.HasSubStream1 = uint8(i[offset]&0x4) > 0
	d.HasSubStream2 = uint8(i[offset]&0x2) > 0
	d.HasSubStream3 = uint8(i[offset]&0x1) > 0
	offset += 1
	if d.HasComponentType {
		d.ComponentType = uint8(i[offset])
		offset += 1
	}
	if d.HasBSID {
		d.BSID = uint8(i[offset])
		offset += 1
	}
	if d.HasMainID {
		d.MainID = uint8(i[offset])
		offset += 1
	}
	if d.HasASVC {
		d.ASVC = uint8(i[offset])
		offset += 1
	}
	if d.HasSubStream1 {
		d.SubStream1 = uint8(i[offset])
		offset += 1
	}
	if d.HasSubStream2 {
		d.SubStream2 = uint8(i[offset])
		offset += 1
	}
	if d.HasSubStream3 {
		d.SubStream3 = uint8(i[offset])
		offset += 1
	}
	for offset < len(i) {
		d.AdditionalInfo = append(d.AdditionalInfo, i[offset])
		offset += 1
	}
	return
}

// DescriptorExtendedEvent represents an extended event descriptor
type DescriptorExtendedEvent struct {
	ISO639LanguageCode   []byte
	Items                []*DescriptorExtendedEventItem
	LastDescriptorNumber uint8
	Number               uint8
	Text                 []byte
}

// DescriptorExtendedEventItem represents an extended event item descriptor
type DescriptorExtendedEventItem struct {
	Content     []byte
	Description []byte
}

func newDescriptorExtendedEvent(i []byte) (d *DescriptorExtendedEvent) {
	// Init
	d = &DescriptorExtendedEvent{}
	var offset int

	// Number
	d.Number = uint8(i[offset] >> 4)

	// Last descriptor number
	d.LastDescriptorNumber = uint8(i[offset] & 0xf)
	offset += 1

	// ISO 639 language code
	d.ISO639LanguageCode = i[offset : offset+3]
	offset += 3

	// Items length
	var itemsLength = int(i[offset])
	offset += 1

	// Items
	var offsetEnd = offset + itemsLength
	for offset < offsetEnd {
		d.Items = append(d.Items, newDescriptorExtendedEventItem(i, &offset))
	}

	// Text length
	var textLength = int(i[offset])
	offset += 1

	// Text
	offsetEnd = offset + textLength
	for offset < offsetEnd {
		d.Text = append(d.Text, i[offset])
		offset += 1
	}
	return
}

func newDescriptorExtendedEventItem(i []byte, offset *int) (d *DescriptorExtendedEventItem) {
	// Init
	d = &DescriptorExtendedEventItem{}

	// Description length
	var descriptionLength = int(i[*offset])
	*offset += 1

	// Description
	var offsetEnd = *offset + descriptionLength
	for *offset < offsetEnd {
		d.Description = append(d.Description, i[*offset])
		*offset += 1
	}

	// Content length
	var contentLength = int(i[*offset])
	*offset += 1

	// Content
	offsetEnd = *offset + contentLength
	for *offset < offsetEnd {
		d.Content = append(d.Content, i[*offset])
		*offset += 1
	}
	return
}

// DescriptorExtension represents an extension descriptor
// Page: 72 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorExtension struct {
	SupplementaryAudio *DescriptorExtensionSupplementaryAudio
	Tag                uint8
}

func newDescriptorExtension(i []byte) (d *DescriptorExtension) {
	// Init
	d = &DescriptorExtension{Tag: uint8(i[0])}

	// Switch on tag
	var b = i[1:]
	switch d.Tag {
	case DescriptorTagExtensionSupplementaryAudio:
		d.SupplementaryAudio = newDescriptorExtensionSupplementaryAudio(b)
	default:
		// TODO Remove this log
		astilog.Debugf("unlisted extension tag 0x%x", d.Tag)
	}
	return
}

// DescriptorExtensionSupplementaryAudio represents a supplementary audio extension descriptor
// Page: 130 | https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorExtensionSupplementaryAudio struct {
	EditorialClassification uint8
	HasLanguageCode         bool
	LanguageCode            []byte
	MixType                 bool
	PrivateData             []byte
}

func newDescriptorExtensionSupplementaryAudio(i []byte) (d *DescriptorExtensionSupplementaryAudio) {
	// Init
	d = &DescriptorExtensionSupplementaryAudio{}
	var offset int

	// Mix type
	d.MixType = i[offset]&0x80 > 0

	// Editorial classification
	d.EditorialClassification = uint8(i[offset] >> 2 & 0x1f)

	// Language code flag
	d.HasLanguageCode = i[offset]&0x1 > 0
	offset += 1

	// Language code
	if d.HasLanguageCode {
		d.LanguageCode = i[offset : offset+3]
		offset += 3
	}

	// Private data
	for offset < len(i) {
		d.PrivateData = append(d.PrivateData, i[offset])
		offset += 1
	}
	return
}

// DescriptorISO639LanguageAndAudioType represents an ISO639 language descriptor
type DescriptorISO639LanguageAndAudioType struct {
	Language []byte
	Type     uint8
}

func newDescriptorISO639LanguageAndAudioType(i []byte) *DescriptorISO639LanguageAndAudioType {
	return &DescriptorISO639LanguageAndAudioType{
		Language: i[0:3],
		Type:     uint8(i[3]),
	}
}

// DescriptorLocalTimeOffset represents a local time offset descriptor
// Page: 84 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorLocalTimeOffset struct {
	Items []*DescriptorLocalTimeOffsetItem
}

// DescriptorLocalTimeOffsetItem represents a local time offset item descriptor
type DescriptorLocalTimeOffsetItem struct {
	CountryCode             []byte
	CountryRegionID         uint8
	LocalTimeOffset         time.Duration
	LocalTimeOffsetPolarity bool
	NextTimeOffset          time.Duration
	TimeOfChange            time.Time
}

func newDescriptorLocalTimeOffset(i []byte) (d *DescriptorLocalTimeOffset) {
	// Init
	d = &DescriptorLocalTimeOffset{}
	var offset int

	// Add items
	for offset < len(i) {
		// Init
		var itm = &DescriptorLocalTimeOffsetItem{}
		d.Items = append(d.Items, itm)

		// Country code
		itm.CountryCode = i[offset : offset+3]
		offset += 3

		// Country region ID
		itm.CountryRegionID = uint8(i[offset] >> 2)

		// Local time offset polarity
		itm.LocalTimeOffsetPolarity = i[offset]&0x1 > 0
		offset += 1

		// Local time offset
		itm.LocalTimeOffset = parseDVBDurationMinutes(i, &offset)

		// Time of change
		itm.TimeOfChange = parseDVBTime(i, &offset)

		// Next time offset
		itm.NextTimeOffset = parseDVBDurationMinutes(i, &offset)
	}
	return
}

// DescriptorMaximumBitrate represents a maximum bitrate descriptor
type DescriptorMaximumBitrate struct {
	Bitrate uint32 // In bytes/second
}

func newDescriptorMaximumBitrate(i []byte) *DescriptorMaximumBitrate {
	return &DescriptorMaximumBitrate{Bitrate: (uint32(i[0]&0x3f)<<16 | uint32(i[1])<<8 | uint32(i[2])) * 50}
}

// DescriptorNetworkName represents a network name descriptor
// Page: 93 | Chapter: 6.2.27 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorNetworkName struct {
	Name []byte
}

func newDescriptorNetworkName(i []byte) *DescriptorNetworkName {
	return &DescriptorNetworkName{Name: i}
}

// DescriptorParentalRating represents a parental rating descriptor
// Page: 93 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorParentalRating struct {
	Items []*DescriptorParentalRatingItem
}

// DescriptorParentalRatingItem represents a parental rating item descriptor
type DescriptorParentalRatingItem struct {
	CountryCode []byte
	Rating      uint8
}

// MinimumAge returns the minimum age for the parental rating
func (d DescriptorParentalRatingItem) MinimumAge() int {
	// Undefined or user defined ratings
	if d.Rating == 0 || d.Rating > 0x10 {
		return 0
	}
	return int(d.Rating) + 3
}

func newDescriptorParentalRating(i []byte) (d *DescriptorParentalRating) {
	// Init
	d = &DescriptorParentalRating{}
	var offset int

	// Add items
	for offset < len(i) {
		d.Items = append(d.Items, &DescriptorParentalRatingItem{
			CountryCode: i[offset : offset+3],
			Rating:      uint8(i[offset+3]),
		})
		offset += 4
	}
	return
}

// DescriptorService represents a service descriptor
// Page: 96 | Chapter: 6.2.33 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorService struct {
	Name     []byte
	Provider []byte
	Type     uint8
}

func newDescriptorService(i []byte) (d *DescriptorService) {
	var offset int
	d = &DescriptorService{Type: uint8(i[offset])}
	offset += 1
	var providerLength = int(i[offset])
	offset += 1
	d.Provider = i[offset : offset+providerLength]
	offset += providerLength
	var nameLength = int(i[offset])
	offset += 1
	d.Name = i[offset : offset+nameLength]
	return
}

// DescriptorShortEvent represents a short event descriptor
// Page: 99 | Chapter: 6.2.37 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorShortEvent struct {
	EventName []byte
	Language  []byte
	Text      []byte
}

func newDescriptorShortEvent(i []byte) (d *DescriptorShortEvent) {
	var offset int
	d = &DescriptorShortEvent{}
	d.Language = i[:3]
	offset += 3
	var length = int(i[offset])
	offset += 1
	d.EventName = i[offset : offset+length]
	offset += length
	length = int(i[offset])
	offset += 1
	d.Text = i[offset : offset+length]
	return
}

// DescriptorStreamIdentifier represents a stream identifier descriptor
// Page: 102 | Chapter: 6.2.39 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorStreamIdentifier struct{ ComponentTag uint8 }

func newDescriptorStreamIdentifier(i []byte) *DescriptorStreamIdentifier {
	return &DescriptorStreamIdentifier{ComponentTag: uint8(i[0])}
}

// DescriptorSubtitling represents a subtitling descriptor
// Page: 103 | Chapter: 6.2.41 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorSubtitling struct {
	Items []*DescriptorSubtitlingItem
}

// DescriptorSubtitlingItem represents subtitling descriptor item
type DescriptorSubtitlingItem struct {
	AncillaryPageID   uint16
	CompositionPageID uint16
	Language          []byte
	Type              uint8
}

func newDescriptorSubtitling(i []byte) (d *DescriptorSubtitling) {
	d = &DescriptorSubtitling{}
	var offset int
	for offset < len(i) {
		itm := &DescriptorSubtitlingItem{}
		itm.Language = i[offset : offset+3]
		offset += 3
		itm.Type = uint8(i[offset])
		offset += 1
		itm.CompositionPageID = uint16(i[offset])<<8 | uint16(i[offset+1])
		offset += 2
		itm.AncillaryPageID = uint16(i[offset])<<8 | uint16(i[offset+1])
		offset += 2
		d.Items = append(d.Items, itm)
	}
	return
}

// DescriptorTeletext represents a teletext descriptor
// Page: 105 | Chapter: 6.2.43 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorTeletext struct {
	Items []*DescriptorTeletextItem
}

// DescriptorTeletextItem represents a teletext descriptor item
type DescriptorTeletextItem struct {
	Language []byte
	Magazine uint8
	Page     uint8
	Type     uint8
}

func newDescriptorTeletext(i []byte) (d *DescriptorTeletext) {
	var offset int
	d = &DescriptorTeletext{}
	for offset < len(i) {
		itm := &DescriptorTeletextItem{}
		itm.Language = i[offset : offset+3]
		offset += 3
		itm.Type = uint8(i[offset]) >> 3
		itm.Magazine = uint8(i[offset] & 0x7)
		offset += 1
		itm.Page = uint8(i[offset])>>4*10 + uint8(i[offset]&0xf)
		offset += 1
		d.Items = append(d.Items, itm)
	}
	return
}

// parseDescriptors parses descriptors
func parseDescriptors(i []byte, offset *int) (o []*Descriptor) {
	// Get length
	var length = int(uint16(i[*offset]&0xf)<<8 | uint16(i[*offset+1]))
	*offset += 2

	// Loop
	if length > 0 {
		length += *offset
		for *offset < length {
			// Init
			var d = &Descriptor{
				Length: uint8(i[*offset+1]),
				Tag:    uint8(i[*offset]),
			}
			*offset += 2

			// Parse data
			if d.Length > 0 {
				// Switch on tag
				var b = i[*offset : *offset+int(d.Length)]
				switch d.Tag {
				case DescriptorTagAC3:
					d.AC3 = newDescriptorAC3(b)
				case DescriptorTagComponent:
					d.Component = newDescriptorComponent(b)
				case DescriptorTagContent:
					d.Content = newDescriptorContent(b)
				case DescriptorTagEnhancedAC3:
					d.EnhancedAC3 = newDescriptorEnhancedAC3(b)
				case DescriptorTagExtendedEvent:
					d.ExtendedEvent = newDescriptorExtendedEvent(b)
				case DescriptorTagExtension:
					d.Extension = newDescriptorExtension(b)
				case DescriptorTagISO639LanguageAndAudioType:
					d.ISO639LanguageAndAudioType = newDescriptorISO639LanguageAndAudioType(b)
				case DescriptorTagLocalTimeOffset:
					d.LocalTimeOffset = newDescriptorLocalTimeOffset(b)
				case DescriptorTagMaximumBitrate:
					d.MaximumBitrate = newDescriptorMaximumBitrate(b)
				case DescriptorTagNetworkName:
					d.NetworkName = newDescriptorNetworkName(b)
				case DescriptorTagParentalRating:
					d.ParentalRating = newDescriptorParentalRating(b)
				case DescriptorTagService:
					d.Service = newDescriptorService(b)
				case DescriptorTagShortEvent:
					d.ShortEvent = newDescriptorShortEvent(b)
				case DescriptorTagStreamIdentifier:
					d.StreamIdentifier = newDescriptorStreamIdentifier(b)
				case DescriptorTagSubtitling:
					d.Subtitling = newDescriptorSubtitling(b)
				case DescriptorTagTeletext:
					d.Teletext = newDescriptorTeletext(b)
				default:
					// TODO Remove this log
					astilog.Debugf("unlisted descriptor tag 0x%x", d.Tag)
				}
				*offset += int(d.Length)
			}
			o = append(o, d)
		}
	}
	return
}
