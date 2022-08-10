package astits

import (
	"fmt"
	"time"

	"github.com/icza/bitio"
)

// Audio types. Page: 683 | Link:
// https://books.google.fr/books?id=6dgWB3-rChYC&printsec=frontcover&hl=fr
const (
	AudioTypeCleanEffects             = 0x1
	AudioTypeHearingImpaired          = 0x2
	AudioTypeVisualImpairedCommentary = 0x3
)

// Data stream alignments. Page: 85 | Chapter:2.6.11 | Link:
// http://ecee.colorado.edu/~ecen5653/ecen5653/papers/iso13818-1.pdf
const (
	DataStreamAligmentAudioSyncWord          = 0x1
	DataStreamAligmentVideoSliceOrAccessUnit = 0x1
	DataStreamAligmentVideoAccessUnit        = 0x2
	DataStreamAligmentVideoGOPOrSEQ          = 0x3
	DataStreamAligmentVideoSEQ               = 0x4
)

// Descriptor tags. Chapter: 6.1 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
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
	DescriptorTagVBIData                    = 0x45
	DescriptorTagVBITeletext                = 0x46
)

// Descriptor extension tags. Chapter: 6.3 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	DescriptorTagExtensionSupplementaryAudio = 0x6
)

// Service types. Chapter: 6.2.33 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	ServiceTypeDigitalTelevisionService = 0x1
)

// Teletext types. Chapter: 6.2.43 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	TeletextTypeAdditionalInformationPage                    = 0x3
	TeletextTypeInitialTeletextPage                          = 0x1
	TeletextTypeProgramSchedulePage                          = 0x4
	TeletextTypeTeletextSubtitlePage                         = 0x2
	TeletextTypeTeletextSubtitlePageForHearingImpairedPeople = 0x5
)

// VBI data service id Chapter: 6.2.47 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
const (
	VBIDataServiceIDClosedCaptioning     = 0x6
	VBIDataServiceIDEBUTeletext          = 0x1
	VBIDataServiceIDInvertedTeletext     = 0x2
	VBIDataServiceIDMonochrome442Samples = 0x7
	VBIDataServiceIDVPS                  = 0x4
	VBIDataServiceIDWSS                  = 0x5
)

// Descriptor represents a descriptor
// TODO Handle UTF8.
type Descriptor struct {
	AC3                        *DescriptorAC3
	AVCVideo                   *DescriptorAVCVideo
	Component                  *DescriptorComponent
	Content                    DescriptorContent
	DataStreamAlignment        DescriptorDataStreamAlignment
	EnhancedAC3                *DescriptorEnhancedAC3
	ExtendedEvent              *DescriptorExtendedEvent
	Extension                  *DescriptorExtension
	ISO639LanguageAndAudioType *DescriptorISO639LanguageAndAudioType
	Length                     uint8
	LocalTimeOffset            DescriptorLocalTimeOffset
	MaximumBitrate             DescriptorMaximumBitrate
	NetworkName                DescriptorNetworkName
	ParentalRating             DescriptorParentalRating
	PrivateDataIndicator       DescriptorPrivateDataIndicator
	PrivateDataSpecifier       DescriptorPrivateDataSpecifier
	Registration               *DescriptorRegistration
	Service                    *DescriptorService
	ShortEvent                 *DescriptorShortEvent
	StreamIdentifier           DescriptorStreamIdentifier
	Subtitling                 DescriptorSubtitling

	// the tag defines the structure of the contained
	// data following the descriptor length.
	Tag         uint8
	Teletext    DescriptorTeletext
	Unknown     *DescriptorUnknown
	UserDefined []byte
	VBIData     DescriptorVBIData
	VBITeletext DescriptorTeletext
}

// DescriptorAC3 represents an AC3 descriptor Chapter: Annex D | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorAC3 struct {
	AdditionalInfo   []byte
	ASVC             uint8
	BSID             uint8
	ComponentType    uint8 // 4 Bits.
	HasASVC          bool
	HasBSID          bool
	HasComponentType bool
	HasMainID        bool
	MainID           uint8
}

func newDescriptorAC3(r *bitio.CountReader, offsetEnd int64) (*DescriptorAC3, error) {
	d := &DescriptorAC3{
		HasASVC:          r.TryReadBool(),
		HasBSID:          r.TryReadBool(),
		HasComponentType: r.TryReadBool(),
		HasMainID:        r.TryReadBool(),
	}
	_ = r.TryReadBits(4) // Reserved.

	if d.HasComponentType {
		d.ComponentType = r.TryReadByte()
	}

	if d.HasBSID {
		d.BSID = r.TryReadByte()
	}

	if d.HasMainID {
		d.MainID = r.TryReadByte()
	}

	if d.HasASVC {
		d.ASVC = r.TryReadByte()
	}

	if r.BitsCount/8 < offsetEnd {
		d.AdditionalInfo = make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, d.AdditionalInfo)
	}

	return d, r.TryError
}

// DescriptorAVCVideo represents an AVC video descriptor.
// No doc found unfortunately, basing the implementation on
// https://github.com/gfto/bitstream/blob/master/mpeg/psi/desc_28.h
type DescriptorAVCVideo struct {
	AVC24HourPictureFlag bool
	AVCStillPresent      bool
	CompatibleFlags      uint8 // 5 bits.
	ConstraintSet0Flag   bool
	ConstraintSet1Flag   bool
	ConstraintSet2Flag   bool
	LevelIDC             uint8
	ProfileIDC           uint8
}

func newDescriptorAVCVideo(r *bitio.CountReader) (*DescriptorAVCVideo, error) {
	d := &DescriptorAVCVideo{}

	d.ProfileIDC = r.TryReadByte()

	d.ConstraintSet0Flag = r.TryReadBool()
	d.ConstraintSet1Flag = r.TryReadBool()
	d.ConstraintSet2Flag = r.TryReadBool()
	d.CompatibleFlags = uint8(r.TryReadBits(5))

	d.LevelIDC = r.TryReadByte()

	d.AVCStillPresent = r.TryReadBool()
	d.AVC24HourPictureFlag = r.TryReadBool()
	// Reserved.
	_ = r.TryReadBits(6)

	return d, r.TryError
}

