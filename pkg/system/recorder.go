package system

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Recorder interface {
	BuildCommand(timingFile, logFile string, shellArgs ...string) (*exec.Cmd, error)
	SupportsTiming() bool
}

func NewRecorder() Recorder {
	return &TtyrecRecorder{}
}

type TtyrecRecorder struct{}

func (t *TtyrecRecorder) BuildCommand(timingFile, logFile string, shellArgs ...string) (*exec.Cmd, error) {
	path, err := exec.LookPath("ttyrec")
	if err != nil {
		return nil, fmt.Errorf("'ttyrec' command not found")
	}

	args := []string{"-a"}
	
	if runtime.GOOS == "darwin" {
		// macOS ttyrec: positional argument for file
		args = append(args, logFile)
	} else {
		// Linux and others: use -f flag to specify output file
		args = append(args, "-f", logFile)
	}
	
	// If shellArgs provided, use -- to separate ttyrec args from shell command
	if len(shellArgs) > 0 {
		args = append(args, "--")
		args = append(args, shellArgs...)
	}
	
	return exec.Command(path, args...), nil
}

func (t *TtyrecRecorder) SupportsTiming() bool {
	return true
}
