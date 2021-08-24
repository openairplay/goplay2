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
			oldPlayTime:  time.Unix(0, 0),
			oldTimeStamp: time.Unix(0, 0),
			skewAverage:  1,
		},
	}, nil
}

func (p *ResamplingFilter) Apply(audioStream audio.Stream, samples []int16, nextTime time.Time, _ uint32, startTs uint32) (int, error) {
	skew := p.skewInfo.calculate(p.clock, nextTime, startTs)
	input := make([]int16, len(samples))
	peeked, _ := audioStream.Peek(input)
	var ratio = 1.0
	if skew < 0.99 {
		ratio = 1.01
	} else if skew > 1.00 {
		ratio = 0.99
	}
	inputFramesUsed, outputFramesGen, err := p.srcState.Process(ratio, input[:peeked], samples)
	if err != nil {
		panic(err)
	}
	_, err = audioStream.Seek(inputFramesUsed)
	return outputFramesGen, err
}
