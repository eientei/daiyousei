package util

import (
	"bytes"
	"encoding/binary"
	"io"
)

func EncodeInt(out io.Writer, value int, size int) (int, error) {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[size-i-1] = byte(value & 0xff)
		value >>= 8
	}
	return out.Write(buf)
}

func EncodeIntLE(out io.Writer, value int, size int) (int, error) {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(value & 0xff)
		value >>= 8
	}
	return out.Write(buf)
}

func EncodeDouble(out io.Writer, value float64) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, value)
	return out.Write(buf.Bytes())
}

func EncodeBoolean(out io.Writer, value bool) (int, error) {
	num := 0
	if (value) {
		num = 1
	}
	return EncodeInt(out, num, 1)
}

func EncodeString(out io.Writer, value string) (int, error) {
	return out.Write([]byte(value))
}