//+build linux

package audio

import (
	"github.com/jfreymuth/pulse"
	"log"
	"time"
)

const audioBufferSize = 10240
const rtpPacketSize = 2048

type PaStream struct {
	client *pulse.Client
	stream *pulse.PlaybackStream
	buffer []int16
	index  int
}

func NewStream() Stream {
	client, err := pulse.NewClient()
	if err != nil {
		log.Panic(err)
	}
	return &PaStream{
		client: client,
	}
}

func (s *PaStream) Init(callBack StreamCallback) error {
	var err error
	streamCallback := func(out []int16) (int, error) {
		var copied = 0
		if s.index + rtpPacketSize < audioBufferSize {
			callBack(s.buffer[s.index: s.index + rtpPacketSize], 0*time.Second, 0*time.Second)
			copied = copy(out, s.buffer[:s.index + rtpPacketSize])
			s.index += rtpPacketSize - copied
		} else {
			copied = copy(out, s.buffer[:s.index])
			s.index -= copied
		}
		copy(s.buffer, s.buffer[copied:])
		if s.index < 0 {
			s.index = 0
		}
		return copied, nil
	}

	s.stream, err = s.client.NewPlayback(pulse.Int16Reader(streamCallback),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(1024),
	)
	if err != nil {
		return err
	}

	s.buffer = make([]int16, audioBufferSize)

	return nil
}

func (s *PaStream) Close() error {
	s.client.Close()
	return nil
}

func (s *PaStream) Start() error {
	s.stream.Start()
	return nil
}

func (s *PaStream) Stop() error {
	s.stream.Stop()
	return nil
}

func (s *PaStream) SetVolume(volume float64) {

}
