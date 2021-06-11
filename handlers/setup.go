package handlers

import (
	"goplay2/audio"
	"goplay2/rtsp"
	"howett.net/plist"
	"strings"
)

type setupStreamRequest struct {
	Type            uint8    `plist:"type"`
	SharedKey       [32]byte `plist:"shk"`
	Spf             uint32   `plist:"spf"`
	CompressionType uint32   `plist:"ct"`
	AudioFormat     uint32   `plist:"audioFormat"`
}

type setupRtsp struct {
	Streams []setupStreamRequest `plist:"streams"`
}

type peerInfos struct {
	Addresses []string `plist:"Addresses"`
	Id        string   `plist:"ID"`
}

type setupEventResponse struct {
	EventPort  uint16    `plist:"eventPort"`
	peerInfo   peerInfos `plist:"timingPeerInfo"`
	TimingPort uint16    `plist:"timingPort"`
}

type setupStream struct {
	BufferSize  uint32 `plist:"audioBufferSize"`
	AudioFormat uint32 `plist:"audioFormat"`
	DataPort    uint16 `plist:"dataPort"`
	ControlPort uint16 `plist:"controlPort"`
	TypeStream  uint8  `plist:"type"`
}

type setupSteamsResponse struct {
	Streams []setupStream `plist:"streams"`
}

const BufferSize = 8388608

func (r *Rstp) OnSetupWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content setupRtsp
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		if content.Streams != nil {
			if s, found := r.streams[req.Path]; found {
				port, err := s.Setup(content.Streams[0].SharedKey[:])
				if err != nil {
					return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
				}
				setupStreamsResponse := &setupSteamsResponse{
					Streams: []setupStream{{
						BufferSize: BufferSize,
						DataPort:   uint16(port),
						TypeStream: content.Streams[0].Type,
					}},
				}

				if body, err := plist.Marshal(*setupStreamsResponse, plist.AutomaticFormat); err == nil {
					return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
						"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
					}, Body: body}, nil
				}
			}
		} else {
			r.streams[req.Path] = audio.NewServer(r.clock, BufferSize)

			setupEventResponse := &setupEventResponse{EventPort: 60003, TimingPort: 0}
			if body, err := plist.Marshal(*setupEventResponse, plist.AutomaticFormat); err == nil {
				return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
					"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
				}, Body: body}, nil
			}
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}

func (r *Rstp) OnSetPeerWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}