// DescriptorComponent represents a component descriptor Chapter: 6.2.8 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorComponent struct {
	StreamContentExt   uint8 // 4 bits.
	StreamContent      uint8 // 4 bits.
	ComponentType      uint8
	ComponentTag       uint8
	ISO639LanguageCode []byte // 3 bytes.
	Text               []byte
}

func newDescriptorComponent(r *bitio.CountReader, offsetEnd int64) (*DescriptorComponent, error) {
	d := &DescriptorComponent{}

	d.StreamContentExt = uint8(r.TryReadBits(4))
	d.StreamContent = uint8(r.TryReadBits(4))

	d.ComponentType = r.TryReadByte()
	d.ComponentTag = r.TryReadByte()

	d.ISO639LanguageCode = make([]byte, 3)
	TryReadFull(r, d.ISO639LanguageCode)

	if r.BitsCount/8 < offsetEnd {
		d.Text = make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, d.Text)
	}

	return d, r.TryError
}

// DescriptorContent represents a content descriptor. Chapter: 6.2.9 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorContent struct {
	Items []*DescriptorContentItem
}

// DescriptorContentItem represents a content item descriptor. Chapter: 6.2.9 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorContentItem struct {
	ContentNibbleLevel1 uint8 // 4 bits.
	ContentNibbleLevel2 uint8 // 4 bits.
	UserByte            uint8
}

func newDescriptorContent(r *bitio.CountReader, offsetEnd int64) (DescriptorContent, error) {
	items := []*DescriptorContentItem{}

	for r.BitsCount/8 < offsetEnd {
		items = append(items, &DescriptorContentItem{
			ContentNibbleLevel1: uint8(r.TryReadBits(4)),
			ContentNibbleLevel2: uint8(r.TryReadBits(4)),
			UserByte:            r.TryReadByte(),
		})
	}

	return DescriptorContent{Items: items}, r.TryError
}

// DescriptorDataStreamAlignment represents a data stream alignment descriptor.
type DescriptorDataStreamAlignment uint8

func newDescriptorDataStreamAlignment(r *bitio.CountReader) (DescriptorDataStreamAlignment, error) {
	typ, err := r.ReadByte()
	return DescriptorDataStreamAlignment(typ), err
}

// DescriptorEnhancedAC3 represents an enhanced AC3 descriptor. Chapter: Annex D | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorEnhancedAC3 struct {
	HasASVC          bool
	HasBSID          bool
	HasComponentType bool
	HasMainID        bool
	HasSubStream1    bool
	HasSubStream2    bool
	HasSubStream3    bool
	MixInfoExists    bool
	ComponentType    uint8
	BSID             uint8
	MainID           uint8
	ASVC             uint8
	SubStream1       uint8
	SubStream2       uint8
	SubStream3       uint8
	AdditionalInfo   []byte
}

func newDescriptorEnhancedAC3(r *bitio.CountReader, offsetEnd int64) (*DescriptorEnhancedAC3, error) {
	d := &DescriptorEnhancedAC3{
		HasASVC:          r.TryReadBool(),
		HasBSID:          r.TryReadBool(),
		HasComponentType: r.TryReadBool(),
		HasMainID:        r.TryReadBool(),
		HasSubStream1:    r.TryReadBool(),
		HasSubStream2:    r.TryReadBool(),
		HasSubStream3:    r.TryReadBool(),
		MixInfoExists:    r.TryReadBool(),
	}

	if d.HasComponentType {
		d.ComponentType = r.TryReadByte()
	}
	if d.HasBSID {
		d.BSID = r.TryReadByte()
	}
	if d.HasMainID {
		d.MainID = r.TryReadByte()
	}
	if d.HasASVC {
		d.ASVC = r.TryReadByte()
	}
	if d.HasSubStream1 {
		d.SubStream1 = r.TryReadByte()
	}
	if d.HasSubStream2 {
		d.SubStream2 = r.TryReadByte()
	}
	if d.HasSubStream3 {
		d.SubStream3 = r.TryReadByte()
	}

	if r.BitsCount/8 < offsetEnd {
		d.AdditionalInfo = make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, d.AdditionalInfo)
	}

	return d, r.TryError
}

// DescriptorExtendedEvent represents an extended event descriptor. Chapter: 6.2.15 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtendedEvent struct {
	Number               uint8  // 4 bits.
	LastDescriptorNumber uint8  // 4 bits.
	ISO639LanguageCode   []byte // 3 bytes.
	Items                []*DescriptorExtendedEventItem
	Text                 []byte
}

// DescriptorExtendedEventItem represents an extended event item descriptor.
// Chapter: 6.2.15 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtendedEventItem struct {
	Content     []byte
	Description []byte
}

func newDescriptorExtendedEvent(r *bitio.CountReader) (*DescriptorExtendedEvent, error) {
	d := &DescriptorExtendedEvent{}

	d.Number = uint8(r.TryReadBits(4))

	d.LastDescriptorNumber = uint8(r.TryReadBits(4))

	d.ISO639LanguageCode = make([]byte, 3)
	TryReadFull(r, d.ISO639LanguageCode)

	itemsLength := r.TryReadByte()
	offsetEnd := r.BitsCount/8 + int64(itemsLength)

	for r.BitsCount/8 < offsetEnd {
		item, err := newDescriptorExtendedEventItem(r)
		if err != nil {
			return nil, fmt.Errorf("creating extended event item failed: %w", err)
		}

		d.Items = append(d.Items, item)
	}

	textLength := r.TryReadByte()
	d.Text = make([]byte, textLength)
	TryReadFull(r, d.Text)

	return d, r.TryError
}

func newDescriptorExtendedEventItem(r *bitio.CountReader) (*DescriptorExtendedEventItem, error) {
	d := &DescriptorExtendedEventItem{}

	descriptionLength := r.TryReadByte()
	d.Description = make([]byte, descriptionLength)
	TryReadFull(r, d.Description)

	contentLength := r.TryReadByte()
	d.Content = make([]byte, contentLength)
	TryReadFull(r, d.Content)

	return d, r.TryError
}

// DescriptorExtension represents an extension descriptor.
// Chapter: 6.2.16 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtension struct {
	SupplementaryAudio *DescriptorExtensionSupplementaryAudio
	Tag                uint8
	Unknown            *[]byte
}

