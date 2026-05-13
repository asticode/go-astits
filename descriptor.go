package astits

import (
	"fmt"
	"time"

	"github.com/asticode/go-astikit"
)

// Audio types
// Page: 683 | https://books.google.fr/books?id=6dgWB3-rChYC&printsec=frontcover&hl=fr
const (
	AudioTypeCleanEffects             = 0x1
	AudioTypeHearingImpaired          = 0x2
	AudioTypeVisualImpairedCommentary = 0x3
)

// Data stream alignments
// Page: 85 | Chapter:2.6.11 | Link: http://ecee.colorado.edu/~ecen5653/ecen5653/papers/iso13818-1.pdf
const (
	DataStreamAligmentAudioSyncWord          = 0x1
	DataStreamAligmentVideoSliceOrAccessUnit = 0x1
	DataStreamAligmentVideoAccessUnit        = 0x2
	DataStreamAligmentVideoGOPOrSEQ          = 0x3
	DataStreamAligmentVideoSEQ               = 0x4
)

// Descriptor tags
// Chapter: 6.1 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	DescriptorTagAC3                        = 0x6a
	DescriptorTagAVCVideo                   = 0x28
	DescriptorTagComponent                  = 0x50
	DescriptorTagContent                    = 0x54
	DescriptorTagDataStreamAlignment        = 0x6
	DescriptorTagEnhancedAC3                = 0x7a
	DescriptorTagExtendedEvent              = 0x4e
	DescriptorTagExtension                  = 0x7f
	DescriptorTagISO639LanguageAndAudioType = 0xa
	DescriptorTagLocalTimeOffset            = 0x58
	DescriptorTagMaximumBitrate             = 0xe
	DescriptorTagNetworkName                = 0x40
	DescriptorTagParentalRating             = 0x55
	DescriptorTagPrivateDataIndicator       = 0xf
	DescriptorTagPrivateDataSpecifier       = 0x5f
	DescriptorTagRegistration               = 0x5
	DescriptorTagService                    = 0x48
	DescriptorTagShortEvent                 = 0x4d
	DescriptorTagStreamIdentifier           = 0x52
	DescriptorTagSubtitling                 = 0x59
	DescriptorTagTeletext                   = 0x56
	DescriptorTagSatelliteDeliverySystem    = 0x43
	DescriptorTagCableDeliverySystem        = 0x44
	DescriptorTagVBIData                    = 0x45
	DescriptorTagVBITeletext                = 0x46
	DescriptorTagTerrestrialDeliverySystem  = 0x5A
)

// Descriptor extension tags
// Chapter: 6.3 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	DescriptorTagExtensionT2DeliverySystem   = 0x04
	DescriptorTagExtensionSupplementaryAudio = 0x6
)

// Service types
// Chapter: 6.2.33 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	ServiceTypeDigitalTelevisionService = 0x1
)

// Teletext types
// Chapter: 6.2.43 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	TeletextTypeAdditionalInformationPage                    = 0x3
	TeletextTypeInitialTeletextPage                          = 0x1
	TeletextTypeProgramSchedulePage                          = 0x4
	TeletextTypeTeletextSubtitlePage                         = 0x2
	TeletextTypeTeletextSubtitlePageForHearingImpairedPeople = 0x5
)

// VBI data service id
// Chapter: 6.2.47 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	VBIDataServiceIDClosedCaptioning     = 0x6
	VBIDataServiceIDEBUTeletext          = 0x1
	VBIDataServiceIDInvertedTeletext     = 0x2
	VBIDataServiceIDMonochrome442Samples = 0x7
	VBIDataServiceIDVPS                  = 0x4
	VBIDataServiceIDWSS                  = 0x5
)

// Descriptor represents a descriptor
// TODO Handle UTF8
type Descriptor struct {
	AC3                        *DescriptorAC3
	AVCVideo                   *DescriptorAVCVideo
	CableDeliverySystem        *DescriptorCableDeliverySystem
	Component                  *DescriptorComponent
	Content                    *DescriptorContent
	DataStreamAlignment        *DescriptorDataStreamAlignment
	EnhancedAC3                *DescriptorEnhancedAC3
	ExtendedEvent              *DescriptorExtendedEvent
	Extension                  *DescriptorExtension
	ISO639LanguageAndAudioType *DescriptorISO639LanguageAndAudioType
	Length                     uint8
	LocalTimeOffset            *DescriptorLocalTimeOffset
	MaximumBitrate             *DescriptorMaximumBitrate
	NetworkName                *DescriptorNetworkName
	ParentalRating             *DescriptorParentalRating
	PrivateDataIndicator       *DescriptorPrivateDataIndicator
	PrivateDataSpecifier       *DescriptorPrivateDataSpecifier
	Registration               *DescriptorRegistration
	SatelliteDeliverySystem    *DescriptorSatelliteDeliverySystem
	Service                    *DescriptorService
	ShortEvent                 *DescriptorShortEvent
	StreamIdentifier           *DescriptorStreamIdentifier
	Subtitling                 *DescriptorSubtitling
	Tag                        uint8 // the tag defines the structure of the contained data following the descriptor length.
	Teletext                   *DescriptorTeletext
	TerrestrialDeliverySystem  *DescriptorTerrestrialDeliverySystem
	Unknown                    *DescriptorUnknown
	UserDefined                []byte
	VBIData                    *DescriptorVBIData
	VBITeletext                *DescriptorTeletext
}

// DescriptorAC3 represents an AC3 descriptor
// Chapter: Annex D | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
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

func newDescriptorAC3(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorAC3, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorAC3{
		HasASVC:          uint8(b&0x10) > 0,
		HasBSID:          uint8(b&0x40) > 0,
		HasComponentType: uint8(b&0x80) > 0,
		HasMainID:        uint8(b&0x20) > 0,
	}

	// Component type
	if d.HasComponentType {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.ComponentType = uint8(b)
	}

	// BSID
	if d.HasBSID {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.BSID = uint8(b)
	}

	// Main ID
	if d.HasMainID {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.MainID = uint8(b)
	}

	// ASVC
	if d.HasASVC {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.ASVC = uint8(b)
	}

	// Additional info
	if i.Offset() < offsetEnd {
		if d.AdditionalInfo, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}
	return
}

// DescriptorAVCVideo represents an AVC video descriptor
// No doc found unfortunately, basing the implementation on https://github.com/gfto/bitstream/blob/master/mpeg/psi/desc_28.h
type DescriptorAVCVideo struct {
	AVC24HourPictureFlag bool
	AVCStillPresent      bool
	CompatibleFlags      uint8
	ConstraintSet0Flag   bool
	ConstraintSet1Flag   bool
	ConstraintSet2Flag   bool
	LevelIDC             uint8
	ProfileIDC           uint8
}

func newDescriptorAVCVideo(i *astikit.BytesIterator) (d *DescriptorAVCVideo, err error) {
	// Init
	d = &DescriptorAVCVideo{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Profile idc
	d.ProfileIDC = uint8(b)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Flags
	d.ConstraintSet0Flag = b&0x80 > 0
	d.ConstraintSet1Flag = b&0x40 > 0
	d.ConstraintSet2Flag = b&0x20 > 0
	d.CompatibleFlags = b & 0x1f

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Level idc
	d.LevelIDC = uint8(b)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// AVC still present
	d.AVCStillPresent = b&0x80 > 0

	// AVC 24 hour picture flag
	d.AVC24HourPictureFlag = b&0x40 > 0
	return
}

// DescriptorComponent represents a component descriptor
// Chapter: 6.2.8 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorComponent struct {
	ComponentTag       uint8
	ComponentType      uint8
	ISO639LanguageCode []byte
	StreamContent      uint8
	StreamContentExt   uint8
	Text               []byte
}

func newDescriptorComponent(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorComponent, err error) {
	// Init
	d = &DescriptorComponent{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Stream content ext
	d.StreamContentExt = uint8(b >> 4)

	// Stream content
	d.StreamContent = uint8(b & 0xf)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Component type
	d.ComponentType = uint8(b)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Component tag
	d.ComponentTag = uint8(b)

	// ISO639 language code
	if d.ISO639LanguageCode, err = i.NextBytes(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Text
	if i.Offset() < offsetEnd {
		if d.Text, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}
	return
}

// DescriptorContent represents a content descriptor
// Chapter: 6.2.9 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorContent struct {
	Items []*DescriptorContentItem
}

// DescriptorContentItem represents a content item descriptor
// Chapter: 6.2.9 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorContentItem struct {
	ContentNibbleLevel1 uint8
	ContentNibbleLevel2 uint8
	UserByte            uint8
}

func newDescriptorContent(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorContent, err error) {
	// Init
	d = &DescriptorContent{}

	// Add items
	for i.Offset() < offsetEnd {
		// Get next bytes
		var bs []byte
		if bs, err = i.NextBytesNoCopy(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Append item
		d.Items = append(d.Items, &DescriptorContentItem{
			ContentNibbleLevel1: uint8(bs[0] >> 4),
			ContentNibbleLevel2: uint8(bs[0] & 0xf),
			UserByte:            uint8(bs[1]),
		})
	}
	return
}

// DescriptorDataStreamAlignment represents a data stream alignment descriptor
type DescriptorDataStreamAlignment struct {
	Type uint8
}

func newDescriptorDataStreamAlignment(i *astikit.BytesIterator) (d *DescriptorDataStreamAlignment, err error) {
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}
	d = &DescriptorDataStreamAlignment{Type: uint8(b)}
	return
}

