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
	Add(Client)
	Remove(Client)
	Broadcast(Data)
	Refresh(Meta)
}

type Registry interface {
	Stream(string) Stream
	Streams()      []Stream
}