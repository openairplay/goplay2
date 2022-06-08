package rtp

import (
	"encoding/binary"
	"github.com/brutella/hc/crypto/chacha20poly1305"
	"github.com/pion/rtp"
	"goplay2/codec"
)

type TcpPacket struct {
	rtp.Packet
	SequenceNumber uint32
}

type Frame struct {
	TcpPacket
	aacData []byte
}

func (p *Frame) PcmData(aacDecoder *codec.AacDecoder, pcm []int16) (int, error) {
	decode, err := aacDecoder.DecodeTo(p.aacData, pcm)
	if err != nil {
		return -1, err
	}
	return decode, nil
}

func NewFrame(rawPacket []byte, sharedKey []byte) (*Frame, error) {
	var err error
	packet := TcpPacket{}
	if err = packet.Unmarshal(rawPacket); err != nil {
		return nil, err
	}
	var seqBytes [4]byte
	copy(seqBytes[1:], rawPacket[1:4])
	packet.Marker = false  // used by apple in sequenceNumber
	packet.PayloadType = 0 // used by apple in sequenceNumber
	packet.SequenceNumber = binary.BigEndian.Uint32(seqBytes[:])
	encryptedData := packet.Payload[:len(packet.Payload)-24]
	nonce := packet.Payload[len(packet.Payload)-8:]
	var mac [16]byte
	copy(mac[:], packet.Payload[len(packet.Payload)-24:len(packet.Payload)-8])
	aad := packet.Raw[4:0xc]
	decrypted, err := chacha20poly1305.DecryptAndVerify(sharedKey,
		nonce, encryptedData, mac, aad)
	if err != nil {
		return nil, err
	}
	return &Frame{TcpPacket: packet,
		aacData: decrypted,
	}, nil
}