// DescriptorEnhancedAC3 represents an enhanced AC3 descriptor
// Chapter: Annex D | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
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

func newDescriptorEnhancedAC3(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorEnhancedAC3, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorEnhancedAC3{
		HasASVC:          uint8(b&0x10) > 0,
		HasBSID:          uint8(b&0x40) > 0,
		HasComponentType: uint8(b&0x80) > 0,
		HasMainID:        uint8(b&0x20) > 0,
		HasSubStream1:    uint8(b&0x4) > 0,
		HasSubStream2:    uint8(b&0x2) > 0,
		HasSubStream3:    uint8(b&0x1) > 0,
		MixInfoExists:    uint8(b&0x8) > 0,
	}

	// Component type
	if d.HasComponentType {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.ComponentType = uint8(b)
	}

	// BSID
	if d.HasBSID {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.BSID = uint8(b)
	}

	// Main ID
	if d.HasMainID {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.MainID = uint8(b)
	}

	// ASVC
	if d.HasASVC {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.ASVC = uint8(b)
	}

	// Substream 1
	if d.HasSubStream1 {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.SubStream1 = uint8(b)
	}

	// Substream 2
	if d.HasSubStream2 {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.SubStream2 = uint8(b)
	}

	// Substream 3
	if d.HasSubStream3 {
		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}
		d.SubStream3 = uint8(b)
	}

	// Additional info
	if i.Offset() < offsetEnd {
		if d.AdditionalInfo, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}
	return
}

// DescriptorExtendedEvent represents an extended event descriptor
// Chapter: 6.2.15 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtendedEvent struct {
	ISO639LanguageCode   []byte
	Items                []*DescriptorExtendedEventItem
	LastDescriptorNumber uint8
	Number               uint8
	Text                 []byte
}

// DescriptorExtendedEventItem represents an extended event item descriptor
// Chapter: 6.2.15 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtendedEventItem struct {
	Content     []byte
	Description []byte
}

func newDescriptorExtendedEvent(i *astikit.BytesIterator) (d *DescriptorExtendedEvent, err error) {
	// Init
	d = &DescriptorExtendedEvent{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Number
	d.Number = uint8(b >> 4)

	// Last descriptor number
	d.LastDescriptorNumber = uint8(b & 0xf)

	// ISO639 language code
	if d.ISO639LanguageCode, err = i.NextBytes(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Items length
	itemsLength := int(b)

	// Items
	offsetEnd := i.Offset() + itemsLength
	for i.Offset() < offsetEnd {
		// Create item
		var item *DescriptorExtendedEventItem
		if item, err = newDescriptorExtendedEventItem(i); err != nil {
			err = fmt.Errorf("astits: creating extended event item failed: %w", err)
			return
		}

		// Append item
		d.Items = append(d.Items, item)
	}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Text length
	textLength := int(b)

	// Text
	if d.Text, err = i.NextBytes(textLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

func newDescriptorExtendedEventItem(i *astikit.BytesIterator) (d *DescriptorExtendedEventItem, err error) {
	// Init
	d = &DescriptorExtendedEventItem{}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Description length
	descriptionLength := int(b)

	// Description
	if d.Description, err = i.NextBytes(descriptionLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Content length
	contentLength := int(b)

	// Content
	if d.Content, err = i.NextBytes(contentLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorExtension represents an extension descriptor
// Chapter: 6.2.16 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtension struct {
	SupplementaryAudio *DescriptorExtensionSupplementaryAudio
	T2DeliverySystem   *DescriptorExtensionT2DeliverySystem
	Tag                uint8
	Unknown            *[]byte
}

func newDescriptorExtension(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorExtension, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorExtension{Tag: uint8(b)}

	// Switch on tag
	switch d.Tag {
	case DescriptorTagExtensionT2DeliverySystem:
		if d.T2DeliverySystem, err = newDescriptorExtensionT2DeliverySystem(i, offsetEnd); err != nil {
			err = fmt.Errorf("astits: parsing T2 Delivery System extension descriptor failed: %w", err)
			return
		}
	case DescriptorTagExtensionSupplementaryAudio:
		if d.SupplementaryAudio, err = newDescriptorExtensionSupplementaryAudio(i, offsetEnd); err != nil {
			err = fmt.Errorf("astits: parsing extension supplementary audio descriptor failed: %w", err)
			return
		}
	default:
		// Get next bytes
		var b []byte
		if b, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Update unknown
		d.Unknown = &b
	}
	return
}

// DescriptorExtensionSupplementaryAudio represents a supplementary audio extension descriptor
// Chapter: 6.4.10 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtensionSupplementaryAudio struct {
	EditorialClassification uint8
	HasLanguageCode         bool
	LanguageCode            []byte
	MixType                 bool
	PrivateData             []byte
}

func newDescriptorExtensionSupplementaryAudio(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorExtensionSupplementaryAudio, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Init
	d = &DescriptorExtensionSupplementaryAudio{
		EditorialClassification: uint8(b >> 2 & 0x1f),
		HasLanguageCode:         b&0x1 > 0,
		MixType:                 b&0x80 > 0,
	}

	// Language code
	if d.HasLanguageCode {
		if d.LanguageCode, err = i.NextBytes(3); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}

	// Private data
	if i.Offset() < offsetEnd {
		if d.PrivateData, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}
	return
}

// DescriptorISO639LanguageAndAudioType represents an ISO639 language descriptor
// https://github.com/gfto/bitstream/blob/master/mpeg/psi/desc_0a.h
// FIXME (barbashov) according to Chapter 2.6.18 ISO/IEC 13818-1:2015 there could be not one, but multiple such descriptors
type DescriptorISO639LanguageAndAudioType struct {
	Language []byte
	Type     uint8
}

// In some actual cases, the length is 3 and the language is described in only 2 bytes
func newDescriptorISO639LanguageAndAudioType(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorISO639LanguageAndAudioType, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorISO639LanguageAndAudioType{
		Language: bs[0 : len(bs)-1],
		Type:     uint8(bs[len(bs)-1]),
	}
	return
}

// DescriptorLocalTimeOffset represents a local time offset descriptor
// Chapter: 6.2.20 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorLocalTimeOffset struct {
	Items []*DescriptorLocalTimeOffsetItem
}

// DescriptorLocalTimeOffsetItem represents a local time offset item descriptor
// Chapter: 6.2.20 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorLocalTimeOffsetItem struct {
	CountryCode             []byte
	CountryRegionID         uint8
	LocalTimeOffset         time.Duration
	LocalTimeOffsetPolarity bool
	NextTimeOffset          time.Duration
	TimeOfChange            time.Time
}

func newDescriptorLocalTimeOffset(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorLocalTimeOffset, err error) {
	// Init
	d = &DescriptorLocalTimeOffset{}

	// Add items
	for i.Offset() < offsetEnd {
		// Create item
		itm := &DescriptorLocalTimeOffsetItem{}

		// Country code
		if itm.CountryCode, err = i.NextBytes(3); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Country region ID
		itm.CountryRegionID = uint8(b >> 2)

		// Local time offset polarity
		itm.LocalTimeOffsetPolarity = b&0x1 > 0

		// Local time offset
		if itm.LocalTimeOffset, err = parseDVBDurationMinutes(i); err != nil {
			err = fmt.Errorf("astits: parsing DVB durationminutes failed: %w", err)
			return
		}

		// Time of change
		if itm.TimeOfChange, err = parseDVBTime(i); err != nil {
			err = fmt.Errorf("astits: parsing DVB time failed: %w", err)
			return
		}

		// Next time offset
		if itm.NextTimeOffset, err = parseDVBDurationMinutes(i); err != nil {
			err = fmt.Errorf("astits: parsing DVB duration minutes failed: %w", err)
			return
		}

		// Append item
		d.Items = append(d.Items, itm)
	}
	return
}

// DescriptorMaximumBitrate represents a maximum bitrate descriptor
type DescriptorMaximumBitrate struct {
	Bitrate uint32 // In bytes/second
}

func newDescriptorMaximumBitrate(i *astikit.BytesIterator) (d *DescriptorMaximumBitrate, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorMaximumBitrate{Bitrate: (uint32(bs[0]&0x3f)<<16 | uint32(bs[1])<<8 | uint32(bs[2])) * 50}
	return
}

// DescriptorNetworkName represents a network name descriptor
// Chapter: 6.2.27 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorNetworkName struct {
	Name []byte
}

func newDescriptorNetworkName(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorNetworkName, err error) {
	// Create descriptor
	d = &DescriptorNetworkName{}

	// Name
	if d.Name, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorParentalRating represents a parental rating descriptor
// Chapter: 6.2.28 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorParentalRating struct {
	Items []*DescriptorParentalRatingItem
}

// DescriptorParentalRatingItem represents a parental rating item descriptor
// Chapter: 6.2.28 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
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

func newDescriptorParentalRating(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorParentalRating, err error) {
	// Create descriptor
	d = &DescriptorParentalRating{}

	// Add items
	for i.Offset() < offsetEnd {
		// Get next bytes
		var bs []byte
		if bs, err = i.NextBytes(4); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Append item
		d.Items = append(d.Items, &DescriptorParentalRatingItem{
			CountryCode: bs[:3],
			Rating:      uint8(bs[3]),
		})
	}
	return
}

// DescriptorPrivateDataIndicator represents a private data Indicator descriptor
type DescriptorPrivateDataIndicator struct {
	Indicator uint32
}

func newDescriptorPrivateDataIndicator(i *astikit.BytesIterator) (d *DescriptorPrivateDataIndicator, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorPrivateDataIndicator{Indicator: uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])}
	return
}

