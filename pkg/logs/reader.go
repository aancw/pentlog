package logs

import (
	"encoding/binary"
	"io"
)

// TtyTextReader adapts a binary ttyrec stream into a plain text stream
// by stripping the timing headers and treating it as a continuous text stream.
type TtyTextReader struct {
	reader io.Reader
	buffer []byte // Current chunk of data being read
	offset int    // Current position in buffer
}

func NewTtyReader(r io.Reader) *TtyTextReader {
	return &TtyTextReader{
		reader: r,
	}
}

func (t *TtyTextReader) Read(p []byte) (n int, err error) {
	// If we have data in the buffer, return it
	if t.offset < len(t.buffer) {
		n = copy(p, t.buffer[t.offset:])
		t.offset += n
		return n, nil
	}

	// Otherwise, read the next ttyrec chunk
	// Header: [sec:4][usec:4][len:4] = 12 bytes
	header := make([]byte, 12)
	_, err = io.ReadFull(t.reader, header)
	if err != nil {
		return 0, err
	}

	// Parse length (little-endian is standard for ttyrec)
	length := binary.LittleEndian.Uint32(header[8:12])

	// Read the data payload
	payload := make([]byte, length)
	_, err = io.ReadFull(t.reader, payload)
	if err != nil {
		return 0, err // Unexpected EOF in middle of chunk
	}

	t.buffer = payload
	t.offset = 0

	// Recursive call to copy data to p
	return t.Read(p)
}
