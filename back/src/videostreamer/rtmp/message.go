package rtmp

import (
	"videostreamer/binutil"
	"bytes"
	"reflect"
	"io"
	"fmt"
	"videostreamer/logger"
)

var (
	genericFields[]string
)

func init() {
	typ := reflect.TypeOf(MessageDesc{})
	for i := 0; i < typ.NumField(); i++ {
		genericFields = append(genericFields, typ.Field(i).Name)
	}
}

func Decode(raw RawMessage) Message {
	var msg Message
	buf := bytes.NewBuffer(raw.Data)
	switch raw.Header.Type {
	case MESSAGE_TYPE_SET_CHUNK_SIZE:
		msg = decodeSetChunkSize(buf, raw)
	case MESSAGE_TYPE_ABORT:
		msg = decodeAbort(buf, raw)
	case MESSAGE_TYPE_ACK:
		msg = decodeAck(buf, raw)
	case MESSAGE_TYPE_USER:
		msg = decodeUser(buf, raw)
	case MESSAGE_TYPE_WINACK:
		msg = decodeWinack(buf, raw)
	case MESSAGE_TYPE_SET_PEER_BAND:
		msg = decodeSetPeerBand(buf, raw)
	//case MESSAGE_TYPE_EDGE:
	case MESSAGE_TYPE_AUDIO:
		msg = decodeAudio(buf, raw)
	case MESSAGE_TYPE_VIDEO:
		msg = decodeVideo(buf, raw)
	//case MESSAGE_TYPE_AMF3_META:
	//case MESSAGE_TYPE_AMF3_SHARED:
	//case MESSAGE_TYPE_AMF3_CMD:
	case MESSAGE_TYPE_AMF0_META:
		msg = decodeAmfMeta(buf, raw)
	//case MESSAGE_TYPE_AMF0_SHARED:
	case MESSAGE_TYPE_AMF0_CMD:
		msg = decodeAmfCmd(buf, raw)
	//case MESSAGE_TYPE_AGGREGATE:
	default:
		panic(fmt.Errorf("Unknown message type during deocde: %v", raw.Header.Type))
	}
	return NewMessage(rawToDesc(raw), msg)
}

func reflectiveSet(desc MessageDesc, iface interface{}) {
	typeid := 0
	switch iface.(type) {
	case *SetChunkSizeMessage:
		typeid = MESSAGE_TYPE_SET_CHUNK_SIZE
	case *AbortMessage:
		typeid = MESSAGE_TYPE_ABORT
	case *AckMessage:
		typeid = MESSAGE_TYPE_ACK
	case *UserMessage:
		typeid = MESSAGE_TYPE_USER
	case *WinAckMessage:
		typeid = MESSAGE_TYPE_WINACK
	case *SetPeerBandMessage:
		typeid = MESSAGE_TYPE_SET_PEER_BAND
	case *AudioMessage:
		typeid = MESSAGE_TYPE_AUDIO
	case *VideoMessage:
		typeid = MESSAGE_TYPE_VIDEO
	case *AmfMetaMessage:
		typeid = MESSAGE_TYPE_AMF0_META
	case *AmfCmdMessage:
		typeid = MESSAGE_TYPE_AMF0_CMD
	}

	if typeid == 0 {
		logger.Error("UNKNWON!", iface)
	}

	src := reflect.ValueOf(desc)
	val := reflect.ValueOf(iface).Elem()

	for _, f := range genericFields {
		val.FieldByName(f + "id").Set(src.FieldByName(f))
	}
	val.FieldByName("Typeid").Set(reflect.ValueOf(uint8(typeid)))
}

func NewMessage(desc MessageDesc, iface interface{}) Message {
	reflectiveSet(desc, iface)
	return iface.(Message)
}

type GenericMessage struct {
	Timestampid uint32
	Streamid    uint32
	Chunkid     uint16
	Typeid      uint8
}

func (msg *GenericMessage) ChunkID() uint16 {
	return msg.Chunkid
}

func (msg *GenericMessage) StreamID() uint32 {
	return msg.Streamid
}

func (msg *GenericMessage) Type() uint8 {
	return msg.Typeid
}

