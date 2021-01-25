package astits

import (
	"context"
	"io"
)

type Muxer struct {
	ctx context.Context
	w   io.Writer
}

func NewMuxer(ctx context.Context, w io.Writer, opts ...func(*Muxer)) *Muxer {
	m := &Muxer{
		ctx: ctx,
		w:   w,
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

}
