package filters

/*
#cgo pkg-config: samplerate
#include <samplerate.h>
#include <string.h>
#include <stdlib.h>

static int resample(SRC_STATE* state, int outputSample, short * input, unsigned long in_len, short * output, unsigned long out_len,
					double ratio, unsigned long * input_frames_used, unsigned long * output_frames_gen) {

	float * inputf = malloc(in_len * sizeof(float));
	src_short_to_float_array(input, inputf, in_len);
	float * outputf = malloc(out_len * sizeof(float));
	memset(outputf, 0, out_len * sizeof(float));

	SRC_DATA srcdata;
	srcdata.src_ratio = ratio;
	srcdata.data_in = inputf;
	srcdata.input_frames = in_len / outputSample;
	srcdata.data_out = outputf;
	srcdata.output_frames = out_len / outputSample;
	srcdata.end_of_input = 0;
	int result =  src_process(state, &srcdata);
	src_float_to_short_array(outputf, output, out_len);

	free(inputf);
	free(outputf);

	if ( input_frames_used != NULL ) {
		*input_frames_used = srcdata.input_frames_used * outputSample;
	}
	if ( output_frames_gen != NULL ) {
		*output_frames_gen = srcdata.output_frames_gen * outputSample;
	}
	return result;
}


*/
import "C"
import (
	"errors"
	"unsafe"
)

type SrcState struct {
	inner          *C.SRC_STATE
	outputChannels int
}

func formatErr(err C.int) error {
	if err == 0 {
		return nil
	}
	return errors.New(C.GoString(C.src_strerror(err)))
}

func NewSrcState(channels int) (*SrcState, error) {
	var err C.int
	state := C.src_new(C.SRC_SINC_BEST_QUALITY, C.int(channels), &err)
	if err != 0 {
		return nil, formatErr(err)
	}
	return &SrcState{
		inner:          state,
		outputChannels: channels,
	}, nil
}

func (s *SrcState) Close() error {
	s.inner = C.src_delete(s.inner)
	return nil
}

func (s *SrcState) Process(ratio float64, input []int16, output []int16) (inputFramesUsed int, outputFramesGen int, err error) {
	var inUsed C.ulong
	var outGen C.ulong

	in := (*C.short)(unsafe.Pointer(&input[0]))
	out := (*C.short)(unsafe.Pointer(&output[0]))

	err = formatErr(C.resample(s.inner, C.int(s.outputChannels), in, C.ulong(len(input)), out,
		C.ulong(len(output)), C.double(ratio), &inUsed, &outGen))

	inputFramesUsed = int(inUsed)
	outputFramesGen = int(outGen)
	return
}
