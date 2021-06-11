package globals

type ControlMessage uint8

const (
	TEARDOWN ControlMessage = iota
	PAUSE_STREAM
)
