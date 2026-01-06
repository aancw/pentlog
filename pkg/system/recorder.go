package system

import (
	"fmt"
	"os/exec"
)

type Recorder interface {
	BuildCommand(timingFile, logFile string) (*exec.Cmd, error)
	SupportsTiming() bool
}

func NewRecorder() Recorder {
	return &TtyrecRecorder{}
}

type TtyrecRecorder struct{}

func (t *TtyrecRecorder) BuildCommand(timingFile, logFile string) (*exec.Cmd, error) {
	path, err := exec.LookPath("ttyrec")
	if err != nil {
		return nil, fmt.Errorf("'ttyrec' command not found")
	}

	return exec.Command(path, "-a", logFile), nil
}

func (t *TtyrecRecorder) SupportsTiming() bool {
	return true
}
