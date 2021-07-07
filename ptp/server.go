package ptp

import (
	"encoding"
	"github.com/albanseurat/go-ptp"
	"log"
	"net"
	"time"
)

const (
	clockIdentity = 0x007343fffe123456
)

type serverMsg struct {
	LocalTimeStamp time.Time
	Msg            interface{}
	Src            *net.UDPAddr
}

type Server struct {
	genConn      *net.UDPConn
	eventConn    *net.UDPConn
	measurements *measures
	clock        *VirtualClock
}

func NewServer(clock *VirtualClock) *Server {
	return &Server{
		measurements: newMeasurements(),
		clock:        clock,
	}
}

func NewDelayRequest(sequenceId uint16) *ptp.DelReqMsg {
	return &ptp.DelReqMsg{
		Header: ptp.Header{
			Flags: ptp.Flags{
				Unicast:   true,
				TwoSteps:  true,
				TimeScale: true,
			},
			MessageType:      ptp.DelayReqMsgType,
			MessageLength:    ptp.HeaderLen + ptp.DelayReqPayloadLen,
			VersionPTP:       ptp.Version2,
			ClockIdentity:    clockIdentity,
			PortNumber:       1,
			SequenceID:       sequenceId,
			LogMessagePeriod: 0x7f,
		},
		OriginTimestamp: time.Unix(0, 0),
	}

}

func (s *Server) Serve() {
	var err error
	s.genConn, err = net.ListenUDP("udp", &net.UDPAddr{Port: 320})
	if err != nil {
		log.Printf("Error while listening port 320 %v\n", err)
		return
	}
	s.eventConn, err = net.ListenUDP("udp", &net.UDPAddr{Port: 319})
	if err != nil {
		log.Printf("Error while listening port 319 %v\n", err)
		return
	}

	packetChan := make(chan serverMsg)

	go handlePtpPacket(s.clock, s.genConn, packetChan)
	go handlePtpPacket(s.clock, s.eventConn, packetChan)

	for {
		msg := <-packetChan
		ptpMsg := msg.Msg
		switch x := ptpMsg.(type) {
		case bool:
			return
		case *ptp.AnnounceMsg:
			s.measurements.currentUTCoffset = time.Duration(x.CurrentUtcOffset) * time.Second
		case *ptp.SyncMsg:
			s.measurements.addSync(x.SequenceID, msg.LocalTimeStamp)
		case *ptp.FollowUpMsg:
			s.measurements.addFollowUp(x.SequenceID, x.PreciseOriginTimestamp)
			binary, _ := NewDelayRequest(x.SequenceID).MarshalBinary()
			binary[0] |= 0x10 // add transport specific to 1 for Apple ...
			if _, err = s.eventConn.WriteToUDP(binary, msg.Src); err != nil {
				log.Printf("Error writing UDP packet %v\n", err)
				return
			}
			s.measurements.addDelayReq(x.SequenceID, s.clock.Now())
		case *ptp.DelRespMsg:
			s.measurements.addDelayResp(x.SequenceID, x.ReceiveTimestamp)
			if result, err := s.measurements.latest(); err == nil {
				s.clock.Offset(result.Offset)
			}
		}
	}
}

func handlePtpPacket(clock *VirtualClock, conn *net.UDPConn, packetChan chan serverMsg) {
	var udpPacket [1024]byte
	var msg encoding.BinaryUnmarshaler
	defer conn.Close()
	defer func() {
		packetChan <- serverMsg{Msg: false}
	}()

	for {
		msgLen, addr, err := conn.ReadFromUDP(udpPacket[:])
		receivedTime := clock.Now()
		if err != nil {
			log.Printf("Error reading UDP packet %v\n", err)
			return
		}
		header := &ptp.Header{}
		if err = parsePacket(udpPacket[:ptp.HeaderLen], header); err != nil {
			return
		}
		switch header.MessageType {
		case ptp.AnnounceMsgType:
			msg = &ptp.AnnounceMsg{}
		case ptp.SignalingMsgType:
			continue
		case ptp.SyncMsgType:
			msg = &ptp.SyncMsg{}
		case ptp.FollowUpMsgType:
			msg = &ptp.FollowUpMsg{}
			msgLen = ptp.HeaderLen + ptp.FollowUpPayloadLen
		case ptp.DelayRespMsgType:
			msg = &ptp.DelRespMsg{}
			msgLen = ptp.HeaderLen + ptp.DelayRespPayloadLen
		}
		if err := parsePacket(udpPacket[:msgLen], msg); err != nil {
			return
		}
		packetChan <- serverMsg{Msg: msg, Src: addr, LocalTimeStamp: receivedTime}
	}
}

func parsePacket(packet []byte, unmarshaler encoding.BinaryUnmarshaler) error {
	if err := unmarshaler.UnmarshalBinary(packet); err != nil {
		log.Printf("Error parsing PTP %v\n", err)
		return err
	}
	return nil
}
