package core

type streamImpl struct {
	name string
	metadata Meta
	clients []Client
}

func NewStream(name string) Stream {
	return &streamImpl{name: name}
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

func (stream *streamImpl) Broadcast(data Data) {
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