// DescriptorPrivateDataSpecifier represents a private data specifier descriptor
type DescriptorPrivateDataSpecifier struct {
	Specifier uint32
}

func newDescriptorPrivateDataSpecifier(i *astikit.BytesIterator) (d *DescriptorPrivateDataSpecifier, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorPrivateDataSpecifier{Specifier: uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])}
	return
}

// DescriptorRegistration represents a registration descriptor
// Page: 84 | http://ecee.colorado.edu/~ecen5653/ecen5653/papers/iso13818-1.pdf
type DescriptorRegistration struct {
	AdditionalIdentificationInfo []byte
	FormatIdentifier             uint32
}

func newDescriptorRegistration(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorRegistration, err error) {
	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorRegistration{FormatIdentifier: uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])}

	// Additional identification info
	if i.Offset() < offsetEnd {
		if d.AdditionalIdentificationInfo, err = i.NextBytes(offsetEnd - i.Offset()); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}
	}
	return
}

// DescriptorService represents a service descriptor
// Chapter: 6.2.33 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorService struct {
	Name     []byte
	Provider []byte
	Type     uint8
}

func newDescriptorService(i *astikit.BytesIterator) (d *DescriptorService, err error) {
	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorService{Type: uint8(b)}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Provider length
	providerLength := int(b)

	// Provider
	if d.Provider, err = i.NextBytes(providerLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Name length
	nameLength := int(b)

	// Name
	if d.Name, err = i.NextBytes(nameLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorShortEvent represents a short event descriptor
// Chapter: 6.2.37 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorShortEvent struct {
	EventName []byte
	Language  []byte
	Text      []byte
}

func newDescriptorShortEvent(i *astikit.BytesIterator) (d *DescriptorShortEvent, err error) {
	// Create descriptor
	d = &DescriptorShortEvent{}

	// Language
	if d.Language, err = i.NextBytes(3); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Event length
	eventLength := int(b)

	// Event name
	if d.EventName, err = i.NextBytes(eventLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Text length
	textLength := int(b)

	// Text
	if d.Text, err = i.NextBytes(textLength); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorStreamIdentifier represents a stream identifier descriptor
// Chapter: 6.2.39 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorStreamIdentifier struct {
	ComponentTag uint8
}

func newDescriptorStreamIdentifier(i *astikit.BytesIterator) (d *DescriptorStreamIdentifier, err error) {
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}
	d = &DescriptorStreamIdentifier{ComponentTag: uint8(b)}
	return
}

// DescriptorSubtitling represents a subtitling descriptor
// Chapter: 6.2.41 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorSubtitling struct {
	Items []*DescriptorSubtitlingItem
}

// DescriptorSubtitlingItem represents subtitling descriptor item
// Chapter: 6.2.41 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorSubtitlingItem struct {
	AncillaryPageID   uint16
	CompositionPageID uint16
	Language          []byte
	Type              uint8
}

func newDescriptorSubtitling(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorSubtitling, err error) {
	// Create descriptor
	d = &DescriptorSubtitling{}

	// Loop
	for i.Offset() < offsetEnd {
		// Create item
		itm := &DescriptorSubtitlingItem{}

		// Language
		if itm.Language, err = i.NextBytes(3); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Type
		itm.Type = uint8(b)

		// Get next bytes
		var bs []byte
		if bs, err = i.NextBytesNoCopy(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Composition page ID
		itm.CompositionPageID = uint16(bs[0])<<8 | uint16(bs[1])

		// Get next bytes
		if bs, err = i.NextBytesNoCopy(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Ancillary page ID
		itm.AncillaryPageID = uint16(bs[0])<<8 | uint16(bs[1])

		// Append item
		d.Items = append(d.Items, itm)
	}
	return
}

// DescriptorTeletext represents a teletext descriptor
// Chapter: 6.2.43 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorTeletext struct {
	Items []*DescriptorTeletextItem
}

// DescriptorTeletextItem represents a teletext descriptor item
// Chapter: 6.2.43 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorTeletextItem struct {
	Language []byte
	Magazine uint8
	Page     uint8
	Type     uint8
}

func newDescriptorTeletext(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorTeletext, err error) {
	// Create descriptor
	d = &DescriptorTeletext{}

	// Loop
	for i.Offset() < offsetEnd {
		// Create item
		itm := &DescriptorTeletextItem{}

		// Language
		if itm.Language, err = i.NextBytes(3); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Type
		itm.Type = uint8(b >> 3)

		// Magazine
		itm.Magazine = uint8(b & 0x7)

		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Page
		itm.Page = uint8(b)>>4*10 + uint8(b&0xf)

		// Append item
		d.Items = append(d.Items, itm)
	}
	return
}

type DescriptorUnknown struct {
	Content []byte
	Tag     uint8
}

func newDescriptorUnknown(i *astikit.BytesIterator, tag, length uint8) (d *DescriptorUnknown, err error) {
	// Create descriptor
	d = &DescriptorUnknown{Tag: tag}

	// Get next bytes
	if d.Content, err = i.NextBytes(int(length)); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorVBIData represents a VBI data descriptor
// Chapter: 6.2.47 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIData struct {
	Services []*DescriptorVBIDataService
}

// DescriptorVBIDataService represents a vbi data service descriptor
// Chapter: 6.2.47 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIDataService struct {
	DataServiceID uint8
	Descriptors   []*DescriptorVBIDataDescriptor
}

// DescriptorVBIDataItem represents a vbi data descriptor item
// Chapter: 6.2.47 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIDataDescriptor struct {
	FieldParity bool
	LineOffset  uint8
}

func newDescriptorVBIData(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorVBIData, err error) {
	// Create descriptor
	d = &DescriptorVBIData{}

	// Loop
	for i.Offset() < offsetEnd {
		// Create service
		srv := &DescriptorVBIDataService{}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Data service ID
		srv.DataServiceID = uint8(b)

		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Data service descriptor length
		dataServiceDescriptorLength := int(b)

		// Data service descriptor
		offsetDataEnd := i.Offset() + dataServiceDescriptorLength
		for i.Offset() < offsetDataEnd {
			// Get next byte
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}

			if srv.DataServiceID == VBIDataServiceIDClosedCaptioning ||
				srv.DataServiceID == VBIDataServiceIDEBUTeletext ||
				srv.DataServiceID == VBIDataServiceIDInvertedTeletext ||
				srv.DataServiceID == VBIDataServiceIDMonochrome442Samples ||
				srv.DataServiceID == VBIDataServiceIDVPS ||
				srv.DataServiceID == VBIDataServiceIDWSS {

				// Append data
				srv.Descriptors = append(srv.Descriptors, &DescriptorVBIDataDescriptor{
					FieldParity: b&0x20 > 0,
					LineOffset:  uint8(b & 0x1f),
				})
			}
		}

		// Append service
		d.Services = append(d.Services, srv)
	}
	return
}

// decodeBCD decodes BCD-encoded bytes into an integer
func decodeBCD(data []byte) uint64 {
	var result uint64
	for _, b := range data {
		hi := uint64((b >> 4) & 0x0F)
		lo := uint64(b & 0x0F)
		result = result*100 + hi*10 + lo
	}
	return result
}

// DescriptorSatelliteDeliverySystem represents a satellite delivery system descriptor
// Chapter: 6.2.13.2 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorSatelliteDeliverySystem struct {
	Frequency        uint32 // in 10 kHz units (BCD decoded)
	OrbitalPosition  uint16 // BCD: e.g. 0x0282 = 28.2°
	WestEastFlag     bool   // false=west, true=east
	Polarization     uint8  // 0=H, 1=V, 2=left, 3=right
	RollOff          uint8  // 0=0.35, 1=0.25, 2=0.20
	ModulationSystem bool   // false=DVB-S, true=DVB-S2
	ModulationType   uint8  // 0=auto, 1=QPSK, 2=8PSK, 3=16-QAM
	SymbolRate       uint32 // in 100 sym/s units (BCD decoded)
	FECInner         uint8  // 0=not defined, 1=1/2, 2=2/3, 3=3/4, 4=5/6, 5=7/8, 6=8/9, 7=3/5, 8=4/5, 9=9/10
}

