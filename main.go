package main

import (
	"videostreamer/logger"
	"videostreamer/syncutil"
	"os"
	"os/signal"
	"videostreamer/core"
	"videostreamer/rtmp"
)

func sigcatch(sig chan os.Signal, latch *syncutil.SyncLatch) {
	if latch.Running {
		signal.Notify(sig, os.Interrupt)
		<-sig
		logger.Newline()
		latch.Terminate()
	}
}

func main() {
	logger.Level(logger.LOG_ALL)
	app := core.NewApplication()
	latch := syncutil.NewSyncLatch()
	go rtmp.Serve(app, latch.SubLatch(), ":1935")

	sig := make(chan os.Signal)
	go sigcatch(sig, latch)
	latch.Await()
	latch.Complete()
	close(sig)
}
