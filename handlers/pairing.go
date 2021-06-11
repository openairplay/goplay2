package handlers

import (
	"bytes"
	"github.com/brutella/hc/crypto"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/util"
	"goplay2/homekit"
	"goplay2/rtsp"
	"howett.net/plist"
	"log"
	"strings"
)

func (r *Rstp) OnPairSetup(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {
	var err error
	var in util.Container
	var out util.Container

	key := conn.NetConn().RemoteAddr().String()
	session := homekit.Server.Context.Get(key).(hap.Session)

	ctrl := session.PairSetupHandler()
	if ctrl == nil {

		if ctrl, err = pair.NewSetupServerController(homekit.Server.Device, homekit.Server.Database); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
		}
		session.SetPairSetupHandler(ctrl)
	}

	if in, err = util.NewTLV8ContainerFromReader(bytes.NewReader(req.Body)); err == nil {
		out, err = ctrl.Handle(in)
	}

	/*var secSession crypto.Cryptographer

	// When key verification is done, switch to a secure session
	// based on the negotiated shared session key
	b := out.GetByte(pair.TagSequence)
	switch pair.VerifyStepType(b) {
	case pair.VerifyStepFinishResponse:
		if secSession, err = crypto.NewSecureSessionFromSharedKey(ctrl.SharedKey()); err == nil {
			session.SetCryptographer(secSession)
		} else {
			return &base.Response{StatusCode: base.StatusInternalServerError}, err
		}
	}*/

	if err != nil {
		return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
	}
	body := out.BytesBuffer().Bytes()

	return &rtsp.Response{StatusCode: rtsp.StatusOK, Body: body}, nil
}

func (r *Rstp) OnPairVerify(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {

	key := conn.NetConn().RemoteAddr().String()
	session := homekit.Server.Context.Get(key).(hap.Session)

	ctlr := session.PairVerifyHandler()
	if ctlr == nil {
		ctlr = pair.NewVerifyServerController(homekit.Server.Database, homekit.Server.Context)
		session.SetPairVerifyHandler(ctlr)
	}

	var err error
	var in util.Container
	var out util.Container
	var secSession crypto.Cryptographer

	if in, err = util.NewTLV8ContainerFromReader(bytes.NewReader(req.Body)); err == nil {
		out, err = ctlr.Handle(in)
	}

	if err != nil {
		return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
	}
	body := out.BytesBuffer().Bytes()

	// When key verification is done, switch to a secure session
	// based on the negotiated shared session key
	b := out.GetByte(pair.TagSequence)
	switch pair.VerifyStepType(b) {
	case pair.VerifyStepFinishResponse:
		if secSession, err = crypto.NewSecureSessionFromSharedKey(ctlr.SharedKey()); err == nil {
			session.SetCryptographer(secSession)
		} else {
			return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
		}
	}

	return &rtsp.Response{StatusCode: rtsp.StatusOK, Body: body}, nil
}

func (r *Rstp) OnPairAdd(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err == nil {
			log.Printf("Concent : %v\n", content)
		}
	}

	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnPairList(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err == nil {
			log.Printf("Concent : %v\n", content)
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}

func (r *Rstp) OnPairConfigure(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content map[string]interface{}
		if _, err := plist.Unmarshal(req.Body, &content); err == nil {
			log.Printf("Concent : %v\n", content)
		}
	}

	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}
