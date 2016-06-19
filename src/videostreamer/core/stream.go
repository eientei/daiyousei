package core

func (stream *Stream) Subscribe(consumer Consumer) {
	stream.Consumers = append(stream.Consumers, consumer)
}

func (stream *Stream) Unsubscribe(consumer Consumer) {
	l := len(stream.Consumers)-1
	for i, c := range stream.Consumers {
		if c == consumer {
			stream.Consumers[i] = stream.Consumers[l]
			stream.Consumers = stream.Consumers[:l]
			break
		}
	}
}

func (stream *Stream) BroadcastVideo(data *VideoData) {
	for _, c := range stream.Consumers {
		c.ConsumeVideo(data)
	}
}

func (stream *Stream) BroadcastAudio(data *AudioData) {
	for _, c := range stream.Consumers {
		c.ConsumeAudio(data)
	}
}

func (stream *Stream) BroadcastMeta(data *MetaData) {
	for _, c := range stream.Consumers {
		c.ConsumeMeta(data)
	}
}