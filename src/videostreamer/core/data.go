package core

import "videostreamer/binutil"

func NewVideoData(time uint32, data []byte) *VideoData {
	return &VideoData{
		Time: time,
		Data: binutil.Dup(data),
	}
}

func NewAudioData(time uint32, data []byte) *AudioData {
	return &AudioData{
		Time: time,
		Data: binutil.Dup(data),
	}
}

func NewMetaData(width uint32, height uint32, framerate uint32) *MetaData {
	return &MetaData{
		Width:     width,
		Height:    height,
		Framerate: framerate,
	}
}