package core

type registryImpl struct {
	streams map[string]Stream
}

func NewRegistry() Registry {
	return &registryImpl{}
}

func (registry *registryImpl) Stream(name string) Stream {
	stream, ok := registry.streams[name]
	if !ok {
		stream = NewStream(name)
		registry.streams[name] = stream
	}
	return stream
}

func (registry *registryImpl) Streams(name string) (streams []Stream) {
	for _, v := range registry.streams {
		streams = append(streams, v)
	}
	return
}