package filters

import (
	"goplay2/audio"
	"math"
	"time"
)

type skewData struct {
	lastSampling time.Time
	count        float64
	skewAverage  float64
}

func (s *skewData) calculate(audioClock *audio.Clock, playTime time.Time, timestamp uint32) float64 {
	e0 := float64(audioClock.PacketTime(int64(timestamp)).Sub(s.lastSampling)) / float64(playTime.Sub(s.lastSampling))
	s.skewAverage += (e0 - s.skewAverage) / s.count
	if s.count < 16 {
		s.count += 1
	}
	if time.Now().Sub(s.lastSampling) > 200*time.Millisecond {
		s.lastSampling = time.Now()
	}
	return math.Max(0.99, math.Min(1.01, s.skewAverage))
}

func (s *skewData) reset(anchorTime time.Time) {
	s.lastSampling = anchorTime
	s.count = 1
	s.skewAverage = 0
}
