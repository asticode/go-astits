package astits

import (
	"fmt"

	"github.com/asticode/go-astikit"
)

// Stream types from https://en.wikipedia.org/wiki/Program-specific_information#Elementary_stream_types
const (
	StreamTypeMPEG1Video                                = 0x01 // ISO/IEC 11172-2
	StreamTypeMPEG2HighRateInterlacedVideo              = 0x02 // Rec. ITU-T H.262 | ISO/IEC 13818-2
	StreamTypeMPEG1Audio                                = 0x03 // ISO/IEC 11172-3
	StreamTypeMPEG2HalvedSampleRateAudio                = 0x04 // ISO/IEC 13818-3
	StreamTypeMPEG2MPEG2TabledData                      = 0x05 // Rec. ITU-T H.222 | ISO/IEC 13818-1
	StreamTypeMPEG2PacketizedData                       = 0x06 // Rec. ITU-T H.222 | ISO/IEC 13818-1 i.e., DVB subtitles/VBI and AC-3
	StreamTypeMHEG                                      = 0x07 // ISO/IEC 13522
	StreamTypeDSMCC                                     = 0x08 // Rec. ITU-T H.222 | ISO/IEC 13818-1
	StreamTypeAuxiliaryDataITUAndISO                    = 0x09 // Rec. ITU-T H.222 | ISO/IEC 13818-1/11172-1
	StreamTypeDSMCCMultiProtocolEncapsulation           = 0x0A // ISO/IEC 13818-6
	StreamTypeDSMCCUNMessages                           = 0x0B // ISO/IEC 13818-6
	StreamTypeDSMCCStreamDescriptors                    = 0x0C // ISO/IEC 13818-6
	StreamTypeDSMCCTabledData                           = 0x0D // ISO/IEC 13818-6
	StreamTypeAuxiliaryDataISO                          = 0x0E // ISO/IEC 13818-1
	StreamTypeAudioADTS                                 = 0x0F // ISO/IEC 13818-7 Audio with ADTS transport syntax
	StreamTypeMPEG4H263Video                            = 0x10 // ISO/IEC 14496-2
	StreamTypeMPEG4LOASMultiFormatFramedAudio           = 0x11 // ISO/IEC 14496-3
	StreamTypeMPEG4FlexMux                              = 0x12 // ISO/IEC 14496-1
	StreamTypeMPEG4FlexMuxInTables                      = 0x13 // ISO/IEC 14496-1 in ISO/IEC 14496 tables
	StreamTypeDSMCCSynchronisedDownloadProtocol         = 0x14 // ISO/IEC 13818-6
	StreamTypePacketisedMetadata                        = 0x15 // Packetized metadata
	StreamTypeSectionedMetadata                         = 0x16 // Sectioned metadata
	StreamTypeDSMCCDataCarouselMetadata                 = 0x17 // ISO/IEC 13818-6
	StreamTypeDSMCCObjectCarouselMetadata               = 0x18 // ISO/IEC 13818-6
	StreamTypeDSMCCSynchronisedDownloadProtocolMetadata = 0x19 // ISO/IEC 13818-6
	StreamTypeIPMP                                      = 0x1A // ISO/IEC 13818-11
	StreamTypeH264Video                                 = 0x1B // Rec. ITU-T H.264 | ISO/IEC 14496-10
	StreamTypeMPEG4RawAudio                             = 0x1C // ISO/IEC 14496-3
	StreamTypeMPEG4Text                                 = 0x1D // ISO/IEC 14496-17
	StreamTypeMPEG4AuxiliaryVideo                       = 0x1E // ISO/IEC 23002-3
	StreamTypeSVCMPEG4AVCSubBitstream                   = 0x1F // ISO/IEC 14496-10
	StreamTypeMVCMPEG4AVCSubBitstream                   = 0x20 // ISO/IEC 14496-10
	StreamTypeJPEG2000Video                             = 0x21 // Rec. ITU-T T.800 | ISO/IEC 15444
	//0x22 and 0x23 are reserved
	StreamTypeH265Video = 0x24 // Rec. ITU-T H.265 | ISO/IEC 23008-2
	//0x25 to 0x41 are reserved
	StreamTypeChineseVideoStandard = 0x42 // Chinese Video Standard
	//0x43 to 0x7e are reserved
	StreamTypeIPMPDRM                                           = 0x7f // ISO/IEC 13818-11
	StreamTypeBluRayDigiCipher2OrPCMAudioWithDES64CBCEncryption = 0x80 // Rec. ITU-T H.262 | ISO/IEC 13818-2
	StreamTypeBluRayAndATSCDolbyDigitalAC3Max6ChannelAudio      = 0x81 // Dolby Digital (AC-3) up to six channel audio for ATSC and Blu-ray
	StreamTypeBluRayDTS6ChannelAudioOrSCTESubtitle              = 0x82 // SCTE subtitle or DTS 6 channel audio for BluRay
	StreamTypeBlueRayDolbyTrueHDAudio                           = 0x83 // Dolby TrueHD lossless audio for Blu-ray
	StreamTypeBluRayDoblyDigitalPlusAC3Max16ChannelAudio        = 0x84 // Dolby Digital Plus (enhanced AC-3) up to 16 channel audio for Blu-ray
	StreamTypeBluRayDTS8ChannelAudio                            = 0x85 // DTS 8 channel audio for Blu-ray
	StreamTypeBluRaySCTE35OrDTS8ChannelAudio                    = 0x86 // SCTE-35[5] digital program insertion cue message or DTS 8 channel lossless audio for Blu-ray
	StreamTypeATSCDoblyDigitalPlusAC3Max16ChannelAudio          = 0x87 // Dolby Digital Plus (enhanced AC-3) up to 16 channel audio for ATSC
	// 0x88 - 0x8F privately defined
	StreamTypeBluRayPresentationGraphicStream = 0x90 // Blu-ray Presentation Graphic Stream (subtitling)
	StreamTypeATSCDSMCCNetworkResourcesTable  = 0x91 // ATSC DSM CC Network Resources table
	// 0x92 - 0xBF privately defined
	StreamTypeDigiCipher2text                                                          = 0xC0 // DigiCipher II text
	StreamTypeDolbyDigitalAC3Max6ChannelAudioWithAES128CBC                             = 0xC1 // Dolby Digital (AC-3) up to six channel audio with AES-128-CBC data encryption
	StreamTypeATSEDSMCCSynchronousDataOrDolbyDigitalPlusMax16ChannelAudioWithAES128CBC = 0xC2 // ATSC DSM CC synchronous data or Dolby Digital Plus up to 16 channel audio with AES-128-CBC data encryption
	// 0xC3 - 0xCE privately defined
	StreamTypeADTSAACWithAES128CBC = 0xCF // ISO/IEC 13818-7
	// 0xD0 privately defined
	StreamTypeBBCDiracVideo = 0xD1 // BBC Dirac (Ultra HD video)
	// 0xD2 - 0xDA privately defined
	StreamTypeAES128CBCSliceEncryption = 0xDB // Rec. ITU-T H.264 and ISO/IEC 14496-10
	// 0xDC - 0xE9 privately defined
	StreamTypeMicrosoftWindowsMediaVideo9 = 0xEA // Microsoft Windows Media Video 9 (lower bit-rate video)
	// 0xEB - 0xFF privately defined

)

