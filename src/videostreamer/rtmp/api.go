package rtmp

import (
	"net"
	"io"
)

const (
	BASIC_TYPE_FULL   = 0x00
	BASIC_TYPE_MEDIUM = 0x01
	BASIC_TYPE_SHORT  = 0x02
	BASIC_TYPE_NONE   = 0x03
)

const (
	MESSAGE_TYPE_SET_CHUNK_SIZE =  1
	MESSAGE_TYPE_ABORT          =  2
	MESSAGE_TYPE_ACK            =  3
	MESSAGE_TYPE_USER           =  4
	MESSAGE_TYPE_WINACK         =  5
	MESSAGE_TYPE_SET_PEER_BAND  =  6
	MESSAGE_TYPE_EDGE           =  7 /* ? */
	MESSAGE_TYPE_AUDIO          =  8
	MESSAGE_TYPE_VIDEO          =  9
	MESSAGE_TYPE_AMF3_META      = 15
	MESSAGE_TYPE_AMF3_SHARED    = 16
	MESSAGE_TYPE_AMF3_CMD       = 17
	MESSAGE_TYPE_AMF0_META      = 18
	MESSAGE_TYPE_AMF0_SHARED    = 19
	MESSAGE_TYPE_AMF0_CMD       = 20
	MESSAGE_TYPE_AGGREGATE      = 22
)

type Basic struct {
	Type  uint8
	Chunk uint16
}

type Header struct {
	Timestamp uint32
	Length    uint32
	Type      uint8
	StreamID  uint32
}

type ChunkedInputStream struct {
	Header Header
	Chunk  uint32
	Data   []byte
	Last   uint32
	Delta  uint32
}

type ChunkedOutputStream struct {
	Header Header
	Chunk  uint32
}

type RawMessage struct {
	Chunk  uint16
	Header Header
	Data   []byte
}

type MessageDesc struct {
	Chunk     uint16
	Timestamp uint32
	Stream    uint32
}

type Message interface {
	Timestamp()       uint32
	StreamID()        uint32
	ChunkID()         uint16
	Encode(io.Writer)
	Type()            uint8
}

type ChunkedComms interface {
	Conn()           net.Conn
	Inbound(uint16)  *ChunkedInputStream
	Outbound(uint16) *ChunkedOutputStream
}