func (msg *GenericMessage) Timestamp() uint32 {
	return msg.Timestampid
}

func (msg *GenericMessage) Encode() []byte {
	return []byte{}
}

type SetChunkSizeMessage struct {
	GenericMessage
	Size uint32
}

func rawToDesc(raw RawMessage) MessageDesc {
	return MessageDesc{raw.Chunk, raw.Header.Timestamp, raw.Header.StreamID}
}

/***** SET_CHUNK_SIZE(1) *****/
func decodeSetChunkSize(buf io.Reader, raw RawMessage) Message {
	return &SetChunkSizeMessage{
		Size: uint32(binutil.ReadInt(buf, 4)),
	}
}
func (msg *SetChunkSizeMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size) & 0x7FFFFFFF, 4)
}


/***** ABORT(2) *****/
type AbortMessage struct {
	GenericMessage
	Stream uint32
}
func decodeAbort(buf io.Reader, raw RawMessage) Message {
	return &AbortMessage{
		Stream: uint32(binutil.ReadInt(buf, 4)),
	}
}
func (msg *AbortMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Stream), 4)
}

/***** ACK(3) *****/
type AckMessage struct {
	GenericMessage
	Size uint32
}
func decodeAck(buf io.Reader, raw RawMessage) Message {
	return &AckMessage{
		Size: uint32(binutil.ReadInt(buf, 4)),
	}
}
func (msg *AckMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
}

/***** USER(4) *****/
type UserMessage struct {
	GenericMessage
	Event  uint16
	First  uint32
	Second uint32
}
func decodeUser(buf io.Reader, raw RawMessage) Message {
	var first, second uint32

	first = uint32(binutil.ReadInt(buf, 4))
	if first == 3 {
		second = uint32(binutil.ReadInt(buf, 4))
	}

	return &UserMessage{
		First: first,
		Second: second,
	}
}
func (msg *UserMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Event), 2)
	binutil.WriteInt(buf, int(msg.First), 4)
	if msg.First == 3 {
		binutil.WriteInt(buf, int(msg.Second), 4)
	}
}

/***** WINACK(5) *****/
type WinAckMessage struct {
	GenericMessage
	Size uint32
}
func decodeWinack(buf io.Reader, raw RawMessage) Message {
	return &WinAckMessage{
		Size: uint32(binutil.ReadInt(buf, 4)),
	}
}
func (msg *WinAckMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
}

/***** SET_PEER_BAND(6) *****/
type SetPeerBandMessage struct {
	GenericMessage
	Size    uint32
	Limtype uint8
}
func decodeSetPeerBand(buf io.Reader, raw RawMessage) Message {
	return &SetPeerBandMessage{
		Size: uint32(binutil.ReadInt(buf, 4)),
		Limtype: uint8(binutil.ReadInt(buf, 1)),
	}
}
func (msg *SetPeerBandMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
	binutil.WriteInt(buf, int(msg.Limtype), 1)
}

/***** AUDIO(8) *****/
type AudioMessage struct {
	GenericMessage
	Data []byte
}
func decodeAudio(buf io.Reader, raw RawMessage) Message {
	return &AudioMessage{
		Data: raw.Data,
	}
}
func (msg *AudioMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}

/***** VIDEO(9) *****/
type VideoMessage struct {
	GenericMessage
	Data []byte
}
func decodeVideo(buf io.Reader, raw RawMessage) Message {
	return &VideoMessage{
		Data: raw.Data,
	}
}
func (msg *VideoMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}

/***** AMF_META(18) *****/
type AmfMetaMessage struct {
	GenericMessage
	Data []byte
}
func decodeAmfMeta(buf io.Reader, raw RawMessage) Message {
	return &AmfMetaMessage{
		Data: raw.Data,
	}
}
func (msg *AmfMetaMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}

/***** AMF_META(20) *****/
type AmfCmdMessage struct {
	GenericMessage
	Data []byte
}
func decodeAmfCmd(buf io.Reader, raw RawMessage) Message {
	return &AmfCmdMessage{
		Data: raw.Data,
	}
}
func (msg *AmfCmdMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}