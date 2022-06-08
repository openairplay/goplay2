// The MIT License (MIT)
//
// Copyright (c) 2016 winlin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// The aac decoder, to decode the encoded aac frame to PCM samples.
package codec

/*

#cgo pkg-config: fdk-aac
#include "fdk-aac/aacdecoder_lib.h"

typedef struct {
	HANDLE_AACDECODER dec;
	// Whether use ADTS mode.
	int is_adts;
	// Init util the first frame decoded.
	CStreamInfo* info;
	// The bits of sample, always 16 for fdkaac.
	int sample_bits;
	// Total filled bytes.
	UINT filled_bytes;
} aacdec_t;

static void _aacdec_init(aacdec_t* h) {
	// For lib-fdkaac, always use 16bits sample.
	// avctx->sample_fmt = AV_SAMPLE_FMT_S16;
	h->sample_bits = 16;
	h->is_adts = 0;
	h->filled_bytes = 0;

	h->dec = NULL;
	h->info = NULL;
}

static int aacdec_init_adts(aacdec_t* h) {
	_aacdec_init(h);

	h->is_adts = 1;

	h->dec = aacDecoder_Open(TT_MP4_ADTS, 1);
	if (!h->dec) {
		return -1;
	}

	return 0;
}

static int aacdec_init_raw(aacdec_t* h, char* asc, int nb_asc) {
	_aacdec_init(h);

	h->dec = aacDecoder_Open(TT_MP4_RAW, 1);
	if (!h->dec) {
		return -1;
	}

	UCHAR* uasc = (UCHAR*)asc;
	UINT unb_asc = (UINT)nb_asc;
	AAC_DECODER_ERROR err = aacDecoder_ConfigRaw(h->dec, &uasc, &unb_asc);
	if (err != AAC_DEC_OK) {
		return err;
	}

	return 0;
}

static void aacdec_close(aacdec_t* h) {
	if (h->dec) {
		aacDecoder_Close(h->dec);
	}
	h->dec = NULL;
}

static int aacdec_fill(aacdec_t* h, char* data, int nb_data, int* pnb_left) {
	h->filled_bytes += nb_data;

	UCHAR* udata = (UCHAR*)data;
	UINT unb_data = (UINT)nb_data;
	UINT unb_left = unb_data;
	AAC_DECODER_ERROR err = aacDecoder_Fill(h->dec, &udata, &unb_data, &unb_left);
	if (err != AAC_DEC_OK) {
		return err;
	}

	if (pnb_left) {
		*pnb_left = (int)unb_left;
	}

	return 0;
}

static int aacdec_pcm_size(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return (int)(h->info->numChannels * h->info->frameSize * h->sample_bits / 8);
}

static int aacdec_decode_frame(aacdec_t* h, char* pcm, int nb_pcm, int* pnb_valid) {
	// when buffer left bytes not enough, directly return not-enough-bits.
	// we requires atleast 7bytes header for adts.
	if (h->is_adts && h->info && h->filled_bytes - h->info->numTotalBytes <= 7) {
		return AAC_DEC_NOT_ENOUGH_BITS;
	}

	INT_PCM* upcm = (INT_PCM*)pcm;
	INT unb_pcm = (INT)nb_pcm;
	AAC_DECODER_ERROR err = aacDecoder_DecodeFrame(h->dec, upcm, unb_pcm, 0);

	// user should fill more bytes then decode.
	if (err == AAC_DEC_NOT_ENOUGH_BITS) {
		return err;
	}
	if (err != AAC_DEC_OK) {
		return err;
	}

	// when decode ok, retrieve the info.
	if (!h->info) {
		h->info = aacDecoder_GetStreamInfo(h->dec);
	}

	// the actual size of pcm.
	if (pnb_valid) {
		*pnb_valid = aacdec_pcm_size(h);
	}

	return 0;
}

*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	aacDecNotEnoughBits = 0x1002
)

type AacDecoder struct {
	m C.aacdec_t
}

func NewAacDecoder() *AacDecoder {
	return &AacDecoder{}
}

// InitRaw Open the decoder in RAW mode with ASC.
// For example, the FLV audio payload is a SequenceHeader(ASC) or RAW AAC data,
// user can init the decoder with ASC and decode the raw data.
// @remark user should never get the info util decode one frame.
func (v *AacDecoder) InitRaw(asc []byte) (err error) {
	p := (*C.char)(unsafe.Pointer(&asc[0]))
	pSize := C.int(len(asc))

	r := C.aacdec_init_raw(&v.m, p, pSize)

	if int(r) != 0 {
		return fmt.Errorf("init RAW decoder failed, code is %d", int(r))
	}

	return nil
}

// De-allocate all resources of an AAC decoder instance.
func (v *AacDecoder) Close() error {
	C.aacdec_close(&v.m)
	return nil
}

func (v *AacDecoder) DecodeTo(input []byte, pcm []int16) (size int, err error) {
	p := (*C.char)(unsafe.Pointer(&input[0]))
	pSize := C.int(len(input))
	leftSize := C.int(0)
	r := C.aacdec_fill(&v.m, p, pSize, &leftSize)
	if int(r) != 0 || int(leftSize) > 0 {
		return -1, fmt.Errorf("fill aac decoder failed, code is %d", int(r))
	}
	if int(leftSize) > 0 {
		return -1, fmt.Errorf("decoder left %v bytes", int(leftSize))
	}
	p = (*C.char)(unsafe.Pointer(&pcm[0]))
	pSize = C.int(len(pcm) * 2)
	validSize := C.int(0)
	r = C.aacdec_decode_frame(&v.m, p, pSize, &validSize)
	if int(r) == aacDecNotEnoughBits {
		return -1, nil
	}
	if int(r) != 0 {
		return -1, fmt.Errorf("decode frame failed, code is %d", int(r))
	}
	return int(validSize), nil
}
