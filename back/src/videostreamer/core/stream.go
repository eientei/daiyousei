package core

type streamImpl struct {
	name string
	metadata Meta
	clients []Client
	keyvideo Data
	keyaudio Data
}

func NewStream(name string) Stream {
	return &streamImpl{
		name: name,
		metadata: NewMeta(0, 0, 0),
	}
}

func (stream *streamImpl) Name() string {
	return stream.name
}

func (stream *streamImpl) Metadata() Meta {
	return stream.metadata
}

func (stream *streamImpl) List() []Client {
	return stream.clients
}

func (stream *streamImpl) Add(client Client) {
	stream.clients = append(stream.clients, client)
}

func (stream *streamImpl) Remove(client Client) {
	l := len(stream.clients)-1
	for i, c := range(stream.clients) {
		if c == client {
			stream.clients[i] = stream.clients[l]
			stream.clients = stream.clients[:l]
			break;
		}
	}
}

func (stream *streamImpl) Broadcast(data Data, key bool) {
	if key {
		switch data.Type() {
		case TYPE_VIDEO:
			stream.keyvideo = data
		case TYPE_AUDIO:
			stream.keyaudio = data
		}
	}
	for _, c := range (stream.clients) {
		c.Consume(data)
	}
}

func (stream *streamImpl) Refresh(metadata Meta) {
	stream.metadata = metadata
	for _, c := range (stream.clients) {
		c.Refresh(metadata)
	}
}

func (stream *streamImpl) KeyAudio() Data {
	return stream.keyaudio
}

func (stream *streamImpl) KeyVideo() Data {
	return stream.keyvideo
}