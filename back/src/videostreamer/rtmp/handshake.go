package rtmp

import (
	"io"
	"videostreamer/check"
	"videostreamer/binutil"
	"fmt"
	"math/rand"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
)

var (
	clientKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	serverKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	clientKey2    = clientKey[:30]
	serverKey2    = serverKey[:36]
	serverVersion = []byte{
		0x0D, 0x0E, 0x0A, 0x0D,
	}
)

func findDigest(buf []byte, mod int) (offs int) {
	for n := 0; n < 4; n++ {
		offs += int(buf[mod+n])
	}
	offs = (offs % 728) + mod + 4
	dig := makeDigest(buf, clientKey2, offs)
	if bytes.Compare(buf[offs:offs+32], dig) != 0 {
		offs = -1
	}
	return
}

func makeDigest(buf []byte, key []byte, offs int) []byte {
	sign := hmac.New(sha256.New, key)
	if offs >= 0 && offs < len(buf) {
		if offs != 0 {
			check.Check1(sign.Write(buf[:offs]))
		}
		if len(buf) != offs+32 {
			check.Check1(sign.Write(buf[offs+32:]))
		}
	} else {
		check.Check1(sign.Write(buf))
	}
	return sign.Sum(nil)
}

func stage1(buf []byte) (dig []byte) {
	if buf[0] != 0x03 {
		panic(fmt.Errorf("First byte of C0 was %#x instead of 0x03", buf[0]))
	}

	roffs := -1
	if roffs = findDigest(buf[1:], 772); roffs == -1 {
		if roffs = findDigest(buf[1:], 8); roffs == -1 {
			panic(fmt.Errorf("Digest was not found in C0"))
		}
	}
	dig = makeDigest(buf[roffs+1:roffs+1+32], serverKey, -1)
	copy(buf[5:9], serverVersion)
	check.Check1(rand.Read(buf[9:]))
	woffs := 0
	for n := 9; n < 13; n++ {
		woffs += int(buf[n])
	}
	woffs = (woffs % 728) + 12
	copy(buf[woffs+1:], makeDigest(buf[1:], serverKey2, woffs))
	return
}

func stage2(buf []byte, dig []byte) {
	check.Check1(rand.Read(buf))
	copy(buf[1536-32:], makeDigest(buf, dig, 1536-32))
	return
}

func Handshake(rw io.ReadWriter) (err error) {
	defer check.CheckPanicHandler(&err)
	var dig []byte
	buf := binutil.ReadBuf(rw, 1537)
	dig = stage1(buf)
	binutil.WriteBuf(rw, buf)
	stage2(buf[1:], dig)
	binutil.WriteBuf(rw, buf[1:])
	binutil.ReadBuf(rw, 1536)
	return
}