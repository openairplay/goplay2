package audio

import (
	"bytes"
	"encoding/binary"
	"github.com/gordonklaus/portaudio"
	"goplay2/globals"
	"goplay2/ptp"
	"log"
)

type PlaybackStatus uint8

const (
	STOPPED PlaybackStatus = iota
	STARTED
	PLAYING
)

type Player struct {
	ControlChannel chan globals.ControlMessage
	clock          *ptp.VirtualClock
	Status         PlaybackStatus
	stream         *portaudio.Stream
	ringBuffer     *Ring
	nextStart      int64
}

func NewPlayer(clock *ptp.VirtualClock, ring *Ring) *Player {
	return &Player{
		clock:          clock,
		ControlChannel: make(chan globals.ControlMessage, 100),
		Status:         STOPPED,
		ringBuffer:     ring,
		nextStart:      0,
	}
}

func (p *Player) Run(s *Server) {
	var err error
	defer s.Close()
	if err := portaudio.Initialize(); err != nil {
		log.Fatalln("PortAudio init error:", err)
	}
	defer portaudio.Terminate()

	out := make([]int16, 2048)
	// TODO : get the framePerBuffer from setup
	p.stream, err = portaudio.OpenDefaultStream(0, 2, 44100, 1024, &out)
	if err != nil {
		log.Println("PortAudio Stream opened false ", err)
		return
	}
	defer p.stream.Close()

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
				return
			}
		default:
			if p.Status != STOPPED && p.clock.Now().Unix() >= p.nextStart {
				if p.Status == STARTED {
					p.Status = PLAYING
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
				if err = p.stream.Write(); err != nil {
					log.Printf("error writing audio :%v\n", err)
					return
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
