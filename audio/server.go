package audio

import (
	"encoding/binary"
	"errors"
	aac "github.com/albanseurat/go-fdkaac"
	"goplay2/globals"
	"io"
	"net"
	"time"
)

type Server struct {
	aacDecoder     *aac.AacDecoder
	sharedKey      []byte
	player         *Player
	controlChannel chan interface{}
}

func NewServer(player *Player) *Server {
	aacDecoder := aac.NewAacDecoder()

	asc := []byte{0x12, 0x10}
	if err := aacDecoder.InitRaw(asc); err != nil {
		globals.ErrLog.Panicf("init decoder failed, err is %s", err)
	}

	return &Server{
		aacDecoder: aacDecoder,
		player:     player,
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
		frame, err := s.decodeToPcm(conn)
		if err != nil {
			globals.ErrLog.Printf("error decoding to pcm %v", err)
			return
		}
		s.player.Push(frame)
	}
}

func (s *Server) decodeToPcm(reader io.Reader) (*PCMFrame, error) {
	var packetSize uint16
	err := binary.Read(reader, binary.BigEndian, &packetSize)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, packetSize-2)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return nil, err
	}
	return NewFrame(s.aacDecoder, s.sharedKey, buffer)
}

func (s *Server) SetRateAnchorTime(rtpTime uint32, networkTime time.Time) {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.START, Param1: networkTime.UnixNano(), Param2: int64(rtpTime)}
}

func (s *Server) Teardown() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.PAUSE}
}

func (s *Server) SetRate0() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.PAUSE}
}

func (s *Server) Flush(fromSeq, untilSeq uint64) {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.SKIP, Param1: int64(fromSeq), Param2: int64(untilSeq)}
}
