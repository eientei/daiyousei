package global


const (
	DATA_VIDEO = 0x01
	DATA_AUDIO = 0x02
)

type MetaData struct {
	Width  int
	Height int
	Name   string
	Rate   int
}

type Data struct {
	Type    int
	Format  int
	Meta    MetaData
	Payload []byte
}
