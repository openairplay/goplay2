package ptp

import "time"

type VirtualClock struct {
	start int64
	diff  time.Duration
	delay int64
}

func NewVirtualClock(delay int64) *VirtualClock {
	return &VirtualClock{start: time.Now().UnixNano(), diff: time.Duration(0), delay: delay}
}

func (v *VirtualClock) Offset(diff time.Duration) {
	v.start += diff.Nanoseconds()
}

func (v *VirtualClock) Now() time.Time {
	return time.Unix(0, time.Now().UnixNano()-v.start-v.delay*1000000)
}
