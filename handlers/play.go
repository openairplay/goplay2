package handlers

import (
	"fmt"
	"goplay2/rtsp"
	"howett.net/plist"
	"strings"
)

type params struct {
	SupportedCommand [][]byte `plist:"mrSupportedCommandsFromSender"`
}

type command struct {
	Type   string `plist:"type"`
	Params params `plist:"params"`
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
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}


func (r *Rstp) OnSetRateAnchorTime(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		fmt.Printf("%v\n", content)
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
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnFlushBuffered(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
		}
		fmt.Printf("%v\n", content)
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}
