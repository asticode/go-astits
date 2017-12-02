package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
	"github.com/asticode/go-astits"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
)

// Flags
var (
	ctx, cancel     = context.WithCancel(context.Background())
	cpuProfiling    = flag.Bool("cp", false, "if yes, cpu profiling is enabled")
	dataTypes       = astiflag.NewStringsMap()
	format          = flag.String("f", "", "the format")
	inputPath       = flag.String("i", "", "the input path")
	memoryProfiling = flag.Bool("mp", false, "if yes, memory profiling is enabled")
)

func main() {
	// Init
	flag.Var(dataTypes, "d", "the datatypes whitelist")
	var s = astiflag.Subcommand()
	flag.Parse()
	astilog.FlagInit()

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
	if r, err = buildReader(); err != nil {
		astilog.Error(errors.Wrap(err, "astits: parsing input failed"))
		return
	}

	// Make sure the reader is closed properly
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}

	// Create the demuxer
	var dmx = astits.New(ctx, r)

	// Switch on subcommand
	switch s {
	case "data":
		// Fetch data
		if err = data(dmx); err != nil {
			astilog.Error(errors.Wrap(err, "astits: fetching data failed"))
			return
		}
	default:
		// Fetch the programs
		var pgms []*Program
		if pgms, err = programs(dmx); err != nil {
			astilog.Error(errors.Wrap(err, "astits: fetching programs failed"))
			return
		}

		// Print
		switch *format {
		case "json":
			var e = json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			if err = e.Encode(pgms); err != nil {
				astilog.Error(errors.Wrap(err, "astits: json encoding to stdout failed"))
				return
			}
		default:
			fmt.Println("Programs are:")
			for _, pgm := range pgms {
				fmt.Printf("* %s\n", pgm)
			}
		}
	}
}

func handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go func() {
		for s := range ch {
			astilog.Debugf("Received signal %s", s)
			switch s {
			case syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				cancel()
			}
			return
		}
	}()
}

func buildReader() (r io.Reader, err error) {
	// Validate input
	if len(*inputPath) <= 0 {
		err = errors.New("Use -i to indicate an input path")
		return
	}

	// Parse input
	var u *url.URL
	if u, err = url.Parse(*inputPath); err != nil {
		err = errors.Wrap(err, "astits: parsing input path failed")
		return
	}

	// Switch on scheme
	switch u.Scheme {
	case "udp":
		// Resolve addr
		var addr *net.UDPAddr
		if addr, err = net.ResolveUDPAddr("udp", u.Host); err != nil {
			err = errors.Wrapf(err, "astits: resolving udp addr %s failed", u.Host)
			return
		}

		// Listen to multicast UDP
		var c *net.UDPConn
		if c, err = net.ListenMulticastUDP("udp", nil, addr); err != nil {
			err = errors.Wrapf(err, "astits: listening on multicast udp addr %s failed", u.Host)
			return
		}
		c.SetReadBuffer(1000)
		r = c
	default:
		// Open file
		var f *os.File
		if f, err = os.Open(*inputPath); err != nil {
			err = errors.Wrapf(err, "astits: opening %s failed", *inputPath)
			return
		}
		r = f
	}
	return
}

func data(dmx *astits.Demuxer) (err error) {
	// Determine which data to log
	var logAll, logEIT, logNIT, logPAT, logPES, logPMT, logSDT, logTOT bool
	if _, ok := dataTypes["all"]; ok {
		logAll = true
	}
	if _, ok := dataTypes["eit"]; ok {
		logEIT = true
	}
	if _, ok := dataTypes["nit"]; ok {
		logNIT = true
	}
	if _, ok := dataTypes["pat"]; ok {
		logPAT = true
	}
	if _, ok := dataTypes["pes"]; ok {
		logPES = true
	}
	if _, ok := dataTypes["pmt"]; ok {
		logPMT = true
	}
	if _, ok := dataTypes["sdt"]; ok {
		logSDT = true
	}
	if _, ok := dataTypes["tot"]; ok {
		logTOT = true
	}

	// Loop through data
	var d *astits.Data
	astilog.Debug("Fetching data...")
	for {
		// Get next data
		if d, err = dmx.NextData(); err != nil {
			if err == astits.ErrNoMorePackets {
				break
			}
			err = errors.Wrap(err, "astits: getting nex data failed")
			return
		}

		// Log data
		if d.EIT != nil && (logAll || logEIT) {
			astilog.Infof("EIT: %d", d.PID)
		} else if d.NIT != nil && (logAll || logNIT) {
			astilog.Infof("NIT: %d", d.PID)
		} else if d.PAT != nil && (logAll || logPAT) {
			astilog.Infof("PAT: %d", d.PID)
		} else if d.PES != nil && (logAll || logPES) {
			astilog.Infof("PES: %d", d.PID)
		} else if d.PMT != nil && (logAll || logPMT) {
			astilog.Infof("PMT: %d", d.PID)
		} else if d.SDT != nil && (logAll || logSDT) {
			astilog.Infof("SDT: %d", d.PID)
		} else if d.TOT != nil && (logAll || logTOT) {
			astilog.Infof("TOT: %d", d.PID)
		}
	}
	return
}

