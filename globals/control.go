package globals

type ControlMessageType uint8

type ControlMessage struct {
	MType  ControlMessageType
	Param1 int64
	Param2 int64
	Paramf float64
}

const (
	PAUSE ControlMessageType = iota
	STOP
	START
	SKIP
	VOLUME
)
