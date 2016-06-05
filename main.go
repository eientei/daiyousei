package main

import (
	"videostreamer/util/logger"
)

func main() {
	logger.Level(logger.LOG_ALL)
	/*
	logger.Level(logger.LOG_ALL)
	buf := make([]byte, 10240)
	buf[0] = 0x03
	err := rtmp.Handshake(bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
	}
	*/
}
