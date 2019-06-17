package rtmp

import(
	"errors"
	"math/rand"
	"time"
	"encoding/binary"
	"bytes"
	"log"
	"go_srs/srs/protocol/skt"
)

type SrsHandshakeBytes struct {
	C0C1 []byte
	S0S1S2 []byte
	C2 []byte
	io *skt.SrsIOReadWriter
}

func NewSrsHandshakeBytes(io_ *skt.SrsIOReadWriter) *SrsHandshakeBytes {
	return &SrsHandshakeBytes{
		io: io_,
	}
}

func (this *SrsHandshakeBytes) ReadC0C1() error {
	if len(this.C0C1) > 0 {
		err := errors.New("handshake read c0c1 failed, already read")
		return err
	}

	this.C0C1 = make([]byte, 1537)
	left := 1537
	for {
		n, err := this.io.Read(this.C0C1[1537-left:1537])
		if err != nil {
			return err
		}
		
		left = left - n
		if left <= 0 {
			return nil
		}
	}
}

func (this *SrsHandshakeBytes) CreateS0S1S2() error {
	if len(this.S0S1S2) > 0 {
		return errors.New("already create")
	}
	rand.Seed(time.Now().UnixNano())
	this.S0S1S2 = make([]byte, 3073)
	//s0 = version
	this.S0S1S2[0] = 0x3
	//s1 for bytes(timestamp)
	binary.Write(bytes.NewBuffer(this.S0S1S2[1:5]), binary.LittleEndian, time.Now().Unix())
	//s1 rand bytes
	if n, err := rand.Read(this.S0S1S2[9:1537]); err != nil || n != 1528 {
		return errors.New("create rand number failed")
	}
	//s2=c1
	copy(this.S0S1S2[1537:], this.C0C1[1:])
	return nil
}

func (this *SrsHandshakeBytes) ReadC2() int {
	if len(this.C2) > 0 {
		return -1
	}

	this.C2 = make([]byte, 1536)
	left := 1536
	for {
		n, err := this.io.Read(this.C2[1536-left:1536])
		if err != nil {
			return -1
		}
		log.Print("read n=", n)
		left = left - n
		if left <= 0 {
			return 0
		}
	}
}

func (this *SrsHandshakeBytes) CheckC2() bool {
	return bytes.Equal(this.C2, this.S0S1S2[1:1537])
}