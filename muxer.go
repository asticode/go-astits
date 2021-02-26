package astits

import (
	"bytes"
	"context"
	"errors"
	"github.com/asticode/go-astikit"
	"io"
)

const (
	StartPID           uint16 = 0x0100
	PMTStartPID        uint16 = 0x1000
	ProgramNumberStart uint16 = 1
)

var (
	MuxerErrorPIDNotFound      = errors.New("PID not found")
	MuxerErrorPIDAlreadyExists = errors.New("PID already exists")
	MuxerErrorPCRPIDInvalid    = errors.New("PCR PID invalid")
)

type Muxer struct {
	ctx context.Context
	w   io.Writer

	packetSize int

	pm             programMap // pid -> programNumber
	pmt            PMTData
	nextPID        uint16
	nextPATVersion psiVersionCounter
	nextPMTVersion psiVersionCounter

	patBytes bytes.Buffer
	pmtBytes bytes.Buffer

	buf bytes.Buffer
}

func NewMuxer(ctx context.Context, w io.Writer, opts ...func(*Muxer)) *Muxer {
	m := &Muxer{
		ctx: ctx,
		w:   w,

		packetSize: MpegTsPacketSize, // no 192-byte packet support yet
		pm:         newProgramMap(),
		pmt: PMTData{
			ElementaryStreams: []*PMTElementaryStream{},
			ProgramNumber:     ProgramNumberStart,
		},
	}

	// TODO multiple programs support
	m.pm.set(PMTStartPID, ProgramNumberStart)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// if es.ElementaryPID is zero, it will be generated automatically
func (m *Muxer) AddElementaryStream(es PMTElementaryStream, isPCRPid bool) error {
	if es.ElementaryPID != 0 {
		for _, oes := range m.pmt.ElementaryStreams {
			if oes.ElementaryPID == es.ElementaryPID {
				return MuxerErrorPIDAlreadyExists
			}
		}
	} else {
		es.ElementaryPID = m.nextPID
		m.nextPID++
	}

	m.pmt.ElementaryStreams = append(m.pmt.ElementaryStreams, &es)
	if isPCRPid {
		m.pmt.PCRPID = es.ElementaryPID
	}

	m.pmtBytes.Reset() // invalidate pmt cache
	return nil
}

func (m *Muxer) RemoveElementaryStream(pid uint16) error {
	foundIdx := -1
	for i, oes := range m.pmt.ElementaryStreams {
		if oes.ElementaryPID == pid {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		return MuxerErrorPIDNotFound
	}

	m.pmt.ElementaryStreams = append(m.pmt.ElementaryStreams[:foundIdx], m.pmt.ElementaryStreams[foundIdx+1:]...)
	return nil
}

func (m *Muxer) WriteTables() (int, error) {
	bytesWritten := 0

	if m.patBytes.Len() != m.packetSize {
		if err := m.generatePAT(); err != nil {
			return bytesWritten, err
		}
	}

	if m.pmtBytes.Len() != m.packetSize {
		if err := m.generatePMT(); err != nil {
			return bytesWritten, err
		}
	}

	n, err := m.w.Write(m.patBytes.Bytes())
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	n, err = m.w.Write(m.pmtBytes.Bytes())
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	return bytesWritten, nil
}

func (m *Muxer) generatePAT() error {
	d := m.pm.toPATData()
	syntax := &PSISectionSyntax{
		Data: &PSISectionSyntaxData{PAT: d},
		Header: &PSISectionSyntaxHeader{
			CurrentNextIndicator: true,
			// TODO support for PAT tables longer than 1 TS packet
			//LastSectionNumber:    0,
			//SectionNumber:        0,
			TableIDExtension: d.TransportStreamID,
			VersionNumber:    uint8(m.nextPATVersion),
		},
	}
	section := PSISection{
		Header: &PSISectionHeader{
			SectionLength:          calcPATSectionLength(d),
			SectionSyntaxIndicator: true,
			TableID:                int(d.TransportStreamID),
		},
		Syntax: syntax,
	}
	psiData := PSIData{
		Sections: []*PSISection{&section},
	}

	m.buf.Reset()
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &m.buf})
	if _, err := writePSIData(w, &psiData); err != nil {
		return err
	}

	m.patBytes.Reset()
	wPacket := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &m.patBytes})

	pkt := Packet{
		Header: &PacketHeader{
			HasPayload:                true,
			PayloadUnitStartIndicator: true,
			PID:                       PIDPAT,
		},
		Payload: m.buf.Bytes(),
	}
	if _, err := writePacket(wPacket, &pkt, m.packetSize); err != nil {
		// FIXME save old PAT and rollback to it here maybe?
		return err
	}

	m.nextPATVersion.increment()
	return nil
}

func (m *Muxer) generatePMT() error {
	hasPCRPID := false
	for _, es := range m.pmt.ElementaryStreams {
		if es.ElementaryPID == m.pmt.PCRPID {
			hasPCRPID = true
			break
		}
	}
	if !hasPCRPID {
		return MuxerErrorPCRPIDInvalid
	}

	syntax := &PSISectionSyntax{
		Data: &PSISectionSyntaxData{PMT: &m.pmt},
		Header: &PSISectionSyntaxHeader{
			CurrentNextIndicator: true,
			// TODO support for PMT tables longer than 1 TS packet
			//LastSectionNumber:    0,
			//SectionNumber:        0,
			TableIDExtension: m.pmt.ProgramNumber,
			VersionNumber:    uint8(m.nextPMTVersion),
		},
	}
	section := PSISection{
		Header: &PSISectionHeader{
			SectionLength:          calcPMTSectionLength(&m.pmt),
			SectionSyntaxIndicator: true,
			TableID:                PSITableTypeIdPMT,
		},
		Syntax: syntax,
	}
	psiData := PSIData{
		Sections: []*PSISection{&section},
	}

	m.buf.Reset()
	w := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &m.buf})
	if _, err := writePSIData(w, &psiData); err != nil {
		return err
	}

	m.pmtBytes.Reset()
	wPacket := astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: &m.pmtBytes})

	pkt := Packet{
		Header: &PacketHeader{
			HasPayload:                true,
			PayloadUnitStartIndicator: true,
			PID:                       PMTStartPID, // FIXME
		},
		Payload: m.buf.Bytes(),
	}
	if _, err := writePacket(wPacket, &pkt, m.packetSize); err != nil {
		// FIXME save old PMT and rollback to it here maybe?
		return err
	}

	m.nextPMTVersion.increment()
	return nil
}
