package audio

type syncAudio struct {
	previousClock     int64
	previousTimeStamp int64
	weightAverage     float64
}

func (s *syncAudio) audioSkew(clock int64, timestamp int64) float64 {
	var skewRatio float64
	defer func() {
		s.previousClock = clock
		s.previousTimeStamp = timestamp
	}()
	if s.previousClock == 0 {
		return -1
	} else {
		skewRatio = float64(clock-s.previousClock) / (float64(timestamp-s.previousTimeStamp) * 22500.0)
	}
	s.weightAverage += (skewRatio - s.weightAverage) / 16

	return s.weightAverage
}
