package share

import (
	"encoding/binary"
	"fmt"
)

const ttyrecHeaderSize = 12

type TtyrecFrame struct {
	Sec     uint32
	Usec    uint32
	Payload []byte
}

type TtyrecParser struct {
	buf []byte
}

func NewTtyrecParser() *TtyrecParser {
	return &TtyrecParser{}
}

func (p *TtyrecParser) Feed(data []byte) {
	p.buf = append(p.buf, data...)
}

func (p *TtyrecParser) Next() (*TtyrecFrame, error) {
	if len(p.buf) < ttyrecHeaderSize {
		return nil, nil
	}

	sec := binary.LittleEndian.Uint32(p.buf[0:4])
	usec := binary.LittleEndian.Uint32(p.buf[4:8])
	payloadLen := binary.LittleEndian.Uint32(p.buf[8:12])

	const maxFramePayload = 4 * 1024 * 1024
	if payloadLen > maxFramePayload {
		return nil, fmt.Errorf("invalid ttyrec frame: payload length %d exceeds max %d", payloadLen, maxFramePayload)
	}

	totalLen := ttyrecHeaderSize + int(payloadLen)
	if len(p.buf) < totalLen {
		return nil, nil
	}

	payload := make([]byte, payloadLen)
	copy(payload, p.buf[ttyrecHeaderSize:totalLen])

	p.buf = p.buf[totalLen:]

	return &TtyrecFrame{
		Sec:     sec,
		Usec:    usec,
		Payload: payload,
	}, nil
}
