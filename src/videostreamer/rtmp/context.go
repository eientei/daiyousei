package rtmp

import (
	"net"
	"videostreamer/core"
	"videostreamer/binutil"
	"videostreamer/check"
	"bytes"
)


func NewRTMPContext(conn net.Conn, app *core.Application) *RTMPContext {
	return &RTMPContext{
		Running:  true,
		Conn:     conn,
		App:      app,
		In:       make(map[uint16]*RawMessage),
		Out:      make(map[uint16]*RawMessage),
		InMsg:    make(chan Message, 16),
		OutMsg:   make(chan Message, 16),
		InChunk:  128,
		OutChunk: 128,
	}
}

func (context *RTMPContext) readBasic() (fmt uint8, chunkid uint16) {
	fst := binutil.ReadInt(context.Conn, 1)
	fmt = uint8((fst>>6) & 0x03)
	chunkid = uint16(fst & 0x3f)
	switch chunkid {
	case 0:
		chunkid = 64 + uint16(binutil.ReadInt(context.Conn, 1))
	case 1:
		chunkid = 64 + uint16(binutil.ReadIntLE(context.Conn, 2))
	}
	return fmt,chunkid
}

func (context *RTMPContext) writeBasic(fmt uint8, chunkid uint16) {
	fst := fmt<<6
	if chunkid >= 320 {
		binutil.WriteInt(context.Conn, int(fst | 1), 1)
		binutil.WriteIntLE(context.Conn, int(chunkid - 320), 2)
	} else if chunkid >= 64 {
		binutil.WriteInt(context.Conn, int(fst), 1)
		binutil.WriteInt(context.Conn, int(chunkid - 64), 1)
	} else {
		binutil.WriteInt(context.Conn, int(fst | uint8(chunkid)), 1)
	}
}

func (context *RTMPContext) rawMessage(chunkid uint16) *RawMessage {
	raw := context.In[chunkid]
	if raw == nil {
		raw = &RawMessage{}
		context.In[chunkid] = raw
	}
	return raw
}

func (context *RTMPContext) prevMessage(chunkid uint16) *RawMessage {
	prev := context.Out[chunkid]
	if prev == nil {
		prev = &RawMessage{}
		context.Out[chunkid] = prev
	}
	return prev
}

func (context *RTMPContext) ReadChunk() (err error) {
	defer check.CheckPanicHandler(&err)
	fmt, chunkid := context.readBasic()
	raw := context.rawMessage(chunkid)
	switch fmt {
	case BASIC_TYPE_FULL:
		raw.Header.Timestamp = uint32(binutil.ReadInt(context.Conn, 3))
		raw.Header.Length    = uint32(binutil.ReadInt(context.Conn, 3))
		raw.Header.Type      =  uint8(binutil.ReadInt(context.Conn, 1))
		raw.Header.StreamID  = uint32(binutil.ReadIntLE(context.Conn, 4))
		if raw.Header.Timestamp == 0xFFFFFF {
			raw.Header.Timestamp = uint32(binutil.ReadInt(context.Conn, 4))
		}
	case BASIC_TYPE_MEDIUM:
		raw.Delta            = uint32(binutil.ReadInt(context.Conn, 3))
		raw.Header.Timestamp = raw.Header.Timestamp + raw.Delta
		raw.Header.Length    = uint32(binutil.ReadInt(context.Conn, 3))
		raw.Header.Type      =  uint8(binutil.ReadInt(context.Conn, 1))
	case BASIC_TYPE_SHORT:
		raw.Delta            = uint32(binutil.ReadInt(context.Conn, 3))
		raw.Header.Timestamp = raw.Header.Timestamp + raw.Delta
	case BASIC_TYPE_NONE:
		if raw.Data.Len() == 0 {
			raw.Header.Timestamp = raw.Header.Timestamp + raw.Delta
		}
	}

	expect := context.InChunk
	if (expect + uint32(raw.Data.Len())) > raw.Header.Length {
		expect = raw.Header.Length - uint32(raw.Data.Len())
	}
	raw.Data.Write(binutil.ReadBuf(context.Conn, int(expect)))

	if raw.Header.Length == uint32(raw.Data.Len()) {
		msg := Decode(raw)
		switch msg.Header().Type {
		case MESSAGE_TYPE_SET_CHUNK_SIZE:
			context.InChunk = msg.(*SetChunkSizeMessage).Size
			break;
		}
		context.InMsg <- msg
		raw.Data.Reset()
	}
	return
}

func (context *RTMPContext) WriteMessage(msg Message) (err error) {
	defer check.CheckPanicHandler(&err)
	var buf bytes.Buffer
	var fmt uint8
	msg.Encode(&buf)
	prev := context.prevMessage(msg.Header().ChunkID)
	if msg.Header().ForceFmt {
		fmt = msg.Header().Format
	}  else {
		fmt = BASIC_TYPE_FULL
		if prev.Header.StreamID == msg.Header().StreamID && msg.Header().StreamID != 0 {
			fmt = BASIC_TYPE_MEDIUM
			if prev.Header.Length == uint32(buf.Len()) && prev.Header.Type == msg.Header().Type {
				fmt = BASIC_TYPE_SHORT
				if prev.Header.Timestamp + prev.Delta == msg.Header().Timestamp {
					fmt = BASIC_TYPE_NONE
				}
			}
			prev.Delta = msg.Header().Timestamp - prev.Header.Timestamp
		}
	}
	prev.Header = *msg.Header()

	context.writeBasic(fmt, msg.Header().ChunkID)

	switch fmt {
	case BASIC_TYPE_FULL:
		ts := msg.Header().Timestamp
		if ts >= 0xFFFFFF {
			ts = 0xFFFFFF
		}
		binutil.WriteInt(context.Conn, int(ts), 3)
		binutil.WriteInt(context.Conn, int(buf.Len()), 3)
		binutil.WriteInt(context.Conn, int(msg.Header().Type), 1)
		binutil.WriteIntLE(context.Conn, int(msg.Header().StreamID), 4)
		if msg.Header().Timestamp >= 0xFFFFFF {
			binutil.WriteInt(context.Conn, int(msg.Header().Timestamp), 4)
		}
	case BASIC_TYPE_MEDIUM:
		binutil.WriteInt(context.Conn, int(prev.Delta), 3)
		binutil.WriteInt(context.Conn, int(buf.Len()), 3)
		binutil.WriteInt(context.Conn, int(msg.Header().Type), 1)
	case BASIC_TYPE_SHORT:
		binutil.WriteInt(context.Conn, int(prev.Delta), 3)
	case BASIC_TYPE_NONE:
	}

	b := buf.Bytes()
	off := uint32(0)
	lim := context.OutChunk
	for {
		if lim > uint32(buf.Len()) {
			lim = uint32(buf.Len())
		}

		binutil.WriteBuf(context.Conn, b[off:lim])
		if (lim == uint32(buf.Len())) {
			break
		}

		off += context.OutChunk
		lim += context.OutChunk
		context.writeBasic(BASIC_TYPE_NONE, msg.Header().ChunkID)
	}
	return
}