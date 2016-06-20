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

func (stream *Stream) Publish() {
	for _, c := range stream.Consumers {
		stream.Bootstrap(c)
	}
	stream.Published = true
}

func (stream *Stream) Bootstrap(c Consumer) {
	c.Publish()
	if stream.Metadata != nil {
		c.ConsumeMeta(stream.Metadata)
	}
	if stream.KeyVideo != nil {
		c.ConsumeVideo(stream.KeyVideo)
	}
	if stream.KeyAudio != nil {
		c.ConsumeAudio(stream.KeyAudio)
	}
}

func (stream *Stream) Unpublish() {
	for _, c := range stream.Consumers {
		c.Unpublish()
	}
	stream.Published = false
}

func (stream *Stream) BroadcastVideo(data *VideoData) {
	if stream.KeyVideo != nil {
		stream.KeyVideo.Time = data.Time
	}
	if !stream.Published {
		return
	}
	for _, c := range stream.Consumers {
		c.ConsumeVideo(data)
	}
}

func (stream *Stream) BroadcastAudio(data *AudioData) {
	if stream.KeyAudio != nil {
		stream.KeyAudio.Time = data.Time
	}
	if !stream.Published {
		return
	}
	for _, c := range stream.Consumers {
		c.ConsumeAudio(data)
	}
}

func (stream *Stream) BroadcastMeta(data *MetaData) {
	if !stream.Published {
		return
	}
	for _, c := range stream.Consumers {
		c.ConsumeMeta(data)
	}
}