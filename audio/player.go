package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"goplay2/globals"
	"goplay2/ptp"
	"io"
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
	SetVolume(volume float64)
}

type Player struct {
	ControlChannel chan globals.ControlMessage
	clock          *ptp.VirtualClock
	Status         PlaybackStatus
	stream         Stream
	ringBuffer     *Ring
	nextStart      int64
	nextFrame      int64
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

func (p *Player) Run() {
	var err error
	if err := p.stream.Init(); err != nil {
		globals.ErrLog.Fatalln("Audio Stream init error:", err)
	}
	defer p.stream.Close()
	out := make([]int16, 2048)

	for {
		select {
		case msg := <-p.ControlChannel:
			switch msg.MType {
			case globals.PAUSE:
				if p.Status == PLAYING {
					if err := p.stream.Stop(); err != nil {
						globals.ErrLog.Printf("error pausing audio :%v\n", err)
						return
					}
				}
				p.Status = STOPPED
			case globals.START:
				p.Status = STARTED
				p.nextStart = msg.Param1
				p.nextFrame = msg.Param2
			case globals.SKIP:
				p.skipUntil(msg.Param1, msg.Param2)
			case globals.VOLUME:
				p.stream.SetVolume(msg.Paramf)
			}
		default:
			if p.Status != STOPPED && p.clock.Now().UnixNano() >= p.nextStart {
				if p.Status == STARTED {
					p.Status = PLAYING
					err = p.stream.Start()
					if err != nil {
						globals.ErrLog.Printf("error while starting flow : %v\n", err)
						return
					}
				}
				frame := p.ringBuffer.Pop().(*PCMFrame)
				err = binary.Read(bytes.NewReader(frame.pcmData), binary.LittleEndian, out)
				if err != nil {
					globals.ErrLog.Printf("error reading data : %v\n", err)
					return
				}
				if err = p.stream.Write(out); err != nil {
					globals.ErrLog.Printf("error writing audio :%v\n", err)
					if err := p.stream.Stop(); err != nil {
						globals.ErrLog.Printf("error stopping audio :%v\n", err)
						return
					}
					p.Status = STARTED
				}
			} else {
				// yield while there is no playing
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

func (p *Player) skipUntil(fromSeq int64, UntilSeq int64) {
	if p.Status == PLAYING {
		p.stream.Stop()
		p.Status = STOPPED
	}
	p.ringBuffer.Flush(func(value interface{}) bool {
		frame := value.(*PCMFrame)
		return frame.SequenceNumber < uint32(fromSeq) || frame.SequenceNumber > uint32(UntilSeq)
	})
}

func (p *Player) Push(frame interface{}) {
	p.ringBuffer.Push(frame)
}

func (p *Player) Reset() {
	p.ringBuffer.Reset()
}