func newDescriptorSatelliteDeliverySystem(i *astikit.BytesIterator) (d *DescriptorSatelliteDeliverySystem, err error) {
	// Get frequency bytes (4 bytes BCD)
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorSatelliteDeliverySystem{
		Frequency: uint32(decodeBCD(bs)),
	}

	// Get orbital position bytes (2 bytes BCD)
	if bs, err = i.NextBytesNoCopy(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Orbital position
	d.OrbitalPosition = uint16(bs[0])<<8 | uint16(bs[1])

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// West/east flag
	d.WestEastFlag = b&0x80 > 0

	// Polarization
	d.Polarization = uint8((b >> 5) & 0x3)

	// Roll off
	d.RollOff = uint8((b >> 3) & 0x3)

	// Modulation system
	d.ModulationSystem = b&0x04 > 0

	// Modulation type
	d.ModulationType = uint8(b & 0x3)

	// Get symbol rate + FEC inner bytes (4 bytes)
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Symbol rate (top 28 bits = BCD, 3.5 bytes)
	srBytes := []byte{bs[0], bs[1], bs[2], (bs[3] >> 4) & 0x0F}
	d.SymbolRate = uint32(decodeBCD(srBytes[:3]))*10 + uint32((bs[3]>>4)&0x0F)

	// FEC inner (bottom 4 bits)
	d.FECInner = uint8(bs[3] & 0x0F)
	return
}

// DescriptorTerrestrialDeliverySystem represents a terrestrial delivery system descriptor
// Chapter: 6.2.13.4 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorTerrestrialDeliverySystem struct {
	Frequency            uint32 // centre frequency in 10 Hz units
	Bandwidth            uint8  // 0=8MHz, 1=7MHz, 2=6MHz, 3=5MHz
	Priority             bool
	TimeSlicingIndicator bool
	MPEFECIndicator      bool
	Constellation        uint8 // 0=QPSK, 1=16-QAM, 2=64-QAM
	HierarchyInformation uint8
	CodeRateHPStream     uint8
	CodeRateLPStream     uint8
	GuardInterval        uint8 // 0=1/32, 1=1/16, 2=1/8, 3=1/4
	TransmissionMode     uint8 // 0=2k, 1=8k, 2=4k
	OtherFrequencyFlag   bool
}

func newDescriptorTerrestrialDeliverySystem(i *astikit.BytesIterator) (d *DescriptorTerrestrialDeliverySystem, err error) {
	// Get frequency bytes (4 bytes, big-endian uint32)
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorTerrestrialDeliverySystem{
		Frequency: uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3]),
	}

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Bandwidth
	d.Bandwidth = uint8((b >> 5) & 0x7)

	// Priority
	d.Priority = b&0x10 > 0

	// Time slicing indicator
	d.TimeSlicingIndicator = b&0x08 > 0

	// MPE-FEC indicator
	d.MPEFECIndicator = b&0x04 > 0

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Constellation
	d.Constellation = uint8((b >> 6) & 0x3)

	// Hierarchy information
	d.HierarchyInformation = uint8((b >> 3) & 0x7)

	// Code rate HP stream
	d.CodeRateHPStream = uint8(b & 0x7)

	// Get next byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Code rate LP stream
	d.CodeRateLPStream = uint8((b >> 5) & 0x7)

	// Guard interval
	d.GuardInterval = uint8((b >> 3) & 0x3)

	// Transmission mode
	d.TransmissionMode = uint8((b >> 1) & 0x3)

	// Other frequency flag
	d.OtherFrequencyFlag = b&0x1 > 0

	// Skip reserved bytes (4 bytes)
	if _, err = i.NextBytes(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}
	return
}

// DescriptorCableDeliverySystem represents a cable delivery system descriptor
// Chapter: 6.2.13.1 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorCableDeliverySystem struct {
	Frequency  uint32 // in 100 Hz units (BCD decoded)
	FECOuter   uint8
	Modulation uint8  // 0=not defined, 1=16-QAM, 2=32-QAM, 3=64-QAM, 4=128-QAM, 5=256-QAM
	SymbolRate uint32 // in 100 sym/s units (BCD decoded)
	FECInner   uint8
}

func newDescriptorCableDeliverySystem(i *astikit.BytesIterator) (d *DescriptorCableDeliverySystem, err error) {
	// Get frequency bytes (4 bytes BCD)
	var bs []byte
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorCableDeliverySystem{
		Frequency: uint32(decodeBCD(bs)),
	}

	// Get reserved + FEC outer bytes (2 bytes)
	if bs, err = i.NextBytesNoCopy(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// FEC outer (bottom 4 bits of second byte)
	d.FECOuter = uint8(bs[1] & 0x0F)

	// Get next byte
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Modulation
	d.Modulation = uint8(b)

	// Get symbol rate + FEC inner bytes (4 bytes)
	if bs, err = i.NextBytesNoCopy(4); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Symbol rate (top 28 bits = BCD, 3.5 bytes)
	d.SymbolRate = uint32(decodeBCD(bs[:3]))*10 + uint32((bs[3]>>4)&0x0F)

	// FEC inner (bottom 4 bits)
	d.FECInner = uint8(bs[3] & 0x0F)
	return
}

// DescriptorExtensionT2DeliverySystem represents a T2 delivery system extension descriptor
// Chapter: 6.4.6 | Link: https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtensionT2DeliverySystem struct {
	PLPID              uint8
	T2SystemID         uint16
	HasExtendedInfo    bool
	SISOorMISO         uint8 // 0=SISO, 1=MISO
	Bandwidth          uint8 // 0=8MHz, 1=7MHz, 2=6MHz, 3=5MHz, 4=10MHz, 5=1.712MHz
	GuardInterval      uint8
	TransmissionMode   uint8
	OtherFrequencyFlag bool
	TFSFlag            bool
	Cells              []T2Cell
}

// T2Cell represents a cell in the T2 delivery system descriptor
type T2Cell struct {
	CellID            uint16
	CentreFrequencies []uint32 // in 10 Hz units
}

func newDescriptorExtensionT2DeliverySystem(i *astikit.BytesIterator, offsetEnd int) (d *DescriptorExtensionT2DeliverySystem, err error) {
	// Get PLP ID
	var b byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("astits: fetching next byte failed: %w", err)
		return
	}

	// Create descriptor
	d = &DescriptorExtensionT2DeliverySystem{
		PLPID: uint8(b),
	}

	// Get T2 system ID (2 bytes)
	var bs []byte
	if bs, err = i.NextBytesNoCopy(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// T2 system ID
	d.T2SystemID = uint16(bs[0])<<8 | uint16(bs[1])

	// Check for extended info
	if i.Offset() < offsetEnd {
		d.HasExtendedInfo = true

		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// SISO/MISO
		d.SISOorMISO = uint8((b >> 6) & 0x3)

		// Bandwidth
		d.Bandwidth = uint8((b >> 2) & 0x0F)

		// Get next byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Guard interval
		d.GuardInterval = uint8((b >> 5) & 0x7)

		// Transmission mode
		d.TransmissionMode = uint8((b >> 2) & 0x7)

		// Other frequency flag
		d.OtherFrequencyFlag = b&0x02 > 0

		// TFS flag
		d.TFSFlag = b&0x01 > 0

		// Cell loop
		for i.Offset() < offsetEnd {
			var cell T2Cell

			// Cell ID (2 bytes)
			if bs, err = i.NextBytesNoCopy(2); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}
			cell.CellID = uint16(bs[0])<<8 | uint16(bs[1])

			// Frequency loop length
			if b, err = i.NextByte(); err != nil {
				err = fmt.Errorf("astits: fetching next byte failed: %w", err)
				return
			}
			freqLoopLength := int(b)

			// Frequencies (4 bytes each)
			freqEnd := i.Offset() + freqLoopLength
			for i.Offset() < freqEnd {
				if bs, err = i.NextBytesNoCopy(4); err != nil {
					err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
					return
				}
				freq := uint32(bs[0])<<24 | uint32(bs[1])<<16 | uint32(bs[2])<<8 | uint32(bs[3])
				cell.CentreFrequencies = append(cell.CentreFrequencies, freq)
			}

			d.Cells = append(d.Cells, cell)
		}
	}
	return
}

