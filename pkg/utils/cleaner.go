package utils

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
)

type CleanReader struct {
	scanner *bufio.Scanner
	buf     []byte
	err     error
}

func NewCleanReader(r io.Reader) *CleanReader {
	return &CleanReader{
		scanner: bufio.NewScanner(r),
	}
}

func (cr *CleanReader) Read(p []byte) (n int, err error) {
	if len(cr.buf) > 0 {
		n = copy(p, cr.buf)
		cr.buf = cr.buf[n:]
		return n, nil
	}

	if cr.err != nil {
		return 0, cr.err
	}

	if cr.scanner.Scan() {
		rawLine := cr.scanner.Text()

		rendered := RenderAnsi(rawLine)

		cr.buf = []byte(rendered + "\n")
		n = copy(p, cr.buf)
		cr.buf = cr.buf[n:]
		return n, nil
	}

	cr.err = cr.scanner.Err()
	if cr.err == nil {
		cr.err = io.EOF
	}
	return 0, cr.err
}

type Cell struct {
	Char  rune
	Style string
}

var (
	reAnsiSeq = regexp.MustCompile(`^\x1b\[([0-9;]*)([A-Za-z])`)
	reOscSeq  = regexp.MustCompile(`^\x1b\][0-9];.*?\x07`)
)

func RenderAnsi(line string) string {
	var buffer []Cell
	cursor := 0
	currentStyle := ""

	runes := []rune(line)
	i := 0
	lenRunes := len(runes)

	for i < lenRunes {
		r := runes[i]

		if r == '\r' {
			cursor = 0
			i++
			continue
		}
		if r == '\b' {
			if cursor > 0 {
				cursor--
			}
			i++
			continue
		}

		// Check for ANSI escape sequence
		if r == '\x1b' {
			// Look ahead for CSI ( [ ... char ) or OSC ( ] ... \07 )
			remainder := string(runes[i:])

			// OSC (Operating System Command) - usually title setting, ignore
			if loc := reOscSeq.FindStringIndex(remainder); loc != nil {
				i += loc[1]
				continue
			}

			// CSI (Control Sequence Introducer)
			if loc := reAnsiSeq.FindStringSubmatchIndex(remainder); loc != nil {
				matchLen := loc[1]
				params := remainder[loc[2]:loc[3]]
				cmd := remainder[loc[4]:loc[5]]

				i += matchLen

				switch cmd {
				case "m": // SGR - Select Graphic Rendition (Colors)
					if params == "" || params == "0" {
						currentStyle = ""
					} else {
						currentStyle = "\x1b[" + params + "m"
					}

				case "K": // EL - Erase in Line
					// 0 (default): cursor to end
					// 1: start to cursor
					// 2: entire line
					mode := 0
					if params != "" {
						if m, err := strconv.Atoi(params); err == nil {
							mode = m
						}
					}

					if mode == 0 {
						if cursor < len(buffer) {
							buffer = buffer[:cursor]
						}
					} else if mode == 1 {
						for k := 0; k < cursor && k < len(buffer); k++ {
							buffer[k] = Cell{Char: ' ', Style: ""}
						}
					} else if mode == 2 {
						buffer = nil
						cursor = 0
					}

				case "D": // Cub - Cursor Left
					count := 1
					if params != "" {
						if c, err := strconv.Atoi(params); err == nil {
							count = c
						}
					}
					cursor -= count
					if cursor < 0 {
						cursor = 0
					}

				case "C": // Cuf - Cursor Right
					count := 1
					if params != "" {
						if c, err := strconv.Atoi(params); err == nil {
							count = c
						}
					}
					cursor += count

				case "G": // CHA - Cursor Horizontal Absolute
					col := 1
					if params != "" {
						if c, err := strconv.Atoi(params); err == nil {
							col = c
						}
					}
					cursor = col - 1
					if cursor < 0 {
						cursor = 0
					}

				}
				continue
			}
		}

		if cursor >= len(buffer) {
			if cursor > CapBufferLimit {

			}
			gap := cursor - len(buffer)
			for k := 0; k < gap; k++ {
				buffer = append(buffer, Cell{Char: ' ', Style: ""})
			}
			buffer = append(buffer, Cell{Char: r, Style: currentStyle})
		} else {
			buffer[cursor] = Cell{Char: r, Style: currentStyle}
		}
		cursor++
		i++
	}

	var out bytes.Buffer
	lastStyle := ""

	for _, cell := range buffer {
		if cell.Style != lastStyle {
			out.WriteString(cell.Style)
			lastStyle = cell.Style
		}
		out.WriteRune(cell.Char)
	}

	if lastStyle != "" && lastStyle != "\x1b[0m" {
		out.WriteString("\x1b[0m")
	}

	return out.String()
}

const CapBufferLimit = 10000
