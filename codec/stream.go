package codec

import (
	"io"
	"time"
)

type StreamCallback func(out []int16, currentTime time.Duration, outputBufferDacTime time.Duration)

type Stream interface {
	io.Closer
	Init(callBack StreamCallback) error
	Start() error
	Stop() error
	SetVolume(volume float64) error
}
