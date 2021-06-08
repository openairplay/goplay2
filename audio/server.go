package audio

import (
	"encoding/binary"
	"fmt"
	"github.com/AlbanSeurat/galac/alac"
	"github.com/brutella/hc/crypto/chacha20poly1305"
	"github.com/gordonklaus/portaudio"
	"github.com/pion/rtp"
	"github.com/winlinvip/go-fdkaac"
	"io"
	"log"
	"net"
)

type Server struct {
	alacDecoder *alac.Decoder
	aacDecoder  *fdkaac.AacDecoder
}

func NewServer() *Server {

	cookie := []byte{
		0x00, 0x00, 0x00, 0x24, 0x61, 0x6c, 0x61, 0x63, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x60, 0x00, 0x10, 0x28, 0x0a, 0x0e, 0x02, 0x00, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xac, 0x44}

	decoder, err := alac.NewDecoder(cookie)
	if err != nil {
		log.Panicf("alac debugger not available : %v", err)
	}

	aacDecoder := fdkaac.NewAacDecoder()

	asc := []byte{0x12, 0x10}
	if err := aacDecoder.InitRaw(asc); err != nil {
		log.Panicf("init decoder failed, err is %s", err)
	}

	return &Server{
		alacDecoder: decoder,
		aacDecoder:  aacDecoder,
	}
}

func (s *Server) Listen(sharedKey []byte, l net.Listener) {
	defer l.Close()

	r, w := io.Pipe()

	go func() {
		s.Play(r)
	}()

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		go s.handleClientStream(sharedKey, conn, w)
	}
}

func (s *Server) handleClientStream(sharedKey []byte, r io.ReadCloser, w io.WriteCloser) {
	defer r.Close()
	defer w.Close()

	for {
		pcmData, err := s.DecodeToPcm(sharedKey, r)
		if err != nil {
			log.Printf("PCM Decode Error : %v\n", err)
			break
		}
		_, err = w.Write(pcmData)
		if err != nil {
			log.Printf("Pipe Write Error : %v\n", err)
			break
		}
	}
}

func (s *Server) DecodeToPcm(sharedKey []byte, reader io.Reader) ([]byte, error) {

	var i uint16
	err := binary.Read(reader, binary.BigEndian, &i)
	if err != nil {
		return nil, err
	} else {

		buffer := make([]byte, i-2)
		if _, err := io.ReadFull(reader, buffer) ; err != nil {
			return nil, err
		}
		packet := rtp.Packet{}
		if err = packet.Unmarshal(buffer) ; err != nil {
			return nil, err
		}
		message := packet.Payload[:len(packet.Payload)-24]
		nonce := packet.Payload[len(packet.Payload)-8:]
		var mac [16]byte
		copy(mac[:], packet.Payload[len(packet.Payload)-24:len(packet.Payload)-8])
		aad := packet.Raw[4:0xc]

		decrypted, err := chacha20poly1305.DecryptAndVerify(sharedKey, nonce, message, mac, aad)
		if err != nil {
			return nil, err
		}
		return s.aacDecoder.Decode(decrypted)
	}
}

func (s *Server) Play(reader io.Reader) {

	if err := portaudio.Initialize(); err != nil {
		log.Fatalln("PortAudio init error:", err)
	}
	defer portaudio.Terminate()

	out := make([]int16, 1024)
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, len(out), &out)
	if err != nil {
		log.Println("PortAudio Stream opened false ", err)
		return
	}
	defer stream.Close()

	err = stream.Start()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		err := binary.Read(reader, binary.LittleEndian, out)
		if err == io.EOF {
			log.Println("error reading pipe", err)
			break
		}
		stream.Write()
	}

}
