package rtmp

import (
	"io"
	"videostreamer/check"
	"videostreamer/binutil"
	"bytes"
)


func ReadBasic(in io.Reader) (basic Basic) {
	fst := binutil.ReadInt(in, 1)
	basic.Type = uint8((fst>>6) & 0x03)
	basic.Chunk = uint16(fst & 0x3f)
	switch basic.Chunk {
	case 0:
		basic.Chunk = 64 + uint16(binutil.ReadInt(in, 1))
	case 1:
		basic.Chunk = 64 + uint16(binutil.ReadIntLE(in, 2))
	}
	return
}

func WriteBasic(out io.Writer, basic Basic) {
	fst := basic.Type<<6

	if basic.Chunk >= 320 {
		binutil.WriteInt(out, int(fst | 1), 1)
		binutil.WriteIntLE(out, int(basic.Chunk - 320), 2)
		return
	}

	if basic.Chunk >= 64 {
		binutil.WriteInt(out, int(fst), 1)
		binutil.WriteInt(out, int(basic.Chunk - 64), 1)
		return
	}

	binutil.WriteInt(out, int(fst | uint8(basic.Chunk)), 1)
	return
}

func syncmsgproc(comms *ChunkedComms, msg Message) {
	switch msg.Type() {
	case MESSAGE_TYPE_SET_CHUNK_SIZE:
		comms.ChunkIn = msg.(*SetChunkSizeMessage).Size
		break;
	}
}

func ReadNext(comms *ChunkedComms, out chan Message) (err error) {
	defer check.CheckPanicHandler(&err)

	comm := comms.Conn
	basic := ReadBasic(comm)
	chunk := comms.Inbound(basic.Chunk)
	datalen := uint32(len(chunk.Data))

	switch basic.Type {
	case BASIC_TYPE_FULL:
		chunk.Header.Timestamp = uint32(binutil.ReadInt(comm, 3))
		chunk.Header.Length    = uint32(binutil.ReadInt(comm, 3))
		chunk.Header.Type      =  uint8(binutil.ReadInt(comm, 1))
		chunk.Header.StreamID  = uint32(binutil.ReadInt(comm, 4))
		if chunk.Header.Timestamp == 0xFFFFFF {
			chunk.Header.Timestamp = uint32(binutil.ReadInt(comm, 4))
		}
		chunk.Delta = 0
		chunk.Last  = chunk.Header.Timestamp
	case BASIC_TYPE_MEDIUM:
		chunk.Delta = uint32(binutil.ReadInt(comm, 3))
		chunk.Header.Timestamp = chunk.Last + chunk.Delta
		chunk.Header.Length    = uint32(binutil.ReadInt(comm, 3))
		chunk.Header.Type      =  uint8(binutil.ReadInt(comm, 1))
		chunk.Last  = chunk.Header.Timestamp
	case BASIC_TYPE_SHORT:
		chunk.Delta = uint32(binutil.ReadInt(comm, 3))
		chunk.Header.Timestamp = chunk.Last + chunk.Delta
		chunk.Last  = chunk.Header.Timestamp
	case BASIC_TYPE_NONE:
		if datalen == 0 {
			chunk.Header.Timestamp = chunk.Last + chunk.Delta
			chunk.Last  = chunk.Header.Timestamp
		}
	}

	max := comms.ChunkIn
	if (datalen + max) > chunk.Header.Length {
		max = chunk.Header.Length - datalen
	}

	chunk.Data = append(chunk.Data, binutil.ReadBuf(comm, int(max))...)
	if chunk.Header.Length == uint32(len(chunk.Data)) {
		msg := Decode(RawMessage{
			Chunk: basic.Chunk,
			Header: chunk.Header,
			Data: chunk.Data,
		})
		syncmsgproc(comms, msg)
		out <- msg
		chunk.Data = []byte{}
	}
	return
}

func WriteNext(comms *ChunkedComms, message Message) (length uint32, err error) {
	defer check.CheckPanicHandler(&err)
	comm := comms.Conn
	chunk := comms.Outbound(message.ChunkID())
	typeid := BASIC_TYPE_FULL

	bbuf := bytes.NewBuffer(nil)
	message.Encode(bbuf)
	buf := bbuf.Bytes()

	length = uint32(len(buf))

	if chunk.Header.StreamID == message.StreamID() && message.StreamID() != 0 {
		typeid = BASIC_TYPE_MEDIUM
		if chunk.Header.Length == length && chunk.Header.Type == message.Type() {
			typeid = BASIC_TYPE_SHORT
			if chunk.Header.Timestamp == message.Timestamp() {
				typeid = BASIC_TYPE_NONE
			}
		}
	}

	delta := message.Timestamp() - chunk.Header.Timestamp
	chunk.Header = Header{
		Timestamp: message.Timestamp(),
		Length:    length,
		Type:      message.Type(),
		StreamID:  message.StreamID(),
	}

	WriteBasic(comm, Basic{
		Type: uint8(typeid),
		Chunk: message.ChunkID(),
	})

	switch typeid {
	case BASIC_TYPE_FULL:
		binutil.WriteInt(comm, int(message.Timestamp()), 3)
		binutil.WriteInt(comm, int(length), 3)
		binutil.WriteInt(comm, int(message.Type()), 1)
		binutil.WriteInt(comm, int(message.StreamID()), 4)
		if message.Timestamp() >= 0xFFFFFF {
			binutil.WriteInt(comm, int(message.Timestamp()), 4)
		}
	case BASIC_TYPE_MEDIUM:
		binutil.WriteInt(comm, int(delta), 3)
		binutil.WriteInt(comm, int(length), 3)
		binutil.WriteInt(comm, int(message.Type()), 1)
	case BASIC_TYPE_SHORT:
		binutil.WriteInt(comm, int(delta), 3)
	case BASIC_TYPE_NONE:
	}

	max := length
	if length > comms.ChunkOut {
		max = comms.ChunkOut
	}

	slice := buf[:max]
	for {
		binutil.WriteBuf(comm, slice)
		if (uint32(cap(slice)) <= comms.ChunkOut) {
			break
		}
		slice = buf[comms.ChunkOut:]
		WriteBasic(comm, Basic{
			Type: uint8(BASIC_TYPE_NONE),
			Chunk: message.ChunkID(),
		})
	}
	return
}