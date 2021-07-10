package handlers

import (
	"fmt"
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/util"
	"goplay2/audio"
	"goplay2/homekit"
	"goplay2/ptp"
	"goplay2/rtsp"
	"log"
)

type Rstp struct {
	streams map[string]*audio.Server
	clock   *ptp.VirtualClock
	pairing *pair.PairingController
}

func NewRstpHandler(deviceName string, clock *ptp.VirtualClock) (*Rstp, error) {

	storage, err := util.NewFileStorage(fmt.Sprintf("%s/db", deviceName))
	if err != nil {
		return nil, err
	}
	ctrl := pair.NewPairingController(db.NewDatabaseWithStorage(storage))
	return &Rstp{streams: make(map[string]*audio.Server), clock: clock, pairing: ctrl}, nil
}

func (r *Rstp) OnConnOpen(conn *rtsp.Conn) {
	log.Printf("conn opened")
	conn.SetNetConn(hap.NewConnection(conn.NetConn(), homekit.Server.Context))
}

func (r *Rstp) OnRequest(conn *rtsp.Conn, request *rtsp.Request) {
	log.Printf("request received : %s %s body %d at %v", request.Method, request.URL, len(request.Body), r.clock.Now().UnixNano())
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
	log.Printf("response sent : body %d at %v", len(resp.Body), r.clock.Now().UnixNano())
}
