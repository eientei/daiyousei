package core

type Meta interface {
	Width()  int
	Height() int
	FPS()    int
}

const (
	TYPE_VIDEO = 0x01
	TYPE_AUDIO = 0x02
)

type Data interface {
	Type()    int
	Time()    uint32
	Payload() []byte
}

type Client interface {
	Consume(Data)
	Refresh(Meta)
}

type Stream interface {
	Name()     string
	Metadata() Meta
	List()     []Client
	KeyVideo() Data
	KeyAudio() Data
	Add(Client)
	Remove(Client)
	Broadcast(Data, bool)
	Refresh(Meta)
}

type Registry interface {
	Stream(string) Stream
	Streams()      []Stream
}