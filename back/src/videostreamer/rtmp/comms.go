package rtmp

import (
	"net"
	"videostreamer/core"
)

func NewChunkedComms(registry core.Registry, conn net.Conn) *ChunkedComms {
	return &ChunkedComms{
		Conn: conn,
		Registry: registry,
		StreamIn: make(map[uint16]*ChunkedInputStream),
		StreamOut: make(map[uint16]*ChunkedOutputStream),
		ChunkIn: 128,
		ChunkOut: 128,
	}
}

func (comms *ChunkedComms) Inbound(chunk uint16) *ChunkedInputStream {
	s, ok := comms.StreamIn[chunk]
	if !ok {
		s = &ChunkedInputStream{
			Last: 0,
		}
		comms.StreamIn[chunk] = s
	}
	return s
}

func (comms *ChunkedComms) Outbound(chunk uint16) *ChunkedOutputStream {
	s, ok := comms.StreamOut[chunk]
	if !ok {
		s = &ChunkedOutputStream{}
		comms.StreamOut[chunk] = s
	}
	return s
}