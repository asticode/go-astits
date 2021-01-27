package astits

import (
	"context"
	"github.com/asticode/go-astikit"
	"io"
)

type Muxer struct {
	ctx context.Context
	w   *astikit.BitsWriter

	packetSize int
}

func NewMuxer(ctx context.Context, w io.Writer, opts ...func(*Muxer)) *Muxer {
	m := &Muxer{
		ctx: ctx,
		w:   astikit.NewBitsWriter(astikit.BitsWriterOptions{Writer: w}),

		packetSize: MpegTsPacketSize, // no 192-byte packet support yet
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Muxer) WritePacket(p *Packet) error {
	if err := m.ctx.Err(); err != nil {
		return err
	}
	// TODO
	return nil
}
