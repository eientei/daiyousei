package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"videostreamer/binutil"
	"reflect"
	"videostreamer/amf"
)

type GenericMessage struct {
	HeaderV Header
}

func (msg *GenericMessage) Encode(buf io.Writer) {
}

func (msg *GenericMessage) Header() *Header {
	return &msg.HeaderV
}

func (msg *GenericMessage) String() string {
	return msg.Header().String() + " ~ " + msg.String()
}

func reflectiveString(msg interface{}) string {
	typ := reflect.TypeOf(msg).Elem()
	val := reflect.ValueOf(msg).Elem()
	var fields string
	first := true
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		pref := ", "
		if first {
			pref = ""
		}
		if fld.Name != "GenericMessage" {
			valfld := val.Field(i)
			fldstr := val.String()
			if valfld.Kind() != reflect.Array {
				//fldstr = fmt.Sprint(valfld.Interface())
			}
			fields += pref + fld.Name + ": " + fldstr
			first = false
		}

	}
	header := reflect.ValueOf(msg).MethodByName("Header").Call([]reflect.Value{})[0].Interface()
	return header.(*Header).String() + " ~ [" + typ.Name() + "]{" + fields + "}"
}

func amfstring(msg interface{}, data []byte) string {
	var res string
	rdr := bytes.NewReader(data)
	for rdr.Len() > 0 {
		val, _ := amf.DecodeAMF(rdr)
		res += fmt.Sprint(val)
		if rdr.Len() > 0 {
			res += ", "
		}
	}
	header := reflect.ValueOf(msg).MethodByName("Header").Call([]reflect.Value{})[0].Interface()
	return header.(*Header).String() + " ~ [" + reflect.TypeOf(msg).Elem().Name() + "](" + res + ")"
}

func Decode(raw *RawMessage) Message {
	var msg Message
	switch raw.Header.Type {
	case MESSAGE_TYPE_SET_CHUNK_SIZE:
		msg = decodeSetChunkSize(&raw.Data)
	case MESSAGE_TYPE_ABORT:
		msg = decodeAbort(&raw.Data)
	case MESSAGE_TYPE_ACK:
		msg = decodeAck(&raw.Data)
	case MESSAGE_TYPE_USER:
		msg = decodeUser(&raw.Data)
	case MESSAGE_TYPE_WINACK:
		msg = decodeWinack(&raw.Data)
	case MESSAGE_TYPE_SET_PEER_BAND:
		msg = decodeSetPeerBand(&raw.Data)
	case MESSAGE_TYPE_AUDIO:
		msg = decodeAudio(&raw.Data)
	case MESSAGE_TYPE_VIDEO:
		msg = decodeVideo(&raw.Data)
	case MESSAGE_TYPE_AMF0_META:
		msg = decodeAmf0Meta(&raw.Data)
	case MESSAGE_TYPE_AMF0_CMD:
		msg = decodeAmf0Cmd(&raw.Data)
	default:
		panic(fmt.Errorf("Unknown message type during deocde: %v", raw.Header.Type))
	}
	reflect.ValueOf(msg).Elem().FieldByName("HeaderV").Set(reflect.ValueOf(raw.Header))
	return msg
}

func NewMessage(header Header, msg interface{}) Message {
	switch msg.(type) {
	case *SetChunkSizeMessage:
		header.Type = MESSAGE_TYPE_SET_CHUNK_SIZE
	case *AbortMessage:
		header.Type = MESSAGE_TYPE_ABORT
	case *AckMessage:
		header.Type = MESSAGE_TYPE_ACK
	case *UserMessage:
		header.Type = MESSAGE_TYPE_USER
	case *WinackMessage:
		header.Type = MESSAGE_TYPE_WINACK
	case *SetPeerBandMessage:
		header.Type = MESSAGE_TYPE_SET_PEER_BAND
	case *AudioMessage:
		header.Type = MESSAGE_TYPE_AUDIO
	case *VideoMessage:
		header.Type = MESSAGE_TYPE_VIDEO
	case *Amf0MetaMessage:
		header.Type = MESSAGE_TYPE_AMF0_META
	case *Amf0CmdMessage:
		header.Type = MESSAGE_TYPE_AMF0_CMD
	}
	reflect.ValueOf(msg).Elem().FieldByName("HeaderV").Set(reflect.ValueOf(header))
	return msg.(Message)
}

///
type SetChunkSizeMessage struct {
	GenericMessage
	Size uint32
}
func decodeSetChunkSize(rdr io.Reader) Message {
	return &SetChunkSizeMessage{
		Size: uint32(binutil.ReadInt(rdr, 4)),
	}
}
func (msg *SetChunkSizeMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size) & 0x7FFFFFFF, 4)
}
func (msg *SetChunkSizeMessage) String() string {
	return reflectiveString(msg)
}

