package rtmp

import (
	"videostreamer/core"
	"videostreamer/check"
	"net"
	"videostreamer/syncutil"
	"videostreamer/logger"
	"time"
	"bytes"
	"videostreamer/amf"
	"videostreamer/binutil"
)

func Serve(wgroup syncutil.SyncService, registry core.Registry, bindaddr string) (err error) {
	defer check.CheckPanicHandler(&err)
	ln := check.Check1(net.Listen("tcp", bindaddr)).(*net.TCPListener)
	defer ln.Close()
	logger.Info("RTMP listener started")

	for wgroup.Running() {
		ln.SetDeadline(time.Now().Add(time.Duration(100)*time.Millisecond))
		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			logger.Error(err)
		}
		go handleconn(wgroup.SubService(), registry, conn)
	}

	wgroup.Wait()
	logger.Info("RTMP listener stopped")
	wgroup.Done()
	return
}

type RTMPClient struct {
	done     bool
	streamid uint32
	stream   core.Stream
	chin     chan Message
	chout    chan Message
	transIn   uint32
	transOut  uint32
	ackIn     uint32
	ackOut    uint32
	hasvid    bool
	hasaud    bool
	publisher bool
}

func closeremove(rtmp *RTMPClient) {
	if rtmp.done {
		return
	}
	if rtmp.stream != nil {
		rtmp.stream.Remove(rtmp)
	}
	close(rtmp.chin)
	close(rtmp.chout)
	rtmp.done = true
}

func (client *RTMPClient) Consume(data core.Data) {
	defer func() {
		v := recover()
		if v != nil {
			closeremove(client)
			logger.Error(v)
		}
	}()
	switch data.Type() {
	case core.TYPE_VIDEO:
		client.chout <- NewMessage(MessageDesc{Chunk: 6, Timestamp: data.Time(), Stream: client.streamid}, &VideoMessage{
			Data: binutil.Dup(data.Payload()),
		})
	case core.TYPE_AUDIO:
		client.chout <- NewMessage(MessageDesc{Chunk: 4, Timestamp: data.Time(), Stream: client.streamid}, &AudioMessage{
			Data: binutil.Dup(data.Payload()),
		})
	}
}

func (client *RTMPClient) Refresh(meta core.Meta) {
	defer func() {
		v := recover()
		if v != nil {
			closeremove(client)
			logger.Error(v)
		}
	}()
	var buf bytes.Buffer
	fw := float64(meta.Width())
	fh := float64(meta.Height())
	ff := float64(meta.FPS())
	amf.EncodeAMF(&buf, "onMetaData")
	amf.EncodeAMF(&buf, struct{
		Width         float64 `name:"width"`
		Height        float64 `name:"height"`
		DisplayWidth  float64 `name:"displayWidth"`
		DisplayHeight float64 `name:"displayHeight"`
		Duration      float64 `name:"duration"`
		Framerate     float64 `name:"framerate"`

	}{fw, fh, fw, fh, 0, ff})
	client.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: client.streamid}, &AmfMetaMessage{Data: binutil.Dup(buf.Bytes())})
}

func handleconn(wgroup syncutil.SyncService, registry core.Registry, conn net.Conn) {
	logger.Info("Client connected")
	err := Handshake(conn)
	if err != nil {
		logger.Info("Handshake failed")
		return
	}

	comms := NewChunkedComms(registry, conn)

	comms.Client = &RTMPClient{
		chin: make(chan Message, 16),
		chout: make(chan Message, 16),
		ackIn: 5000000,
		ackOut: 5000000,
	}

	go recv(comms, wgroup.SubService())
	go send(comms, wgroup.SubService())
	go hndl(comms, wgroup.SubService())

	wgroup.Wait()

	rtmp := comms.Client.(*RTMPClient)
	closeremove(rtmp)
	conn.Close()
	logger.Info("Client disconnected")
	wgroup.Done()
}

func recv(comms *ChunkedComms, wgroup syncutil.SyncService) {
	var err error
	rtmp := comms.Client.(*RTMPClient)
	for wgroup.Running() && !rtmp.done {
		comms.Conn.SetReadDeadline(time.Now().Add(func recv(context *RTMPContext, latch *syncutil.SyncLatch) {

		}time.Duration(100) * time.Millisecond))
		err = ReadNext(comms, rtmp.chin)
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			break
		}
	}
	closeremove(rtmp)
	wgroup.Done()
}

