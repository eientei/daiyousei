package rtmp

import (
	"io"
	"videostreamer/util"
	"fmt"
	"videostreamer/util/logger"
	"crypto/hmac"
	"crypto/sha256"
	"bytes"
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
	clientKey2 = clientKey[:30]
	serverKey2 = serverKey[:36]
	serverVersion = []byte{
		0x0D, 0x0E, 0x0A, 0x0D,
	}
)

func makeDigest(buf []byte, key []byte, offs int) []byte {
	sign := hmac.New(sha256.New, key)
	if offs >= 0 && offs < len(buf) {
		if offs != 0 {
			sign.Write(buf[:offs])
		}
		if len(buf) != offs + 32 {
			sign.Write(buf[offs+32:])
		}
	} else {
		sign.Write(buf)
	}
	return sign.Sum(nil)
}

func findDigest(buf []byte, mod int) (offs int) {
	for n := 0; n < 4; n++ {
		offs += int(buf[mod + n])
	}

	offs = (offs % 728) + mod + 4

	dig := makeDigest(buf, clientKey2, offs)
	if bytes.Compare(buf[offs:offs+32], dig) != 0 {
		offs = -1
	}
	return
}

func stage0(in io.Reader) (dig []byte, err error) {
	var (
		C0   []byte
		time []byte
		vers []byte
		offs int
	)

	C0, err = util.ReadBuf(in, 1537)
	if err != nil {
		return
	}
	if (C0[0] != 0x03) {
		err = fmt.Errorf("First byte of C0 was %v instead of 0x03", C0[0])
		return
	}
	time = C0[1:5]
	vers = C0[5:9]
	logger.Debugf("Handshake C0 received with time = %v and version = %v", time, vers)

	if offs = findDigest(C0[1:], 772); offs == -1 {
		if offs = findDigest(C0[1:], 8); offs == -1 {
			err = fmt.Errorf("handshake: digest was not found")
			return
		}
	}

	logger.Debugf("Handshake offset: %v", offs)
	dig = makeDigest(C0[offs+1:offs+1+32], serverKey, -1)
	return
}

func Handshake(in io.Reader) (err error) {
	logger.Debug("Handshake requested")
	var (
		dig []byte
	)
	if dig, err = stage0(in); err != nil {
		return
	}
	logger.Debug(dig)
	return
}