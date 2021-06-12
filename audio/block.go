package audio

type RingBuffer struct {
	inputChannel  <-chan *PCMFrame
	outputChannel chan *PCMFrame
}

func NewRingBuffer(inputChannel chan *PCMFrame, bufferSize int) *RingBuffer {
	return &RingBuffer{inputChannel, make(chan *PCMFrame, bufferSize)}
}

func (r *RingBuffer) Run() {
	for v := range r.inputChannel {
		select {
		case r.outputChannel <- v:
		default:
			<-r.outputChannel
			r.outputChannel <- v
		}
	}
	close(r.outputChannel)
}

func (r *RingBuffer) Close() {
	close(r.outputChannel)
}
