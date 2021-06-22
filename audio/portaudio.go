//+build !linux

package audio

import (
	"github.com/gordonklaus/portaudio"
)

type PortAudioStream struct {
	out    []int16
	stream *portaudio.Stream
}

func NewStream() Stream {
	return &PortAudioStream{}
}

func (s *PortAudioStream) Init() error {
	var err error
	if err = portaudio.Initialize(); err != nil {
		return err
	}
	s.out = make([]int16, 2048)
	// TODO : get the framePerBuffer from setup
	s.stream, err = portaudio.OpenDefaultStream(0, 2, 44100, 1024, &s.out)
	if err != nil {
		return err
	}
	return nil
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

func (s *PortAudioStream) Write(output []int16) error {
	copy(s.out, output)
	return s.stream.Write()
}