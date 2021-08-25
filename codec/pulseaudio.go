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
	creation time.Time
	client *pulse.Client
	stream *pulse.PlaybackStream
	sink   *pulse.Sink
	index  int
}

func (s *PaStream) AudioTime() time.Duration {
	return time.Now().Sub(s.creation)
}

func (s *PaStream) FilterVolume(out []int16) int {
	return len(out)
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
		creation: time.Now(),
	}
}

func (s *PaStream) Init(callBack StreamCallback) error {
	var err error
	s.sink, err = s.client.SinkByID(config.Config.PulseSink)
	if err != nil {
		s.sink, err = s.client.DefaultSink()
	}
	if err != nil {
		return err
	}
	pulseAudioCallBack := func(out []int16) (int, error){
		audioTime := s.AudioTime()
		callBack(out, audioTime, audioTime)
		return len(out), nil
	}
	s.stream, err = s.client.NewPlayback(pulse.Int16Reader(pulseAudioCallBack),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(1024),
		pulse.PlaybackSink(s.sink),
	)
	if err != nil {
		return err
	}
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
