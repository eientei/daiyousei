package util

import (
	"strings"
	"strconv"
	"encoding/hex"
)

type Repoch chan Handlerback

type Handlerback struct {
	Back chan string
	Addr string
}

func convert(addr string) (str string, err error) {
	var (
		val  int
	)
	colon := strings.Index(addr, ":")
	parts := strings.Split(addr[:colon], ".")
	buf := make([]byte, len(parts))
	for i, part := range parts {
		val, err = strconv.Atoi(part)
		buf[i] = byte(val)
	}
	str = hex.EncodeToString(buf)
	return
}

func repohandler(repoch Repoch) {
	clients := make(map[string]int)
	for {
		ch := <-repoch
		con, _ := convert(ch.Addr)
		clients[con]++
		ch.Back <- con + "#" + strconv.Itoa(clients[con])
	}
}

func GenerateID(repoch Repoch, str string) string {
	ch := make(chan string)
	defer close(ch)
	repoch <- Handlerback{Back: ch, Addr: str}
	return <- ch
}

func MakeIDChan() Repoch {
	repoch := make(Repoch)
	go repohandler(repoch)
	return repoch
}