// PMTData represents a PMT data
// https://en.wikipedia.org/wiki/Program-specific_information
type PMTData struct {
	ElementaryStreams  []*PMTElementaryStream
	PCRPID             uint16        // The packet identifier that contains the program clock reference used to improve the random access accuracy of the stream's timing that is derived from the program timestamp. If this is unused. then it is set to 0x1FFF (all bits on).
	ProgramDescriptors []*Descriptor // Program descriptors
	ProgramNumber      uint16
}

// PMTElementaryStream represents a PMT elementary stream
type PMTElementaryStream struct {
	ElementaryPID               uint16        // The packet identifier that contains the stream type data.
	ElementaryStreamDescriptors []*Descriptor // Elementary stream descriptors
	StreamType                  uint8         // This defines the structure of the data contained within the elementary packet identifier.
}

// parsePMTSection parses a PMT section
func parsePMTSection(i *astikit.BytesIterator, offsetSectionsEnd int, tableIDExtension uint16) (d *PMTData, err error) {
	// Create data
	d = &PMTData{ProgramNumber: tableIDExtension}

	// Get next bytes
	var bs []byte
	if bs, err = i.NextBytes(2); err != nil {
		err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
		return
	}

	// PCR PID
	d.PCRPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

	// Program descriptors
	if d.ProgramDescriptors, err = parseDescriptors(i); err != nil {
		err = fmt.Errorf("astits: parsing descriptors failed: %w", err)
		return
	}

	// Loop until end of section data is reached
	for i.Offset() < offsetSectionsEnd {
		// Create stream
		e := &PMTElementaryStream{}

		// Get next byte
		var b byte
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("astits: fetching next byte failed: %w", err)
			return
		}

		// Stream type
		e.StreamType = uint8(b)

		// Get next bytes
		if bs, err = i.NextBytes(2); err != nil {
			err = fmt.Errorf("astits: fetching next bytes failed: %w", err)
			return
		}

		// Elementary PID
		e.ElementaryPID = uint16(bs[0]&0x1f)<<8 | uint16(bs[1])

		// Elementary descriptors
		if e.ElementaryStreamDescriptors, err = parseDescriptors(i); err != nil {
			err = fmt.Errorf("astits: parsing descriptors failed: %w", err)
			return
		}

		// Add elementary stream
		d.ElementaryStreams = append(d.ElementaryStreams, e)
	}
	return
}

func (p *PMTData) Serialise(b []byte) (int, error) {
	if len(b) <= 4 {
		return 0, ErrNoRoomInBuffer
	}
	b[0] = 0x7<<5 | uint8(0x1f&(p.PCRPID>>8))
	b[1] = uint8(0xff & p.PCRPID)
	program_info_length := 0
	idx := 4
	for i := range p.ProgramDescriptors {
		n, err := p.ProgramDescriptors[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
		program_info_length += n
	}
	for i := range p.ElementaryStreams {
		n, err := p.ElementaryStreams[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
	}
	b[2] = 0xf0 | uint8(0x3&(uint8(program_info_length)>>8))
	b[3] = uint8(program_info_length)
	return idx, nil
}

func (pes *PMTElementaryStream) Serialise(b []byte) (int, error) {
	if len(b) <= 5 {
		return 0, ErrNoRoomInBuffer
	}
	b[0] = pes.StreamType
	b[1] = 0x7<<5 | uint8(0x1f&(pes.ElementaryPID>>8))
	b[2] = uint8(0xff & pes.ElementaryPID)
	es_info_length := 0
	idx := 5
	for i := range pes.ElementaryStreamDescriptors {
		n, err := pes.ElementaryStreamDescriptors[i].Serialise(b[idx:])
		if err != nil {
			return idx, err
		}
		idx += n
		es_info_length += n
	}
	b[3] = 0xf0 | (uint8(0x3 & (es_info_length >> 8)))
	b[4] = uint8(es_info_length)
	return idx, nil
}