func programs(dmx *astits.Demuxer) (o []*Program, err error) {
	// Loop through data
	var d *astits.Data
	var pgmsToProcess = make(map[uint16]bool)
	var pgms = make(map[uint16]*Program)
	astilog.Debug("Fetching data...")
	for {
		// Get next data
		if d, err = dmx.NextData(); err != nil {
			if err == astits.ErrNoMorePackets {
				var pgmsNotProcessed []string
				for n := range pgms {
					pgmsNotProcessed = append(pgmsNotProcessed, strconv.Itoa(int(n)))
				}
				err = fmt.Errorf("astits: no PMT found for program(s) %s", strings.Join(pgmsNotProcessed, ", "))
			} else {
				err = errors.Wrap(err, "astits: getting next data failed")
			}
			return
		}

		// Check data
		if d.PAT != nil && len(pgms) == 0 {
			// Build programs list
			for _, p := range d.PAT.Programs {
				// Program number 0 is reserved to NIT
				if p.ProgramNumber > 0 {
					pgmsToProcess[p.ProgramNumber] = true
					pgms[p.ProgramNumber] = newProgram(p.ProgramNumber, p.ProgramMapID)
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
	Descriptors []string `json:"descriptors,omitempty"`
	ID          uint16   `json:"id,omitempty"`
	Type        uint8    `json:"type,omitempty"`
}

func newProgram(id, mapID uint16) *Program {
	return &Program{
		ID:    id,
		MapID: mapID,
	}
}

func newStream(id uint16, _type uint8) *Stream {
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
	case astits.StreamTypeLowerBitrateVideo:
		t = "Lower bitrate video"
	case astits.StreamTypeMPEG1Audio:
		t = "MPEG-1 audio"
	case astits.StreamTypeMPEG2HalvedSampleRateAudio:
		t = "MPEG-2 halved sample rate audio"
	case astits.StreamTypeMPEG2PacketizedData:
		t = "DVB subtitles/VBI or AC-3"
	}

	// Output
	o = fmt.Sprintf("[%d] - Type: %s", s.ID, t)
	for _, d := range s.Descriptors {
		o += fmt.Sprintf(" - %s", d)
	}
	return
}

func descriptorToString(d *astits.Descriptor) string {
	switch d.Tag {
	case astits.DescriptorTagAC3:
		return fmt.Sprintf("AC3 asvc: %d | bsid: %d | component type: %d | mainid: %d | info: %s", d.AC3.ASVC, d.AC3.BSID, d.AC3.ComponentType, d.AC3.MainID, d.AC3.AdditionalInfo)
	case astits.DescriptorTagISO639LanguageAndAudioType:
		return fmt.Sprintf("Language %s and audio type %d", d.ISO639LanguageAndAudioType.Language, d.ISO639LanguageAndAudioType.Type)
	case astits.DescriptorTagMaximumBitrate:
		return fmt.Sprintf("Maximum bitrate: %d", d.MaximumBitrate.Bitrate)
	case astits.DescriptorTagNetworkName:
		return fmt.Sprintf("Network name: %s", d.NetworkName.Name)
	case astits.DescriptorTagService:
		return fmt.Sprintf("Service %s from provider %s", d.Service.Name, d.Service.Provider)
	case astits.DescriptorTagShortEvent:
		return fmt.Sprintf("Short event %s for language %s with text %s", d.ShortEvent.EventName, d.ShortEvent.Language, d.ShortEvent.Text)
	case astits.DescriptorTagStreamIdentifier:
		return fmt.Sprintf("Stream identifier component tag: %d", d.StreamIdentifier.ComponentTag)
	case astits.DescriptorTagSubtitling:
		var os []string
		for _, i := range d.Subtitling.Items {
			os = append(os, fmt.Sprintf("Subtitling composition page %d and ancillary page %d: %s", i.CompositionPageID, i.AncillaryPageID, i.Language))
		}
		return strings.Join(os, " - ")
	case astits.DescriptorTagTeletext:
		var os []string
		for _, t := range d.Teletext.Items {
			os = append(os, fmt.Sprintf("Teletext page %01d%02d: %s", t.Magazine, t.Page, t.Language))
		}
		return strings.Join(os, " - ")
	}
	return fmt.Sprintf("unlisted descriptor tag 0x%x", d.Tag)
}
