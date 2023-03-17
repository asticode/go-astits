package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astits"
	"github.com/pkg/profile"
)

// Flags
var (
	ctx, cancel     = context.WithCancel(context.Background())
	cpuProfiling    = flag.Bool("cp", false, "if yes, cpu profiling is enabled")
	dataTypes       = astikit.NewFlagStrings()
	format          = flag.String("f", "", "the format")
	inputPath       = flag.String("i", "", "the input path")
	memoryProfiling = flag.Bool("mp", false, "if yes, memory profiling is enabled")
)

func main() {
	// Init
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s <data|packets|default>:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Var(dataTypes, "d", "the datatypes whitelist (all, pat, pmt, pes, eit, nit, sdt, tot)")
	cmd := astikit.FlagCmd()
	flag.Parse()

	// Handle signals
	handleSignals()

	// Start profiling
	if *cpuProfiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	} else if *memoryProfiling {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	}

	// Build the reader
	var r io.Reader
	var err error
	if r, err = buildReader(ctx); err != nil {
		log.Fatal(fmt.Errorf("astits: parsing input failed: %w", err))
	}

	// Make sure the reader is closed properly
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}

	// Create the demuxer
	var dmx = astits.NewDemuxer(ctx, r, astits.DemuxerOptLogger(log.Default()))

	// Switch on command
	switch cmd {
	case "data":
		// Fetch data
		if err = data(dmx); err != nil {
			if !errors.Is(err, astits.ErrNoMorePackets) {
				log.Fatal(fmt.Errorf("astits: fetching data failed: %w", err))
			}
		}
	case "packets":
		// Fetch packets
		if err = packets(dmx); err != nil {
			if !errors.Is(err, astits.ErrNoMorePackets) {
				log.Fatal(fmt.Errorf("astits: fetching packets failed: %w", err))
			}
		}
	default:
		// Fetch the programs
		var pgms []*Program
		if pgms, err = programs(dmx); err != nil {
			if !errors.Is(err, astits.ErrNoMorePackets) {
				log.Fatal(fmt.Errorf("astits: fetching programs failed: %w", err))
			}
		}

		// Print
		switch *format {
		case "json":
			var e = json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			if err = e.Encode(pgms); err != nil {
				log.Fatal(fmt.Errorf("astits: json encoding to stdout failed: %w", err))
			}
		default:
			fmt.Println("Programs are:")
			for _, pgm := range pgms {
				log.Printf("* %s\n", pgm)
			}
		}
	}
}

func handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go func() {
		for s := range ch {
			if s != syscall.SIGURG {
				log.Printf("Received signal %s\n", s)
			}
			switch s {
			case syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				cancel()
				return
			}
		}
	}()
}

func buildReader(ctx context.Context) (r io.Reader, err error) {
	// Validate input
	if len(*inputPath) <= 0 {
		err = errors.New("use -i to indicate an input path")
		return
	}

	// Parse input
	var u *url.URL
	if u, err = url.Parse(*inputPath); err != nil {
		err = fmt.Errorf("astits: parsing input path failed: %w", err)
		return
	}

	// Switch on scheme
	switch u.Scheme {
	case "udp":
		// Resolve addr
		var addr *net.UDPAddr
		if addr, err = net.ResolveUDPAddr("udp", u.Host); err != nil {
			err = fmt.Errorf("astits: resolving udp addr %s failed: %w", u.Host, err)
			return
		}

		// Listen to multicast UDP
		var c *net.UDPConn
		if c, err = net.ListenMulticastUDP("udp", nil, addr); err != nil {
			err = fmt.Errorf("astits: listening on multicast udp addr %s failed: %w", u.Host, err)
			return
		}
		c.SetReadBuffer(4096)
		r = c
	default:
		// Open file
		var f *os.File
		if f, err = os.Open(*inputPath); err != nil {
			err = fmt.Errorf("astits: opening %s failed: %w", *inputPath, err)
			return
		}
		r = f
	}
	return
}

