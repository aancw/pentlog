package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

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
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	stripped := re.ReplaceAllString(str, "")

	var stack []rune
	for _, r := range stripped {
		if r == '\x08' { // Backspace
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		} else if r == '\r' { // Carriage Return

			continue
		} else if r >= 0x20 || r == '\n' || r == '\t' {
			stack = append(stack, r)
		}
	}

	return string(stack)
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
