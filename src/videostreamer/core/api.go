package core

type MetaData struct {
	Width     uint32
	Height    uint32
	Framerate uint32
}

type VideoData struct {
	Time uint32
	Data []byte
}

type AudioData struct {
	Time uint32
	Data []byte
}

type Stream struct {
	Name      string
	Metadata  *MetaData
	Consumers []Consumer
	KeyVideo  *VideoData
	KeyAudio  *AudioData
	Published bool
}

type Application struct {
	Streams map[string]*Stream
}

type Consumer interface {
	ConsumeVideo(*VideoData)
	ConsumeAudio(*AudioData)
	ConsumeMeta(*MetaData)
	Publish()
	Unpublish()
}