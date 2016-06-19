package rtmp

import (
	"videostreamer/core"
	"videostreamer/syncutil"
	"videostreamer/check"
	"net"
	"videostreamer/logger"
	"io"
	"bytes"
	"videostreamer/amf"
)

func Serve(app *core.Application, latch *syncutil.SyncLatch, addr string) {
	logger.Info("RTMP server started")
	ln := check.Check1(net.Listen("tcp", addr)).(*net.TCPListener)
	latch.Handle(func() {
		ln.Close()
	})
	for latch.Running {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		go connection(latch.SubLatch(), conn, app)
	}

	latch.Await()
	latch.Complete()
	logger.Info("RTMP server done")
	return
}

func connection(latch *syncutil.SyncLatch, conn net.Conn, app *core.Application) {
	err := Handshake(conn)
	latch.Handle(func() {
		conn.Close()
	})

	if err != nil {
		logger.Info("Handshake failed")
		latch.Complete()
		return
	}

	logger.Info("Clinet connected")

	context := NewRTMPContext(conn, app)

	go recv(context, latch.SubLatch())
	go send(context, latch.SubLatch())
	go hndl(context, latch.SubLatch())

	latch.Await()
	latch.Complete()
	logger.Info("Clinet disconnected")
}

func recv(context *RTMPContext, latch *syncutil.SyncLatch) {
	for latch.Running {
		if err := context.ReadChunk(); err != nil {
			if err != io.EOF && latch.Running {
				logger.Error(err)
			}
			break
		}
	}
	close(context.InMsg)
	close(context.OutMsg)
	if context.Stream != nil && context.Client != nil {
		context.Stream.Unsubscribe(context.Client)
	}
	latch.Complete()
}

func send(context *RTMPContext, latch *syncutil.SyncLatch) {
	for latch.Running {
		msg := <- context.OutMsg
		if msg == nil {
			break
		}
		//logger.Debug("<-", msg)
		/*
		if (msg.Header().Type == MESSAGE_TYPE_VIDEO) {
			dt := msg.(*VideoMessage).Data
			logger.Debug("^^", len(dt), dt[len(dt)-10:])
		}
		*/
		context.WriteMessage(msg)
	}
	latch.Complete()
}

func hndl(context *RTMPContext, latch *syncutil.SyncLatch) {
	for latch.Running {
		msg := <- context.InMsg
		if msg == nil {
			break
		}
		//logger.Debug("->", msg)
		switch msg.Header().Type {
		case MESSAGE_TYPE_WINACK:
			context.InAck = msg.(*WinackMessage).Size
		case MESSAGE_TYPE_AMF0_CMD:
			handlecmd(context, msg.(*Amf0CmdMessage))
		case MESSAGE_TYPE_AMF0_META:
			handlemeta(context, msg.(*Amf0MetaMessage))
		case MESSAGE_TYPE_USER:
			handleuser(context, msg.(*UserMessage))
		case MESSAGE_TYPE_AUDIO:
			if context.Stream != nil {
				context.Stream.BroadcastAudio(core.NewAudioData(msg.Header().Timestamp, msg.(*AudioMessage).Data))
			}
		case MESSAGE_TYPE_VIDEO:
			if context.Stream != nil {
				data := msg.(*VideoMessage).Data
				viddata := core.NewVideoData(msg.Header().Timestamp, data)
				if data[1] == 0 {
					context.Stream.KeyVideo = viddata
				}
				context.Stream.BroadcastVideo(viddata)
			}
		}
	}
	latch.Complete()
}


type RTMPClient struct {
	Context *RTMPContext
}

func (client *RTMPClient) ConsumeVideo(data *core.VideoData) {
	if (!client.Context.WasVideo) {
		if ((data.Data[0] & 0xf0) >> 4) == 1 {
			client.Context.WasVideo = true
		} else {
			return
		}
	}
	client.Context.OutMsg <- NewMessage(Header{ChunkID:6, Timestamp: data.Time}, &VideoMessage{Data: data.Data})
}

func (client *RTMPClient) ConsumeAudio(data *core.AudioData) {
	client.Context.OutMsg <- NewMessage(Header{ChunkID:4, Timestamp: data.Time}, &AudioMessage{Data: data.Data})
}

func (client *RTMPClient) ConsumeMeta(data *core.MetaData) {
	client.Context.OutMsg <- makeMetadata(data)
}

