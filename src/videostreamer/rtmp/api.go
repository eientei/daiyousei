package rtmp

import (
	"io"
	"net"
	"videostreamer/core"
	"bytes"
)

const (
	BASIC_TYPE_FULL   = 0
	BASIC_TYPE_MEDIUM = 1
	BASIC_TYPE_SHORT  = 2
	BASIC_TYPE_NONE   = 3
)

const (
	MESSAGE_TYPE_SET_CHUNK_SIZE =  1
	MESSAGE_TYPE_ABORT          =  2
	MESSAGE_TYPE_ACK            =  3
	MESSAGE_TYPE_USER           =  4
	MESSAGE_TYPE_WINACK         =  5
	MESSAGE_TYPE_SET_PEER_BAND  =  6
	MESSAGE_TYPE_EDGE           =  7
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
const (
	USER_EVENT_STREAM_BEGIN       = 0
	USER_EVENT_STREAM_EOF         = 1
	USER_EVENT_STREAM_DRY         = 2
	USER_EVENT_SET_BUFFER_LENGTH  = 3
	USER_EVENT_STREAM_IS_RECORDED = 4
	USER_EVENT_PING_REQUEST       = 6
	USER_EVENT_PING_RESPONSE      = 7
)

type Header struct {
	Format    uint8
	ChunkID   uint16
	Timestamp uint32
	Length    uint32
	Type      uint8
	StreamID  uint32
}

type RawMessage struct {
	Header Header
	Delta  uint32
	Data   bytes.Buffer
}

type Message interface {
	Header()          *Header
	Encode(io.Writer)
	String()          string
}

type RTMPContext struct {
	Running  bool
	Conn     net.Conn
	App      *core.Application
	Client   core.Consumer
	Stream   *core.Stream
	WasVideo bool
	In       map[uint16]*RawMessage
	Out      map[uint16]*RawMessage
	InMsg    chan Message
	OutMsg   chan Message
	InChunk  uint32
	OutChunk uint32
	InAck    uint32
	OutAck   uint32
	InTrans  uint32
	OutTrans uint32
}