package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

func OpenFile(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func GetTerminalWidth() int {
	ws := &winsize{}
	ret, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(ret) == 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 80
}

func CenterBlock(lines []string) []string {
	width := GetTerminalWidth()
	var centered []string

	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	padding := (width - maxLen) / 2
	if padding < 0 {
		padding = 0
	}
	padStr := strings.Repeat(" ", padding)

	for _, line := range lines {
		centered = append(centered, padStr+line)
	}
	return centered
}

func PrintCenteredBlock(lines []string) {
	centered := CenterBlock(lines)
	for _, line := range centered {
		fmt.Println(line)
	}
}

func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func StripANSI(str string) string {
	var buffer []rune
	cursor := 0

	// State machine states
	const (
		StateNormal = iota
		StateEscape
		StateCSI
	)
	state := StateNormal
	var csiArgs []int
	var currentArg int
	var hasArg bool

	for _, r := range str {
		switch state {
		case StateNormal:
			if r == '\x1b' {
				state = StateEscape
			} else if r == '\x08' { // Backspace
				if cursor > 0 {
					cursor--
				}
			} else if r == '\r' { // Carriage Return
				cursor = 0
			} else if r == '\n' {
				// We treat newline as append-only in this context
				buffer = append(buffer, r)
				cursor = len(buffer)
			} else if r >= 0x20 || r == '\t' {
				// Regular character
				if cursor < len(buffer) {
					buffer[cursor] = r
				} else {
					buffer = append(buffer, r)
				}
				cursor++
			}
		case StateEscape:
			if r == '[' {
				state = StateCSI
				csiArgs = []int{}
				currentArg = 0
				hasArg = false
			} else {
				state = StateNormal
			}
		case StateCSI:
			if r >= '0' && r <= '9' {
				currentArg = currentArg*10 + int(r-'0')
				hasArg = true
			} else if r == ';' {
				csiArgs = append(csiArgs, currentArg)
				currentArg = 0
				hasArg = false // arg separator
			} else if r >= 0x40 && r <= 0x7E {
				// Final byte of CSI
				if hasArg {
					csiArgs = append(csiArgs, currentArg)
				}

				// Process recognized commands
				switch r {
				case 'K': // Erase in Line
					mode := 0
					if len(csiArgs) > 0 {
						mode = csiArgs[0]
					}
					switch mode {
					case 0: // Erase from cursor to end
						if cursor < len(buffer) {
							buffer = buffer[:cursor]
						}
					case 1: // Erase from start to cursor
						for i := 0; i < cursor && i < len(buffer); i++ {
							buffer[i] = ' '
						}
					case 2: // Erase entire line
						buffer = []rune{}
						cursor = 0
					}
				case 'G': // Cursor Horizontal Absolute
					col := 1
					if len(csiArgs) > 0 {
						col = csiArgs[0]
					}
					if col < 1 {
						col = 1
					}
					cursor = col - 1
					for len(buffer) < cursor {
						buffer = append(buffer, ' ')
					}
				}

				state = StateNormal
			}
		}
	}

	return string(buffer)
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}
	runes := []rune(str)
	if len(runes) > length {
		return string(runes[:length])
	}
	return str
}

func PrintBox(title string, lines []string) {
	width := GetTerminalWidth()
	contentWidth := 0
	for _, line := range lines {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}
	if len(title)+2 > contentWidth {
		contentWidth = len(title) + 2
	}

	boxWidth := contentWidth + 4

	leftMargin := (width - boxWidth) / 2
	if leftMargin < 0 {
		leftMargin = 0
	}
	margin := strings.Repeat(" ", leftMargin)

	topLeft := "┌"
	topRight := "┐"
	bottomLeft := "└"
	bottomRight := "┘"
	horizontal := "─"
	vertical := "│"

	fmt.Printf("%s%s%s%s\n", margin, topLeft, strings.Repeat(horizontal, boxWidth-2), topRight)

	if title != "" {
		fmt.Printf("%s%s %- *s %s\n", margin, vertical, boxWidth-4, title, vertical)
		fmt.Printf("%s%s%s%s\n", margin, "├", strings.Repeat(horizontal, boxWidth-2), "┤")
	}

	for _, line := range lines {
		fmt.Printf("%s%s %- *s %s\n", margin, vertical, boxWidth-4, line, vertical)
	}

	fmt.Printf("%s%s%s%s\n", margin, bottomLeft, strings.Repeat(horizontal, boxWidth-2), bottomRight)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cKiB", float64(b)/float64(div), "KMGTPE"[exp])
}
