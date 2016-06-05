package util

import (
	"io"
	"encoding/binary"
	"bytes"
)

func DecodeBuf(in io.Reader, size int) []byte {
	buf := make([]byte, size)
	in.Read(buf)
	return buf
}

func DecodeInt(in io.Reader, size int) (ret int) {
	buf := DecodeBuf(in, size)
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[i])
	}
	return
}

func DecodeDouble(in io.Reader) float64 {
	var res float64
	binary.Read(bytes.NewBuffer(DecodeBuf(in, 8)), binary.BigEndian, &res)
	return res
}

func DecodeIntLE(in io.Reader, size int) (ret int) {
	buf := DecodeBuf(in, size)
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[size-i-1])
	}
	return
}

func DecodeString(in io.Reader, size int) string {
	return string(DecodeBuf(in, size))
}