// parseDescriptors parses descriptors
func parseDescriptors(i *astikit.BytesIterator) (o []*Descriptor, err error) {
	// Get next 2 bytes
	var bs []byte
	if bs, err = i.NextBytesNoCopy(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// Get length
	length := int(uint16(bs[0]&0xf)<<8 | uint16(bs[1]))

	// Loop
	if length > 0 {
		offsetEnd := i.Offset() + length
		for i.Offset() < offsetEnd {
			// Get next 2 bytes
			if bs, err = i.NextBytesNoCopy(2); err != nil {
				err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
				return
			}

			// Create descriptor
			d := &Descriptor{
				Length: uint8(bs[1]),
				Tag:    uint8(bs[0]),
			}

			// Parse data
			if d.Length > 0 {
				// Unfortunately there's no way to be sure the real descriptor length is the same as the one indicated
				// previously therefore we must fetch bytes in descriptor functions and seek at the end
				offsetDescriptorEnd := i.Offset() + int(d.Length)

				// User defined
				if d.Tag >= 0x80 && d.Tag <= 0xfe {
					// Get next bytes
					if d.UserDefined, err = i.NextBytes(int(d.Length)); err != nil {
						err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
						return
					}
				} else {
					// Switch on tag
					switch d.Tag {
					case DescriptorTagAC3:
						if d.AC3, err = newDescriptorAC3(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing AC3 descriptor failed: %w", err)
							return
						}
					case DescriptorTagAVCVideo:
						if d.AVCVideo, err = newDescriptorAVCVideo(i); err != nil {
							err = fmt.Errorf("astits: parsing AVC Video descriptor failed: %w", err)
							return
						}
					case DescriptorTagComponent:
						if d.Component, err = newDescriptorComponent(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Component descriptor failed: %w", err)
							return
						}
					case DescriptorTagContent:
						if d.Content, err = newDescriptorContent(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Content descriptor failed: %w", err)
							return
						}
					case DescriptorTagDataStreamAlignment:
						if d.DataStreamAlignment, err = newDescriptorDataStreamAlignment(i); err != nil {
							err = fmt.Errorf("astits: parsing Data Stream Alignment descriptor failed: %w", err)
							return
						}
					case DescriptorTagEnhancedAC3:
						if d.EnhancedAC3, err = newDescriptorEnhancedAC3(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Enhanced AC3 descriptor failed: %w", err)
							return
						}
					case DescriptorTagExtendedEvent:
						if d.ExtendedEvent, err = newDescriptorExtendedEvent(i); err != nil {
							err = fmt.Errorf("astits: parsing Extended event descriptor failed: %w", err)
							return
						}
					case DescriptorTagExtension:
						if d.Extension, err = newDescriptorExtension(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Extension descriptor failed: %w", err)
							return
						}
					case DescriptorTagISO639LanguageAndAudioType:
						if d.ISO639LanguageAndAudioType, err = newDescriptorISO639LanguageAndAudioType(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing ISO639 Language and Audio Type descriptor failed: %w", err)
							return
						}
					case DescriptorTagLocalTimeOffset:
						if d.LocalTimeOffset, err = newDescriptorLocalTimeOffset(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Local Time Offset descriptor failed: %w", err)
							return
						}
					case DescriptorTagMaximumBitrate:
						if d.MaximumBitrate, err = newDescriptorMaximumBitrate(i); err != nil {
							err = fmt.Errorf("astits: parsing Maximum Bitrate descriptor failed: %w", err)
							return
						}
					case DescriptorTagNetworkName:
						if d.NetworkName, err = newDescriptorNetworkName(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Network Name descriptor failed: %w", err)
							return
						}
					case DescriptorTagParentalRating:
						if d.ParentalRating, err = newDescriptorParentalRating(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Parental Rating descriptor failed: %w", err)
							return
						}
					case DescriptorTagPrivateDataIndicator:
						if d.PrivateDataIndicator, err = newDescriptorPrivateDataIndicator(i); err != nil {
							err = fmt.Errorf("astits: parsing Private Data Indicator descriptor failed: %w", err)
							return
						}
					case DescriptorTagPrivateDataSpecifier:
						if d.PrivateDataSpecifier, err = newDescriptorPrivateDataSpecifier(i); err != nil {
							err = fmt.Errorf("astits: parsing Private Data Specifier descriptor failed: %w", err)
							return
						}
					case DescriptorTagRegistration:
						if d.Registration, err = newDescriptorRegistration(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Registration descriptor failed: %w", err)
							return
						}
					case DescriptorTagSatelliteDeliverySystem:
						if d.SatelliteDeliverySystem, err = newDescriptorSatelliteDeliverySystem(i); err != nil {
							err = fmt.Errorf("astits: parsing Satellite Delivery System descriptor failed: %w", err)
							return
						}
					case DescriptorTagTerrestrialDeliverySystem:
						if d.TerrestrialDeliverySystem, err = newDescriptorTerrestrialDeliverySystem(i); err != nil {
							err = fmt.Errorf("astits: parsing Terrestrial Delivery System descriptor failed: %w", err)
							return
						}
					case DescriptorTagCableDeliverySystem:
						if d.CableDeliverySystem, err = newDescriptorCableDeliverySystem(i); err != nil {
							err = fmt.Errorf("astits: parsing Cable Delivery System descriptor failed: %w", err)
							return
						}
					case DescriptorTagService:
						if d.Service, err = newDescriptorService(i); err != nil {
							err = fmt.Errorf("astits: parsing Service descriptor failed: %w", err)
							return
						}
					case DescriptorTagShortEvent:
						if d.ShortEvent, err = newDescriptorShortEvent(i); err != nil {
							err = fmt.Errorf("astits: parsing Short Event descriptor failed: %w", err)
							return
						}
					case DescriptorTagStreamIdentifier:
						if d.StreamIdentifier, err = newDescriptorStreamIdentifier(i); err != nil {
							err = fmt.Errorf("astits: parsing Stream Identifier descriptor failed: %w", err)
							return
						}
					case DescriptorTagSubtitling:
						if d.Subtitling, err = newDescriptorSubtitling(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Subtitling descriptor failed: %w", err)
							return
						}
					case DescriptorTagTeletext:
						if d.Teletext, err = newDescriptorTeletext(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing Teletext descriptor failed: %w", err)
							return
						}
					case DescriptorTagVBIData:
						if d.VBIData, err = newDescriptorVBIData(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing VBI Date descriptor failed: %w", err)
							return
						}
					case DescriptorTagVBITeletext:
						if d.VBITeletext, err = newDescriptorTeletext(i, offsetDescriptorEnd); err != nil {
							err = fmt.Errorf("astits: parsing VBI Teletext descriptor failed: %w", err)
							return
						}
					default:
						if d.Unknown, err = newDescriptorUnknown(i, d.Tag, d.Length); err != nil {
							err = fmt.Errorf("astits: parsing unknown descriptor failed: %w", err)
							return
						}
					}
				}

				// Seek in iterator to make sure we move to the end of the descriptor since its content may be
				// corrupted
				i.Seek(offsetDescriptorEnd)
			}
			o = append(o, d)
		}
	}
	return
}

func calcDescriptorUserDefinedLength(d []byte) uint8 {
	if d == nil {
		return 0
	}
	return uint8(len(d))
}

func writeDescriptorUserDefined(w *astikit.BitsWriter, d []byte) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d)

	return b.Err()
}

func calcDescriptorAC3Length(d *DescriptorAC3) uint8 {
	if d == nil {
		return 0
	}
	ret := 1 // flags

	if d.HasComponentType {
		ret++
	}
	if d.HasBSID {
		ret++
	}
	if d.HasMainID {
		ret++
	}
	if d.HasASVC {
		ret++
	}

	ret += len(d.AdditionalInfo)

	return uint8(ret)
}

func writeDescriptorAC3(w *astikit.BitsWriter, d *DescriptorAC3) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.HasComponentType)
	b.Write(d.HasBSID)
	b.Write(d.HasMainID)
	b.Write(d.HasASVC)
	b.WriteN(uint8(0xff), 4)

	if d.HasComponentType {
		b.Write(d.ComponentType)
	}
	if d.HasBSID {
		b.Write(d.BSID)
	}
	if d.HasMainID {
		b.Write(d.MainID)
	}
	if d.HasASVC {
		b.Write(d.ASVC)
	}
	b.Write(d.AdditionalInfo)

	return b.Err()
}

func calcDescriptorAVCVideoLength(d *DescriptorAVCVideo) uint8 {
	if d == nil {
		return 0
	}
	return 4
}

func writeDescriptorAVCVideo(w *astikit.BitsWriter, d *DescriptorAVCVideo) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.ProfileIDC)

	b.Write(d.ConstraintSet0Flag)
	b.Write(d.ConstraintSet1Flag)
	b.Write(d.ConstraintSet2Flag)
	b.WriteN(d.CompatibleFlags, 5)

	b.Write(d.LevelIDC)

	b.Write(d.AVCStillPresent)
	b.Write(d.AVC24HourPictureFlag)
	b.WriteN(uint8(0xff), 6)

	return b.Err()
}

func calcDescriptorComponentLength(d *DescriptorComponent) uint8 {
	if d == nil {
		return 0
	}
	return uint8(6 + len(d.Text))
}

func writeDescriptorComponent(w *astikit.BitsWriter, d *DescriptorComponent) error {
	b := astikit.NewBitsWriterBatch(w)

	b.WriteN(d.StreamContentExt, 4)
	b.WriteN(d.StreamContent, 4)

	b.Write(d.ComponentType)
	b.Write(d.ComponentTag)

	b.WriteBytesN(d.ISO639LanguageCode, 3, 0)

	b.Write(d.Text)

	return b.Err()
}

func calcDescriptorContentLength(d *DescriptorContent) uint8 {
	if d == nil {
		return 0
	}
	return uint8(2 * len(d.Items))
}

func writeDescriptorContent(w *astikit.BitsWriter, d *DescriptorContent) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Items {
		b.WriteN(item.ContentNibbleLevel1, 4)
		b.WriteN(item.ContentNibbleLevel2, 4)
		b.Write(item.UserByte)
	}

	return b.Err()
}

func calcDescriptorDataStreamAlignmentLength(d *DescriptorDataStreamAlignment) uint8 {
	if d == nil {
		return 0
	}
	return 1
}

func writeDescriptorDataStreamAlignment(w *astikit.BitsWriter, d *DescriptorDataStreamAlignment) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Type)

	return b.Err()
}