func packets(dmx *astits.Demuxer) (err error) {
	// Loop through packets
	var p *astits.Packet
	log.Println("Fetching packets...")
	for {
		// Get next packet
		if p, err = dmx.NextPacket(); err != nil {
			if err == astits.ErrNoMorePackets {
				break
			}
			err = fmt.Errorf("astits: getting next packet failed: %w", err)
			return
		}

		// Log packet
		log.Printf("PKT: %d\n", p.Header.PID)
		log.Printf("  Continuity Counter: %v\n", p.Header.ContinuityCounter)
		log.Printf("  Payload Unit Start Indicator: %v\n", p.Header.PayloadUnitStartIndicator)
		log.Printf("  Has Payload: %v\n", p.Header.HasPayload)
		log.Printf("  Has Adaptation Field: %v\n", p.Header.HasAdaptationField)
		log.Printf("  Transport Error Indicator: %v\n", p.Header.TransportErrorIndicator)
		log.Printf("  Transport Priority: %v\n", p.Header.TransportPriority)
		log.Printf("  Transport Scrambling Control: %v\n", p.Header.TransportScramblingControl)
		if p.Header.HasAdaptationField {
			log.Printf("  Adaptation Field: %+v\n", p.AdaptationField)
		}
	}
	return nil
}

func data(dmx *astits.Demuxer) (err error) {
	// Determine which data to log
	var logAll, logEIT, logNIT, logPAT, logPES, logPMT, logSDT, logTOT bool
	if _, ok := dataTypes.Map["all"]; ok {
		logAll = true
	}
	if _, ok := dataTypes.Map["eit"]; ok {
		logEIT = true
	}
	if _, ok := dataTypes.Map["nit"]; ok {
		logNIT = true
	}
	if _, ok := dataTypes.Map["pat"]; ok {
		logPAT = true
	}
	if _, ok := dataTypes.Map["pes"]; ok {
		logPES = true
	}
	if _, ok := dataTypes.Map["pmt"]; ok {
		logPMT = true
	}
	if _, ok := dataTypes.Map["sdt"]; ok {
		logSDT = true
	}
	if _, ok := dataTypes.Map["tot"]; ok {
		logTOT = true
	}

	// Loop through data
	var d *astits.DemuxerData
	log.Println("Fetching data...")
	for {
		// Get next data
		if d, err = dmx.NextData(); err != nil {
			if err == astits.ErrNoMorePackets {
				break
			}
			err = fmt.Errorf("astits: getting next data failed: %w", err)
			return
		}

		// Log data
		if d.EIT != nil && (logAll || logEIT) {
			log.Printf("EIT: %d\n", d.PID)
			log.Println(eventsToString(d.EIT.Events))
		} else if d.NIT != nil && (logAll || logNIT) {
			log.Printf("NIT: %d\n", d.PID)
		} else if d.PAT != nil && (logAll || logPAT) {
			log.Printf("PAT: %d\n", d.PID)
			log.Printf("  Transport Stream ID: %v\n", d.PAT.TransportStreamID)
			log.Println("  Programs:")
			for _, p := range d.PAT.Programs {
				log.Printf("    %+v\n", p)
			}
		} else if d.PES != nil && (logAll || logPES) {
			log.Printf("PES: %d\n", d.PID)
			log.Printf("  Stream ID: %v\n", d.PES.Header.StreamID)
			log.Printf("  Packet Length: %v\n", d.PES.Header.PacketLength)
			log.Printf("  Optional Header: %+v\n", d.PES.Header.OptionalHeader)
		} else if d.PMT != nil && (logAll || logPMT) {
			log.Printf("PMT: %d\n", d.PID)
			log.Printf("  ProgramNumber: %v\n", d.PMT.ProgramNumber)
			log.Printf("  PCR PID: %v\n", d.PMT.PCRPID)
			log.Println("  Elementary Streams:")
			for _, s := range d.PMT.ElementaryStreams {
				log.Printf("    %+v\n", s)
			}
			log.Println("  Program Descriptors:")
			for _, d := range d.PMT.ProgramDescriptors {
				log.Printf("    %+v\n", d)
			}
		} else if d.SDT != nil && (logAll || logSDT) {
			log.Printf("SDT: %d\n", d.PID)
		} else if d.TOT != nil && (logAll || logTOT) {
			log.Printf("TOT: %d\n", d.PID)
		}
	}
	return
}

