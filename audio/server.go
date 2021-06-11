package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/gordonklaus/portaudio"
	"github.com/smallnest/ringbuffer"
	"github.com/winlinvip/go-fdkaac"
	"goplay2/globals"
	"goplay2/ptp"
	"io"
	"log"
	"net"
	"time"
)

type PlaybackStatus uint8

const (
	CLOSED PlaybackStatus = iota
	STARTED
	PAUSED
	SKIPPING
)

type Server struct {
	aacDecoder    *fdkaac.AacDecoder
	clock         *ptp.VirtualClock
	controlChan   chan globals.ControlMessage
	playbackChan  chan globals.ControlMessage
	ringBuffer    *BlockingRingBuffer
	sharedKey     []byte
	timerChannel  *time.Timer
	status        PlaybackStatus
	stream        *portaudio.Stream
	SkippingValue uint32
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

	timer := time.NewTimer(time.Hour)
	timer.Stop()

	return &Server{
		aacDecoder:   aacDecoder,
		clock:        clock,
		controlChan:  make(chan globals.ControlMessage),
		playbackChan: make(chan globals.ControlMessage),
		ringBuffer:   NewBlockingReader(bufferSize),
		timerChannel: timer,
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
	go func() {
		s.play()
	}()
	conn, err := l.Accept()
	if err != nil {
		log.Println("Error accepting: ", err.Error())
	}
	defer conn.Close()
	for {
		select {
		case <-s.controlChan:
			s.ringBuffer.Close()
			log.Println("closing control")
			return
		default:
			if _, err := io.Copy(s.ringBuffer, conn); err != nil && err != ringbuffer.ErrIsFull {
				log.Printf("error copying data into ring buffer %v", err)
				return
			}
		}
	}
}

func (s *Server) decodeToPcm(reader io.Reader) (*PCMFrame, error) {
	var packetSize uint16
	for {
		err := binary.Read(reader, binary.BigEndian, &packetSize)
		if err != nil {
			return nil, err
		}
		buffer := make([]byte, packetSize-2)
		if _, err := io.ReadFull(reader, buffer); err != nil {
			return nil, err
		}
		packet, err := NewFrame(s.aacDecoder, s.sharedKey, buffer)
		if err != nil {
			return nil, err
		}
		if packet.SequenceNumber > s.SkippingValue {
			return packet, nil
		}
	}
}

func (s *Server) play() {
	var err error
	if err := portaudio.Initialize(); err != nil {
		log.Fatalln("PortAudio init error:", err)
	}
	defer portaudio.Terminate()

	out := make([]int16, 2048)
	// TODO : get the framePerBuffer from setup
	s.stream, err = portaudio.OpenDefaultStream(0, 2, 44100, 1024, &out)
	if err != nil {
		log.Println("PortAudio Stream opened false ", err)
		return
	}
	defer s.stream.Close()

	for {
		pcmFrame, err := s.decodeToPcm(s.ringBuffer)
		if err != nil {
			log.Printf("PCM Decode Error : %v\n", err)
			return
		}

		err = binary.Read(bytes.NewReader(pcmFrame.pcmData), binary.LittleEndian, out)
		if err != nil {
			log.Printf("error reading data : %v\n", err)
			return
		}

		switch s.status {
		case CLOSED:
			s.status = STARTED
			<-s.timerChannel.C
			err = s.stream.Start()
			if err != nil {
				log.Fatalln(err)
			}
		}

		select {
		case <-s.playbackChan:
			s.status = CLOSED
			if err := s.stream.Stop(); err != nil {
				log.Printf("error stoping audio :%v\n", err)
			}
		default:
			if err = s.stream.Write(); err != nil {
				log.Printf("error writing audio :%v\n", err)
				return
			}
		}

	}
}

func (s *Server) Start(rtpTime uint32, networkTime time.Time) {
	log.Printf("RtpTime : %v, networkTime %v ", rtpTime, networkTime)
	s.timerChannel.Reset(networkTime.Sub(s.clock.Now()))
}

func (s *Server) Pause() {
	s.playbackChan <- globals.PAUSE_STREAM
}

func (s *Server) Stop() {
	s.controlChan <- globals.TEARDOWN
}

func (s *Server) Flush(sequenceNumber uint64) {
	s.SkippingValue = uint32(sequenceNumber)
}