func calcDescriptorEnhancedAC3Length(d *DescriptorEnhancedAC3) uint8 {
	if d == nil {
		return 0
	}
	ret := 1 // flags

	if d.HasComponentType {
		ret++
	}
	if d.HasBSID {
		ret++
	}
	if d.HasMainID {
		ret++
	}
	if d.HasASVC {
		ret++
	}
	if d.HasSubStream1 {
		ret++
	}
	if d.HasSubStream2 {
		ret++
	}
	if d.HasSubStream3 {
		ret++
	}

	ret += len(d.AdditionalInfo)

	return uint8(ret)
}

func writeDescriptorEnhancedAC3(w *astikit.BitsWriter, d *DescriptorEnhancedAC3) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.HasComponentType)
	b.Write(d.HasBSID)
	b.Write(d.HasMainID)
	b.Write(d.HasASVC)
	b.Write(d.MixInfoExists)
	b.Write(d.HasSubStream1)
	b.Write(d.HasSubStream2)
	b.Write(d.HasSubStream3)

	if d.HasComponentType {
		b.Write(d.ComponentType)
	}
	if d.HasBSID {
		b.Write(d.BSID)
	}
	if d.HasMainID {
		b.Write(d.MainID)
	}
	if d.HasASVC {
		b.Write(d.ASVC)
	}
	if d.HasSubStream1 {
		b.Write(d.SubStream1)
	}
	if d.HasSubStream2 {
		b.Write(d.SubStream2)
	}
	if d.HasSubStream3 {
		b.Write(d.SubStream3)
	}

	b.Write(d.AdditionalInfo)

	return b.Err()
}

func calcDescriptorExtendedEventLength(d *DescriptorExtendedEvent) (descriptorLength, lengthOfItems uint8) {
	if d == nil {
		return 0, 0
	}
	ret := 1 + 3 + 1 // numbers, language and items length

	itemsRet := 0
	for _, item := range d.Items {
		itemsRet += 1 // description length
		itemsRet += len(item.Description)
		itemsRet += 1 // content length
		itemsRet += len(item.Content)
	}

	ret += itemsRet

	ret += 1 // text length
	ret += len(d.Text)

	return uint8(ret), uint8(itemsRet)
}

func writeDescriptorExtendedEvent(w *astikit.BitsWriter, d *DescriptorExtendedEvent) error {
	b := astikit.NewBitsWriterBatch(w)

	var lengthOfItems uint8

	_, lengthOfItems = calcDescriptorExtendedEventLength(d)

	b.WriteN(d.Number, 4)
	b.WriteN(d.LastDescriptorNumber, 4)

	b.WriteBytesN(d.ISO639LanguageCode, 3, 0)

	b.Write(lengthOfItems)
	for _, item := range d.Items {
		b.Write(uint8(len(item.Description)))
		b.Write(item.Description)
		b.Write(uint8(len(item.Content)))
		b.Write(item.Content)
	}

	b.Write(uint8(len(d.Text)))
	b.Write(d.Text)

	return b.Err()
}

func calcDescriptorExtensionSupplementaryAudioLength(d *DescriptorExtensionSupplementaryAudio) int {
	if d == nil {
		return 0
	}
	ret := 1
	if d.HasLanguageCode {
		ret += 3
	}
	ret += len(d.PrivateData)
	return ret
}

func calcDescriptorExtensionLength(d *DescriptorExtension) uint8 {
	if d == nil {
		return 0
	}
	ret := 1 // tag

	switch d.Tag {
	case DescriptorTagExtensionT2DeliverySystem:
		ret += calcDescriptorExtensionT2DeliverySystemLength(d.T2DeliverySystem)
	case DescriptorTagExtensionSupplementaryAudio:
		ret += calcDescriptorExtensionSupplementaryAudioLength(d.SupplementaryAudio)
	default:
		if d.Unknown != nil {
			ret += len(*d.Unknown)
		}
	}

	return uint8(ret)
}

func writeDescriptorExtensionSupplementaryAudio(w *astikit.BitsWriter, d *DescriptorExtensionSupplementaryAudio) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.MixType)
	b.WriteN(d.EditorialClassification, 5)
	b.Write(true) // reserved
	b.Write(d.HasLanguageCode)

	if d.HasLanguageCode {
		b.WriteBytesN(d.LanguageCode, 3, 0)
	}

	b.Write(d.PrivateData)

	return b.Err()
}

func writeDescriptorExtensionT2DeliverySystem(w *astikit.BitsWriter, d *DescriptorExtensionT2DeliverySystem) error {
	b := astikit.NewBitsWriterBatch(w)

	// PLP ID
	b.Write(d.PLPID)

	// T2 system ID
	b.Write(uint8(d.T2SystemID >> 8))
	b.Write(uint8(d.T2SystemID & 0xFF))

	if d.HasExtendedInfo {
		// SISO/MISO (2 bits) + bandwidth (4 bits) + reserved (2 bits)
		var byte3 uint8
		byte3 |= (d.SISOorMISO & 0x3) << 6
		byte3 |= (d.Bandwidth & 0x0F) << 2
		// reserved bits 1-0 set to 0
		b.Write(byte3)

		// Guard interval (3 bits) + transmission mode (3 bits) + other freq flag (1) + TFS flag (1)
		var byte4 uint8
		byte4 |= (d.GuardInterval & 0x7) << 5
		byte4 |= (d.TransmissionMode & 0x7) << 2
		if d.OtherFrequencyFlag {
			byte4 |= 0x02
		}
		if d.TFSFlag {
			byte4 |= 0x01
		}
		b.Write(byte4)

		// Cells
		for _, cell := range d.Cells {
			b.Write(uint8(cell.CellID >> 8))
			b.Write(uint8(cell.CellID & 0xFF))
			b.Write(uint8(len(cell.CentreFrequencies) * 4))
			for _, freq := range cell.CentreFrequencies {
				b.Write(freq)
			}
		}
	}

	return b.Err()
}

func writeDescriptorExtension(w *astikit.BitsWriter, d *DescriptorExtension) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Tag)

	switch d.Tag {
	case DescriptorTagExtensionT2DeliverySystem:
		err := writeDescriptorExtensionT2DeliverySystem(w, d.T2DeliverySystem)
		if err != nil {
			return err
		}
	case DescriptorTagExtensionSupplementaryAudio:
		err := writeDescriptorExtensionSupplementaryAudio(w, d.SupplementaryAudio)
		if err != nil {
			return err
		}
	default:
		if d.Unknown != nil {
			b.Write(*d.Unknown)
		}
	}

	return b.Err()
}

func calcDescriptorISO639LanguageAndAudioTypeLength(d *DescriptorISO639LanguageAndAudioType) uint8 {
	if d == nil {
		return 0
	}
	return 3 + 1 // language code + type
}

func writeDescriptorISO639LanguageAndAudioType(w *astikit.BitsWriter, d *DescriptorISO639LanguageAndAudioType) error {
	b := astikit.NewBitsWriterBatch(w)

	b.WriteBytesN(d.Language, 3, 0)
	b.Write(d.Type)

	return b.Err()
}

func calcDescriptorLocalTimeOffsetLength(d *DescriptorLocalTimeOffset) uint8 {
	if d == nil {
		return 0
	}
	return uint8(13 * len(d.Items))
}

func writeDescriptorLocalTimeOffset(w *astikit.BitsWriter, d *DescriptorLocalTimeOffset) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Items {
		b.WriteBytesN(item.CountryCode, 3, 0)

		b.WriteN(item.CountryRegionID, 6)
		b.WriteN(uint8(0xff), 1)
		b.Write(item.LocalTimeOffsetPolarity)

		if _, err := writeDVBDurationMinutes(w, item.LocalTimeOffset); err != nil {
			return err
		}
		if _, err := writeDVBTime(w, item.TimeOfChange); err != nil {
			return err
		}
		if _, err := writeDVBDurationMinutes(w, item.NextTimeOffset); err != nil {
			return err
		}
	}

	return b.Err()
}

func calcDescriptorMaximumBitrateLength(d *DescriptorMaximumBitrate) uint8 {
	if d == nil {
		return 0
	}
	return 3
}

func writeDescriptorMaximumBitrate(w *astikit.BitsWriter, d *DescriptorMaximumBitrate) error {
	b := astikit.NewBitsWriterBatch(w)

	b.WriteN(uint8(0xff), 2)
	b.WriteN(uint32(d.Bitrate/50), 22)

	return b.Err()
}

func calcDescriptorNetworkNameLength(d *DescriptorNetworkName) uint8 {
	if d == nil {
		return 0
	}
	return uint8(len(d.Name))
}

func writeDescriptorNetworkName(w *astikit.BitsWriter, d *DescriptorNetworkName) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Name)

	return b.Err()
}

func calcDescriptorParentalRatingLength(d *DescriptorParentalRating) uint8 {
	if d == nil {
		return 0
	}
	return uint8(4 * len(d.Items))
}

func writeDescriptorParentalRating(w *astikit.BitsWriter, d *DescriptorParentalRating) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Items {
		b.WriteBytesN(item.CountryCode, 3, 0)
		b.Write(item.Rating)
	}

	return b.Err()
}

func calcDescriptorPrivateDataIndicatorLength(d *DescriptorPrivateDataIndicator) uint8 {
	if d == nil {
		return 0
	}
	return 4
}