func send(comms *ChunkedComms, wgroup syncutil.SyncService) {
	rtmp := comms.Client.(*RTMPClient)
	for wgroup.Running() && !rtmp.done {
		select {
		case msg, ok := <-rtmp.chout:
			if !ok {
				break
			}
			logger.Debugf("<- [%d:%d][%d][%s]",
				msg.ChunkID(),
				msg.StreamID(),
				msg.Timestamp(),
				MessageTypeNames[msg.Type()])
			bts, _ := WriteNext(comms, msg)
			rtmp.transOut += bts
		}
	/*
		msg, ok := <-rtmp.chout
		if !ok {
			break
		}
		logger.Debugf("<- [%d:%d][%d][%s]",
			msg.ChunkID(),
			msg.StreamID(),
			msg.Timestamp(),
			MessageTypeNames[msg.Type()])
		bts, _ := WriteNext(comms, msg)
		rtmp.transOut += bts
		*/
	}
	wgroup.Done()
}

func hndl(comms *ChunkedComms, wgroup syncutil.SyncService) {
	defer func() {
		v := recover()
		if v != nil {
			logger.Error(v)
		}
	}()
	rtmp := comms.Client.(*RTMPClient)
	for wgroup.Running() && !rtmp.done {
		msg, ok := <-rtmp.chin
		if !ok {
			break
		}
		logger.Debugf("-> [%d:%d][%d][%s]",
			msg.ChunkID(),
			msg.StreamID(),
			msg.Timestamp(),
			MessageTypeNames[msg.Type()])
		switch msg.Type() {
		case MESSAGE_TYPE_WINACK:
			rtmp.ackIn = msg.(*WinAckMessage).Size
		case MESSAGE_TYPE_AMF0_CMD:
			if err := handleCmd(comms, msg.(*AmfCmdMessage)); err != nil {
				logger.Error(err)
			}
		case MESSAGE_TYPE_AMF0_META:
			if err := handleMeta(comms, msg.(*AmfMetaMessage)); err != nil {
				logger.Error(err)
			}
		case MESSAGE_TYPE_USER:
			usrmsg := msg.(*UserMessage)
			logger.Debugf("^ [USER][%s] %d %d", UserEventNames[usrmsg.Event], usrmsg.First, usrmsg.Second);
			if err := handleUser(comms, usrmsg); err != nil {
				logger.Error(err)
			}
		case MESSAGE_TYPE_AUDIO:
			if rtmp.publisher {
				comms.Client.(*RTMPClient).stream.Broadcast(core.NewData(core.TYPE_AUDIO, msg.Timestamp(), binutil.Dup(msg.(*AudioMessage).Data)), !rtmp.hasaud)
				rtmp.hasaud = false
			}
		case MESSAGE_TYPE_VIDEO:
			if rtmp.publisher {
				comms.Client.(*RTMPClient).stream.Broadcast(core.NewData(core.TYPE_VIDEO, msg.Timestamp(), binutil.Dup(msg.(*VideoMessage).Data)), !rtmp.hasvid)
				rtmp.hasvid = false
			}
		}
	}
	wgroup.Done()
}