func newDescriptorExtension(r *bitio.CountReader, offsetEnd int64) (*DescriptorExtension, error) {
	d := &DescriptorExtension{}
	d.Tag = r.TryReadByte()

	var err error
	switch d.Tag {
	case DescriptorTagExtensionSupplementaryAudio:
		if d.SupplementaryAudio, err = newDescriptorExtensionSupplementaryAudio(r, offsetEnd); err != nil {
			return nil, err
		}
	default:
		unknown := make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, unknown)
		d.Unknown = &unknown
	}
	return d, r.TryError
}

// DescriptorExtensionSupplementaryAudio represents
// a supplementary audio extension descriptor.
// Chapter: 6.4.10 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorExtensionSupplementaryAudio struct {
	EditorialClassification uint8 // 5 bits.
	HasLanguageCode         bool
	MixType                 bool
	LanguageCode            []byte // 3 bytes.
	PrivateData             []byte
}

func newDescriptorExtensionSupplementaryAudio(
	r *bitio.CountReader, offsetEnd int64,
) (*DescriptorExtensionSupplementaryAudio, error) {
	d := &DescriptorExtensionSupplementaryAudio{}

	d.MixType = r.TryReadBool()
	d.EditorialClassification = uint8(r.TryReadBits(5))
	_ = r.TryReadBool() // Reserved.
	d.HasLanguageCode = r.TryReadBool()

	// Language code
	if d.HasLanguageCode {
		d.LanguageCode = make([]byte, 3)
		TryReadFull(r, d.LanguageCode)
	}

	if r.BitsCount/8 < offsetEnd {
		d.PrivateData = make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, d.PrivateData)
	}

	return d, r.TryError
}

// DescriptorISO639LanguageAndAudioType represents an ISO639 language descriptor
// https://github.com/gfto/bitstream/blob/master/mpeg/psi/desc_0a.h
// FIXME (barbashov) according to Chapter 2.6.18 ISO/IEC 13818-1:2015
// there could be not one, but multiple such descriptors.
type DescriptorISO639LanguageAndAudioType struct {
	Language []byte
	Type     uint8
}

// newDescriptorISO639LanguageAndAudioType In some actual cases,
// the length is 3 and the language is described in only 2 bytes.
func newDescriptorISO639LanguageAndAudioType(
	r *bitio.CountReader, offsetEnd int64,
) (*DescriptorISO639LanguageAndAudioType, error) {
	offset := uint8(offsetEnd - r.BitsCount/8)
	language := make([]byte, offset-1)
	TryReadFull(r, language)

	d := &DescriptorISO639LanguageAndAudioType{
		Language: language,
		Type:     r.TryReadByte(),
	}
	return d, r.TryError
}

// DescriptorLocalTimeOffset represents a local time offset descriptor
// Chapter: 6.2.20 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorLocalTimeOffset []*DescriptorLocalTimeOffsetItem

// DescriptorLocalTimeOffsetItem represents a local time offset item descriptor
// Chapter: 6.2.20 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorLocalTimeOffsetItem struct {
	CountryCode             []byte // 3 bytes.
	CountryRegionID         uint8  // 6 bits.
	LocalTimeOffset         time.Duration
	LocalTimeOffsetPolarity bool
	TimeOfChange            time.Time
	NextTimeOffset          time.Duration
}

func newDescriptorLocalTimeOffset(r *bitio.CountReader, offsetEnd int64) (DescriptorLocalTimeOffset, error) {
	d := DescriptorLocalTimeOffset{}

	for r.BitsCount/8 < offsetEnd {
		item := &DescriptorLocalTimeOffsetItem{}
		var err error

		item.CountryCode = make([]byte, 3)
		TryReadFull(r, item.CountryCode)

		item.CountryRegionID = uint8(r.TryReadBits(6))
		_ = r.TryReadBool() // Reserved.
		item.LocalTimeOffsetPolarity = r.TryReadBool()

		if item.LocalTimeOffset, err = parseDVBDurationMinutes(r); err != nil {
			return nil, fmt.Errorf("parsing localTimeOffset failed: %w", err)
		}

		if item.TimeOfChange, err = parseDVBTime(r); err != nil {
			return nil, fmt.Errorf("parsing timeOfChange failed: %w", err)
		}

		if item.NextTimeOffset, err = parseDVBDurationMinutes(r); err != nil {
			return nil, fmt.Errorf("parsing NextTimeOffset failed: %w", err)
		}

		d = append(d, item)
	}
	return d, r.TryError
}

// DescriptorMaximumBitrate represents a maximum bitrate descriptor.
// ISO/IEC 13818-1 Chapter: 2.6.26 .
type DescriptorMaximumBitrate struct {
	Bitrate uint32 // In bytes/second. 22 bits.
}

func newDescriptorMaximumBitrate(r *bitio.CountReader) (d DescriptorMaximumBitrate, err error) {
	r.TryReadBits(2) // Reserved.

	bitrate := uint32(r.TryReadBits(22))
	return DescriptorMaximumBitrate{Bitrate: bitrate}, r.TryError
}

// DescriptorNetworkName represents a network name descriptor.
// Chapter: 6.2.27 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorNetworkName struct {
	Name []byte
}

func newDescriptorNetworkName(r *bitio.CountReader, offsetEnd int64) (d DescriptorNetworkName, err error) {
	name := make([]byte, offsetEnd-r.BitsCount/8)
	TryReadFull(r, name)
	return DescriptorNetworkName{Name: name}, r.TryError
}

// DescriptorParentalRating represents a parental rating descriptor.
// Chapter: 6.2.28 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorParentalRating struct {
	Items []*DescriptorParentalRatingItem
}

// DescriptorParentalRatingItem represents a parental rating item descriptor.
// Chapter: 6.2.28 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorParentalRatingItem struct {
	CountryCode []byte // 3 bytes.
	Rating      uint8
}

// MinimumAge returns the minimum age for the parental rating.
func (d DescriptorParentalRatingItem) MinimumAge() int {
	// Undefined or user defined ratings
	if d.Rating == 0 || d.Rating > 0x10 {
		return 0
	}
	return int(d.Rating) + 3
}

func newDescriptorParentalRating(r *bitio.CountReader, offsetEnd int64) (DescriptorParentalRating, error) {
	items := []*DescriptorParentalRatingItem{}

	for r.BitsCount/8 < offsetEnd {
		country := make([]byte, 3)
		TryReadFull(r, country)

		rating := r.TryReadByte()

		items = append(items, &DescriptorParentalRatingItem{
			CountryCode: country,
			Rating:      rating,
		})
	}
	return DescriptorParentalRating{Items: items}, r.TryError
}

