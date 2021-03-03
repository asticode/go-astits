package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astits"
	"log"
	"os"
	"path"
	"time"
)

const (
	ioBufSize = 10 * 1024 * 1024
)

type muxerOut struct {
	f *os.File
	w *bufio.Writer
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Split TS file into multiple files each holding one elementary stream")
		fmt.Fprintf(flag.CommandLine.Output(), "%s [FLAGS] INPUT_FILE:\n", os.Args[0])
		flag.PrintDefaults()
	}
	outDir := flag.String("o", "out", "Output dir, 'out' by default")
	inputFile := astikit.FlagCmd()
	flag.Parse()

	infile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer infile.Close()

	_, err = os.Stat(*outDir)
	if !os.IsNotExist(err) {
		log.Fatalf("can't write to `%s': already exists", *outDir)
	}

	if err = os.MkdirAll(*outDir, os.ModePerm); err != nil {
		log.Fatalf("%v", err)
	}

	demux := astits.New(
		context.Background(),
		bufio.NewReaderSize(infile, ioBufSize),
	)

	var pat *astits.PATData
	// key is program number
	pmts := map[uint16]*astits.PMTData{}
	gotAllPMTs := false
	// key is pid
	muxers := map[uint16]*astits.Muxer{}
	outfiles := map[uint16]muxerOut{}

	pmtsPrinted := false

	lastRateOutput := time.Now()
	bytesWritten := 0

	for {
		d, err := demux.NextData()
		if err != nil {
			if err == astits.ErrNoMorePackets {
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
				_, ok := pmts[p.ProgramNumber]
				if !ok {
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
					_, ok := muxers[es.ElementaryPID]
					if ok {
						continue
					}

					esFilename := path.Join(*outDir, fmt.Sprintf("%d.ts", es.ElementaryPID))
					outfile, err := os.Create(esFilename)
					if err != nil {
						log.Fatalf("%v", err)
					}

					bufWriter := bufio.NewWriterSize(outfile, ioBufSize)
					mux := astits.NewMuxer(context.Background(), bufWriter)
					err = mux.AddElementaryStream(*es, true)
					if err != nil {
						log.Fatalf("%v", err)
					}

					outfiles[es.ElementaryPID] = muxerOut{
						f: outfile,
						w: bufWriter,
					}
					muxers[es.ElementaryPID] = mux

					if !pmtsPrinted {
						log.Printf("\t\tES PID %d type %s",
							es.ElementaryPID, astits.StreamTypeString(es.StreamType),
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

		//log.Printf("PID %d Payload start, bytes %s", pid, hex.EncodeToString(d.PES.Data[:32]))

		af := d.FirstPacket.AdaptationField

		if af != nil && af.HasPCR {
			af.HasPCR = false
		}

		var pcr *astits.ClockReference
		if d.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorBothPresent {
			pcr = d.PES.Header.OptionalHeader.DTS
		} else if d.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorOnlyPTS {
			pcr = d.PES.Header.OptionalHeader.PTS
		}

		if pcr != nil {
			if af == nil {
				af = &astits.PacketAdaptationField{}
			}
			af.HasPCR = true
			af.PCR = pcr
		}

		n, err := mux.WritePayload(pid, af, d.PES.Header, d.PES.Data)
		if err != nil {
			log.Fatalf("%v", err)
		}

		bytesWritten += n
	}

	timeDiff := time.Since(lastRateOutput)
	lastRateOutput = time.Now()
	log.Printf("%d bytes written at rate %.02f mb/s", bytesWritten, (float64(bytesWritten)/1024.0/1024.0)/timeDiff.Seconds())

	for _, f := range outfiles {
		if err = f.w.Flush(); err != nil {
			log.Printf("Error flushing %s: %v", f.f.Name(), err)
		}
		if err = f.f.Close(); err != nil {
			log.Printf("Error closing %s: %v", f.f.Name(), err)
		}
	}

	log.Printf("Done")
}
