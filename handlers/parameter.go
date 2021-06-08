package handlers

import (
	"bufio"
	"bytes"
	"goplay2/rtsp"
	"strings"
)

func (r *Rstp) OnGetParameterWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "text/parameters") {

		var body string
		scanner := bufio.NewScanner(bytes.NewReader(req.Body))
		for scanner.Scan() {
			switch scanner.Text() {
			case "volume":
				body += "volume: -999\r\n"
			}
		}
		return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
			"Content-Type": rtsp.HeaderValue{"text/parameters"},
		}, Body: []byte(body)}, nil

	}

	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}


func (r *Rstp) OnSetParameterWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "text/parameters") {
		return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
	}

	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
