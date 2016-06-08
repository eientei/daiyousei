package global

import (
	"sync"
)

type Client interface {
	Context() Context
	Stream()  Stream
	Close()
	ID()      *ID
}

type ClientPublisher interface {
	Client
	Receive() *Data
}

type ClientConsumer interface {
	Client
	Send(data *Data)
}

type clientImpl struct {
	context Context
	stream  Stream
	id      *ID
	ch      chan *Data
	closer  sync.Once
}

func (client *clientImpl) String() string {
	pref := "Client"
	if (client.stream != nil) {
		if (client.stream.Publisher().ID() == client.id) {
			pref = "Publisher"
		} else {
			pref = "Consumer"
		}
	}
	return pref + "[" + client.id.Ident + "]"
}


func (client *clientImpl) Context() Context {
	return client.context
}

func (client *clientImpl) Stream() Stream {
	return client.stream
}

func (client *clientImpl) Close() {
	client.closer.Do(func() {
		close(client.ch)
		if (client.stream != nil) {
			if (client.stream.Publisher().ID() == client.id) {
				client.stream.SetPublisher(nil)
			} else {
				client.stream.RemoveConsumer(client)
			}
			client.stream = nil
		}
	})
}

func (client *clientImpl) ID() *ID {
	return client.id
}


func (client *clientImpl) Receive() *Data {
	return <- client.ch
}

func (client *clientImpl) Send(data *Data) {
	client.ch <- data
}