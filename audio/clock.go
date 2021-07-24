package audio

import (
	"goplay2/ptp"
	"time"
)

type Clock struct {
	virtualClock   *ptp.VirtualClock
	realAnchorTime time.Time
	anchorRtpTime  int64
	currentRtpTime int64
}

// AnchorTime setup anchor time using the real clock
// virtualAnchorTime - network Time (using PTP) of the anchorTime
// rtpTime Timestamp - monotonic counter of frames
func (c *Clock) AnchorTime(virtualAnchorTime int64, rtpTime int64) {

	c.anchorRtpTime = rtpTime
	c.currentRtpTime = rtpTime

	realNowTime := time.Now()
	virtualNowTime := c.virtualClock.Now()

	c.realAnchorTime = time.Unix(0, realNowTime.UnixNano()+(virtualAnchorTime-virtualNowTime.UnixNano()))
}

// PacketTime returns the real clock time at which time a RTP packet should be played
func (c *Clock) PacketTime(frameRtpTime int64) time.Time {
	return time.Unix(0, 0)
}

func (c *Clock) CurrentRtpTime() int64 {
	return c.currentRtpTime
}

func (c *Clock) IncRtpTime() {
	c.currentRtpTime += 1024
}
