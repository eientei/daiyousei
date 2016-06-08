package global

import (
	"net"
	"videostreamer/util"
)

type Context interface {
	Identify(conn net.Conn) *ID
	Open(name string) Stream
	Close(stream Stream)
	Publish(id *ID, stream Stream) ClientPublisher
	Consume(id *ID, stream Stream) ClientConsumer
}

type ID struct {
	Conn  net.Conn
	Ident string
}

type contextImpl struct {
	streams map[string]Stream
	repoids util.Repoch
}

func (id ID) String() string {
	return id.Ident
}

func NewContext() Context {
	context :=        new(contextImpl)
	context.streams = make(map[string]Stream)
	context.repoids = util.MakeIDChan()
	return context
}

func (context *contextImpl) Identify(conn net.Conn) *ID {
	return &ID{ Conn: conn, Ident: util.GenerateID(context.repoids, conn.RemoteAddr().String()) }
}

func (context *contextImpl) Open(name string) Stream {
	stream := context.streams[name]
	if stream == nil {
		stream = &streamImpl{
			context: context,
		}
		context.streams[name] = stream
	}
	return stream
}

func (context *contextImpl) Close(stream Stream) {
	stream.Publisher().Close()
	for _, client := range stream.Consumers() {
		client.Close()
	}
}

func newClient(context Context, id *ID, stream Stream) Client {
	return &clientImpl{
		context: context,
		stream: stream,
		id: id,
		ch: make(chan *Data),
	}
}

func (context *contextImpl) Publish(id *ID, stream Stream) ClientPublisher {
	client := newClient(context, id, stream).(ClientPublisher)
	stream.SetPublisher(client)
	return client
}

func (context *contextImpl) Consume(id *ID, stream Stream) ClientConsumer {
	client := newClient(context, id, stream).(ClientConsumer)
	stream.AddConsumer(client)
	return client
}