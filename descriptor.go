package astits

import (
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
	DescriptorTagExtendedEvent              = 0x4e
	DescriptorTagISO639LanguageAndAudioType = 0xa
	DescriptorTagMaximumBitrate             = 0xe
	DescriptorTagNetworkName                = 0x40
	DescriptorTagService                    = 0x48
	DescriptorTagShortEvent                 = 0x4d
	DescriptorTagStreamIdentifier           = 0x52
	DescriptorTagSubtitling                 = 0x59
	DescriptorTagTeletext                   = 0x56
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
	ExtendedEvent              *DescriptorExtendedEvent
	ISO639LanguageAndAudioType *DescriptorISO639LanguageAndAudioType
	Length                     uint8
	MaximumBitrate             *DescriptorMaximumBitrate
	NetworkName                *DescriptorNetworkName
	Service                    *DescriptorService
	ShortEvent                 *DescriptorShortEvent
	StreamIdentifier           *DescriptorStreamIdentifier
	Subtitling                 *DescriptorSubtitling
	Tag                        uint8 // the tag defines the structure of the contained data following the descriptor length.
	Teletext                   *DescriptorTeletext
}

// DescriptorAC3 represents an AC3 descriptor
// Page: 165 | https://books.google.fr/books?id=6dgWB3-rChYC&printsec=frontcover&hl=fr
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

// DescriptorMaximumBitrate represents a maximum bitrate descriptor
type DescriptorMaximumBitrate struct {
	Bitrate uint32 // In bytes/second
}

func newDescriptorMaximumBitrate(i []byte) *DescriptorMaximumBitrate {
	return &DescriptorMaximumBitrate{Bitrate: (uint32(i[0]&0x3f)<<16 | uint32(i[1])<<8 | uint32(i[2])) * 50}
}

// DescriptorNetworkName represents a network name descriptor
// Page: 93 | Chapter: 6.2.27 | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
type DescriptorNetworkName struct{ Name []byte }

func newDescriptorNetworkName(i []byte) *DescriptorNetworkName {
	return &DescriptorNetworkName{Name: i}
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
				case DescriptorTagExtendedEvent:
					d.ExtendedEvent = newDescriptorExtendedEvent(b)
				case DescriptorTagISO639LanguageAndAudioType:
					d.ISO639LanguageAndAudioType = newDescriptorISO639LanguageAndAudioType(b)
				case DescriptorTagMaximumBitrate:
					d.MaximumBitrate = newDescriptorMaximumBitrate(b)
				case DescriptorTagNetworkName:
					d.NetworkName = newDescriptorNetworkName(b)
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
