package handlers

import "goplay2/rtsp"

func (r *Rstp) OnPostWeb(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {

	switch req.Path {
	case "pair-setup":
		return r.OnPairSetup(conn, req)
	case "pair-verify":
		return r.OnPairVerify(conn, req)
	case "pair-add":
		return r.OnPairAdd(conn, req)
	case "pair-remove":
		return r.OnPairRemove(conn, req)
	case "pair-list":
		return r.OnPairList(conn, req)
	case "configure":
		return r.OnPairConfigure(req)
	case "fp-setup":
		return r.OnFpSetup(req)
	case "command":
		return r.OnCommand(req)
	case "audioMode":
		return r.OnAudioMode(req)
	case "feedback":
		return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}