// DescriptorPrivateDataIndicator represents a private data Indicator descriptor.
type DescriptorPrivateDataIndicator uint32

func newDescriptorPrivateDataIndicator(r *bitio.CountReader) (DescriptorPrivateDataIndicator, error) {
	data := uint32(r.TryReadBits(32))
	return DescriptorPrivateDataIndicator(data), r.TryError
}

// DescriptorPrivateDataSpecifier represents a private data specifier descriptor.
type DescriptorPrivateDataSpecifier struct {
	Specifier uint32
}

func newDescriptorPrivateDataSpecifier(r *bitio.CountReader) (DescriptorPrivateDataSpecifier, error) {
	specifier := uint32(r.TryReadBits(32))
	return DescriptorPrivateDataSpecifier{Specifier: specifier}, r.TryError
}

// DescriptorRegistration represents a registration descriptor.
// Page: 84 | Link:
// http://ecee.colorado.edu/~ecen5653/ecen5653/papers/iso13818-1.pdf
type DescriptorRegistration struct {
	FormatIdentifier             uint32
	AdditionalIdentificationInfo []byte
}

func newDescriptorRegistration(r *bitio.CountReader, offsetEnd int64) (*DescriptorRegistration, error) {
	d := &DescriptorRegistration{}

	d.FormatIdentifier = uint32(r.TryReadBits(32))

	if r.BitsCount/8 < offsetEnd {
		d.AdditionalIdentificationInfo = make([]byte, offsetEnd-r.BitsCount/8)
		TryReadFull(r, d.AdditionalIdentificationInfo)
	}

	return d, r.TryError
}

// DescriptorService represents a service descriptor.
// Chapter: 6.2.33 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorService struct {
	Type     uint8
	Provider []byte
	Name     []byte
}

func newDescriptorService(r *bitio.CountReader) (*DescriptorService, error) {
	d := &DescriptorService{}

	d.Type = r.TryReadByte()

	providerLength := r.TryReadByte()
	d.Provider = make([]byte, providerLength)
	TryReadFull(r, d.Provider)

	nameLength := r.TryReadByte()
	d.Name = make([]byte, nameLength)
	TryReadFull(r, d.Name)

	return d, r.TryError
}

// DescriptorShortEvent represents a short event descriptor.
// Chapter: 6.2.37 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorShortEvent struct {
	Language  []byte // 3 bytes.
	EventName []byte
	Text      []byte
}

func newDescriptorShortEvent(r *bitio.CountReader) (*DescriptorShortEvent, error) {
	d := &DescriptorShortEvent{}

	d.Language = make([]byte, 3)
	TryReadFull(r, d.Language)

	eventLength := r.TryReadByte()
	d.EventName = make([]byte, eventLength)
	TryReadFull(r, d.EventName)

	textLength := r.TryReadByte()
	d.Text = make([]byte, textLength)
	TryReadFull(r, d.Text)

	return d, r.TryError
}

// DescriptorStreamIdentifier represents a stream identifier descriptor.
// Chapter: 6.2.39 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorStreamIdentifier struct {
	ComponentTag uint8
}

func newDescriptorStreamIdentifier(r *bitio.CountReader) (DescriptorStreamIdentifier, error) {
	identifier, err := r.ReadByte()
	return DescriptorStreamIdentifier{ComponentTag: identifier}, err
}

// DescriptorSubtitling represents a subtitling descriptor.
// Chapter: 6.2.41 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorSubtitling struct {
	Items []*DescriptorSubtitlingItem
}

// DescriptorSubtitlingItem represents subtitling descriptor item.
// Chapter: 6.2.41 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorSubtitlingItem struct {
	Language          []byte // 3 bytes.
	Type              uint8
	CompositionPageID uint16
	AncillaryPageID   uint16
}

func newDescriptorSubtitling(r *bitio.CountReader, offsetEnd int64) (DescriptorSubtitling, error) {
	items := []*DescriptorSubtitlingItem{}

	for r.BitsCount/8 < offsetEnd {
		item := &DescriptorSubtitlingItem{}

		item.Language = make([]byte, 3)
		TryReadFull(r, item.Language)

		item.Type = r.TryReadByte()

		item.CompositionPageID = uint16(r.TryReadBits(16))

		item.AncillaryPageID = uint16(r.TryReadBits(16))

		items = append(items, item)
	}

	return DescriptorSubtitling{Items: items}, r.TryError
}

// DescriptorTeletext represents a teletext descriptor.
// Chapter: 6.2.43 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorTeletext struct {
	Items []*DescriptorTeletextItem
}

// DescriptorTeletextItem represents a teletext descriptor item.
// Chapter: 6.2.43 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorTeletextItem struct {
	Language []byte
	Type     uint8 // 5 bits.
	Magazine uint8 // 3 bits.
	Page     uint8
}

func newDescriptorTeletext(r *bitio.CountReader, offsetEnd int64) (DescriptorTeletext, error) {
	items := []*DescriptorTeletextItem{}

	for r.BitsCount/8 < offsetEnd {
		item := &DescriptorTeletextItem{}

		item.Language = make([]byte, 3)
		TryReadFull(r, item.Language)

		item.Type = uint8(r.TryReadBits(5))

		item.Magazine = uint8(r.TryReadBits(3))

		b := r.TryReadByte()

		// Optimization?
		item.Page = b>>4*10 + b&0xf
		// w.TryWriteBits(item.Page/10, 4)
		// w.TryWriteBits(item.Page%10, 4)

		items = append(items, item)
	}
	return DescriptorTeletext{Items: items}, r.TryError
}

// DescriptorUnknown .
type DescriptorUnknown struct {
	Tag     uint8
	Content []byte
}

func newDescriptorUnknown(r *bitio.CountReader, tag, length uint8) (*DescriptorUnknown, error) {
	d := &DescriptorUnknown{Tag: tag}

	d.Content = make([]byte, length)
	TryReadFull(r, d.Content)
	return d, r.TryError
}

// DescriptorVBIData represents a VBI data descriptor.
// Chapter: 6.2.47 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIData []*DescriptorVBIDataService

// DescriptorVBIDataService represents a vbi data service descriptor.
// Chapter: 6.2.47 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIDataService struct {
	DataServiceID uint8
	Descriptors   []*DescriptorVBIDataDescriptor
}

