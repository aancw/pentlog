package recorder

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

type Frame struct {
	TimestampUsec int64
	Data          []byte
}

type FrameReader struct {
	file   *os.File
	header [12]byte
}

func NewFrameReader(path string) (*FrameReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FrameReader{file: file}, nil
}

func (r *FrameReader) Close() error {
	return r.file.Close()
}

func (r *FrameReader) ReadFrame() (*Frame, error) {
	_, err := io.ReadFull(r.file, r.header[:])
	if err != nil {
		return nil, err
	}

	sec := binary.LittleEndian.Uint32(r.header[0:4])
	usec := binary.LittleEndian.Uint32(r.header[4:8])
	length := binary.LittleEndian.Uint32(r.header[8:12])

	data := make([]byte, length)
	_, err = io.ReadFull(r.file, data)
	if err != nil {
		return nil, err
	}

	return &Frame{
		TimestampUsec: int64(sec)*1000000 + int64(usec),
		Data:          data,
	}, nil
}

func ParseTTYRec(path string) ([]Frame, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var frames []Frame
	reader := bytes.NewReader(data)

	for reader.Len() >= 12 {
		var sec, usec, length uint32
		if err := binary.Read(reader, binary.LittleEndian, &sec); err != nil {
			break
		}
		if err := binary.Read(reader, binary.LittleEndian, &usec); err != nil {
			break
		}
		if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
			break
		}

		if reader.Len() < int(length) {
			break
		}

		frameData := make([]byte, length)
		if _, err := reader.Read(frameData); err != nil {
			break
		}

		timestamp := int64(sec)*1000000 + int64(usec)
		frames = append(frames, Frame{
			TimestampUsec: timestamp,
			Data:          frameData,
		})
	}

	return frames, nil
}
