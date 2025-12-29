package system

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Recorder interface {
	BuildCommand(timingFile, logFile string) (*exec.Cmd, error)
	SupportsTiming() bool
}

func NewRecorder() Recorder {
	switch runtime.GOOS {
	case "darwin":
		return &MacRecorder{}
	default:
		return &LinuxRecorder{}
	}
}

type LinuxRecorder struct{}

func (l *LinuxRecorder) BuildCommand(timingFile, logFile string) (*exec.Cmd, error) {
	scriptPath, err := exec.LookPath("script")
	if err != nil {
		return nil, fmt.Errorf("'script' command not found")
	}

	return exec.Command(scriptPath, "--quiet", "--flush", "--append", "-t"+timingFile, logFile), nil
}

func (l *LinuxRecorder) SupportsTiming() bool {
	return true
}

type MacRecorder struct{}

func (m *MacRecorder) BuildCommand(timingFile, logFile string) (*exec.Cmd, error) {
	scriptPath, err := exec.LookPath("script")
	if err != nil {
		return nil, fmt.Errorf("'script' command not found")
	}

	// macOS/BSD flags: -q (quiet) -F (flush) -a (append)
	// Note: Standard BSD script on macOS does not easily support separate timing files like Linux.
	return exec.Command(scriptPath, "-q", "-F", "-a", logFile), nil
}

func (m *MacRecorder) SupportsTiming() bool {
	return false
}
