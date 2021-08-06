package audio

import (
	"goplay2/config"
	"goplay2/globals"
	"time"
)

type FilterSync struct {
	clock    *Clock
	metrics  *config.Metrics
	untilSeq uint32

	skewInfo struct {
		oldPlayTime  time.Time
		oldTimeStamp time.Time
		skewAverage  float64
	}
}

func (p *FilterSync) apply(playTime time.Time, sequence uint32, startTs uint32) TimingDecision {
	driftTime := p.clock.PacketTime(int64(startTs)).Sub(playTime)
	p.metrics.Drift(driftTime)
	if sequence <= p.untilSeq || driftTime < -23*time.Millisecond {
		p.metrics.Drop()
		return DISCARD
	} else if driftTime > 23*time.Millisecond {
		p.metrics.Silence()
		return DELAY
	}
	p.skew(playTime, startTs)
	return PLAY
}

func (p *FilterSync) FlushSequence(resetValue uint32) {
	p.untilSeq = resetValue
}

// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.125.8673&rep=rep1&type=pdf
// if improvement are needed : https://hal.archives-ouvertes.fr/hal-02158803/document
func (p *FilterSync) skew(playTime time.Time, timestamp uint32) float64 {
	realTimeStamp := p.clock.PacketTime(int64(timestamp))

	e0 := float64(playTime.Sub(p.skewInfo.oldPlayTime)) / float64(realTimeStamp.Sub(p.skewInfo.oldTimeStamp))
	p.skewInfo.skewAverage += (e0 - p.skewInfo.skewAverage) / 16

	p.skewInfo.oldTimeStamp = realTimeStamp
	p.skewInfo.oldPlayTime = playTime
	globals.MetricLog.Printf("Audio skew : %v  - ts : %v , play : %v\n", p.skewInfo.skewAverage, timestamp, playTime)
	return p.skewInfo.skewAverage
}

func (f * FilterSync) GetLatestAudioSkew() float64 {
	return f.skewInfo.skewAverage
}