// DescriptorVBIDataDescriptor represents a vbi data descriptor item.
// Chapter: 6.2.47 | Link:
// https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.15.01_60/en_300468v011501p.pdf
type DescriptorVBIDataDescriptor struct {
	FieldParity bool
	LineOffset  uint8 // 5 bits.
}

func newDescriptorVBIData(r *bitio.CountReader, offsetEnd int64) DescriptorVBIData {
	d := DescriptorVBIData{}

	for r.BitsCount/8 < offsetEnd {
		srv := &DescriptorVBIDataService{}

		srv.DataServiceID = r.TryReadByte()

		dataServiceDescriptorLength := r.TryReadByte()

		offsetDataEnd := r.BitsCount/8 + int64(dataServiceDescriptorLength)
		for r.BitsCount/8 < offsetDataEnd {
			if srv.DataServiceID == VBIDataServiceIDClosedCaptioning ||
				srv.DataServiceID == VBIDataServiceIDEBUTeletext ||
				srv.DataServiceID == VBIDataServiceIDInvertedTeletext ||
				srv.DataServiceID == VBIDataServiceIDMonochrome442Samples ||
				srv.DataServiceID == VBIDataServiceIDVPS ||
				srv.DataServiceID == VBIDataServiceIDWSS {
				_ = r.TryReadBits(2) // Reserved.

				srv.Descriptors = append(srv.Descriptors, &DescriptorVBIDataDescriptor{
					FieldParity: r.TryReadBool(),
					LineOffset:  uint8(r.TryReadBits(5)),
				})
			}
		}
		d = append(d, srv)
	}

	return d
}

// parseDescriptors parses descriptors.
func parseDescriptors(r *bitio.CountReader) ([]*Descriptor, error) {
	var o []*Descriptor

	length := int64(r.TryReadBits(12))

	if length <= 0 {
		return o, nil
	}

	offsetEnd := r.BitsCount/8 + length
	for r.BitsCount/8 < offsetEnd {
		d := &Descriptor{
			Tag:    r.TryReadByte(),
			Length: r.TryReadByte(),
		}

		if r.TryError != nil {
			return nil, r.TryError
		}

		if d.Length <= 0 {
			continue
		}

		// Parse data.
		// Unfortunately there's no way to be sure the real descriptor
		// length is the same as the one indicated previously therefore
		// we must fetch bytes in descriptor functions and seek at the end.
		offsetDescriptorEnd := r.BitsCount/8 + int64(d.Length)

		// User defined
		if d.Tag >= 0x80 && d.Tag <= 0xfe {
			d.UserDefined = make([]byte, d.Length)
			TryReadFull(r, d.UserDefined)

			// Make sure we move to the end of the descriptor
			// since its content may be corrupted.
			if offsetDescriptorEnd > r.BitsCount/8 {
				skip := make([]byte, offsetDescriptorEnd-r.BitsCount/8)
				TryReadFull(r, skip)
			}

			o = append(o, d)
			continue
		}

		err := parseDescriptor(d, r, offsetDescriptorEnd)
		if err != nil {
			return nil, err
		}

		o = append(o, d)
	}
	return o, r.TryError
}

func parseDescriptor( //nolint:funlen,gocognit,gocyclo
	d *Descriptor,
	r *bitio.CountReader,
	offsetDescriptorEnd int64,
) error {
	var err error

	switch d.Tag {
	case DescriptorTagAC3:
		if d.AC3, err = newDescriptorAC3(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing AC3 descriptor failed: %w", err)
		}
	case DescriptorTagAVCVideo:
		if d.AVCVideo, err = newDescriptorAVCVideo(r); err != nil {
			return fmt.Errorf("parsing AVC Video descriptor failed: %w", err)
		}
	case DescriptorTagComponent:
		if d.Component, err = newDescriptorComponent(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Component descriptor failed: %w", err)
		}
	case DescriptorTagContent:
		if d.Content, err = newDescriptorContent(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Content descriptor failed: %w", err)
		}
	case DescriptorTagDataStreamAlignment:
		if d.DataStreamAlignment, err = newDescriptorDataStreamAlignment(r); err != nil {
			return fmt.Errorf("parsing Data Stream Alignment descriptor failed: %w", err)
		}
	case DescriptorTagEnhancedAC3:
		if d.EnhancedAC3, err = newDescriptorEnhancedAC3(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Enhanced AC3 descriptor failed: %w", err)
		}
	case DescriptorTagExtendedEvent:
		if d.ExtendedEvent, err = newDescriptorExtendedEvent(r); err != nil {
			return fmt.Errorf("parsing Extended event descriptor failed: %w", err)
		}
	case DescriptorTagExtension:
		if d.Extension, err = newDescriptorExtension(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Extension descriptor failed: %w", err)
		}
	case DescriptorTagISO639LanguageAndAudioType:
		if d.ISO639LanguageAndAudioType, err = newDescriptorISO639LanguageAndAudioType(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing ISO639 Language and Audio Type descriptor failed: %w", err)
		}
	case DescriptorTagLocalTimeOffset:
		if d.LocalTimeOffset, err = newDescriptorLocalTimeOffset(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Local Time Offset descriptor failed: %w", err)
		}
	case DescriptorTagMaximumBitrate:
		if d.MaximumBitrate, err = newDescriptorMaximumBitrate(r); err != nil {
			return fmt.Errorf("parsing Maximum Bitrate descriptor failed: %w", err)
		}
	case DescriptorTagNetworkName:
		if d.NetworkName, err = newDescriptorNetworkName(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Network Name descriptor failed: %w", err)
		}
	case DescriptorTagParentalRating:
		if d.ParentalRating, err = newDescriptorParentalRating(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Parental Rating descriptor failed: %w", err)
		}
	case DescriptorTagPrivateDataIndicator:
		if d.PrivateDataIndicator, err = newDescriptorPrivateDataIndicator(r); err != nil {
			return fmt.Errorf("parsing Private Data Indicator descriptor failed: %w", err)
		}
	case DescriptorTagPrivateDataSpecifier:
		if d.PrivateDataSpecifier, err = newDescriptorPrivateDataSpecifier(r); err != nil {
			return fmt.Errorf("parsing Private Data Specifier descriptor failed: %w", err)
		}
	case DescriptorTagRegistration:
		if d.Registration, err = newDescriptorRegistration(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Registration descriptor failed: %w", err)
		}
	case DescriptorTagService:
		if d.Service, err = newDescriptorService(r); err != nil {
			return fmt.Errorf("parsing Service descriptor failed: %w", err)
		}
	case DescriptorTagShortEvent:
		if d.ShortEvent, err = newDescriptorShortEvent(r); err != nil {
			return fmt.Errorf("parsing Short Event descriptor failed: %w", err)
		}
	case DescriptorTagStreamIdentifier:
		if d.StreamIdentifier, err = newDescriptorStreamIdentifier(r); err != nil {
			return fmt.Errorf("parsing Stream Identifier descriptor failed: %w", err)
		}
	case DescriptorTagSubtitling:
		if d.Subtitling, err = newDescriptorSubtitling(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Subtitling descriptor failed: %w", err)
		}
	case DescriptorTagTeletext:
		if d.Teletext, err = newDescriptorTeletext(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing Teletext descriptor failed: %w", err)
		}
	case DescriptorTagVBIData:
		d.VBIData = newDescriptorVBIData(r, offsetDescriptorEnd)
	case DescriptorTagVBITeletext:
		if d.VBITeletext, err = newDescriptorTeletext(r, offsetDescriptorEnd); err != nil {
			return fmt.Errorf("parsing VBI Teletext descriptor failed: %w", err)
		}
	default:
		if d.Unknown, err = newDescriptorUnknown(r, d.Tag, d.Length); err != nil {
			return fmt.Errorf("parsing unknown descriptor failed: %w", err)
		}
	}

	// Make sure we move to the end of the descriptor
	// since its content may be corrupted.
	if offsetDescriptorEnd > r.BitsCount/8 {
		seek := make([]byte, offsetDescriptorEnd-r.BitsCount/8)
		TryReadFull(r, seek)
	}

	return nil
}

