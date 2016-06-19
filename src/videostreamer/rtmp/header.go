package rtmp

import "fmt"

func (header *Header) String() string {
	return fmt.Sprintf("(%d:%d %d %d)", header.Format, header.ChunkID, header.Timestamp, header.StreamID)
}