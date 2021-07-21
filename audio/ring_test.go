package audio

import (
	"fmt"
	"testing"
	"time"
)

func TestRing_Pop(t *testing.T) {

	buffer := NewRing(2)

	go func() {
		time.Sleep(1 * time.Second)
		buffer.Push(0)
	}()
	_, err := buffer.TryPop()
	value := buffer.Pop()

	if value != 0 || err != ErrIsEmpty {
		t.Fail()
	}
}

func TestRing_Push(t *testing.T) {

	buffer := NewRing(2)

	buffer.Push(0)
	buffer.Push(1)
	go func() {
		time.Sleep(1 * time.Second)
		buffer.Pop()
	}()
	err := buffer.TryPush(2)
	buffer.Push(2)

	if err != ErrIsFull {
		t.Fail()
	}

}

func TestRing_Flush(t *testing.T) {

	buffer := NewRing(10)
	for i := 0; i < 8; i++ {
		buffer.Push(i)
	}
	buffer.Flush(func(value interface{}) bool {
		return value.(int) > 0 && value.(int) < 3
	})
	fmt.Printf("buffer : %v\n", buffer)
	if buffer.Length() != 2 {
		t.Fail()
	}

}
