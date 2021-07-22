package audio

import (
	"errors"
	"sync"
)

var (
	ErrIsFull  = errors.New("ring is full")
	ErrIsEmpty = errors.New("ring is empty")
)

type Ring struct {
	buf    []interface{}
	size   int
	r      int // next position to read
	w      int // next position to write
	isFull bool
	mu     sync.Mutex
	wcd    *sync.Cond
	rcd    *sync.Cond
}

// NewRing New returns a new Ring whose buffer has the given size.
func NewRing(size int) *Ring {
	rwmu := sync.Mutex{}
	return &Ring{
		buf:  make([]interface{}, size),
		size: size,
		wcd:  sync.NewCond(&rwmu),
		rcd:  sync.NewCond(&rwmu),
	}
}

func (r *Ring) TryPop() (b interface{}, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.w == r.r && !r.isFull {
		return nil, ErrIsEmpty
	}
	b = r.buf[r.r]
	r.r++
	if r.r == r.size {
		r.r = 0
	}
	r.isFull = false
	r.wcd.Signal()
	return b, err
}

func (r *Ring) TryPush(c interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.w == r.r && r.isFull {
		return ErrIsFull
	}
	r.buf[r.w] = c
	r.w++

	if r.w == r.size {
		r.w = 0
	}
	if r.w == r.r {
		r.isFull = true
	} else {
		r.rcd.Signal()
	}
	return nil
}

func (r *Ring) TryPeek() (b interface{}, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.w == r.r && !r.isFull {
		return nil, ErrIsEmpty
	}
	b = r.buf[r.r]
	return b, nil
}

func (r *Ring) Flush(predicate func(interface{}) bool) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.w == r.r && !r.isFull {
		return 0
	}
	writerPos := 0
	result := make([]interface{}, len(r.buf))
	pos := r.r + 1
	max := r.w
	for pos != max {
		v := r.buf[pos]
		if v != nil && predicate(v) {
			result[writerPos] = v
			writerPos++
		}
		pos++
		if pos == r.size {
			pos = 0
		}
	}
	r.buf = result
	r.r = 0
	r.w = writerPos
	return writerPos
}

func (r *Ring) Push(c interface{}) {
	err := r.TryPush(c)
	r.wcd.L.Lock()
	for err == ErrIsFull {
		r.wcd.Wait()
		err = r.TryPush(c)
	}
	r.wcd.L.Unlock()
}

func (r *Ring) Pop() interface{} {
	value, err := r.TryPop()
	r.rcd.L.Lock()
	for err == ErrIsEmpty {
		r.rcd.Wait()
		value, err = r.TryPop()
	}
	r.rcd.L.Unlock()
	return value
}

func (r *Ring) Peek() interface{} {
	value, err := r.TryPeek()
	r.rcd.L.Lock()
	for err == ErrIsEmpty {
		r.rcd.Wait()
		value, err = r.TryPeek()
	}
	r.rcd.L.Unlock()
	return value
}

// Length return the length of available read bytes.
func (r *Ring) Length() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.w == r.r {
		if r.isFull {
			return r.size
		}
		return 0
	}

	if r.w > r.r {
		return r.w - r.r
	}

	return r.size - r.r + r.w
}

// Capacity returns the size of the underlying buffer.
func (r *Ring) Capacity() int {
	return r.size
}

// Free returns the length of available bytes to write.
func (r *Ring) Free() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.w == r.r {
		if r.isFull {
			return 0
		}
		return r.size
	}

	if r.w < r.r {
		return r.r - r.w
	}

	return r.size - r.w + r.r
}

// IsFull returns this Ring is full.
func (r *Ring) IsFull() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.isFull
}

// IsEmpty returns this Ring is empty.
func (r *Ring) IsEmpty() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return !r.isFull && r.w == r.r
}

// Reset the read pointer and writer pointer to zero.
func (r *Ring) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.r = 0
	r.w = 0
	r.isFull = false
}
