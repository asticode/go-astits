package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/pkg/profile"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astits"
)

const (
	ioBufSize = 10 * 1024 * 1024
)

type muxerOut struct {
	name   string
	closer io.Closer
	*bufio.Writer
}

func newMuxerOut(name string, discard bool) (*muxerOut, error) {
	var w io.Writer
	var c io.Closer
	if !discard {
		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		name = f.Name()
		c = f
		w = f
	} else {
		name += " --discard--"
		w = io.Discard
	}
	return &muxerOut{name, c, bufio.NewWriterSize(w, ioBufSize)}, nil
}

func (m *muxerOut) Close() error {
	if err := m.Flush(); err != nil {
		log.Printf("Error flushing %s: %v", m.name, err)
	}
	if m.closer != nil {
		if err := m.closer.Close(); err != nil {
			return fmt.Errorf("error closing %s: %w", m.name, err)
		}
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Split TS file into multiple files each holding one elementary stream")
		fmt.Fprintf(flag.CommandLine.Output(), "%s INPUT_FILE [FLAGS]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	memoryProfiling := flag.Bool("mp", false, "if yes, memory profiling is enabled")
	cpuProfiling := flag.Bool("cp", false, "if yes, cpu profiling is enabled")
	discard := flag.Bool("discard", false, "if yes, output will be passed to discard (profiling/debug only)")
	outDir := flag.String("o", "out", "Output dir, 'out' by default")
	inputFile := astikit.FlagCmd()
	flag.Parse()

	if *cpuProfiling {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	} else if *memoryProfiling {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	}

	infile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer infile.Close()

	if !*discard {
		if _, err = os.Stat(*outDir); !os.IsNotExist(err) {
			log.Fatalf("can't write to '%s': already exists", *outDir)
		}

		if err = os.MkdirAll(*outDir, os.ModePerm); err != nil {
			log.Fatalf("%v", err)
		}
	}

	demux := astits.NewDemuxer(
		context.Background(),
		bufio.NewReaderSize(infile, ioBufSize),
	)

	var pat *astits.PATData
	// key is program number
	pmts := map[uint16]*astits.PMTData{}
	gotAllPMTs := false
	// key is pid
	muxers := map[uint16]*astits.Muxer{}

	pmtsPrinted := false

	timeStarted := time.Now()
	bytesWritten := 0

	var d *astits.DemuxerData
	for {
		if d, err = demux.NextData(); err != nil {
			if errors.Is(err, astits.ErrNoMorePackets) {
				break
			}
			log.Fatalf("%v", err)
		}

		if d.PAT != nil {
			pat = d.PAT
			gotAllPMTs = false
			continue
		}

		if d.PMT != nil {
			pmts[d.PMT.ProgramNumber] = d.PMT

			gotAllPMTs = true
			for _, p := range pat.Programs {
				if _, ok := pmts[p.ProgramNumber]; !ok {
					gotAllPMTs = false
					break
				}
			}

			if !gotAllPMTs {
				continue
			}

			if !pmtsPrinted {
				log.Printf("Got all PMTs")
			}
			for _, pmt := range pmts {
				if !pmtsPrinted {
					log.Printf("\tProgram %d PCR PID %d", pmt.ProgramNumber, pmt.PCRPID)
				}
				for _, es := range pmt.ElementaryStreams {
					if _, ok := muxers[es.ElementaryPID]; ok {
						continue
					}

					esFilename := path.Join(*outDir, fmt.Sprintf("%d.ts", es.ElementaryPID))
					var outWriter *muxerOut
					if outWriter, err = newMuxerOut(esFilename, *discard); err != nil {
						log.Fatalf("%v", err)
					}
					defer func() {
						if err = outWriter.Close(); err != nil {
							log.Print(err)
						}
					}()

					mux := astits.NewMuxer(context.Background(), outWriter)
					if err = mux.AddElementaryStream(*es); err != nil {
						log.Fatalf("%v", err)
					}
					mux.SetPCRPID(es.ElementaryPID)
					muxers[es.ElementaryPID] = mux

					if !pmtsPrinted {
						log.Printf("\t\tES PID %d type %s",
							es.ElementaryPID, es.StreamType.String(),
						)
					}
				}
			}

			pmtsPrinted = true
			continue
		}

		if !gotAllPMTs {
			continue
		}

		if d.PES == nil {
			continue
		}

		pid := d.FirstPacket.Header.PID
		mux, ok := muxers[pid]
		if !ok {
			log.Printf("Got payload for unknown PID %d", pid)
			continue
		}

		af := d.FirstPacket.AdaptationField

		if af != nil && af.HasPCR {
			af.HasPCR = false
		}

		var pcr *astits.ClockReference
		switch d.PES.Header.OptionalHeader.PTSDTSIndicator {
		case astits.PTSDTSIndicatorOnlyPTS:
			pcr = d.PES.Header.OptionalHeader.PTS
		case astits.PTSDTSIndicatorBothPresent:
			pcr = d.PES.Header.OptionalHeader.DTS
		}

		if pcr != nil {
			if af == nil {
				af = &astits.PacketAdaptationField{}
			}
			af.HasPCR = true
			af.PCR = pcr
		}

		var written int
		if written, err = mux.WriteData(&astits.MuxerData{
			PID:             pid,
			AdaptationField: af,
			PES:             d.PES,
		}); err != nil {
			log.Fatalf("%v", err)
		}

		bytesWritten += written
	}

	timeDiff := time.Since(timeStarted)
	log.Printf("%d bytes written at rate %.02f mb/s", bytesWritten, (float64(bytesWritten)/1024.0/1024.0)/timeDiff.Seconds())

	log.Printf("Done")
}
