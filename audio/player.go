package audio

import (
	"goplay2/codec"
	"goplay2/config"
	"goplay2/globals"
	"goplay2/rtp"
	"log"
	"time"
)

type PlaybackStatus uint8

const (
	STOPPED PlaybackStatus = iota
	PLAYING
)

type Player struct {
	ControlChannel chan globals.ControlMessage
	clock          *Clock
	filter         Filter
	Status         PlaybackStatus
	stream         codec.Stream
	ring           *Ring
	aacDecoder     *codec.AacDecoder
	untilSeq       uint32
}

func NewPlayer(audioClock *Clock, filter Filter) *Player {
	aacDecoder := codec.NewAacDecoder()
	asc := []byte{0x12, 0x10}
	if err := aacDecoder.InitRaw(asc); err != nil {
		globals.ErrLog.Panicf("init decoder failed, err is %s", err)
	}
	player := &Player{
		clock:          audioClock,
		filter:         filter,
		ControlChannel: make(chan globals.ControlMessage, 100),
		aacDecoder:     aacDecoder,
		stream:         codec.NewStream(),
		Status:         STOPPED,
		ring:           New(globals.BufferSize / 2048),
		untilSeq:       0,
	}
	return player
}

func (p *Player) audioSync(reader Stream, samples []int16, nextTime time.Time, sequence uint32, startTs uint32) (int, error) {
	if sequence <= p.untilSeq {
		return 0, nil
	}
	if config.Config.DisableAudioSync {
		return reader.Read(samples)
	} else {
		return p.filter.Apply(reader, samples, nextTime, sequence, startTs)
	}
}

func (p *Player) callBack(out []int16, currentTime time.Duration, outputBufferDacTime time.Duration) (int, error) {
	playTime := p.clock.PlayTime(currentTime, outputBufferDacTime)
	size, err := p.ring.TryRead(out, playTime, p.audioSync)
	if err == ErrIsEmpty {
		p.fillSilence(out)
	} else if size < len(out) {
		p.fillSilence(out[size:])
	}
	return len(out), nil
}

func (p *Player) Run() {
	var err error
	if err := p.stream.Init(p.callBack); err != nil {
		globals.ErrLog.Fatalln("Audio Stream init error:", err)
	}
	defer p.stream.Close()
	p.clock.AudioTime(p.stream.AudioTime(), time.Now())
	for {
		select {
		case msg := <-p.ControlChannel:
			switch msg.MType {
			case globals.STOP:
				if p.Status == PLAYING {
					if err := p.stream.Stop(); err != nil {
						globals.ErrLog.Printf("error pausing audio :%v\n", err)
						return
					}
				}
				p.Reset()
				p.Status = STOPPED
			case globals.PAUSE:
				if p.Status == PLAYING {
					if err := p.stream.Stop(); err != nil {
						globals.ErrLog.Printf("error pausing audio :%v\n", err)
						return
					}
				}
				p.Status = STOPPED
			case globals.START:
				if p.Status == STOPPED {
					err = p.stream.Start()
					if err != nil {
						globals.ErrLog.Printf("error while starting flow : %v\n", err)
						return
					}
				}
				p.Status = PLAYING
				p.clock.AnchorTime(msg.Param1, msg.Param2)
			case globals.SKIP:
				p.skipUntil(msg.Param1, msg.Param2)
			case globals.VOLUME:
				if err := p.stream.SetVolume(msg.Paramf); err != nil {
					globals.ErrLog.Printf("error while setting new volume : %v", err)
				}
			}
		}
	}
}

func (p *Player) skipUntil(fromSeq int64, untilSeq int64) {
	log.Printf("drop from sequence %v to %v\n", fromSeq, untilSeq)
	// TODO : use also timestamp to have better precision
	p.ring.Filter(func(sequence uint32, startTs uint32) bool {
		return sequence > uint32(fromSeq) && sequence < uint32(untilSeq)
	})
	// some data are possibly not yet in the buffer - reader should skip them afterwards (during async callback)
	p.untilSeq = uint32(untilSeq)
}

func (p *Player) Push(frame *rtp.Frame) {
	var pcmBuffer = make([]int16, 2048)
	_, err := frame.PcmData(p.aacDecoder, pcmBuffer)
	if err != nil {
		globals.ErrLog.Printf("error decoding the packet %v", err)
	}
	p.ring.Write(pcmBuffer, frame.SequenceNumber, frame.Timestamp)
}

func (p *Player) Reset() {
	p.ring.Reset()
	p.untilSeq = 0
}

func (p *Player) fillSilence(out []int16) {
	for i := range out {
		out[i] = 0
	}
}
