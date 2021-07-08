package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"goplay2/globals"
	"goplay2/ptp"
	"io"
	"log"
	"time"
)

type PlaybackStatus uint8

const (
	STOPPED PlaybackStatus = iota
	STARTED
	PLAYING
)

var underflow = errors.New("audio underflow")

type Stream interface {
	io.Closer
	Init() error
	Write([]int16) error
	Start() error
	Stop() error
}

type Player struct {
	ControlChannel chan globals.ControlMessage
	clock          *ptp.VirtualClock
	Status         PlaybackStatus
	stream         Stream
	ringBuffer     *Ring
	nextStart      int64
}

func NewPlayer(clock *ptp.VirtualClock, ring *Ring) *Player {
	return &Player{
		clock:          clock,
		ControlChannel: make(chan globals.ControlMessage, 100),
		stream:         NewStream(),
		Status:         STOPPED,
		ringBuffer:     ring,
		nextStart:      0,
	}
}

func (p *Player) Run(s *Server) {
	var err error
	defer s.Close()
	if err := p.stream.Init(); err != nil {
		log.Fatalln("Audio Stream init error:", err)
	}
	defer p.stream.Close()
	out := make([]int16, 2048)

	for {
		select {
		case msg := <-p.ControlChannel:
			switch msg.MType {
			case globals.PAUSE:
				if err := p.stream.Stop(); err != nil {
					log.Printf("error stoping audio :%v\n", err)
					return
				}
				p.Status = STOPPED
			case globals.START:
				p.Status = STARTED
			case globals.WAIT:
				p.nextStart = msg.Param1
			case globals.SKIP:
				p.skipUntil(msg.Param1, msg.Param2)
			case globals.STOP:
				log.Printf("Stopping audio player")
				return
			}
		default:
			if p.Status != STOPPED && p.clock.Now().UnixNano() >= p.nextStart {
				if p.Status == STARTED {
					p.Status = PLAYING
					log.Printf("%v Starting streaming with anchor time %v at %v\n", time.Now(), p.nextStart, p.clock.Now().UnixNano())
					err = p.stream.Start()
					if err != nil {
						log.Printf("error while starting flow : %v\n", err)
						return
					}
				}
				frame := p.ringBuffer.Pop()
				err = binary.Read(bytes.NewReader(frame.(*PCMFrame).pcmData), binary.LittleEndian, out)
				if err != nil {
					log.Printf("error reading data : %v\n", err)
					return
				}
				if err = p.stream.Write(out); err != nil {
					log.Printf("error writing audio :%v\n", err)
					if err != underflow {
						return
					}
					if err := p.stream.Stop(); err != nil {
						log.Printf("error stoping audio :%v\n", err)
						return
					}
					p.Status = STARTED
				}
			}
		}
	}
}

func (p *Player) skipUntil(fromSeq int64, UntilSeq int64) {
	if p.Status == PLAYING {
		log.Printf("Stopping streaming for skipping to sequence %v\n", UntilSeq)
		p.stream.Stop()
		p.Status = STOPPED
	}
	p.ringBuffer.Flush(func(value interface{}) bool {
		frame := value.(*PCMFrame)
		return frame.SequenceNumber < uint32(fromSeq) || frame.SequenceNumber > uint32(UntilSeq)
	})
}
