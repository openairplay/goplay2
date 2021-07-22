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
	nextFrame      int64
	syncLogger     *log.Logger
}

func NewPlayer(syncFile io.Writer, clock *ptp.VirtualClock, ring *Ring) *Player {

	return &Player{
		clock:          clock,
		ControlChannel: make(chan globals.ControlMessage, 100),
		stream:         NewStream(),
		Status:         STOPPED,
		ringBuffer:     ring,
		nextStart:      0,
		syncLogger:     log.New(syncFile, "SYNC: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (p *Player) Run() {
	var err error

	syncA := syncAudio{}
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
			}
		default:
			if p.Status != STOPPED && p.clock.Now().UnixNano() >= p.nextStart {
				if p.Status == STARTED {
					p.Status = PLAYING
					p.syncLogger.Printf("%v Starting streaming with anchor time %v at %v\n", time.Now(), p.nextStart, p.clock.Now().UnixNano())
					err = p.stream.Start()
					if err != nil {
						globals.ErrLog.Printf("error while starting flow : %v\n", err)
						return
					}
				}
				frame := p.ringBuffer.Pop().(*PCMFrame)
				// https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.125.8673&rep=rep1&type=pdf
				p.syncLogger.Printf("Frame timestamp : %v , now : %v - skew ratio: %v\n",
					frame.Timestamp, p.clock.Now().UnixNano(), syncA.audioSkew(p.clock.Now().UnixNano(), int64(frame.Timestamp)))
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
		p.syncLogger.Printf("Stopping streaming for skipping to sequence %v\n", UntilSeq)
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
