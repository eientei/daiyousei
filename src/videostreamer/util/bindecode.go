package util

import (
	"io"
	"encoding/binary"
	"bytes"
)

func ReadBuf(in io.Reader, size int) (ret []byte, err error) {
	ret = make([]byte, size)
	_, err = in.Read(ret)
	return
}

func DecodeInt(in io.Reader, size int) (ret int, err error) {
	var buf []byte;
	if buf, err = ReadBuf(in, size); err != nil {
		return
	}
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[i])
	}
	return
}

func DecodeDouble(in io.Reader) (ret float64, err error) {
	var (
		buf []byte
	)
	if buf, err = ReadBuf(in, 8); err != nil {
		return
	}
	err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &ret)
	return
}

func DecodeIntLE(in io.Reader, size int) (ret int, err error) {
	var buf []byte
	if buf, err = ReadBuf(in, size); err != nil {
		return
	}
	for i := 0; i < size; i++ {
		ret <<= 8
		ret |= int(buf[size-i-1])
	}
	return
}

func DecodeString(in io.Reader, size int) (ret string, err error)  {
	var buf []byte
	if buf, err = ReadBuf(in, size); err != nil {
		return
	}
	ret = string(buf)
	return
}