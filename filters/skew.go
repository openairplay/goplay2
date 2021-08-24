package filters

import (
	"goplay2/audio"
	"goplay2/globals"
	"time"
)

type skewData struct {
	lastSampling time.Time
	oldPlayTime  time.Time
	oldTimeStamp time.Time
	skewAverage  float64
}

// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.125.8673&rep=rep1&type=pdf
// if improvement are needed : https://hal.archives-ouvertes.fr/hal-02158803/document
func (s *skewData) calculate(audioClock *audio.Clock, playTime time.Time, timestamp uint32) float64 {
	if time.Now().Sub(s.lastSampling) > 200*time.Millisecond {
		s.lastSampling = time.Now()
		realTimeStamp := audioClock.PacketTime(int64(timestamp))
		e0 := float64(playTime.Sub(s.oldPlayTime)) / float64(realTimeStamp.Sub(s.oldTimeStamp))
		s.skewAverage += (e0 - s.skewAverage) / 16
		s.oldTimeStamp = realTimeStamp
		s.oldPlayTime = playTime
		globals.MetricLog.Printf("Audio skew : %v  - ts : %v , play : %v\n", s.skewAverage, timestamp, playTime)
	}
	return s.skewAverage
}
