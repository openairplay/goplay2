//+build linux

package codec

import (
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
	"goplay2/config"
	"log"
	"math"
	"time"
)

const (
	audioBufferSize = 10240
	rtpPacketSize   = 2048

	paVolumeMuted = 0
	paVolumeNorm  = 0x10000
	paVolumeMax   = math.MaxUint32 / 2
)

type PaStream struct {
	client *pulse.Client
	stream *pulse.PlaybackStream
	sink   *pulse.Sink
	buffer []int16
	index  int
}

func dbToLinearVolume(volume float64) uint32 {
	if math.IsInf(volume, -1) || volume <= math.Inf(-1) {
		return paVolumeMuted
	}
	volume = math.Pow(10, volume/20)
	if volume < 0 {
		return paVolumeMuted
	}
	return uint32(math.Max(paVolumeMuted, math.Min(paVolumeMax, math.Round(math.Cbrt(volume)*paVolumeNorm))))
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
		if s.index+rtpPacketSize < audioBufferSize {
			callBack(s.buffer[s.index:s.index+rtpPacketSize], 0*time.Second, 0*time.Second)
			copied = copy(out, s.buffer[:s.index+rtpPacketSize])
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
	s.sink, err = s.client.SinkByID(config.Config.PulseSink)
	if err != nil {
		s.sink, err = s.client.DefaultSink()
	}
	if err != nil {
		return err
	}
	s.stream, err = s.client.NewPlayback(pulse.Int16Reader(streamCallback),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(1024),
		pulse.PlaybackSink(s.sink),
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

type SetSinkVolume struct {
	SinkIndex uint32
	SinkName  string
	Volume    uint32
}

func (*SetSinkVolume) command() uint32 {
	return proto.OpSetSinkVolume
}

func (s *PaStream) SetVolume(volume float64) error {

	linearVolume := dbToLinearVolume(volume)
	vols := make(proto.ChannelVolumes, 2)

	vols[0] = linearVolume
	vols[1] = linearVolume

	return s.client.RawRequest(&proto.SetSinkInputVolume{
		SinkInputIndex: s.stream.StreamInputIndex(),
		ChannelVolumes: vols,
	}, nil)
}
