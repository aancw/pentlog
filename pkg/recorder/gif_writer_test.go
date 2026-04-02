package recorder

import (
	"image"
	"image/gif"
	"os"
	"testing"
)

func TestGIFWriter(t *testing.T) {
	palette := buildPalette()
	width := 100
	height := 50

	tmpFile, err := os.CreateTemp("", "test_gif_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	outFile, err := os.Create(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	gw := NewGIFWriter(outFile, palette, width, height)

	img1 := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img1.SetColorIndex(x, y, 0)
		}
	}

	img2 := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img2.SetColorIndex(x, y, 7)
		}
	}

	if err := gw.WriteFrame(img1, 10); err != nil {
		t.Fatal(err)
	}

	if err := gw.WriteFrame(img2, 20); err != nil {
		t.Fatal(err)
	}

	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	outFile.Close()

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(data) < 50 {
		t.Fatalf("GIF file too small: %d bytes", len(data))
	}

	if data[0] != 'G' || data[1] != 'I' || data[2] != 'F' {
		t.Fatal("GIF header not found")
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	decoded, err := gif.DecodeAll(file)
	if err != nil {
		t.Fatalf("Failed to decode GIF: %v", err)
	}

	if len(decoded.Image) < 2 {
		t.Fatalf("Expected at least 2 frames, got %d", len(decoded.Image))
	}

	t.Logf("Generated streaming GIF: %d bytes, %d frames", len(data), len(decoded.Image))
}
