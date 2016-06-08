package global

type Stream interface {
	Context() Context
	Publisher() ClientPublisher
	Consumers() []ClientConsumer
	SetPublisher(pub ClientPublisher)
	AddConsumer(sub ClientConsumer)
	RemoveConsumer(sub ClientConsumer)
	Broadcast(*Data)
}

type streamImpl struct {
	context   Context
	publisher ClientPublisher
	consumers []ClientConsumer
}

func (stream *streamImpl) Context() Context {
	return stream.context
}

func (stream *streamImpl) Publisher() ClientPublisher {
	return stream.publisher
}

func (stream *streamImpl) Consumers() []ClientConsumer {
	return stream.consumers
}

func (stream *streamImpl) SetPublisher(pub ClientPublisher) {
	stream.publisher = pub
}

func (stream *streamImpl) AddConsumer(sub ClientConsumer) {
	stream.consumers = append(stream.consumers, sub)
}

func (stream *streamImpl) RemoveConsumer(sub ClientConsumer) {
	no := -1
	for i, client := range stream.consumers {
		if (client == sub) {
			no = i
			break
		}
	}
	if no > -1 {
		siz := len(stream.consumers)-1
		stream.consumers[no] = stream.consumers[siz]
		stream.consumers = stream.consumers[:siz]
	}
}

func (stream *streamImpl) Broadcast(data *Data) {
	for _, client := range stream.consumers {
		client.Send(data)
	}
}