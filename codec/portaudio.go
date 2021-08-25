//+build !linux

package codec

import (
	"github.com/gordonklaus/portaudio"
	"math"
	"time"
)

type PortAudioStream struct {
	out              []int16
	stream           *portaudio.Stream
	volumeMultiplier float64
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

func (s *PortAudioStream) SetVolume(volumeLevelDb float64) error {
	s.volumeMultiplier = 1.0 * math.Pow(10, volumeLevelDb/20.0)
	return nil
}

func (s *PortAudioStream) FilterVolume(out []int16) int {
	for index, sample := range out {
		out[index] = int16(float64(sample) * s.volumeMultiplier)
	}
	return len(out)
}
