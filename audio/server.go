package audio

import (
	"encoding/binary"
	"errors"
	"goplay2/globals"
	"goplay2/rtp"
	"io"
	"net"
	"time"
)

type Server struct {
	sharedKey      []byte
	player         *Player
	controlChannel chan interface{}
}

func NewServer(player *Player) *Server {

	return &Server{
		player: player,
	}
}

func (s *Server) Setup(sharedKey []byte) (int, error) {
	var err error
	s.sharedKey = sharedKey
	listener, err := net.Listen("tcp", ":")
	if err != nil {
		return -1, err
	}
	go func() {
		s.control(listener)
	}()
	switch a := listener.Addr().(type) {
	case *net.TCPAddr:
		return a.Port, nil
	}
	return -1, errors.New("port not defined")
}

func (s *Server) control(l net.Listener) {
	defer l.Close()
	defer s.player.Reset()
	conn, err := l.Accept()
	if err != nil {
		globals.ErrLog.Println("Error accepting: ", err.Error())
	}
	defer conn.Close()
	for {
		frame, err := s.decodeToFrame(conn)
		if err != nil {
			globals.ErrLog.Printf("error parsing the packet %v", err)
			return
		}
		s.player.Push(frame)
	}
}

func (s *Server) decodeToFrame(reader io.Reader) (*rtp.Frame, error) {
	var packetSize uint16
	err := binary.Read(reader, binary.BigEndian, &packetSize)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, packetSize-2)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return nil, err
	}
	return rtp.NewFrame(buffer, s.sharedKey)
}

func (s *Server) SetRateAnchorTime(rtpTime uint32, networkTime time.Time) {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.START, Param1: networkTime.UnixNano(), Param2: int64(rtpTime)}
}

func (s *Server) Teardown() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.STOP}
}

func (s *Server) SetRate0() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.PAUSE}
}

func (s *Server) Flush(fromSeq, untilSeq uint64) {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.SKIP, Param1: int64(fromSeq), Param2: int64(untilSeq)}
}
