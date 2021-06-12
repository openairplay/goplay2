package globals

type ControlMessageType uint8

type ControlMessage struct {
	MType ControlMessageType
	Value int64
}

const (
	PAUSE ControlMessageType = iota
	START
	WAIT
	SKIP
	STOP
)
