package ptp

import (
	"fmt"
	"time"
)

type mData struct {
	sendTS    time.Time
	receiveTS time.Time
}

type MeasurementResult struct {
	Delay              time.Duration
	Offset             time.Duration
	ServerToClientDiff time.Duration
	ClientToServerDiff time.Duration
	Timestamp          time.Time
}

type measures struct {
	currentUTCoffset time.Duration
	serverToClient   map[uint16]*mData
	clientToServer   map[uint16]*mData
}

// addSync stores ts and seq of SYNC packet
func (m *measures) addSync(seq uint16, ts time.Time) {
	v, found := m.serverToClient[seq]
	if found {
		v.receiveTS = ts
	} else {
		m.serverToClient[seq] = &mData{receiveTS: ts}
	}
}

// addFollowUp stores ts and seq of FOLLOW_UP packet
func (m *measures) addFollowUp(seq uint16, ts time.Time) {
	v, found := m.serverToClient[seq]
	if found {
		v.sendTS = ts
	} else {
		m.serverToClient[seq] = &mData{sendTS: ts}
	}
}

// addDelayReq stores ts and seq of DELAY_REQ packet
func (m *measures) addDelayReq(seq uint16, ts time.Time) {
	v, found := m.clientToServer[seq]
	if found {
		v.sendTS = ts
	} else {
		m.clientToServer[seq] = &mData{sendTS: ts}
	}
}

// addDelayResp stores ts and seq of DELAY_RESP packet and updates history with latest measures
func (m *measures) addDelayResp(seq uint16, ts time.Time) {

	v, found := m.clientToServer[seq]
	if found {
		v.receiveTS = ts
	} else {
		m.clientToServer[seq] = &mData{receiveTS: ts}
	}
}

// we take last complete sample of sync/followup data and last complete sample of delay req/resp data
// to calculate delay and offset
func (m *measures) latest() (*MeasurementResult, error) {
	var lastServerToClient *mData
	var lastClientToServer *mData
	for _, v := range m.serverToClient {
		if v.receiveTS.IsZero() || v.sendTS.IsZero() {
			continue
		}
		if lastServerToClient == nil || v.receiveTS.After(lastServerToClient.receiveTS) {
			lastServerToClient = v
		}
	}
	for _, v := range m.clientToServer {
		if v.receiveTS.IsZero() || v.sendTS.IsZero() {
			continue
		}
		if lastClientToServer == nil || v.receiveTS.After(lastClientToServer.receiveTS) {
			lastClientToServer = v
		}
	}
	if lastServerToClient == nil {
		return nil, fmt.Errorf("no sync/followup data yet")
	}
	if lastClientToServer == nil {
		return nil, fmt.Errorf("no delay data yet")
	}

	m.serverToClient = map[uint16]*mData{}
	m.clientToServer = map[uint16]*mData{}

	clientToServerDiff := lastClientToServer.receiveTS.Sub(lastClientToServer.sendTS)
	serverToClientDiff := lastServerToClient.receiveTS.Sub(lastServerToClient.sendTS)
	delay := (clientToServerDiff + serverToClientDiff) / 2
	offset := serverToClientDiff - delay
	// or this expression of same formula
	// offset := (serverToClientDiff - clientToServerDiff)/2

	return &MeasurementResult{
		Delay:              delay,
		Offset:             offset,
		ServerToClientDiff: serverToClientDiff,
		ClientToServerDiff: clientToServerDiff,
		Timestamp:          lastClientToServer.receiveTS,
	}, nil
}

func newMeasurements() *measures {
	return &measures{
		serverToClient: map[uint16]*mData{},
		clientToServer: map[uint16]*mData{},
	}
}
