package handlers

import (
	"goplay2/rtsp"
	"howett.net/plist"
	"log"
	"math"
	"math/big"
	"strings"
	"time"
)

type params struct {
	SupportedCommand [][]byte `plist:"mrSupportedCommandsFromSender"`
}

type command struct {
	Type   string `plist:"type"`
	Params params `plist:"params"`
}

type setRateAnchorTime struct {
	Rate          uint32 `plist:"rate"`
	RtpTime       uint32 `plist:"rtpTime"`
	Fraction      uint64 `plist:"networkTimeFrac"`
	Seconds       uint64 `plist:"networkTimeSecs"`
	ClockIdentity uint64 `plist:"networkTimeId"`
}

func (s *setRateAnchorTime) StartTime() time.Time {
	mantissa := big.NewFloat(float64(s.Fraction))
	result := big.NewFloat(0).SetMantExp(mantissa, -64)
	result.Mul(result, big.NewFloat(math.Pow(10, 9)))
	value, _ := result.Int64()
	return time.Unix(int64(s.Seconds), value)
}

func (r *Rstp) OnRecordWeb(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnAudioMode(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		log.Printf("%v\n", content)
	}

	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnSetRateAnchorTime(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content setRateAnchorTime
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		if s, found := r.streams[req.Path]; found {
			if content.Rate == 0 {
				s.SetRate0()
			} else {
				s.SetRateAnchorTime(content.RtpTime, content.StartTime())
			}
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnCommand(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content command
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		for _, value := range content.Params.SupportedCommand {
			var commandContent map[string]interface{}
			if _, err := plist.Unmarshal(value, &commandContent); err != nil {
				return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
			}
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnTeardownWeb(req *rtsp.Request) (*rtsp.Response, error) {
	if s, found := r.streams[req.Path]; found {
		s.Teardown()
		delete(r.streams, req.Path)
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnFlushBuffered(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		if s, found := r.streams[req.Path]; found {
			if fromSeq, found := content["flushFromSeq"] ; found {
				s.Flush(fromSeq.(uint64), content["flushUntilSeq"].(uint64))
			} else {
				s.Flush(0, content["flushUntilSeq"].(uint64))
			}
		}
		log.Printf("%v\n", content)
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}