func programs(dmx *astits.Demuxer) (o []*Program, err error) {
	// Loop through data
	var d *astits.DemuxerData
	var pgmsToProcess = make(map[uint16]bool)
	var pgms = make(map[uint16]*Program)
	log.Println("Fetching data...")
	for {
		// Get next data
		if d, err = dmx.NextData(); err != nil {
			if err == astits.ErrNoMorePackets {
				err = nil
				break
			}
			err = fmt.Errorf("astits: getting next data failed: %w", err)
			return
		}

		// Check data
		if d.PAT != nil {
			// Build programs list
			for _, p := range d.PAT.Programs {
				// Program number 0 is reserved to NIT
				if p.ProgramNumber > 0 {
					// Program has not already been added
					if _, ok := pgms[p.ProgramNumber]; !ok {
						pgmsToProcess[p.ProgramNumber] = true
						pgms[p.ProgramNumber] = newProgram(p.ProgramNumber, p.ProgramMapID)
					}
				}
			}
		} else if d.PMT != nil {
			// Program has already been processed
			if _, ok := pgmsToProcess[d.PMT.ProgramNumber]; !ok {
				continue
			}

			// Update program
			for _, dsc := range d.PMT.ProgramDescriptors {
				pgms[d.PMT.ProgramNumber].Descriptors = append(pgms[d.PMT.ProgramNumber].Descriptors, descriptorToString(dsc))
			}

			// Add elementary streams
			for _, es := range d.PMT.ElementaryStreams {
				var s = newStream(es.ElementaryPID, es.StreamType)
				for _, d := range es.ElementaryStreamDescriptors {
					s.Descriptors = append(s.Descriptors, descriptorToString(d))
				}
				pgms[d.PMT.ProgramNumber].Streams = append(pgms[d.PMT.ProgramNumber].Streams, s)
			}

			// Update list of programs to process
			delete(pgmsToProcess, d.PMT.ProgramNumber)

			// All PMTs have been processed
			if len(pgmsToProcess) == 0 {
				break
			}
		}
	}

	// Build final data
	for _, p := range pgms {
		o = append(o, p)
	}
	return
}

// Program represents a program
type Program struct {
	Descriptors []string  `json:"descriptors,omitempty"`
	ID          uint16    `json:"id,omitempty"`
	MapID       uint16    `json:"map_id,omitempty"`
	Streams     []*Stream `json:"streams,omitempty"`
}

// Stream represents a stream
type Stream struct {
	Descriptors []string          `json:"descriptors,omitempty"`
	ID          uint16            `json:"id,omitempty"`
	Type        astits.StreamType `json:"type,omitempty"`
}

func newProgram(id, mapID uint16) *Program {
	return &Program{
		ID:    id,
		MapID: mapID,
	}
}

func newStream(id uint16, _type astits.StreamType) *Stream {
	return &Stream{
		ID:   id,
		Type: _type,
	}
}

// String implements the Stringer interface
func (p Program) String() (o string) {
	o = fmt.Sprintf("[%d] - Map ID: %d", p.ID, p.MapID)
	for _, d := range p.Descriptors {
		o += fmt.Sprintf(" - %s", d)
	}
	for _, s := range p.Streams {
		o += fmt.Sprintf("\n  * %s", s.String())
	}
	return
}

// String implements the Stringer interface
func (s Stream) String() (o string) {
	// Get type
	var t = fmt.Sprintf("unlisted stream type %d", s.Type)
	switch s.Type {
	case astits.StreamTypeMPEG1Audio:
		t = "MPEG-1 audio"
	case astits.StreamTypeMPEG2HalvedSampleRateAudio:
		t = "MPEG-2 halved sample rate audio"
	case astits.StreamTypeMPEG2PacketizedData:
		t = "DVB subtitles/VBI or AC-3"
	case astits.StreamTypeADTS:
		t = "ADTS"
	case astits.StreamTypeH264Video:
		t = "H264 video"
	case astits.StreamTypeH265Video:
		t = "H265 video"
	}

	// Output
	o = fmt.Sprintf("[%d] - Type: %s", s.ID, t)
	for _, d := range s.Descriptors {
		o += fmt.Sprintf(" - %s", d)
	}
	return
}

func eventsToString(es []*astits.EITDataEvent) string {
	var os []string
	for idx, e := range es {
		os = append(os, eventToString(idx, e))
	}
	return strings.Join(os, "\n")
}

func eventToString(idx int, e *astits.EITDataEvent) (s string) {
	s += fmt.Sprintf("- #%d | id: %d | start: %s | duration: %s | status: %s\n", idx+1, e.EventID, e.StartTime.Format("15:04:05"), e.Duration, runningStatusToString(e.RunningStatus))
	var os []string
	for _, d := range e.Descriptors {
		os = append(os, "  - "+descriptorToString(d))
	}
	return s + strings.Join(os, "\n")
}

