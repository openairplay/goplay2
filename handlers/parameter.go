package handlers

import (
	"bufio"
	"bytes"
	"goplay2/rtsp"
	"strings"
	"goplay2/config"
	"os"
	"os/exec"
	"math"
	"strconv"
	"io/ioutil"
)

func (r *Rstp) OnGetParameterWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "text/parameters") {

		var body string
		scanner := bufio.NewScanner(bytes.NewReader(req.Body))
		for scanner.Scan() {
			switch scanner.Text() {
			case "volume":
				if _, err := os.Stat("./" + config.Config.DeviceName + "/volume"); err == nil {
					vol, _ := ioutil.ReadFile("./" + config.Config.DeviceName + "/volume")
					body += string(vol)
				} else {
					body += "volume: -999\r\n"
				}
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
		if(strings.Contains(string(req.Body), "volume")) {
			ioutil.WriteFile("./" + config.Config.DeviceName + "/volume", []byte(req.Body), 0644)
			vol, _ := strconv.ParseFloat(strings.Trim(strings.Split(string(req.Body), ":")[1], " \r\n"), 64)
			vol = math.Abs(vol)
			if (vol == 144) {
				vol = 0
			} else {
				vol = math.Floor((30 - vol) / 30 * 100)
			}
			if(config.Config.AlsaMixerName != "disabled") {
				cmd := exec.Command("amixer", "sset", config.Config.AlsaMixerName, strconv.FormatFloat(vol, 'f', 0, 64) + "%")
				cmd.Run()
			}
		}
		return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
	}

	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
