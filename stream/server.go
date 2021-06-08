package stream

import (
	"errors"
	"goplay2/audio"
	"net"
)

type Server struct {
	Id          string
	audioServer *audio.Server
	listener    net.Listener
}

func (s *Server) Start(sharedKey []byte) (int, error) {
	var err error
	s.listener, err = net.Listen("tcp", ":")
	if err != nil {
		return -1, err
	}
	go func() {
		s.audioServer.Listen(sharedKey, s.listener)
	}()
	switch a := s.listener.Addr().(type) {
	case *net.TCPAddr:
		return a.Port, nil
	}
	return -1, errors.New("port not defined")
}

func NewServer(id string) *Server {
	return &Server{Id: id, audioServer: audio.NewServer()}
}

func (s *Server) Close() error {
	return s.listener.Close()
}
