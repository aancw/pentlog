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

	reControl := regexp.MustCompile(`[\x08\r]`)
	return reControl.ReplaceAllString(stripped, "")
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
