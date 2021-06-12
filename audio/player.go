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
	ringBuffer     *RingBuffer
	nextStart      int64
}

func NewPlayer(clock *ptp.VirtualClock, ring *RingBuffer) *Player {
	return &Player{
		clock:          clock,
		ControlChannel: make(chan globals.ControlMessage, 100),
		Status:         STOPPED,
		ringBuffer:     ring,
		nextStart:      0,
	}
}

func (p *Player) Run() {
	var err error
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
				p.nextStart = msg.Value
			case globals.SKIP:
				p.skipUntil(msg.Value)
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
				frame := <-p.ringBuffer.outputChannel
				err = binary.Read(bytes.NewReader(frame.pcmData), binary.LittleEndian, out)
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

func (p *Player) skipUntil(sequenceId int64) {
	for frame := range p.ringBuffer.outputChannel {
		if frame.SequenceNumber > uint32(sequenceId) {
			return
		}
	}
}
