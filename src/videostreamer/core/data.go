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
	payload []byte
}

func NewData(typeid int, payload []byte) Data {
	return &dataImpl{
		typeid: typeid,
		payload:payload,
	}
}

func (meta *dataImpl) Type() int {
	return meta.typeid
}

func (meta *dataImpl) Payload() []byte {
	return meta.payload
}