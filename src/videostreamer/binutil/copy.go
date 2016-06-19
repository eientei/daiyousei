package binutil

func Dup(buf []byte) (res []byte) {
	res = make([]byte, len(buf))
	copy(res, buf)
	return
}