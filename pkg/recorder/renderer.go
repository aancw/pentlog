package recorder

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"io"
	"os"

	"github.com/james4k/terminal"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type RenderConfig struct {
	Cols      int
	Rows      int
	Speed     float64
	MaxFrames int
}

func DefaultConfig() RenderConfig {
	return RenderConfig{
		Cols:      80,
		Rows:      24,
		Speed:     1.0,
		MaxFrames: 0,
	}
}

var ansiPalette = []color.Color{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{170, 0, 0, 255},
	color.RGBA{0, 170, 0, 255},
	color.RGBA{170, 85, 0, 255},
	color.RGBA{0, 0, 170, 255},
	color.RGBA{170, 0, 170, 255},
	color.RGBA{0, 170, 170, 255},
	color.RGBA{170, 170, 170, 255},
	color.RGBA{85, 85, 85, 255},
	color.RGBA{255, 85, 85, 255},
	color.RGBA{85, 255, 85, 255},
	color.RGBA{255, 255, 85, 255},
	color.RGBA{85, 85, 255, 255},
	color.RGBA{255, 85, 255, 255},
	color.RGBA{85, 255, 255, 255},
	color.RGBA{255, 255, 255, 255},
}

func buildPalette() color.Palette {
	p := make(color.Palette, 0, 256)
	p = append(p, ansiPalette...)

	for i := 16; i < 232; i++ {
		idx := i - 16
		r := uint8((idx/36)%6) * 51
		g := uint8((idx/6)%6) * 51
		b := uint8(idx%6) * 51
		p = append(p, color.RGBA{r, g, b, 255})
	}

	for i := 232; i < 256; i++ {
		v := uint8((i-232)*10 + 8)
		p = append(p, color.RGBA{v, v, v, 255})
	}

	return p
}

func RenderToGIF(inputPath, outputPath string, cfg RenderConfig) error {
	frames, err := ParseTTYRec(inputPath)
	if err != nil {
		return err
	}

	if len(frames) == 0 {
		return nil
	}

	state := &terminal.State{}
	vt, err := terminal.Create(state, io.NopCloser(bytes.NewBuffer(nil)))
	if err != nil {
		return err
	}
	defer vt.Close()
	vt.Resize(cfg.Cols, cfg.Rows)

	palette := buildPalette()
	charW, charH := 7, 13
	imgW := cfg.Cols*charW + 20
	imgH := cfg.Rows*charH + 20

	var gifImages []*image.Paletted
	var delays []int

	baseTime := frames[0].TimestampUsec
	var lastCaptureTime int64

	for i, frame := range frames {
		vt.Write(frame.Data)

		elapsed := frame.TimestampUsec - baseTime
		scaledElapsed := int64(float64(elapsed) / cfg.Speed)

		var delayCs int
		if i == 0 {
			delayCs = 10
		} else {
			diff := scaledElapsed - lastCaptureTime
			delayCs = int(diff / 10000)
		}

		if delayCs < 2 {
			delayCs = 2
		}
		if delayCs > 500 {
			delayCs = 500
		}

		img := renderFrame(state, cfg.Cols, cfg.Rows, imgW, imgH, charW, charH, palette)
		gifImages = append(gifImages, img)
		delays = append(delays, delayCs)
		lastCaptureTime = scaledElapsed

		if cfg.MaxFrames > 0 && len(gifImages) >= cfg.MaxFrames {
			break
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return gif.EncodeAll(outFile, &gif.GIF{
		Image: gifImages,
		Delay: delays,
	})
}

func renderFrame(state *terminal.State, cols, rows, imgW, imgH, charW, charH int, palette color.Palette) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, imgW, imgH), palette)

	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.Set(x, y, palette[0])
		}
	}

	curX, curY := state.Cursor()
	face := basicfont.Face7x13

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			ch, fg, bg := state.Cell(col, row)

			px := 10 + col*charW
			py := 10 + row*charH

			bgIdx := colorToIndex(bg, true)
			if state.CursorVisible() && row == curY && col == curX {
				bgIdx = 15
			}

			for dy := 0; dy < charH; dy++ {
				for dx := 0; dx < charW; dx++ {
					img.SetColorIndex(px+dx, py+dy, bgIdx)
				}
			}

			if ch > 32 {
				fgIdx := colorToIndex(fg, false)
				drawRune(img, face, ch, px, py+charH-2, palette[fgIdx])
			}
		}
	}

	return img
}

func colorToIndex(c terminal.Color, isBg bool) uint8 {
	switch c {
	case terminal.DefaultBG:
		return 0
	case terminal.DefaultFG:
		return 7
	default:
		if c >= 0 && c < 256 {
			return uint8(c)
		}
		if isBg {
			return 0
		}
		return 7
	}
}

func drawRune(img *image.Paletted, face font.Face, r rune, x, y int, clr color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(string(r))
}