func handleCmd(comms *ChunkedComms, msg *AmfCmdMessage) (err error) {
	rtmp := comms.Client.(*RTMPClient)
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	name := check.Check1(amf.DecodeAMF(rdr)).(string)
	logger.Debug("^^", name)

	switch name {
	case "connect":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)

		rtmp.chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &WinAckMessage{Size: 5000000})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetPeerBandMessage{Size: 5000000})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetChunkSizeMessage{Size: 128})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &UserMessage{
			Event: USER_EVENT_STREAM_BEGIN,
			First: 1,
		})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, struct{
			FmtVer string  `name:"fmtVer"`
			Caps   float64 `name:"capabilities"`
		}{"FMS/3,0,1,123", 31})
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
			ObjEnc float64 `name:"objectEncoding"`
		}{"status", "NetConnection.Connect.Success", "Connection Success.", 0})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &AmfCmdMessage{Data: binutil.Dup(buf.Bytes())})
	case "createStream":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)
		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, 1)
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &AmfCmdMessage{Data: binutil.Dup(buf.Bytes())})
	case "play":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		stream := comms.Registry.Stream(streamname)
		streamid := comms.Client.(*RTMPClient).streamid // check if play message id is more useful

		rtmp.chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetChunkSizeMessage{Size: comms.ChunkOut})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &UserMessage{Event: USER_EVENT_STREAM_BEGIN, First: streamid})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
		}{"status", "NetStream.Play.Start", "Start live."})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfCmdMessage{Data: binutil.Dup(buf.Bytes())})

		buf.Reset()
		amf.EncodeAMF(&buf, "|RtmpSampleAccess")
		amf.EncodeAMF(&buf, true)
		amf.EncodeAMF(&buf, true)
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfMetaMessage{Data: binutil.Dup(buf.Bytes())})

		buf.Reset()
		fw := float64(stream.Metadata().Width())
		fh := float64(stream.Metadata().Height())
		ff := float64(stream.Metadata().FPS())
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
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfMetaMessage{Data: buf.Bytes()})
		if stream.KeyVideo() != nil {
			rtmp.Consume(stream.KeyVideo())
		}
		if stream.KeyAudio() != nil {
			rtmp.Consume(stream.KeyAudio())
		}
		stream.Add(comms.Client)
	case "publish":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		comms.Client.(*RTMPClient).stream = comms.Registry.Stream(streamname)
		streamid := comms.Client.(*RTMPClient).streamid
		rtmp.publisher = true

		rtmp.chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: streamid}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
		}{"status", "NetStream.Publish.Start", "Start publising."})
		rtmp.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfCmdMessage{Data: binutil.Dup(buf.Bytes())})
	}

	return
}

func handleMeta(comms *ChunkedComms, msg *AmfMetaMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	amf.DecodeAMF(rdr)
	amf.DecodeAMF(rdr)
	metaraw := check.Check1(amf.DecodeAMF(rdr)).(amf.AMFMap)
	meta := core.NewMeta(int(metaraw["width"].(float64)), int(metaraw["height"].(float64)), int(metaraw["framerate"].(float64)))
	comms.Client.(*RTMPClient).stream.Refresh(meta)
	return
}

func handleUser(comms *ChunkedComms, msg *UserMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	switch msg.Event {
	case USER_EVENT_STREAM_BEGIN:
		comms.Client.(*RTMPClient).streamid = msg.StreamID()
	}
	return
}

