package audio

import (
	"goplay2/ptp"
	"time"
)

type Clock struct {
	virtualClock        *ptp.VirtualClock
	firstFrameTime      int64
	firstFrameTimestamp int64
	previousDrift       int64
}

func (c *Clock) NowMediaTime() time.Time {
	return c.virtualClock.Now()
}
