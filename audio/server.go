package audio

import (
	"encoding/binary"
	"errors"
	"github.com/winlinvip/go-fdkaac"
	"goplay2/globals"
	"goplay2/ptp"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	aacDecoder   *fdkaac.AacDecoder
	ringBuffer   *RingBuffer
	sharedKey    []byte
	inputChannel chan *PCMFrame
	player       *Player
}

func NewServer(clock *ptp.VirtualClock, bufferSize int) *Server {

	/*cookie := []byte{
		0x00, 0x00, 0x00, 0x24, 0x61, 0x6c, 0x61, 0x63, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x60, 0x00, 0x10, 0x28, 0x0a, 0x0e, 0x02, 0x00, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xac, 0x44}

	decoder, err := alac.NewDecoder(cookie)
	if err != nil {
		log.Panicf("alac debugger not available : %v", err)
	}*/

	aacDecoder := fdkaac.NewAacDecoder()

	asc := []byte{0x12, 0x10}
	if err := aacDecoder.InitRaw(asc); err != nil {
		log.Panicf("init decoder failed, err is %s", err)
	}

	inputChannel := make(chan *PCMFrame)

	// Divided by 100 -> average size of a RTP packet
	// TODO : create a circular blocked list for better flush handling
	buffer := NewRingBuffer(inputChannel, bufferSize/100)

	return &Server{
		aacDecoder:   aacDecoder,
		inputChannel: inputChannel,
		ringBuffer:   buffer,
		player:       NewPlayer(clock, buffer),
	}
}

func (s *Server) Setup(sharedKey []byte) (int, error) {
	var err error
	s.sharedKey = sharedKey
	listener, err := net.Listen("tcp", ":")
	if err != nil {
		return -1, err
	}
	go s.ringBuffer.Run()
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
	conn, err := l.Accept()
	if err != nil {
		log.Println("Error accepting: ", err.Error())
	}
	defer conn.Close()

	go func() {
		s.player.Run()
	}()
	for {
		frame, err := s.decodeToPcm(conn)
		if err != nil {
			log.Printf("error copying data into ring buffer %v", err)
			return
		}
		s.inputChannel <- frame

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
	// TODO send message for SKIP to TIMESTAMP (find the sequence and then send the sequence from timetamp in buffer
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.WAIT, Value: networkTime.Unix()}
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.START}
}

func (s *Server) Teardown() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.STOP}
}

func (s *Server) SetRate0() {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.PAUSE}
}

func (s *Server) Flush(sequenceId uint64) {
	s.player.ControlChannel <- globals.ControlMessage{MType: globals.SKIP, Value: int64(sequenceId)}
}

