package main

import (
	"videostreamer/rtmp"
	"videostreamer/util/logger"
	"videostreamer/global"
)

func main() {
	logger.Level(logger.LOG_ALL)
	context := global.NewContext()
	rtmp.Server(":1935", context)

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