func writeDescriptorPrivateDataIndicator(w *astikit.BitsWriter, d *DescriptorPrivateDataIndicator) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Indicator)

	return b.Err()
}

func calcDescriptorPrivateDataSpecifierLength(d *DescriptorPrivateDataSpecifier) uint8 {
	if d == nil {
		return 0
	}
	return 4
}

func writeDescriptorPrivateDataSpecifier(w *astikit.BitsWriter, d *DescriptorPrivateDataSpecifier) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Specifier)

	return b.Err()
}

func calcDescriptorRegistrationLength(d *DescriptorRegistration) uint8 {
	if d == nil {
		return 0
	}
	return uint8(4 + len(d.AdditionalIdentificationInfo))
}

func writeDescriptorRegistration(w *astikit.BitsWriter, d *DescriptorRegistration) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.FormatIdentifier)
	b.Write(d.AdditionalIdentificationInfo)

	return b.Err()
}

func calcDescriptorServiceLength(d *DescriptorService) uint8 {
	if d == nil {
		return 0
	}
	ret := 3 // type and lengths
	ret += len(d.Name)
	ret += len(d.Provider)
	return uint8(ret)
}

func writeDescriptorService(w *astikit.BitsWriter, d *DescriptorService) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Type)
	b.Write(uint8(len(d.Provider)))
	b.Write(d.Provider)
	b.Write(uint8(len(d.Name)))
	b.Write(d.Name)

	return b.Err()
}

func calcDescriptorShortEventLength(d *DescriptorShortEvent) uint8 {
	if d == nil {
		return 0
	}
	ret := 3 + 1 + 1 // language code and lengths
	ret += len(d.EventName)
	ret += len(d.Text)
	return uint8(ret)
}

func writeDescriptorShortEvent(w *astikit.BitsWriter, d *DescriptorShortEvent) error {
	b := astikit.NewBitsWriterBatch(w)

	b.WriteBytesN(d.Language, 3, 0)

	b.Write(uint8(len(d.EventName)))
	b.Write(d.EventName)

	b.Write(uint8(len(d.Text)))
	b.Write(d.Text)

	return b.Err()
}

func calcDescriptorStreamIdentifierLength(d *DescriptorStreamIdentifier) uint8 {
	if d == nil {
		return 0
	}
	return 1
}

func writeDescriptorStreamIdentifier(w *astikit.BitsWriter, d *DescriptorStreamIdentifier) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.ComponentTag)

	return b.Err()
}

func calcDescriptorSubtitlingLength(d *DescriptorSubtitling) uint8 {
	if d == nil {
		return 0
	}
	return uint8(8 * len(d.Items))
}

func writeDescriptorSubtitling(w *astikit.BitsWriter, d *DescriptorSubtitling) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Items {
		b.WriteBytesN(item.Language, 3, 0)
		b.Write(item.Type)
		b.Write(item.CompositionPageID)
		b.Write(item.AncillaryPageID)
	}

	return b.Err()
}

func calcDescriptorTeletextLength(d *DescriptorTeletext) uint8 {
	if d == nil {
		return 0
	}
	return uint8(5 * len(d.Items))
}

func writeDescriptorTeletext(w *astikit.BitsWriter, d *DescriptorTeletext) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Items {
		b.WriteBytesN(item.Language, 3, 0)
		b.WriteN(item.Type, 5)
		b.WriteN(item.Magazine, 3)
		b.WriteN(item.Page/10, 4)
		b.WriteN(item.Page%10, 4)
	}

	return b.Err()
}

func calcDescriptorVBIDataLength(d *DescriptorVBIData) uint8 {
	if d == nil {
		return 0
	}
	return uint8(3 * len(d.Services))
}

func writeDescriptorVBIData(w *astikit.BitsWriter, d *DescriptorVBIData) error {
	b := astikit.NewBitsWriterBatch(w)

	for _, item := range d.Services {
		b.Write(item.DataServiceID)

		if item.DataServiceID == VBIDataServiceIDClosedCaptioning ||
			item.DataServiceID == VBIDataServiceIDEBUTeletext ||
			item.DataServiceID == VBIDataServiceIDInvertedTeletext ||
			item.DataServiceID == VBIDataServiceIDMonochrome442Samples ||
			item.DataServiceID == VBIDataServiceIDVPS ||
			item.DataServiceID == VBIDataServiceIDWSS {

			b.Write(uint8(len(item.Descriptors))) // each descriptor is 1 byte
			for _, desc := range item.Descriptors {
				b.WriteN(uint8(0xff), 2)
				b.Write(desc.FieldParity)
				b.WriteN(desc.LineOffset, 5)
			}
		} else {
			// let's put one reserved byte
			b.Write(uint8(1))
			b.Write(uint8(0xff))
		}
	}

	return b.Err()
}

func calcDescriptorUnknownLength(d *DescriptorUnknown) uint8 {
	if d == nil {
		return 0
	}
	return uint8(len(d.Content))
}

func writeDescriptorUnknown(w *astikit.BitsWriter, d *DescriptorUnknown) error {
	b := astikit.NewBitsWriterBatch(w)

	b.Write(d.Content)

	return b.Err()
}

// encodeBCD encodes an integer into BCD bytes
func encodeBCD(val uint64, numBytes int) []byte {
	result := make([]byte, numBytes)
	for i := numBytes - 1; i >= 0; i-- {
		lo := val % 10
		val /= 10
		hi := val % 10
		val /= 10
		result[i] = byte(hi<<4 | lo)
	}
	return result
}

func calcDescriptorSatelliteDeliverySystemLength(d *DescriptorSatelliteDeliverySystem) uint8 {
	if d == nil {
		return 0
	}
	return 11
}

func writeDescriptorSatelliteDeliverySystem(w *astikit.BitsWriter, d *DescriptorSatelliteDeliverySystem) error {
	b := astikit.NewBitsWriterBatch(w)

	// Frequency (4 BCD bytes)
	b.Write(encodeBCD(uint64(d.Frequency), 4))

	// Orbital position (2 bytes, raw)
	b.Write(uint8(d.OrbitalPosition >> 8))
	b.Write(uint8(d.OrbitalPosition & 0xFF))

	// West/east, polarization, roll-off, modulation system, modulation type
	var flags uint8
	if d.WestEastFlag {
		flags |= 0x80
	}
	flags |= (d.Polarization & 0x3) << 5
	flags |= (d.RollOff & 0x3) << 3
	if d.ModulationSystem {
		flags |= 0x04
	}
	flags |= d.ModulationType & 0x3
	b.Write(flags)

	// Symbol rate (top 28 bits BCD = 7 digits) + FEC inner (bottom 4 bits)
	sr := uint64(d.SymbolRate)
	srBytes := encodeBCD(sr/10, 3)
	lastNibble := uint8(sr % 10)
	b.Write(srBytes)
	b.Write(lastNibble<<4 | (d.FECInner & 0x0F))

	return b.Err()
}

func calcDescriptorTerrestrialDeliverySystemLength(d *DescriptorTerrestrialDeliverySystem) uint8 {
	if d == nil {
		return 0
	}
	return 11
}

func writeDescriptorTerrestrialDeliverySystem(w *astikit.BitsWriter, d *DescriptorTerrestrialDeliverySystem) error {
	b := astikit.NewBitsWriterBatch(w)

	// Centre frequency (uint32 big-endian)
	b.Write(d.Frequency)

	// Bandwidth (3 bits), priority (1), time slicing (1), MPE-FEC (1), reserved (2)
	var byte4 uint8
	byte4 |= (d.Bandwidth & 0x7) << 5
	if d.Priority {
		byte4 |= 0x10
	}
	if d.TimeSlicingIndicator {
		byte4 |= 0x08
	}
	if d.MPEFECIndicator {
		byte4 |= 0x04
	}
	// reserved bits 1-0 set to 0
	b.Write(byte4)

	// Constellation (2 bits), hierarchy (3), code rate HP (3)
	var byte5 uint8
	byte5 |= (d.Constellation & 0x3) << 6
	byte5 |= (d.HierarchyInformation & 0x7) << 3
	byte5 |= d.CodeRateHPStream & 0x7
	b.Write(byte5)

	// Code rate LP (3 bits), guard interval (2), transmission mode (2), other freq flag (1)
	var byte6 uint8
	byte6 |= (d.CodeRateLPStream & 0x7) << 5
	byte6 |= (d.GuardInterval & 0x3) << 3
	byte6 |= (d.TransmissionMode & 0x3) << 1
	if d.OtherFrequencyFlag {
		byte6 |= 0x01
	}
	b.Write(byte6)

	// Reserved (4 bytes)
	b.Write(uint32(0xFFFFFFFF))

	return b.Err()
}

func calcDescriptorCableDeliverySystemLength(d *DescriptorCableDeliverySystem) uint8 {
	if d == nil {
		return 0
	}
	return 11
}

