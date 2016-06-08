package rtmp

import (
	"net"
	"videostreamer/util/logger"
	"videostreamer/global"
	"fmt"
)

func handle(conn net.Conn, context global.Context) {
	defer conn.Close()

	id := context.Identify(conn)

	logger.Debugf("[%s] Client connected", id)
	Handshake(conn)
	logger.Debugf("[%s] Handhaked", id)

	stream := context.Open("abc")
	context.Publish(id, stream)
	fmt.Println(stream.Publisher())
}

func Server(addr string, context global.Context) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error(err)
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error(err)
			return
		}
		go handle(conn, context)
	}
}
