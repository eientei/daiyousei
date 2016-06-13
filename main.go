package main

import (
	"videostreamer/logger"
	"videostreamer/rtmp"
	"fmt"
)

func main() {
	logger.Level(logger.LOG_ALL)

	fmt.Println(rtmp.NewMessage(rtmp.MessageDesc{3, 4, 32}, &rtmp.SetChunkSizeMessage{
		Size: 1,
	}))
}
