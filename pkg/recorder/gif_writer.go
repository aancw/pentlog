package recorder

import (
	"bufio"
	"compress/lzw"
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

type GIFWriter struct {
	writer     *bufio.Writer
	palette    color.Palette
	width      int
	height     int
	frameCount int
}

func NewGIFWriter(w io.Writer, palette color.Palette, width, height int) *GIFWriter {
	gw := &GIFWriter{
		writer:  bufio.NewWriterSize(w, 64*1024),
		palette: palette,
		width:   width,
		height:  height,
	}
	gw.writeHeader()
	return gw
}

func (gw *GIFWriter) writeHeader() {
	gw.writer.WriteString("GIF89a")
	binary.Write(gw.writer, binary.LittleEndian, uint16(gw.width))
	binary.Write(gw.writer, binary.LittleEndian, uint16(gw.height))

	paletteBits := 7
	gw.writer.WriteByte(0xF0 | uint8(paletteBits))
	gw.writer.WriteByte(0)
	gw.writer.WriteByte(0)

	for _, c := range gw.palette {
		r, g, b, _ := c.RGBA()
		gw.writer.WriteByte(uint8(r >> 8))
		gw.writer.WriteByte(uint8(g >> 8))
		gw.writer.WriteByte(uint8(b >> 8))
	}

	gw.writer.WriteByte(0x21)
	gw.writer.WriteByte(0xFF)
	gw.writer.WriteByte(11)
	gw.writer.WriteString("NETSCAPE2.0")
	gw.writer.WriteByte(3)
	gw.writer.WriteByte(1)
	binary.Write(gw.writer, binary.LittleEndian, uint16(0))
	gw.writer.WriteByte(0)
}

func (gw *GIFWriter) WriteFrame(img *image.Paletted, delayCs int) error {
	gw.writer.WriteByte(0x21)
	gw.writer.WriteByte(0xF9)
	gw.writer.WriteByte(4)
	gw.writer.WriteByte(0)
	binary.Write(gw.writer, binary.LittleEndian, uint16(delayCs))
	gw.writer.WriteByte(0)
	gw.writer.WriteByte(0)

	gw.writer.WriteByte(0x2C)
	binary.Write(gw.writer, binary.LittleEndian, uint16(0))
	binary.Write(gw.writer, binary.LittleEndian, uint16(0))
	binary.Write(gw.writer, binary.LittleEndian, uint16(gw.width))
	binary.Write(gw.writer, binary.LittleEndian, uint16(gw.height))
	gw.writer.WriteByte(0)

	minCodeSize := 8
	gw.writer.WriteByte(uint8(minCodeSize))

	blockWriter := &gifBlockWriter{writer: gw.writer}
	lzwWriter := lzw.NewWriter(blockWriter, lzw.LSB, minCodeSize)

	for y := 0; y < gw.height; y++ {
		for x := 0; x < gw.width; x++ {
			lzwWriter.Write([]byte{img.ColorIndexAt(x, y)})
		}
	}
	lzwWriter.Close()
	blockWriter.Flush()

	gw.frameCount++
	return gw.writer.Flush()
}

func (gw *GIFWriter) Close() error {
	gw.writer.WriteByte(0x3B)
	return gw.writer.Flush()
}

type gifBlockWriter struct {
	writer   *bufio.Writer
	blockBuf [255]byte
	blockPos int
}

func (w *gifBlockWriter) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		if w.blockPos == 255 {
			w.writer.WriteByte(255)
			w.writer.Write(w.blockBuf[:])
			w.blockPos = 0
		}

		copyLen := 255 - w.blockPos
		if copyLen > len(p) {
			copyLen = len(p)
		}
		copy(w.blockBuf[w.blockPos:], p[:copyLen])
		w.blockPos += copyLen
		p = p[copyLen:]
		n += copyLen
	}
	return n, nil
}

func (w *gifBlockWriter) Flush() {
	if w.blockPos > 0 {
		w.writer.WriteByte(uint8(w.blockPos))
		w.writer.Write(w.blockBuf[:w.blockPos])
		w.blockPos = 0
	}
	w.writer.WriteByte(0)
}
