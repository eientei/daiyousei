package binutil

import (
	"io"
	"videostreamer/check"
	"encoding/binary"
)

func WriteBuf(out io.Writer, value []byte) {
	check.Check1(out.Write(value))
}

func WriteInt(out io.Writer, value int, size int) {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[size-i-1] = byte(value & 0xff)
		value >>= 8
	}
	WriteBuf(out, buf)
}

func WriteIntLE(out io.Writer, value int, size int) {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(value & 0xff)
		value >>= 8
	}
	WriteBuf(out, buf)
}

func WriteDouble64(out io.Writer, value float64) {
	check.Check0(binary.Write(out, binary.BigEndian, value))
}

func WriteBoolean(out io.Writer, value bool) {
	num := 0
	if value {
		num = 1
	}
	WriteInt(out, num, 1)
}

func WriteString(out io.Writer, value string) {
	WriteBuf(out, []byte(value))
}