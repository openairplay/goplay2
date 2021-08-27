package filters

import "C"
import (
	"goplay2/audio"
	"goplay2/codec"
	"goplay2/config"
	"time"
)

type ResamplingFilter struct {
	clock    *audio.Clock
	metrics  *config.Metrics
	srcState *SrcState
	skewInfo skewData
}

func NewResamplingFilter(clock *audio.Clock, metrics *config.Metrics) (*ResamplingFilter, error) {
	srcState, err := NewSrcState(codec.OutputChannel)

	if err != nil {
		return nil, err
	}
	return &ResamplingFilter{
		clock:    clock,
		metrics:  metrics,
		srcState: srcState,
		skewInfo: skewData{
			skewAverage: 0,
			count:       1,
		},
	}, nil
}

func (p *ResamplingFilter) Reset(clock *audio.Clock) {
	p.skewInfo.reset(clock.AnchorTime())
}

func (p *ResamplingFilter) Apply(audioStream audio.Stream, samples []int16, nextTime time.Time, _ uint32, startTs uint32) (int, error) {
	driftTime := p.clock.PacketTime(int64(startTs)).Sub(nextTime)
	p.metrics.Drift(driftTime)
	if driftTime < -150*time.Millisecond {
		// drop packet if too old
		return 0, nil
	} else if driftTime > 150*time.Millisecond {
		// add silence if really too young
		return 0, audio.ErrIsEmpty
	}
	skew := p.skewInfo.calculate(p.clock, nextTime, startTs)
	input := make([]int16, len(samples))
	peeked, _ := audioStream.Peek(input)
	inputFramesUsed, outputFramesGen, err := p.srcState.Process(skew, input[:peeked], samples)
	if err != nil {
		panic(err)
	}
	_, err = audioStream.Seek(inputFramesUsed)
	return outputFramesGen, err
}
