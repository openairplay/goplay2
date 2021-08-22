package audio

import "time"

type Filter interface {
	Apply(nextTime time.Time, sequence uint32, startTs uint32) TimingDecision
}
