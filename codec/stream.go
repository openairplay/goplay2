package codec

import (
	"io"
	"time"
)

const (
	OutputChannel = 2
	SampleRate    = 44100
)

type StreamCallback func(out []int16, currentTime time.Duration, outputBufferDacTime time.Duration) (int, error)

type Stream interface {
	io.Closer
	Init(callBack StreamCallback) error
	Start() error
	Stop() error
	SetVolume(volume float64) error
	AudioTime() time.Duration
}
