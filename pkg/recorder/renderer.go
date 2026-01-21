package recorder

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"regexp"
	"strings"

	"github.com/muesli/termenv"
	"github.com/vito/vt100"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type RenderConfig struct {
	Cols       int
	Rows       int
	Speed      float64
	MaxFrames  int
	Resolution string // "720p" or "1080p"
}

func DefaultConfig() RenderConfig {
	return RenderConfig{
		Cols:       160,
		Rows:       45,
		Speed:      1.0,
		MaxFrames:  0,
		Resolution: "720p",
	}
}

var promptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\(pentlog:[^)]*\)\s*`),
	regexp.MustCompile(`\[pentlog:[^\]]*\]\s*`),
}

func filterPromptFromData(data []byte) []byte {
	result := data
	for _, pattern := range promptPatterns {
		result = pattern.ReplaceAll(result, []byte{})
	}
	return result
}

// Improved ANSI color palette with better Kali Linux terminal colors
var ansi16 = []color.RGBA{
	{0, 0, 0, 255},       // Black
	{205, 49, 49, 255},   // Red (improved)
	{13, 188, 121, 255},  // Green (improved)
	{229, 229, 16, 255},  // Yellow (improved)
	{36, 114, 200, 255},  // Blue (improved)
	{188, 63, 188, 255},  // Magenta (improved)
	{17, 168, 205, 255},  // Cyan (improved)
	{229, 229, 229, 255}, // White (improved)
	{102, 102, 102, 255}, // Bright Black (Gray)
	{241, 76, 76, 255},   // Bright Red
	{35, 209, 139, 255},  // Bright Green
	{245, 245, 67, 255},  // Bright Yellow
	{59, 142, 234, 255},  // Bright Blue
	{214, 112, 214, 255}, // Bright Magenta
	{41, 184, 219, 255},  // Bright Cyan
	{255, 255, 255, 255}, // Bright White
}

func buildPalette() color.Palette {
	p := make(color.Palette, 0, 256)
	for _, c := range ansi16 {
		p = append(p, c)
	}
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

	term := vt100.NewVT100(cfg.Rows, cfg.Cols)
	palette := buildPalette()

	// Create font face based on resolution
	fontSize := 12.0
	if cfg.Resolution == "1080p" {
		fontSize = 14.0
	}
	fontFace, err := createFontFace(fontSize)
	if err != nil {
		return err
	}

	// Calculate character dimensions from font metrics
	charW := fontFace.Metrics().Height.Ceil() * 6 / 10 // Approximate monospace width
	charH := fontFace.Metrics().Height.Ceil()
	paddingX, paddingY := 12, 12
	imgW := cfg.Cols*charW + paddingX*2
	imgH := cfg.Rows*charH + paddingY*2

	var gifImages []*image.Paletted
	var delays []int

	baseTime := frames[0].TimestampUsec
	var lastCaptureTime int64
	var lastContent string

	for i, frame := range frames {
		filteredData := filterPromptFromData(frame.Data)
		reader := bytes.NewReader(filteredData)
		for reader.Len() > 0 {
			cmd, err := vt100.Decode(reader)
			if err != nil {
				break
			}
			term.Process(cmd)
		}

		currentContent := terminalContentHash(term, cfg.Cols, cfg.Rows)
		if currentContent == lastContent && i > 0 {
			continue
		}
		lastContent = currentContent

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
		if delayCs > 300 {
			delayCs = 300
		}

		img := renderTerminal(term, cfg.Cols, cfg.Rows, imgW, imgH, charW, charH, paddingX, paddingY, palette, fontFace)
		gifImages = append(gifImages, img)
		delays = append(delays, delayCs)
		lastCaptureTime = scaledElapsed

		if cfg.MaxFrames > 0 && len(gifImages) >= cfg.MaxFrames {
			break
		}
	}

	if len(gifImages) == 0 {
		img := renderTerminal(term, cfg.Cols, cfg.Rows, imgW, imgH, charW, charH, paddingX, paddingY, palette, fontFace)
		gifImages = append(gifImages, img)
		delays = append(delays, 100)
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

func terminalContentHash(term *vt100.VT100, cols, rows int) string {
	var b strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			b.WriteRune(term.Content[row][col])
		}
	}
	return b.String()
}

func createFontFace(size float64) (font.Face, error) {
	parsedFont, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func renderTerminal(term *vt100.VT100, cols, rows, imgW, imgH, charW, charH, padX, padY int, palette color.Palette, face font.Face) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, imgW, imgH), palette)

	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.SetColorIndex(x, y, 0)
		}
	}

	cursorY, cursorX := term.Cursor.Y, term.Cursor.X

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			px := padX + col*charW
			py := padY + row*charH

			ch := term.Content[row][col]

			format := term.Format[row][col]
			bgIdx := colorToIdx(format.Bg, true)
			fgIdx := colorToIdx(format.Fg, false)

			if format.Reverse {
				bgIdx, fgIdx = fgIdx, bgIdx
			}

			if row == cursorY && col == cursorX {
				bgIdx = 15
				fgIdx = 0
			}

			for dy := 0; dy < charH; dy++ {
				for dx := 0; dx < charW; dx++ {
					img.SetColorIndex(px+dx, py+dy, bgIdx)
				}
			}

			if ch > 32 && ch != ' ' {
				drawGlyph(img, face, ch, px, py+charH-2, palette[fgIdx])
			}
		}
	}

	return img
}

func colorToIdx(c termenv.Color, isBg bool) uint8 {
	if c == nil {
		if isBg {
			return 0
		}
		return 7
	}

	switch v := c.(type) {
	case termenv.ANSIColor:
		return uint8(v)
	case termenv.ANSI256Color:
		return uint8(v)
	case termenv.RGBColor:
		return 7
	}

	if isBg {
		return 0
	}
	return 7
}

func drawGlyph(img *image.Paletted, face font.Face, r rune, x, y int, clr color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(string(r))
}
