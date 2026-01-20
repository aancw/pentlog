package recorder

import (
	"bytes"
	"encoding/binary"
	"os"
)

type Frame struct {
	TimestampUsec int64
	Data          []byte
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
