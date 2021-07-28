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

#cgo darwin CFLAGS: -I/usr/local/Cellar/fdk-aac/2.0.2/include/fdk-aac
#cgo darwin LDFLAGS: -L/usr/local/Cellar/fdk-aac/2.0.2/lib -lfdk-aac -lm

#cgo linux LDFLAGS: -L/usr/local/lib -l:libfdk-aac.a -lm

#ifdef __linux__
#include "fdk-aac/aacdecoder_lib.h"
#else
#include "aacdecoder_lib.h"
#endif

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

static int aacdec_sample_bits(aacdec_t* h) {
	return h->sample_bits;
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

static int aacdec_sample_rate(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->sampleRate;
}

static int aacdec_frame_size(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->frameSize;
}

static int aacdec_num_channels(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numChannels;
}

static int aacdec_aac_sample_rate(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->aacSampleRate;
}

static int aacdec_profile(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->profile;
}

static int aacdec_audio_object_type(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->aot;
}

static int aacdec_channel_config(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->channelConfig;
}

static int aacdec_bitrate(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->bitRate;
}

static int aacdec_aac_samples_per_frame(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->aacSamplesPerFrame;
}

static int aacdec_aac_num_channels(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->aacNumChannels;
}

static int aacdec_extension_audio_object_type(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->extAot;
}

static int aacdec_extension_sampling_rate(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->extSamplingRate;
}

static int aacdec_num_lost_access_units(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numLostAccessUnits;
}

static int aacdec_num_total_bytes(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numTotalBytes;
}

static int aacdec_num_bad_bytes(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numBadBytes;
}

static int aacdec_num_total_access_units(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numTotalAccessUnits;
}

static int aacdec_num_bad_access_units(aacdec_t* h) {
	if (!h->info) {
		return 0;
	}
	return h->info->numBadAccessUnits;
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

// Open the decoder in RAW mode with ASC.
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

// Open the decoder in ADTS mode without ASC,
// we never know the stream info util got the first frame,
// because the codec info is insert at begin of each frame.
// @remark The frame to Decode() is muxed in ADTS format.
func (v *AacDecoder) InitAdts() (err error) {
	r := C.aacdec_init_adts(&v.m)

	if int(r) != 0 {
		return fmt.Errorf("init ADTS decoder failed, code is %d", int(r))
	}

	return nil
}

// De-allocate all resources of an AAC decoder instance.
func (v *AacDecoder) Close() error {
	C.aacdec_close(&v.m)
	return nil
}

// Fill the buffer of decoder then decode.
// @remark we always expect all input are consumed by decoder.
func (v *AacDecoder) fill(input []byte) (err error) {
	p := (*C.char)(unsafe.Pointer(&input[0]))
	pSize := C.int(len(input))
	leftSize := C.int(0)

	r := C.aacdec_fill(&v.m, p, pSize, &leftSize)

	if int(r) != 0 {
		return fmt.Errorf("fill aac decoder failed, code is %d", int(r))
	}

	if int(leftSize) > 0 {
		return fmt.Errorf("decoder left %v bytes", int(leftSize))
	}

	return
}

// Decode one audio frame.
// @param the frame contains encoded aac frame, optional can be nil.
// @eturn when pcm is nil, should fill more bytes and decode again.
func (v *AacDecoder) Decode(frame []byte) (pcm []byte, err error) {
	if len(frame) > 0 {
		if err = v.fill(frame); err != nil {
			return
		}
	}

	nbPcm := int(C.aacdec_pcm_size(&v.m))
	if nbPcm == 0 {
		nbPcm = 50 * 1024
	}
	pcm = make([]byte, nbPcm)

	p := (*C.char)(unsafe.Pointer(&pcm[0]))
	pSize := C.int(nbPcm)
	validSize := C.int(0)

	r := C.aacdec_decode_frame(&v.m, p, pSize, &validSize)

	if int(r) == aacDecNotEnoughBits {
		return nil, nil
	}

	if int(r) != 0 {
		return nil, fmt.Errorf("decode frame failed, code is %d", int(r))
	}

	return pcm[0:int(validSize)], nil
}

// The bits of a sample, the fdk aac always use 16bits sample.
func (v *AacDecoder) SampleBits() int {
	return int(C.aacdec_sample_bits(&v.m))
}

// The samplerate in Hz of the fully decoded PCM audio signal (after SBR processing).
// @remark The only really relevant ones for the user.
func (v *AacDecoder) SampleRate() int {
	return int(C.aacdec_sample_rate(&v.m))
}

// The frame size of the decoded PCM audio signal.
//		1024 or 960 for AAC-LC
//		2048 or 1920 for HE-AAC (v2)
//		512 or 480 for AAC-LD and AAC-ELD
// @remark The only really relevant ones for the user.
func (v *AacDecoder) FrameSize() int {
	return int(C.aacdec_frame_size(&v.m))
}

// The number of output audio channels in the decoded and interleaved PCM audio signal.
// @remark The only really relevant ones for the user.
func (v *AacDecoder) NumChannels() int {
	return int(C.aacdec_num_channels(&v.m))
}

// sampling rate in Hz without SBR (from configuration info).
// @remark Decoder internal members.
func (v *AacDecoder) AacSampleRate() int {
	return int(C.aacdec_aac_sample_rate(&v.m))
}

// MPEG-2 profile (from file header) (-1: not applicable (e. g. MPEG-4)).
// @remark Decoder internal members.
func (v *AacDecoder) Profile() int {
	return int(C.aacdec_profile(&v.m))
}

// Audio Object Type (from ASC): is set to the appropriate value for MPEG-2 bitstreams (e. g. 2 for AAC-LC).
// @remark Decoder internal members.
func (v *AacDecoder) AudioObjectType() int {
	return int(C.aacdec_audio_object_type(&v.m))
}

// Channel configuration (0: PCE defined, 1: mono, 2: stereo, ...
// @remark Decoder internal members.
func (v *AacDecoder) ChannelConfig() int {
	return int(C.aacdec_channel_config(&v.m))
}

// Instantaneous bit rate.
// @remark Decoder internal members.
func (v *AacDecoder) Bitrate() int {
	return int(C.aacdec_bitrate(&v.m))
}

// Samples per frame for the AAC core (from ASC).
//		1024 or 960 for AAC-LC
//		512 or 480 for AAC-LD and AAC-ELD
// @remark Decoder internal members.
func (v *AacDecoder) AacSamplesPerFrame() int {
	return int(C.aacdec_aac_samples_per_frame(&v.m))
}

// The number of audio channels after AAC core processing (before PS or MPS processing).
//		CAUTION: This are not the final number of output channels!
// @remark Decoder internal members.
func (v *AacDecoder) AacNumChannels() int {
	return int(C.aacdec_aac_num_channels(&v.m))
}

// Extension Audio Object Type (from ASC)
// @remark Decoder internal members.
func (v *AacDecoder) ExtensionAudioObjectType() int {
	return int(C.aacdec_extension_audio_object_type(&v.m))
}

// Extension sampling rate in Hz (from ASC)
// @remark Decoder internal members.
func (v *AacDecoder) ExtensionSamplingRate() int {
	return int(C.aacdec_extension_sampling_rate(&v.m))
}

// This integer will reflect the estimated amount of lost access units in case aacDecoder_DecodeFrame()
// returns AAC_DEC_TRANSPORT_SYNC_ERROR. It will be < 0 if the estimation failed.
// @remark Statistics.
func (v *AacDecoder) NumLostAccessUnits() int {
	return int(C.aacdec_num_lost_access_units(&v.m))
}

// This is the number of total bytes that have passed through the decoder.
// @remark Statistics.
func (v *AacDecoder) NumTotalBytes() int {
	return int(C.aacdec_num_total_bytes(&v.m))
}

// This is the number of total bytes that were considered with errors from numTotalBytes.
// @remark Statistics.
func (v *AacDecoder) NumBadBytes() int {
	return int(C.aacdec_num_bad_bytes(&v.m))
}

// This is the number of total access units that have passed through the decoder.
// @remark Statistics.
func (v *AacDecoder) NumTotalAccessUnits() int {
	return int(C.aacdec_num_total_access_units(&v.m))
}

// This is the number of total access units that were considered with errors from numTotalBytes.
// @remark Statistics.
func (v *AacDecoder) NumBadAccessUnits() int {
	return int(C.aacdec_num_bad_access_units(&v.m))
}
