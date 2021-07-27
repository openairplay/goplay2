package audio

import (
	"goplay2/codec"
	"goplay2/ptp"
	"time"
)

type Clock struct {
	networkClock *ptp.VirtualClock
	// audio times
	realAudioAnchorTime time.Time
	audioAnchorTime     time.Duration
	// network times
	realNetworkAnchorTime time.Time
	anchorRtpTime         int64
	networkAnchorTime     int64
}

func NewClock(networkClock *ptp.VirtualClock) *Clock {
	return &Clock{
		networkClock: networkClock,
	}
}

func (c *Clock) virtualToRealTime(elapsed int64) time.Time {
	return time.Unix(0, time.Now().UnixNano()+(elapsed-c.networkClock.Now().UnixNano()))
}

// AnchorTime setup anchor time using the real clock
// virtualAnchorTime - network Time (using PTP) of the anchorTime
// rtpTime Timestamp - monotonic counter of frames
func (c *Clock) AnchorTime(virtualAnchorTime int64, rtpTime int64) {
	c.anchorRtpTime = rtpTime
	c.networkAnchorTime = virtualAnchorTime
	c.realNetworkAnchorTime = c.virtualToRealTime(virtualAnchorTime)
}

// PacketTime returns the real clock time at which time a RTP packet should be played
func (c *Clock) PacketTime(frameRtpTime int64) time.Time {
	elapsedTime := float64(frameRtpTime-c.anchorRtpTime) * 1.0 / float64(codec.SampleRate) * 1e9 // convert to nanoseconds
	return c.virtualToRealTime(c.networkAnchorTime + int64(elapsedTime))
}

func (c *Clock) AudioTime(anchorAudioTime time.Duration, realAnchorTime time.Time) {
	c.audioAnchorTime = anchorAudioTime
	c.realAudioAnchorTime = realAnchorTime
}

func (c *Clock) PlayTime(nowAudioTime time.Duration, playbackTime time.Duration) time.Time {
	now := time.Now()
	return now.Add(playbackTime - nowAudioTime)
}
