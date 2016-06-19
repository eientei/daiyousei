package core

type registryImpl struct {
	streams map[string]Stream
}

func NewRegistry() Registry {
	return &registryImpl{
		streams: make(map[string]Stream),
	}
}

func (registry *registryImpl) Stream(name string) Stream {
	stream, ok := registry.streams[name]
	if !ok {
		stream = NewStream(name)
		registry.streams[name] = stream
	}
	return stream
}

func (registry *registryImpl) Streams() (streams []Stream) {
	for _, v := range registry.streams {
		streams = append(streams, v)
	}
	return
}