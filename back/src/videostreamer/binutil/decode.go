package binutil

import (
	"io"
	"videostreamer/check"
	"encoding/binary"
)

func ReadBuf(in io.Reader, size int) (buf []byte) {
	buf = make([]byte, size)
	check.Check1(in.Read(buf))
	return
}

func ReadInt(in io.Reader, size int) (ret int) {
	buf := ReadBuf(in, size)
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[i])
	}
	return
}

func ReadIntLE(in io.Reader, size int) (ret int) {
	buf := ReadBuf(in, size)
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[size-i-1])
	}
	return
}

func ReadDobule64(in io.Reader) (ret float64) {
	check.Check0(binary.Read(in, binary.BigEndian, &ret))
	return
}

func ReadString(in io.Reader, size int) string {
	return string(ReadBuf(in, size))
}