package filters

import (
	"goplay2/audio"
	"goplay2/config"
	"goplay2/globals"
	"time"
)

type skewData struct {
	oldPlayTime  time.Time
	oldTimeStamp time.Time
	skewAverage  float64
}

type Filter struct {
	clock    *audio.Clock
	metrics  *config.Metrics
	skewInfo skewData
}

func NewFilter(clock *audio.Clock, metrics *config.Metrics) *Filter {
	return &Filter{
		clock:   clock,
		metrics: metrics,
		skewInfo: skewData{
			oldPlayTime:  time.Unix(0, 0),
			oldTimeStamp: time.Unix(0, 0),
			skewAverage:  1,
		},
	}
}

func (p *Filter) AddDrop(nextTime time.Time, sequence uint32, startTs uint32) audio.TimingDecision {
	driftTime := p.clock.PacketTime(int64(startTs)).Sub(nextTime)
	p.metrics.Drift(driftTime)
	if driftTime < -23*time.Millisecond {
		p.metrics.Drop()
		return audio.DISCARD
	} else if driftTime > 23*time.Millisecond {
		p.metrics.Silence()
		return audio.DELAY
	}
	p.skew(nextTime, startTs)
	return audio.PLAY
}

func (p *Filter) Resample(_ time.Time, _ uint32, _ uint32) audio.TimingDecision {
	return audio.PLAY
}

// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.125.8673&rep=rep1&type=pdf
// if improvement are needed : https://hal.archives-ouvertes.fr/hal-02158803/document
func (p *Filter) skew(playTime time.Time, timestamp uint32) float64 {
	realTimeStamp := p.clock.PacketTime(int64(timestamp))
	e0 := float64(playTime.Sub(p.skewInfo.oldPlayTime)) / float64(realTimeStamp.Sub(p.skewInfo.oldTimeStamp))
	p.skewInfo.skewAverage += (e0 - p.skewInfo.skewAverage) / 16
	p.skewInfo.oldTimeStamp = realTimeStamp
	p.skewInfo.oldPlayTime = playTime
	globals.MetricLog.Printf("Audio skew : %v  - ts : %v , play : %v\n", p.skewInfo.skewAverage, timestamp, playTime)
	return p.skewInfo.skewAverage
}
