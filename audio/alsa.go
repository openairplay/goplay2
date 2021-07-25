//+build never

package audio

import (
	"goplay2/alsa"
	"goplay2/config"
	"math"
	"os/exec"
	"strconv"
)

type AlsaSteam struct {
	out    []int16
	device *alsa.PlaybackDevice
}

func NewStream() Stream {
	return &AlsaSteam{}
}

func (s *AlsaSteam) Init(callBack StreamCallback) error {
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

