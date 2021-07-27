//+build !linux

package codec

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

func (s *PortAudioStream) Init(callBack StreamCallback) error {
	var err error
	if err = portaudio.Initialize(); err != nil {
		return err
	}
	portAudioCallback := func(out []int16, info portaudio.StreamCallbackTimeInfo) {
		callBack(out, info.CurrentTime, info.OutputBufferDacTime)
	}
	//TODO : get the framePerBuffer from setup
	s.stream, err = portaudio.OpenDefaultStream(0, OutputChannel, SampleRate, 1024, portAudioCallback)
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

func (s *PortAudioStream) AudioTime() time.Duration {
	return s.stream.Time()
}

func (s *PortAudioStream) SetVolume(_ float64) error {
	return nil
	// to nothing on mac
}
