package audio

import (
	"container/list"
	"errors"
	"goplay2/codec"
	"sync"
	"time"
)

var (
	ErrIsFull    = errors.New("ring is full")
	ErrIsEmpty   = errors.New("ring is empty")
	ErrIsPartial = errors.New("buffer is partial")
)

type markedBuffer struct {
	sequence uint32
	startTs  uint32
	buffer   []int16
}

func (b *markedBuffer) len() int {
	return len(b.buffer)
}

func (b *markedBuffer) data() []int16 {
	return b.buffer
}

func (b *markedBuffer) read(samples []int16) (int, error) {
	copied := copy(samples, b.buffer)
	if copied < len(b.buffer) {
		b.buffer = b.buffer[copied:]
		b.startTs += uint32(copied)
		return copied, ErrIsPartial
	} else {
		return copied, nil
	}
}

type TimingDecision uint8

const (
	PLAY    TimingDecision = iota
	DISCARD                // will drop the frame
	DELAY                  // will play silence
)

type TimingPredicate func(playTime time.Time, sequence uint32, startTs uint32) TimingDecision

// Ring is a circular buffer that implement io.ReaderWriter interface.
type Ring struct {
	buffers *list.List
	size    int
	mu      sync.Mutex
	wcd     *sync.Cond
	rcd     *sync.Cond
}

// New returns a new Ring whose buffer has the given size.
func New(size int) *Ring {
	rwmu := sync.Mutex{}
	return &Ring{
		buffers: list.New(),
		size:    size,
		wcd:     sync.NewCond(&rwmu),
	}
}

func (r *Ring) Write(samples []int16, sequence uint32, ts uint32) {
	err := r.TryWrite(samples, sequence, ts)
	r.wcd.L.Lock()
	for err == ErrIsFull {
		r.wcd.Wait()
		err = r.TryWrite(samples, sequence, ts)
	}
	r.wcd.L.Unlock()
}

func (r *Ring) TryWrite(samples []int16, sequence uint32, ts uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.buffers.Len() == r.size {
		return ErrIsFull
	}
	r.buffers.PushFront(&markedBuffer{sequence: sequence, startTs: ts, buffer: samples})
	return nil
}

func (r *Ring) TryRead(samples []int16, playTime time.Time, predicate TimingPredicate) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.buffers.Len() == 0 {
		return 0, ErrIsEmpty
	}
	n := 0
	var err error = nil
	var size int
	for r.buffers.Len() > 0 && n < len(samples) {
		back := r.buffers.Back()
		elem := back.Value.(*markedBuffer)
		command := predicate(playTime, elem.sequence, elem.startTs)
		playTime.Add(time.Duration(n*1e9/codec.SampleRate) * time.Nanosecond)
		if command == PLAY {
			size, err = elem.read(samples[n:])
			n += size
		} else if command == DELAY {
			return 0, ErrIsEmpty
		}
		if err == nil {
			r.buffers.Remove(back)
		}
	}
	r.wcd.Signal()
	return n, nil
}

// Reset the read pointer and writer pointer to zero.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buffers.Init()
	r.wcd.Signal()
}

func (r *Ring) Filter(predicate func(sequence uint32, startTs uint32) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for e := r.buffers.Front(); e != nil; e = e.Next() {
		elem := e.Value.(*markedBuffer)
		if predicate(elem.sequence, elem.startTs) {
			prev := e.Prev()
			r.buffers.Remove(e)
			if prev == nil {
				e = r.buffers.Front()
			} else {
				e = prev
			}
		}
	}
}
