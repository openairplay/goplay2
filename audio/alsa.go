//+build linux

package audio

import (
	"errors"
	"github.com/albanseurat/goalsa"
	"goplay2/config"
)

type AlsaSteam struct {
	out    []int16
	device *alsa.PlaybackDevice
}

func NewStream() Stream {
	return &AlsaSteam{}
}

func (s *AlsaSteam) Init() error {
	var err error
	if s.device, err = alsa.NewPlaybackDevice(config.Config.AlsaPortName, 2, alsa.FormatS16LE, 44100, alsa.BufferParams{}); err != nil {
		return err
	}
	return nil
}

func (s *AlsaSteam) Close() error {
	s.device.Close()
	return nil
}

func (s *AlsaSteam) Start() error {
	return s.device.Prepare()
}

func (s *AlsaSteam) Stop() error {
	return s.device.Drop()
}

func (s *AlsaSteam) Write(output []int16) error {
	ret, err := s.device.Write(output)
	if err == alsa.ErrUnderrun {
		return underflow
	}
	if err != nil {
		return err
	}
	if ret != len(output) {
		return errors.New("number of bytes written is not good")
	}
	return nil
}

// TODO : use real alsa function rather than launching a command
func (s *AlsaSteam) SetVolume(volume float64) {
	vol := math.Abs(volume)
	if vol == 144 {
		vol = 0
	} else {
		vol = math.Floor((30 - vol) / 30 * 100)
	}
	if config.Config.AlsaMixerName != "disabled" {
		cmd := exec.Command("amixer", "sset", config.Config.AlsaMixerName, strconv.FormatFloat(vol, 'f', 0, 64)+"%")
		cmd.Run()
	}
}