func makeMetadata(data *core.MetaData) Message {
	buf := bytes.Buffer{}
	fw := float64(data.Width)
	fh := float64(data.Height)
	ff := float64(data.Framerate)
	amf.EncodeAMF(&buf, "onMetaData")
	amf.EncodeAMF(&buf, struct{
		Width         float64 `name:"width"`
		Height        float64 `name:"height"`
		DisplayWidth  float64 `name:"displayWidth"`
		DisplayHeight float64 `name:"displayHeight"`
		Duration      float64 `name:"duration"`
		Framerate     float64 `name:"framerate"`
		Videocodecid  float64 `name:"videocodecid"`
		Audiocodecid  float64 `name:"audiocodecid"`

	}{fw, fh, fw, fh, -1, ff, 7, 10})
	return NewMessage(Header{ChunkID: 3}, &Amf0MetaMessage{Data: buf.Bytes()})
}

func handlecmd(context *RTMPContext, msg *Amf0CmdMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	name := check.Check1(amf.DecodeAMF(rdr)).(string)
	switch name {
	case "connect":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)

		context.OutMsg <- NewMessage(Header{ChunkID: 2}, &WinackMessage{Size: 5000000})
		context.OutMsg <- NewMessage(Header{ChunkID: 2}, &SetPeerBandMessage{Size: 5000000})
		context.OutMsg <- NewMessage(Header{ChunkID: 2}, &SetChunkSizeMessage{Size: 128})
		context.OutMsg <- NewMessage(Header{ChunkID: 2}, &UserMessage{
			Event: USER_EVENT_STREAM_BEGIN,
			First: 1,
		})

		buf := bytes.Buffer{}
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, struct {
			FmtVer string  `name:"fmtVer"`
			Caps   float64 `name:"capabilities"`
		}{"FMS/3,0,1,123", 31})
		amf.EncodeAMF(&buf, struct {
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
			ObjEnc float64 `name:"objectEncoding"`
		}{"status", "NetConnection.Connect.Success", "Connection Success.", 0})
		context.OutMsg <- NewMessage(Header{ChunkID: 3}, &Amf0CmdMessage{Data: buf.Bytes()})
	case "createStream":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)
		buf := bytes.Buffer{}
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, 1)
		context.OutMsg <- NewMessage(Header{ChunkID: 3}, &Amf0CmdMessage{Data: buf.Bytes()})
	case "play":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		context.Stream = context.App.AcquireStream(streamname)
		context.Client = &RTMPClient{Context: context}
		context.OutMsg <- NewMessage(Header{ChunkID: 2}, &SetChunkSizeMessage{Size: context.OutChunk})
		context.OutMsg <- NewMessage(Header{ChunkID: 3}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})
		context.OutMsg <- NewMessage(Header{ChunkID: 3}, &UserMessage{Event: USER_EVENT_STREAM_BEGIN, First: 0})

		buf := bytes.Buffer{}
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct {
			Level string  `name:"level"`
			Code  string  `name:"code"`
			Desc  string  `name:"description"`
		}{"status", "NetStream.Play.Start", "Start live."})
		context.OutMsg <- NewMessage(Header{ChunkID: 5}, &Amf0CmdMessage{Data: buf.Bytes()})

		buf = bytes.Buffer{}
		amf.EncodeAMF(&buf, "|RtmpSampleAccess")
		amf.EncodeAMF(&buf, true)
		amf.EncodeAMF(&buf, true)
		context.OutMsg <- NewMessage(Header{ChunkID: 5}, &Amf0MetaMessage{Data: buf.Bytes()})

		if context.Stream.Metadata != nil {
			context.Client.ConsumeMeta(context.Stream.Metadata)
		}

		if context.Stream.KeyVideo != nil {
			context.Client.ConsumeVideo(context.Stream.KeyVideo)
		}
		if context.Stream.KeyAudio != nil {
			context.Client.ConsumeAudio(context.Stream.KeyAudio)
		}

		context.Stream.Subscribe(context.Client)
	case "publish":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		context.Stream = context.App.AcquireStream(streamname)

		context.OutMsg <- NewMessage(Header{ChunkID: 3}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})

		buf := bytes.Buffer{}
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
		}{"status", "NetStream.Publish.Start", "Start publising."})
		context.OutMsg <- NewMessage(Header{ChunkID: 5}, &Amf0CmdMessage{Data: buf.Bytes()})
	}
	return
}

func handlemeta(context *RTMPContext, msg *Amf0MetaMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	amf.DecodeAMF(rdr)
	amf.DecodeAMF(rdr)
	raw := check.Check1(amf.DecodeAMF(rdr)).(amf.AMFMap)
	context.Stream.Metadata = core.NewMetaData(uint32(raw["width"].(float64)), uint32(raw["height"].(float64)), uint32(raw["framerate"].(float64)))
	context.Stream.BroadcastMeta(context.Stream.Metadata)
	return
}

func handleuser(context *RTMPContext, msg *UserMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	return
}