func runningStatusToString(s uint8) string {
	switch s {
	case astits.RunningStatusNotRunning:
		return "not running"
	case astits.RunningStatusPausing:
		return "pausing"
	case astits.RunningStatusRunning:
		return "running"
	}
	return "unknown"
}

func descriptorToString(d *astits.Descriptor) string {
	switch d.Tag {
	case astits.DescriptorTagAC3:
		return fmt.Sprintf("[AC3] ac3 asvc: %d | bsid: %d | component type: %d | mainid: %d | info: %s", d.AC3.ASVC, d.AC3.BSID, d.AC3.ComponentType, d.AC3.MainID, d.AC3.AdditionalInfo)
	case astits.DescriptorTagComponent:
		return fmt.Sprintf("[Component] language: %s | text: %s | component tag: %d | component type: %d | stream content: %d | stream content ext: %d", d.Component.ISO639LanguageCode, d.Component.Text, d.Component.ComponentTag, d.Component.ComponentType, d.Component.StreamContent, d.Component.StreamContentExt)
	case astits.DescriptorTagContent:
		var os []string
		for _, i := range d.Content.Items {
			os = append(os, fmt.Sprintf("content nibble 1: %d | content nibble 2: %d | user byte: %d", i.ContentNibbleLevel1, i.ContentNibbleLevel2, i.UserByte))
		}
		return "[Content] " + strings.Join(os, " - ")
	case astits.DescriptorTagExtendedEvent:
		s := fmt.Sprintf("[Extended event] language: %s | text: %s", d.ExtendedEvent.ISO639LanguageCode, d.ExtendedEvent.Text)
		for _, i := range d.ExtendedEvent.Items {
			s += fmt.Sprintf(" | %s: %s", i.Description, i.Content)
		}
		return s
	case astits.DescriptorTagISO639LanguageAndAudioType:
		return fmt.Sprintf("[ISO639 language and audio type] language: %s | audio type: %d", d.ISO639LanguageAndAudioType.Language, d.ISO639LanguageAndAudioType.Type)
	case astits.DescriptorTagMaximumBitrate:
		return fmt.Sprintf("[Maximum bitrate] maximum bitrate: %d", d.MaximumBitrate.Bitrate)
	case astits.DescriptorTagNetworkName:
		return fmt.Sprintf("[Network name] network name: %s", d.NetworkName.Name)
	case astits.DescriptorTagParentalRating:
		var os []string
		for _, i := range d.ParentalRating.Items {
			os = append(os, fmt.Sprintf("country: %s | rating: %d | minimum age: %d", i.CountryCode, i.Rating, i.MinimumAge()))
		}
		return "[Parental rating] " + strings.Join(os, " - ")
	case astits.DescriptorTagPrivateDataSpecifier:
		return fmt.Sprintf("[Private data specifier] specifier: %d", d.PrivateDataSpecifier.Specifier)
	case astits.DescriptorTagService:
		return fmt.Sprintf("[Service] service %s | provider: %s", d.Service.Name, d.Service.Provider)
	case astits.DescriptorTagShortEvent:
		return fmt.Sprintf("[Short event] language: %s | name: %s | text: %s", d.ShortEvent.Language, d.ShortEvent.EventName, d.ShortEvent.Text)
	case astits.DescriptorTagStreamIdentifier:
		return fmt.Sprintf("[Stream identifier] stream identifier component tag: %d", d.StreamIdentifier.ComponentTag)
	case astits.DescriptorTagSubtitling:
		var os []string
		for _, i := range d.Subtitling.Items {
			os = append(os, fmt.Sprintf("subtitling composition page: %d | ancillary page %d: %s", i.CompositionPageID, i.AncillaryPageID, i.Language))
		}
		return "[Subtitling] " + strings.Join(os, " - ")
	case astits.DescriptorTagTeletext:
		var os []string
		for _, t := range d.Teletext.Items {
			os = append(os, fmt.Sprintf("Teletext page %01d%02d: %s", t.Magazine, t.Page, t.Language))
		}
		return "[Teletext] " + strings.Join(os, " - ")
	}
	return fmt.Sprintf("unlisted descriptor tag 0x%x", d.Tag)
}