/*
func Serve(wgroup syncutil.SyncService, registry core.Registry, bindaddr string) (err error) {
	defer wgroup.Wait()
	defer wgroup.Done()
	defer check.CheckPanicHandler(&err)
	ln := check.Check1(net.Listen("tcp", bindaddr)).(*net.TCPListener)
	defer ln.Close()

	for wgroup.Running() {
		ln.SetDeadline(time.Now().Add(time.Duration(100)*time.Millisecond))
		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			logger.Error(err)
		}
		go handleconn(wgroup.SubService(), registry, conn)
	}
	logger.Info("RTMP listener stopping")
	return
}

type RTMPClient struct {
	closed   bool
	chout    chan Message
	streamid uint32
	stream   core.Stream
}

func (client *RTMPClient) Consume(data core.Data) {
	if (client.closed) {
		return
	}
	switch data.Type() {
	case core.TYPE_VIDEO:
		client.chout <- NewMessage(MessageDesc{Chunk: 6, Timestamp: data.Time(), Stream: client.streamid}, &VideoMessage{
			Data: data.Payload(),
		})
	case core.TYPE_AUDIO:
		client.chout <- NewMessage(MessageDesc{Chunk: 4, Timestamp: data.Time(), Stream: client.streamid}, &AudioMessage{
			Data: data.Payload(),
		})
	}
}

func (client *RTMPClient) Refresh(meta core.Meta) {
	var buf bytes.Buffer
	fw := float64(meta.Width())
	fh := float64(meta.Height())
	ff := float64(meta.FPS())
	amf.EncodeAMF(&buf, "onMetaData")
	amf.EncodeAMF(&buf, struct{
		Width         float64 `name:"width"`
		Height        float64 `name:"height"`
		DisplayWidth  float64 `name:"displayWidth"`
		DisplayHeight float64 `name:"displayHeight"`
		Duration      float64 `name:"duration"`
		Framerate     float64 `name:"framerate"`

	}{fw, fh, fw, fh, 0, ff})
	client.chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: client.streamid}, &AmfMetaMessage{Data: buf.Bytes()})
}

func handleconn(wgroup syncutil.SyncService, registry core.Registry, conn net.Conn) {
	defer conn.Close()
	defer wgroup.Done()
	logger.Info("Client connected")
	reftime, err := Handshake(conn)
	if err != nil {
		logger.Info("Handshake failed")
		return
	}
	comms := NewChunkedComms(registry, reftime, conn)
	chin := make(chan Message, 128)
	chout := make(chan Message, 128)
	comms.Client = &RTMPClient{
		chout: chout,
	}
	go handlemsg(comms, wgroup, chin, chout)
	for wgroup.Running() {
		more := true
		for more {
			select {
			case msg := <-chout:
				logger.Debugf("<- [%d:%d][%d][%s]",
					msg.ChunkID(),
					msg.StreamID(),
					msg.Timestamp(),
					MessageTypeNames[msg.Type()])
				WriteNext(comms, msg)
			default:
				more = false
			}
		}
		comms.Conn.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Millisecond))
		err = ReadNext(comms, chin)
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			break
		}
	}
	stream := comms.Client.(*RTMPClient).stream
	if stream != nil {
		stream.Remove(comms.Client)
	}
	comms.Client.(*RTMPClient).closed = true
	logger.Info("Client disconnected")
}

func handlemsg(comms *ChunkedComms, wgroup syncutil.SyncService, chin chan Message, chout chan Message) {
	logger.Info("Message handler started")
	for wgroup.Running() {
		select {
		case inmsg := <- chin:
			if inmsg == nil {
				continue
			}
			logger.Debugf("-> [%d:%d][%d][%s]",
				inmsg.ChunkID(),
				inmsg.StreamID(),
				inmsg.Timestamp(),
				MessageTypeNames[inmsg.Type()])
			switch inmsg.Type() {
			case MESSAGE_TYPE_WINACK:
				comms.AckIn = inmsg.(*WinAckMessage).Size
			case MESSAGE_TYPE_AMF0_CMD:
				if err := handleCmd(comms, chout, inmsg.(*AmfCmdMessage)); err != nil {
					logger.Error(err)
				}
			case MESSAGE_TYPE_AMF0_META:
				if err := handleMeta(comms, chout, inmsg.(*AmfMetaMessage)); err != nil {
					logger.Error(err)
				}
			case MESSAGE_TYPE_USER:
				usrmsg := inmsg.(*UserMessage)
				logger.Debugf("^ [USER][%s] %d %d", UserEventNames[usrmsg.Event], usrmsg.First, usrmsg.Second);
				if err := handleUser(comms, chout, usrmsg); err != nil {
					logger.Error(err)
				}
			case MESSAGE_TYPE_AUDIO:
				comms.Client.(*RTMPClient).stream.Broadcast(core.NewData(core.TYPE_AUDIO, inmsg.Timestamp(), inmsg.(*AudioMessage).Data))
			case MESSAGE_TYPE_VIDEO:
				comms.Client.(*RTMPClient).stream.Broadcast(core.NewData(core.TYPE_VIDEO, inmsg.Timestamp(), inmsg.(*VideoMessage).Data))
			}
		}
	}
	logger.Info("Message handler stopped")
}

func handleCmd(comms *ChunkedComms, chout chan Message, msg *AmfCmdMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	name := check.Check1(amf.DecodeAMF(rdr)).(string)
	logger.Debug("^^", name)
	switch name {
	case "connect":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)
		//data := check.Check1(amf.DecodeAMF(rdr)).(amf.AMFMap)

		chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &WinAckMessage{Size: 5000000})
		chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetPeerBandMessage{Size: 5000000})
		//chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetChunkSizeMessage{Size: 128})
		chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &UserMessage{
			Event: USER_EVENT_STREAM_BEGIN,
			First: 1,
		})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, struct{
			FmtVer string  `name:"fmtVer"`
			Caps   float64 `name:"capabilities"`
		}{"FMS/3,0,1,123", 31})
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
			ObjEnc float64 `name:"objectEncoding"`
		}{"status", "NetConnection.Connect.Success", "Connection Success.", 0})
		chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &AmfCmdMessage{Data: buf.Bytes()})
	case "createStream":
		serial := check.Check1(amf.DecodeAMF(rdr)).(float64)
		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "_result")
		amf.EncodeAMF(&buf, serial)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, 1)
		chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &AmfCmdMessage{Data: buf.Bytes()})
	case "play":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		stream := comms.Registry.Stream(streamname)
		streamid := comms.Client.(*RTMPClient).streamid // check if play message id is more useful

		chout <- NewMessage(MessageDesc{Chunk: 2, Timestamp: 0, Stream: 0}, &SetChunkSizeMessage{Size: comms.ChunkOut})
		chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})
		chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: 0}, &UserMessage{Event: USER_EVENT_STREAM_BEGIN, First: streamid})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
		}{"status", "NetStream.Play.Start", "Start live."})
		chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfCmdMessage{Data: buf.Bytes()})

		buf.Reset()
		amf.EncodeAMF(&buf, "|RtmpSampleAccess")
		amf.EncodeAMF(&buf, true)
		amf.EncodeAMF(&buf, true)
		chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfMetaMessage{Data: buf.Bytes()})

		buf.Reset()
		fw := float64(stream.Metadata().Width())
		fh := float64(stream.Metadata().Height())
		ff := float64(stream.Metadata().FPS())
		amf.EncodeAMF(&buf, "onMetaData")
		amf.EncodeAMF(&buf, struct{
			Width         float64 `name:"width"`
			Height        float64 `name:"height"`
			DisplayWidth  float64 `name:"displayWidth"`
			DisplayHeight float64 `name:"displayHeight"`
			Duration      float64 `name:"duration"`
			Framerate     float64 `name:"framerate"`

		}{fw, fh, fw, fh, 0, ff})
		chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfMetaMessage{Data: buf.Bytes()})
		stream.Add(comms.Client)
	case "publish":
		amf.DecodeAMF(rdr) // serial
		amf.DecodeAMF(rdr) // nil
		streamname := check.Check1(amf.DecodeAMF(rdr)).(string)
		comms.Client.(*RTMPClient).stream = comms.Registry.Stream(streamname)
		streamid := comms.Client.(*RTMPClient).streamid

		chout <- NewMessage(MessageDesc{Chunk: 3, Timestamp: 0, Stream: streamid}, &UserMessage{Event: USER_EVENT_STREAM_IS_RECORDED, First: 0})

		var buf bytes.Buffer
		amf.EncodeAMF(&buf, "onStatus")
		amf.EncodeAMF(&buf, 0)
		amf.EncodeAMF(&buf, nil)
		amf.EncodeAMF(&buf, struct{
			Level  string  `name:"level"`
			Code   string  `name:"code"`
			Desc   string  `name:"description"`
		}{"status", "NetStream.Publish.Start", "Start publising."})
		chout <- NewMessage(MessageDesc{Chunk: 5, Timestamp: 0, Stream: streamid}, &AmfCmdMessage{Data: buf.Bytes()})
	}
	return
}

func handleMeta(comms *ChunkedComms, chout chan Message, msg *AmfMetaMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	rdr := bytes.NewReader(msg.Data)
	amf.DecodeAMF(rdr)
	amf.DecodeAMF(rdr)
	metaraw := check.Check1(amf.DecodeAMF(rdr)).(amf.AMFMap)
	meta := core.NewMeta(int(metaraw["width"].(float64)), int(metaraw["height"].(float64)), int(metaraw["framerate"].(float64)))
	comms.Client.(*RTMPClient).stream.Refresh(meta)
	return
}

func handleUser(comms *ChunkedComms, chout chan Message, msg *UserMessage) (err error) {
	defer check.CheckPanicHandler(&err)
	switch msg.Event {
	case USER_EVENT_STREAM_BEGIN:
		comms.Client.(*RTMPClient).streamid = msg.Streamid
	}
	return
}
*/