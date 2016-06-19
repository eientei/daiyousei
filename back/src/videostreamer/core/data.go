package core

type metaImpl struct {
	width  int
	height int
	fps    int
}

func NewMeta(width int, height int, fps int) Meta {
	return &metaImpl{
		width: width,
		height: height,
		fps: fps,
	}
}

func (meta *metaImpl) Width() int {
	return meta.width
}

func (meta *metaImpl) Height() int {
	return meta.height
}

func (meta *metaImpl) FPS() int {
	return meta.fps
}


type dataImpl struct {
	typeid  int
	time    uint32
	payload []byte
}

func NewData(typeid int, time uint32, payload []byte) Data {
	return &dataImpl{
		typeid: typeid,
		time: time,
		payload:payload,
	}
}

func (data *dataImpl) Type() int {
	return data.typeid
}

func (data *dataImpl) Time() uint32 {
	return data.time;
}

func (data *dataImpl) Payload() []byte {
	return data.payload
}