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

	d := astits.New(context.Background(), bufio.NewReaderSize(infile, ioBufSize))

	var pat *astits.PATData
	// key is program number
	pmts := map[uint16]*astits.PMTData{}
	gotAllPMTs := false
	// key is pid
	muxers := map[uint16]*astits.Muxer{}
	outfiles := map[uint16]*os.File{}

	pmtsPrinted := false

	lastRateOutput := time.Now()
	bytesWritten := 0

	for {
		d, err := d.NextData()
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
					defer outfile.Close() // ouch!

					mux := astits.NewMuxer(context.Background(), bufio.NewWriterSize(outfile, ioBufSize))
					err = mux.AddElementaryStream(*es, true)
					if err != nil {
						log.Fatalf("%v", err)
					}

					outfiles[es.ElementaryPID] = outfile
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

		n, err := mux.WritePayload(pid, d.FirstPacket.AdaptationField, d.PES.Header, d.PES.Data)
		if err != nil {
			log.Fatalf("%v", err)
		}

		bytesWritten += n
		timeDiff := time.Since(lastRateOutput)
		if timeDiff > 5*time.Second {
			lastRateOutput = time.Now()
			log.Printf("%.02f mb/s", (float64(bytesWritten)/1024.0/1024.0)/timeDiff.Seconds())
			bytesWritten = 0
		}
	}

	log.Printf("Done")
}
