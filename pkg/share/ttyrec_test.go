package share

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func makeTtyrecFrame(sec, usec uint32, payload []byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, sec)
	binary.Write(&buf, binary.LittleEndian, usec)
	binary.Write(&buf, binary.LittleEndian, uint32(len(payload)))
	buf.Write(payload)
	return buf.Bytes()
}

func TestParserSingleFrame(t *testing.T) {
	p := NewTtyrecParser()
	frame := makeTtyrecFrame(1, 0, []byte("hello"))
	p.Feed(frame)

	f, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected frame, got nil")
	}
	if string(f.Payload) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(f.Payload))
	}
	if f.Sec != 1 || f.Usec != 0 {
		t.Fatalf("unexpected timestamp: %d.%d", f.Sec, f.Usec)
	}

	f, err = p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != nil {
		t.Fatal("expected nil, got frame")
	}
}

func TestParserMultipleFrames(t *testing.T) {
	p := NewTtyrecParser()
	var data []byte
	data = append(data, makeTtyrecFrame(1, 0, []byte("aaa"))...)
	data = append(data, makeTtyrecFrame(2, 500, []byte("bbb"))...)
	p.Feed(data)

	f1, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f1 == nil || string(f1.Payload) != "aaa" {
		t.Fatalf("expected 'aaa', got %v", f1)
	}

	f2, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f2 == nil || string(f2.Payload) != "bbb" {
		t.Fatalf("expected 'bbb', got %v", f2)
	}

	f3, _ := p.Next()
	if f3 != nil {
		t.Fatal("expected nil")
	}
}

func TestParserPartialFrame(t *testing.T) {
	p := NewTtyrecParser()
	frame := makeTtyrecFrame(1, 0, []byte("hello world"))

	p.Feed(frame[:6])
	f, _ := p.Next()
	if f != nil {
		t.Fatal("expected nil for partial header")
	}

	p.Feed(frame[6:14])
	f, _ = p.Next()
	if f != nil {
		t.Fatal("expected nil for partial payload")
	}

	p.Feed(frame[14:])
	f, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil || string(f.Payload) != "hello world" {
		t.Fatalf("expected 'hello world', got %v", f)
	}
}

func TestParserInvalidLength(t *testing.T) {
	p := NewTtyrecParser()
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(5*1024*1024))
	p.Feed(buf.Bytes())

	_, err := p.Next()
	if err == nil {
		t.Fatal("expected error for oversized payload")
	}
}

func TestParserEmptyPayload(t *testing.T) {
	p := NewTtyrecParser()
	frame := makeTtyrecFrame(0, 0, []byte{})
	p.Feed(frame)

	f, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected frame")
	}
	if len(f.Payload) != 0 {
		t.Fatalf("expected empty payload, got %d bytes", len(f.Payload))
	}
}
