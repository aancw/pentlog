package utils

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
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
	reAnsiSeq = regexp.MustCompile(`^\x1b\[([?0-9;]*)([A-Za-z])`)
	reOscSeq  = regexp.MustCompile(`^\x1b\][0-9]*;.*?(?:\x07|\x1b\\)`)
	reAltSeq  = regexp.MustCompile(`^\x1b([=><78]|[()][0-9A-Za-z])`)

	// Regex to match the TUI block. using (?s) for dot-all (match newlines)
	reTuiBlock = regexp.MustCompile(`(?s)\x1b]99;PENTLOG_TUI_START\x07.*?\x1b]99;PENTLOG_TUI_END\x07`)
)

func ParseAnsi(line string) []Cell {
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

			if loc := reOscSeq.FindStringIndex(remainder); loc != nil {
				// Validate OSC doesn't contain shell metacharacters
				oscContent := remainder[loc[0]:loc[1]]
				if strings.ContainsAny(oscContent, "`;$()") {
					// Skip potentially malicious OSC sequence
					i += loc[1] // Skip the entire sequence
					continue
				}

				i += loc[1]
				continue
			}

			// Alternative Sequences (Keypad, Character Set, Cursor Save/Restore)
			if loc := reAltSeq.FindStringIndex(remainder); loc != nil {
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
					if cursor >= CapBufferLimit {
						cursor = CapBufferLimit // prevent unbounded growth
					}

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
				// Prevent buffer from growing indefinitely if cursor is way out
				cursor = CapBufferLimit
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
	return buffer
}

func RenderAnsi(line string) string {
	buffer := ParseAnsi(line)
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

func RenderAnsiHTML(line string) string {
	buffer := ParseAnsi(line)
	var out bytes.Buffer
	currentClass := ""

	for _, cell := range buffer {
		styleClass := getStyleClass(cell.Style)

		if styleClass != currentClass {
			if currentClass != "" {
				out.WriteString("</span>")
			}
			if styleClass != "" {
				out.WriteString("<span class=\"" + styleClass + "\">")
			}
			currentClass = styleClass
		}

		switch cell.Char {
		case '&':
			out.WriteString("&amp;")
		case '<':
			out.WriteString("&lt;")
		case '>':
			out.WriteString("&gt;")
		case '"':
			out.WriteString("&quot;")
		case '\'':
			out.WriteString("&#39;")
		default:
			out.WriteRune(cell.Char)
		}
	}

	if currentClass != "" {
		out.WriteString("</span>")
	}

	return out.String()
}

func RenderPlain(line string) string {
	buffer := ParseAnsi(line)
	var out bytes.Buffer

	for _, cell := range buffer {
		out.WriteRune(cell.Char)
	}

	return strings.TrimRight(out.String(), " ")
}

func getStyleClass(ansi string) string {
	if ansi == "" || ansi == "\x1b[0m" {
		return ""
	}

	// Simplified parsing for common colors
	// \x1b[31;1m -> class="red bold"
	// We can use regex to extract numbers

	// Just standard colors mapping for now
	// 30-37: black, red, green, yellow, blue, magenta, cyan, white
	// 90-97: bright versions

	var classes []string

	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(ansi, -1)

	for _, m := range matches {
		code, _ := strconv.Atoi(m)
		switch {
		case code == 1:
			classes = append(classes, "ansi-bold")
		case code >= 30 && code <= 37:
			colors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
			classes = append(classes, "ansi-"+colors[code-30])
		case code >= 90 && code <= 97:
			colors := []string{"bright-black", "bright-red", "bright-green", "bright-yellow", "bright-blue", "bright-magenta", "bright-cyan", "bright-white"}
			classes = append(classes, "ansi-"+colors[code-90])
		}
	}

	if len(classes) == 0 {
		return ""
	}
	return strings.Join(classes, " ")
}

const CapBufferLimit = 10000

// CleanTuiMarkers removes the interactive TUI blocks from the log data
// based on the special OSC markers injected by the vuln command.
func CleanTuiMarkers(input []byte) []byte {
	return reTuiBlock.ReplaceAll(input, []byte(""))
}
