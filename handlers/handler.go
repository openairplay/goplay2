package handlers

import (
	"github.com/brutella/hc/hap"
	"goplay2/audio"
	"goplay2/homekit"
	"goplay2/ptp"
	"goplay2/rtsp"
	"log"
	"time"
)

type Rstp struct {
	streams map[string]*audio.Server
	clock   *ptp.VirtualClock
}

func NewRstpHandler(clock *ptp.VirtualClock) *Rstp {
	return &Rstp{streams: make(map[string]*audio.Server), clock: clock}
}

func (r *Rstp) OnConnOpen(conn *rtsp.Conn) {
	log.Printf("conn opened")
	conn.SetNetConn(hap.NewConnection(conn.NetConn(), homekit.Server.Context))
}

func (r *Rstp) OnRequest(conn *rtsp.Conn, request *rtsp.Request) {
	log.Printf("request received : %s %s body %d at %s", request.Method, request.URL, len(request.Body), time.Now().Format("2006-01-02T15:04:05.999999999Z07:00"))
}

func (r *Rstp) Handle(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {
	switch req.Method {
	case "GET":
		return r.OnGetWeb(req)
	case "POST":
		return r.OnPostWeb(conn, req)
	case "SETUP":
		return r.OnSetupWeb(req)
	case "GET_PARAMETER":
		return r.OnGetParameterWeb(req)
	case "SET_PARAMETER":
		return r.OnSetParameterWeb(req)
	case "RECORD":
		return r.OnRecordWeb(req)
	case "SETPEERS":
		return r.OnSetPeerWeb(req)
	case "SETRATEANCHORTIME":
		return r.OnSetRateAnchorTime(req)
	case "FLUSHBUFFERED":
		return r.OnFlushBuffered(req)
	case "TEARDOWN":
		return r.OnTeardownWeb(req)
	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}

func (r *Rstp) OnResponse(conn *rtsp.Conn, resp *rtsp.Response) {
	log.Printf("response sent : body %d at %s", len(resp.Body), time.Now().Format("2006-01-02T15:04:05.999999999Z07:00"))
}