func calcDescriptorUserDefinedLength(d []byte) uint8 {
	return uint8(len(d))
}

func calcDescriptorAC3Length(d *DescriptorAC3) uint8 {
	ret := 1 // flags.

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

func writeDescriptorAC3(w *bitio.Writer, d *DescriptorAC3) error {
	w.TryWriteBool(d.HasComponentType)
	w.TryWriteBool(d.HasBSID)
	w.TryWriteBool(d.HasMainID)
	w.TryWriteBool(d.HasASVC)
	w.TryWriteBits(0xff, 4) // Reserved.

	if d.HasComponentType {
		w.TryWriteByte(d.ComponentType)
	}
	if d.HasBSID {
		w.TryWriteByte(d.BSID)
	}
	if d.HasMainID {
		w.TryWriteByte(d.MainID)
	}
	if d.HasASVC {
		w.TryWriteByte(d.ASVC)
	}
	w.TryWrite(d.AdditionalInfo)

	return w.TryError
}

func calcDescriptorAVCVideoLength(d *DescriptorAVCVideo) uint8 {
	return 4
}

func writeDescriptorAVCVideo(w *bitio.Writer, d *DescriptorAVCVideo) error {
	w.TryWriteByte(d.ProfileIDC)

	w.TryWriteBool(d.ConstraintSet0Flag)
	w.TryWriteBool(d.ConstraintSet1Flag)
	w.TryWriteBool(d.ConstraintSet2Flag)
	w.TryWriteBits(uint64(d.CompatibleFlags), 5)

	w.TryWriteByte(d.LevelIDC)

	w.TryWriteBool(d.AVCStillPresent)
	w.TryWriteBool(d.AVC24HourPictureFlag)
	w.TryWriteBits(uint64(0xff), 6) // Reserved.

	return w.TryError
}

func calcDescriptorComponentLength(d *DescriptorComponent) uint8 {
	return uint8(6 + len(d.Text))
}

func writeDescriptorComponent(w *bitio.Writer, d *DescriptorComponent) error {
	w.TryWriteBits(uint64(d.StreamContentExt), 4)
	w.TryWriteBits(uint64(d.StreamContent), 4)

	w.TryWriteByte(d.ComponentType)
	w.TryWriteByte(d.ComponentTag)

	w.TryWrite(d.ISO639LanguageCode)

	w.TryWrite(d.Text)

	return w.TryError
}

func calcDescriptorContentLength(d DescriptorContent) uint8 {
	return uint8(2 * len(d.Items))
}

func writeDescriptorContent(w *bitio.Writer, d DescriptorContent) error {
	for _, item := range d.Items {
		w.TryWriteBits(uint64(item.ContentNibbleLevel1), 4)
		w.TryWriteBits(uint64(item.ContentNibbleLevel2), 4)
		w.TryWriteByte(item.UserByte)
	}
	return w.TryError
}

func calcDescriptorDataStreamAlignmentLength(d DescriptorDataStreamAlignment) uint8 {
	return 1
}

func writeDescriptorDataStreamAlignment(w *bitio.Writer, d DescriptorDataStreamAlignment) error {
	return w.WriteByte(uint8(d))
}

func calcDescriptorEnhancedAC3Length(d *DescriptorEnhancedAC3) uint8 {
	ret := 1 // flags.

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

func writeDescriptorEnhancedAC3(w *bitio.Writer, d *DescriptorEnhancedAC3) error {
	w.TryWriteBool(d.HasComponentType)
	w.TryWriteBool(d.HasBSID)
	w.TryWriteBool(d.HasMainID)
	w.TryWriteBool(d.HasASVC)
	w.TryWriteBool(d.MixInfoExists)
	w.TryWriteBool(d.HasSubStream1)
	w.TryWriteBool(d.HasSubStream2)
	w.TryWriteBool(d.HasSubStream3)

	if d.HasComponentType {
		w.TryWriteByte(d.ComponentType)
	}
	if d.HasBSID {
		w.TryWriteByte(d.BSID)
	}
	if d.HasMainID {
		w.TryWriteByte(d.MainID)
	}
	if d.HasASVC {
		w.TryWriteByte(d.ASVC)
	}
	if d.HasSubStream1 {
		w.TryWriteByte(d.SubStream1)
	}
	if d.HasSubStream2 {
		w.TryWriteByte(d.SubStream2)
	}
	if d.HasSubStream3 {
		w.TryWriteByte(d.SubStream3)
	}

	w.TryWrite(d.AdditionalInfo)

	return w.TryError
}

func calcDescriptorExtendedEventLength(d *DescriptorExtendedEvent) (descriptorLength, lengthOfItems uint8) {
	ret := 1 + 3 + 1 // numbers, language and items length.

	itemsRet := 0
	for _, item := range d.Items {
		itemsRet++ // description length
		itemsRet += len(item.Description)
		itemsRet++ // content length
		itemsRet += len(item.Content)
	}

	ret += itemsRet

	ret++ // text length
	ret += len(d.Text)

	return uint8(ret), uint8(itemsRet)
}

func writeDescriptorExtendedEvent(w *bitio.Writer, d *DescriptorExtendedEvent) error {
	var lengthOfItems uint8

	_, lengthOfItems = calcDescriptorExtendedEventLength(d)

	w.TryWriteBits(uint64(d.Number), 4)
	w.TryWriteBits(uint64(d.LastDescriptorNumber), 4)

	w.TryWrite(d.ISO639LanguageCode)

	w.TryWriteByte(lengthOfItems)
	for _, item := range d.Items {
		w.TryWriteByte(uint8(len(item.Description)))
		w.TryWrite(item.Description)
		w.TryWriteByte(uint8(len(item.Content)))
		w.TryWrite(item.Content)
	}

	w.TryWriteByte(uint8(len(d.Text)))
	w.TryWrite(d.Text)

	return w.TryError
}

func calcDescriptorExtensionSupplementaryAudioLength(d *DescriptorExtensionSupplementaryAudio) int {
	ret := 1
	if d.HasLanguageCode {
		ret += 3
	}
	ret += len(d.PrivateData)
	return ret
}

func calcDescriptorExtensionLength(d *DescriptorExtension) uint8 {
	ret := 1 // tag.

	switch d.Tag {
	case DescriptorTagExtensionSupplementaryAudio:
		ret += calcDescriptorExtensionSupplementaryAudioLength(d.SupplementaryAudio)
	default:
		if d.Unknown != nil {
			ret += len(*d.Unknown)
		}
	}

	return uint8(ret)
}

func writeDescriptorExtensionSupplementaryAudio(w *bitio.Writer, d *DescriptorExtensionSupplementaryAudio) error {
	w.TryWriteBool(d.MixType)
	w.TryWriteBits(uint64(d.EditorialClassification), 5)
	w.TryWriteBool(true) // Reserved.
	w.TryWriteBool(d.HasLanguageCode)

	if d.HasLanguageCode {
		w.TryWrite(d.LanguageCode)
	}

	w.TryWrite(d.PrivateData)

	return w.TryError
}

func writeDescriptorExtension(w *bitio.Writer, d *DescriptorExtension) error {
	if err := w.WriteByte(d.Tag); err != nil {
		return err
	}

	switch d.Tag {
	case DescriptorTagExtensionSupplementaryAudio:
		err := writeDescriptorExtensionSupplementaryAudio(w, d.SupplementaryAudio)
		if err != nil {
			return err
		}
	default:
		if d.Unknown != nil {
			if _, err := w.Write(*d.Unknown); err != nil {
				return err
			}
		}
	}
	return nil
}

func calcDescriptorISO639LanguageAndAudioTypeLength(d *DescriptorISO639LanguageAndAudioType) uint8 {
	return 3 + 1 // language code + type.
}

func writeDescriptorISO639LanguageAndAudioType(w *bitio.Writer, d *DescriptorISO639LanguageAndAudioType) error {
	w.TryWrite(d.Language)
	w.TryWriteByte(d.Type)

	return w.TryError
}

func calcDescriptorLocalTimeOffsetLength(d DescriptorLocalTimeOffset) uint8 {
	return uint8(13 * len(d))
}

func writeDescriptorLocalTimeOffset(w *bitio.Writer, d DescriptorLocalTimeOffset) error {
	for _, item := range d {
		w.TryWrite(item.CountryCode)

		w.TryWriteBits(uint64(item.CountryRegionID), 6)
		w.TryWriteBits(0xff, 1) // Reserved.
		w.TryWriteBool(item.LocalTimeOffsetPolarity)

		if err := writeDVBDurationMinutes(w, item.LocalTimeOffset); err != nil {
			return fmt.Errorf("writing LocalTimeOffset failed: %w", err)
		}
		if _, err := writeDVBTime(w, item.TimeOfChange); err != nil {
			return fmt.Errorf("writing TimeOfChange failed: %w", err)
		}
		if err := writeDVBDurationMinutes(w, item.NextTimeOffset); err != nil {
			return fmt.Errorf("writing NextTimeOffset failed: %w", err)
		}
	}

	return w.TryError
}

func calcDescriptorMaximumBitrateLength(d DescriptorMaximumBitrate) uint8 {
	return 3
}

func writeDescriptorMaximumBitrate(w *bitio.Writer, d DescriptorMaximumBitrate) error {
	w.TryWriteBits(0xff, 2) // Reserved.
	w.TryWriteBits(uint64(d.Bitrate), 22)

	return w.TryError
}

func calcDescriptorNetworkNameLength(name DescriptorNetworkName) uint8 {
	return uint8(len(name.Name))
}

func writeDescriptorNetworkName(w *bitio.Writer, d DescriptorNetworkName) error {
	_, err := w.Write(d.Name)
	return err
}

func calcDescriptorParentalRatingLength(d DescriptorParentalRating) uint8 {
	return uint8(4 * len(d.Items))
}

func writeDescriptorParentalRating(w *bitio.Writer, d DescriptorParentalRating) error {
	for _, item := range d.Items {
		w.TryWrite(item.CountryCode)
		w.TryWriteByte(item.Rating)
	}
	return w.TryError
}

func calcDescriptorPrivateDataIndicatorLength(d DescriptorPrivateDataIndicator) uint8 {
	return 4
}

func writeDescriptorPrivateDataIndicator(w *bitio.Writer, d DescriptorPrivateDataIndicator) error {
	return w.WriteBits(uint64(d), 32)
}

func calcDescriptorPrivateDataSpecifierLength(d DescriptorPrivateDataSpecifier) uint8 {
	return 4
}

func writeDescriptorPrivateDataSpecifier(w *bitio.Writer, d DescriptorPrivateDataSpecifier) error {
	return w.WriteBits(uint64(d.Specifier), 32)
}

func calcDescriptorRegistrationLength(d *DescriptorRegistration) uint8 {
	return uint8(4 + len(d.AdditionalIdentificationInfo))
}

func writeDescriptorRegistration(w *bitio.Writer, d *DescriptorRegistration) error {
	w.TryWriteBits(uint64(d.FormatIdentifier), 32)
	w.TryWrite(d.AdditionalIdentificationInfo)

	return w.TryError
}

func calcDescriptorServiceLength(d *DescriptorService) uint8 {
	ret := 3 // type and lengths
	ret += len(d.Name)
	ret += len(d.Provider)
	return uint8(ret)
}

func writeDescriptorService(w *bitio.Writer, d *DescriptorService) error {
	w.TryWriteByte(d.Type)
	w.TryWriteByte(uint8(len(d.Provider)))
	w.TryWrite(d.Provider)
	w.TryWriteByte(uint8(len(d.Name)))
	w.TryWrite(d.Name)

	return w.TryError
}

func calcDescriptorShortEventLength(d *DescriptorShortEvent) uint8 {
	ret := 3 + 1 + 1 // Language code and lengths.
	ret += len(d.EventName)
	ret += len(d.Text)
	return uint8(ret)
}

func writeDescriptorShortEvent(w *bitio.Writer, d *DescriptorShortEvent) error {
	w.TryWrite(d.Language)

	w.TryWriteByte(uint8(len(d.EventName)))
	w.TryWrite(d.EventName)

	w.TryWriteByte(uint8(len(d.Text)))
	w.TryWrite(d.Text)

	return w.TryError
}

func calcDescriptorStreamIdentifierLength(d DescriptorStreamIdentifier) uint8 {
	return 1
}

func writeDescriptorStreamIdentifier(w *bitio.Writer, d DescriptorStreamIdentifier) error {
	return w.WriteByte(d.ComponentTag)
}

func calcDescriptorSubtitlingLength(d DescriptorSubtitling) uint8 {
	return uint8(8 * len(d.Items))
}

func writeDescriptorSubtitling(w *bitio.Writer, d DescriptorSubtitling) error {
	for _, item := range d.Items {
		w.TryWrite(item.Language)
		w.TryWriteByte(item.Type)
		w.TryWriteBits(uint64(item.CompositionPageID), 16)
		w.TryWriteBits(uint64(item.AncillaryPageID), 16)
	}
	return w.TryError
}

func calcDescriptorTeletextLength(d DescriptorTeletext) uint8 {
	return uint8(5 * len(d.Items))
}

func writeDescriptorTeletext(w *bitio.Writer, d DescriptorTeletext) error {
	for _, item := range d.Items {
		w.TryWrite(item.Language)
		w.TryWriteBits(uint64(item.Type), 5)
		w.TryWriteBits(uint64(item.Magazine), 3)
		w.TryWriteBits(uint64(item.Page/10), 4)
		w.TryWriteBits(uint64(item.Page%10), 4)
	}
	return w.TryError
}

func calcDescriptorVBIDataLength(d DescriptorVBIData) uint8 {
	return uint8(3 * len(d))
}

func writeDescriptorVBIData(w *bitio.Writer, d DescriptorVBIData) error {
	for _, item := range d {
		w.TryWriteByte(item.DataServiceID)

		if item.DataServiceID == VBIDataServiceIDClosedCaptioning ||
			item.DataServiceID == VBIDataServiceIDEBUTeletext ||
			item.DataServiceID == VBIDataServiceIDInvertedTeletext ||
			item.DataServiceID == VBIDataServiceIDMonochrome442Samples ||
			item.DataServiceID == VBIDataServiceIDVPS ||
			item.DataServiceID == VBIDataServiceIDWSS {
			w.TryWriteByte(uint8(len(item.Descriptors))) // Each descriptor is 1 byte.
			for _, desc := range item.Descriptors {
				w.TryWriteBits(0xff, 2) // Reserved.
				w.TryWriteBool(desc.FieldParity)
				w.TryWriteBits(uint64(desc.LineOffset), 5)
			}
		} else {
			// Let's put one reserved byte.
			w.TryWriteByte(1)
			w.TryWriteByte(0xff)
		}
	}

	return w.TryError
}

func calcDescriptorUnknownLength(d *DescriptorUnknown) uint8 {
	return uint8(len(d.Content))
}

func writeDescriptorUnknown(w *bitio.Writer, d *DescriptorUnknown) error {
	_, err := w.Write(d.Content)
	return err
}

func calcDescriptorLength(d *Descriptor) uint8 { //nolint:funlen
	if d.Tag >= 0x80 && d.Tag <= 0xfe {
		return calcDescriptorUserDefinedLength(d.UserDefined)
	}

	switch d.Tag {
	case DescriptorTagAC3:
		return calcDescriptorAC3Length(d.AC3)
	case DescriptorTagAVCVideo:
		return calcDescriptorAVCVideoLength(d.AVCVideo)
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

func writeDescriptor(w *bitio.Writer, d *Descriptor) (int, error) { //nolint:funlen
	length := calcDescriptorLength(d)

	w.TryWriteByte(d.Tag)
	w.TryWriteByte(length)
	if w.TryError != nil {
		return 0, w.TryError
	}

	written := int(length) + 2

	if d.Tag >= 0x80 && d.Tag <= 0xfe {
		_, err := w.Write(d.UserDefined)
		return written, err
	}

	switch d.Tag {
	case DescriptorTagAC3:
		return written, writeDescriptorAC3(w, d.AC3)
	case DescriptorTagAVCVideo:
		return written, writeDescriptorAVCVideo(w, d.AVCVideo)
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
		length += 2 // Tag and length.
		length += uint16(calcDescriptorLength(d))
	}
	return length
}

func writeDescriptors(w *bitio.Writer, ds []*Descriptor) (int, error) {
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

func writeDescriptorsWithLength(w *bitio.Writer, ds []*Descriptor) (int, error) {
	length := calcDescriptorsLength(ds)

	w.TryWriteBits(0xff, 4)            // Reserved.
	w.TryWriteBits(uint64(length), 12) // program_info_length.

	if w.TryError != nil {
		return 0, w.TryError
	}

	written, err := writeDescriptors(w, ds)
	if err != nil {
		return 0, fmt.Errorf("writing descriptors failed: %w", err)
	}

	written += 2
	return written, nil
}