///
type AbortMessage struct {
	GenericMessage
	Stream uint32
}
func decodeAbort(rdr io.Reader) Message {
	return &AbortMessage{
		Stream: uint32(binutil.ReadInt(rdr, 4)),
	}
}
func (msg *AbortMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Stream), 4)
}
func (msg *AbortMessage) String() string {
	return reflectiveString(msg)
}

///
type AckMessage struct {
	GenericMessage
	Size uint32
}
func decodeAck(rdr io.Reader) Message {
	return &AckMessage{
		Size: uint32(binutil.ReadInt(rdr, 4)),
	}
}
func (msg *AckMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
}
func (msg *AckMessage) String() string {
	return reflectiveString(msg)
}

///
type UserMessage struct {
	GenericMessage
	Event  uint16
	First  uint32
	Second uint32
}
func decodeUser(rdr io.Reader) Message {
	var usr UserMessage
	usr.Event = uint16(binutil.ReadInt(rdr, 2))
	usr.First = uint32(binutil.ReadInt(rdr, 4))
	if usr.Event == USER_EVENT_SET_BUFFER_LENGTH {
		usr.Second = uint32(binutil.ReadInt(rdr, 4))
	}
	return &usr
}
func (msg *UserMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Event), 2)
	binutil.WriteInt(buf, int(msg.First), 4)
	if msg.Event == USER_EVENT_SET_BUFFER_LENGTH {
		binutil.WriteInt(buf, int(msg.Second), 4)
	}
}
func (msg *UserMessage) String() string {
	var typname string
	switch msg.Event {
	case USER_EVENT_STREAM_BEGIN:
		typname = "STREAM_BEGIN"
	case USER_EVENT_STREAM_EOF:
		typname = "STREAM_EOF"
	case USER_EVENT_STREAM_DRY:
		typname = "STREAM_DRY"
	case USER_EVENT_SET_BUFFER_LENGTH:
		typname = "SET_BUFFER_LENGTH"
	case USER_EVENT_STREAM_IS_RECORDED:
		typname = "STREAM_IS_RECORDED"
	case USER_EVENT_PING_REQUEST:
		typname = "PING_REQUEST"
	case USER_EVENT_PING_RESPONSE:
		typname = "PING_RESPONSE"
	}
	return reflectiveString(msg) + "{" + typname + "}"
}

///
type WinackMessage struct {
	GenericMessage
	Size uint32
}
func decodeWinack(rdr io.Reader) Message {
	return &WinackMessage{
		Size: uint32(binutil.ReadInt(rdr, 4)),
	}
}
func (msg *WinackMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
}
func (msg *WinackMessage) String() string {
	return reflectiveString(msg)
}

///
type SetPeerBandMessage struct {
	GenericMessage
	Size uint32
	Type uint8
}
func decodeSetPeerBand(rdr io.Reader) Message {
	return &SetPeerBandMessage{
		Size: uint32(binutil.ReadInt(rdr, 4)),
		Type: uint8(binutil.ReadInt(rdr, 1)),
	}
}
func (msg *SetPeerBandMessage) Encode(buf io.Writer) {
	binutil.WriteInt(buf, int(msg.Size), 4)
	binutil.WriteInt(buf, int(msg.Type), 1)
}
func (msg *SetPeerBandMessage) String() string {
	return reflectiveString(msg)
}

///
type AudioMessage struct {
	GenericMessage
	Data []byte
}
func decodeAudio(rdr io.Reader) Message {
	len := rdr.(*bytes.Buffer).Len()
	return &AudioMessage{
		Data: binutil.ReadBuf(rdr, len),
	}
}
func (msg *AudioMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}
func (msg *AudioMessage) String() string {
	return reflectiveString(msg)
}

///
type VideoMessage struct {
	GenericMessage
	Data []byte
}
func decodeVideo(rdr io.Reader) Message {
	len := rdr.(*bytes.Buffer).Len()
	return &VideoMessage{
		Data: binutil.ReadBuf(rdr, len),
	}
}
func (msg *VideoMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}
func (msg *VideoMessage) String() string {
	return reflectiveString(msg)
}

///
type Amf0MetaMessage struct {
	GenericMessage
	Data []byte
}
func decodeAmf0Meta(rdr io.Reader) Message {
	len := rdr.(*bytes.Buffer).Len()
	return &Amf0MetaMessage{
		Data: binutil.ReadBuf(rdr, len),
	}
}
func (msg *Amf0MetaMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}
func (msg *Amf0MetaMessage) String() string {
	return amfstring(msg, msg.Data)
}

///
type Amf0CmdMessage struct {
	GenericMessage
	Data []byte
}
func decodeAmf0Cmd(rdr io.Reader) Message {
	len := rdr.(*bytes.Buffer).Len()
	return &Amf0CmdMessage{
		Data: binutil.ReadBuf(rdr, len),
	}
}
func (msg *Amf0CmdMessage) Encode(buf io.Writer) {
	binutil.WriteBuf(buf, msg.Data)
}
func (msg *Amf0CmdMessage) String() string {
	return amfstring(msg, msg.Data)
}
