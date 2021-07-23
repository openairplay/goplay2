//+build !linux

package audio

import (
	"github.com/gordonklaus/portaudio"
	"time"
)

type PortAudioStream struct {
	out    []int16
	stream *portaudio.Stream
}

func NewStream() Stream {
	return &PortAudioStream{}
}

func (s *PortAudioStream) Init(callBack func (out []int16, info portaudio.StreamCallbackTimeInfo)) error {
	var err error
	if err = portaudio.Initialize(); err != nil {
		return err
	}
	// TODO : get the framePerBuffer from setup
	s.stream, err = portaudio.OpenDefaultStream(0, 2, 44100, 1024, callBack)
	if err != nil {
		return err
	}
	return nil
}

func (s * PortAudioStream) CurrentTime() time.Duration {
	return s.stream.Time()
}

func (s *PortAudioStream) Close() error {
	err := s.stream.Close()
	if err != nil {
		return err
	}
	return portaudio.Terminate()
}

func (s *PortAudioStream) Start() error {
	return s.stream.Start()
}

func (s *PortAudioStream) Stop() error {
	return s.stream.Stop()
}

func (s *PortAudioStream) SetVolume(volume float64) {
	// to nothing on mac
}