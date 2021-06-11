package audio

import (
	"github.com/smallnest/ringbuffer"
	"io"
	"sync"
)

type BlockingRingBuffer struct {
	buf    *ringbuffer.RingBuffer
	cond   *sync.Cond
	closed bool
}

func NewBlockingReader(size int) *BlockingRingBuffer {
	m := sync.Mutex{}
	return &BlockingRingBuffer{
		cond: sync.NewCond(&m),
		buf:  ringbuffer.New(size),
	}
}

func (br *BlockingRingBuffer) Write(b []byte) (ln int, err error) {
	ln, err = br.buf.Write(b)
	br.cond.Broadcast()
	return
}

func (br *BlockingRingBuffer) Read(b []byte) (ln int, err error) {
	ln, err = br.buf.Read(b)
	br.cond.L.Lock()
	for err == ringbuffer.ErrIsEmpty {
		if br.closed {
			return 0, io.EOF
		}
		br.cond.Wait()
		ln, err = br.buf.Read(b)
	}
	br.cond.L.Unlock()
	return
}

func (br *BlockingRingBuffer) Close() error {
	br.closed = true
	br.buf.Reset()
	br.cond.Broadcast()
	return nil
}