func writeDescriptorCableDeliverySystem(w *astikit.BitsWriter, d *DescriptorCableDeliverySystem) error {
	b := astikit.NewBitsWriterBatch(w)

	// Frequency (4 BCD bytes)
	b.Write(encodeBCD(uint64(d.Frequency), 4))

	// Reserved (12 bits) + FEC outer (4 bits) = 2 bytes
	b.Write(uint8(0xFF))
	b.Write(uint8(0xF0 | (d.FECOuter & 0x0F)))

	// Modulation
	b.Write(d.Modulation)

	// Symbol rate (top 28 bits BCD) + FEC inner (bottom 4 bits)
	sr := uint64(d.SymbolRate)
	srBytes := encodeBCD(sr/10, 3)
	lastNibble := uint8(sr % 10)
	b.Write(srBytes)
	b.Write(lastNibble<<4 | (d.FECInner & 0x0F))

	return b.Err()
}

func calcDescriptorExtensionT2DeliverySystemLength(d *DescriptorExtensionT2DeliverySystem) int {
	if d == nil {
		return 0
	}
	ret := 3 // PLP ID (1) + T2 system ID (2)
	if d.HasExtendedInfo {
		ret += 2 // SISO/MISO + bandwidth byte (1) + guard/transmission byte (1)
		for _, cell := range d.Cells {
			ret += 3 // cell ID (2) + freq loop length (1)
			ret += 4 * len(cell.CentreFrequencies)
		}
	}
	return ret
}

func calcDescriptorLength(d *Descriptor) uint8 {
	if d.Tag >= 0x80 && d.Tag <= 0xfe {
		return calcDescriptorUserDefinedLength(d.UserDefined)
	}

	switch d.Tag {
	case DescriptorTagAC3:
		return calcDescriptorAC3Length(d.AC3)
	case DescriptorTagAVCVideo:
		return calcDescriptorAVCVideoLength(d.AVCVideo)
	case DescriptorTagSatelliteDeliverySystem:
		return calcDescriptorSatelliteDeliverySystemLength(d.SatelliteDeliverySystem)
	case DescriptorTagCableDeliverySystem:
		return calcDescriptorCableDeliverySystemLength(d.CableDeliverySystem)
	case DescriptorTagTerrestrialDeliverySystem:
		return calcDescriptorTerrestrialDeliverySystemLength(d.TerrestrialDeliverySystem)
	case DescriptorTagComponent:
		return calcDescriptorComponentLength(d.Component)
	case DescriptorTagContent:
		return calcDescriptorContentLength(d.Content)
	case DescriptorTagDataStreamAlignment:
		return calcDescriptorDataStreamAlignmentLength(d.DataStreamAlignment)
	case DescriptorTagEnhancedAC3:
		return calcDescriptorEnhancedAC3Length(d.EnhancedAC3)
	case DescriptorTagExtendedEvent:
		ret, _ := calcDescriptorExtendedEventLength(d.ExtendedEvent)
		return ret
	case DescriptorTagExtension:
		return calcDescriptorExtensionLength(d.Extension)
	case DescriptorTagISO639LanguageAndAudioType:
		return calcDescriptorISO639LanguageAndAudioTypeLength(d.ISO639LanguageAndAudioType)
	case DescriptorTagLocalTimeOffset:
		return calcDescriptorLocalTimeOffsetLength(d.LocalTimeOffset)
	case DescriptorTagMaximumBitrate:
		return calcDescriptorMaximumBitrateLength(d.MaximumBitrate)
	case DescriptorTagNetworkName:
		return calcDescriptorNetworkNameLength(d.NetworkName)
	case DescriptorTagParentalRating:
		return calcDescriptorParentalRatingLength(d.ParentalRating)
	case DescriptorTagPrivateDataIndicator:
		return calcDescriptorPrivateDataIndicatorLength(d.PrivateDataIndicator)
	case DescriptorTagPrivateDataSpecifier:
		return calcDescriptorPrivateDataSpecifierLength(d.PrivateDataSpecifier)
	case DescriptorTagRegistration:
		return calcDescriptorRegistrationLength(d.Registration)
	case DescriptorTagService:
		return calcDescriptorServiceLength(d.Service)
	case DescriptorTagShortEvent:
		return calcDescriptorShortEventLength(d.ShortEvent)
	case DescriptorTagStreamIdentifier:
		return calcDescriptorStreamIdentifierLength(d.StreamIdentifier)
	case DescriptorTagSubtitling:
		return calcDescriptorSubtitlingLength(d.Subtitling)
	case DescriptorTagTeletext:
		return calcDescriptorTeletextLength(d.Teletext)
	case DescriptorTagVBIData:
		return calcDescriptorVBIDataLength(d.VBIData)
	case DescriptorTagVBITeletext:
		return calcDescriptorTeletextLength(d.VBITeletext)
	}

	return calcDescriptorUnknownLength(d.Unknown)
}

func writeDescriptor(w *astikit.BitsWriter, d *Descriptor) (int, error) {
	b := astikit.NewBitsWriterBatch(w)
	length := calcDescriptorLength(d)

	b.Write(d.Tag)
	b.Write(length)

	if err := b.Err(); err != nil {
		return 0, err
	}

	written := int(length) + 2

	if d.Length == 0 {
		return written, nil
	}

	if d.Tag >= 0x80 && d.Tag <= 0xfe {
		return written, writeDescriptorUserDefined(w, d.UserDefined)
	}

	switch d.Tag {
	case DescriptorTagAC3:
		return written, writeDescriptorAC3(w, d.AC3)
	case DescriptorTagAVCVideo:
		return written, writeDescriptorAVCVideo(w, d.AVCVideo)
	case DescriptorTagSatelliteDeliverySystem:
		return written, writeDescriptorSatelliteDeliverySystem(w, d.SatelliteDeliverySystem)
	case DescriptorTagCableDeliverySystem:
		return written, writeDescriptorCableDeliverySystem(w, d.CableDeliverySystem)
	case DescriptorTagTerrestrialDeliverySystem:
		return written, writeDescriptorTerrestrialDeliverySystem(w, d.TerrestrialDeliverySystem)
	case DescriptorTagComponent:
		return written, writeDescriptorComponent(w, d.Component)
	case DescriptorTagContent:
		return written, writeDescriptorContent(w, d.Content)
	case DescriptorTagDataStreamAlignment:
		return written, writeDescriptorDataStreamAlignment(w, d.DataStreamAlignment)
	case DescriptorTagEnhancedAC3:
		return written, writeDescriptorEnhancedAC3(w, d.EnhancedAC3)
	case DescriptorTagExtendedEvent:
		return written, writeDescriptorExtendedEvent(w, d.ExtendedEvent)
	case DescriptorTagExtension:
		return written, writeDescriptorExtension(w, d.Extension)
	case DescriptorTagISO639LanguageAndAudioType:
		return written, writeDescriptorISO639LanguageAndAudioType(w, d.ISO639LanguageAndAudioType)
	case DescriptorTagLocalTimeOffset:
		return written, writeDescriptorLocalTimeOffset(w, d.LocalTimeOffset)
	case DescriptorTagMaximumBitrate:
		return written, writeDescriptorMaximumBitrate(w, d.MaximumBitrate)
	case DescriptorTagNetworkName:
		return written, writeDescriptorNetworkName(w, d.NetworkName)
	case DescriptorTagParentalRating:
		return written, writeDescriptorParentalRating(w, d.ParentalRating)
	case DescriptorTagPrivateDataIndicator:
		return written, writeDescriptorPrivateDataIndicator(w, d.PrivateDataIndicator)
	case DescriptorTagPrivateDataSpecifier:
		return written, writeDescriptorPrivateDataSpecifier(w, d.PrivateDataSpecifier)
	case DescriptorTagRegistration:
		return written, writeDescriptorRegistration(w, d.Registration)
	case DescriptorTagService:
		return written, writeDescriptorService(w, d.Service)
	case DescriptorTagShortEvent:
		return written, writeDescriptorShortEvent(w, d.ShortEvent)
	case DescriptorTagStreamIdentifier:
		return written, writeDescriptorStreamIdentifier(w, d.StreamIdentifier)
	case DescriptorTagSubtitling:
		return written, writeDescriptorSubtitling(w, d.Subtitling)
	case DescriptorTagTeletext:
		return written, writeDescriptorTeletext(w, d.Teletext)
	case DescriptorTagVBIData:
		return written, writeDescriptorVBIData(w, d.VBIData)
	case DescriptorTagVBITeletext:
		return written, writeDescriptorTeletext(w, d.VBITeletext)
	}

	return written, writeDescriptorUnknown(w, d.Unknown)
}

func calcDescriptorsLength(ds []*Descriptor) uint16 {
	length := uint16(0)
	for _, d := range ds {
		length += 2 // tag and length
		length += uint16(calcDescriptorLength(d))
	}
	return length
}

func writeDescriptors(w *astikit.BitsWriter, ds []*Descriptor) (int, error) {
	written := 0

	for _, d := range ds {
		n, err := writeDescriptor(w, d)
		if err != nil {
			return 0, err
		}
		written += n
	}

	return written, nil
}

func writeDescriptorsWithLength(w *astikit.BitsWriter, ds []*Descriptor) (int, error) {
	length := calcDescriptorsLength(ds)
	b := astikit.NewBitsWriterBatch(w)

	b.WriteN(uint8(0xff), 4) // reserved
	b.WriteN(length, 12)     // program_info_length

	if err := b.Err(); err != nil {
		return 0, err
	}

	written, err := writeDescriptors(w, ds)
	return written + 2, err // 2 for length
}
