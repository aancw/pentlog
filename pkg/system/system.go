package system

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/config"
	"runtime"
)

func CheckDependencies() error {
	deps := []string{"script"}
	if runtime.GOOS == "linux" {
		deps = append(deps, "scriptreplay")
	}

	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH", dep)
		}
	}
	return nil
}

func CheckReplayDependencies() error {
	if runtime.GOOS == "darwin" {
		return fmt.Errorf("session replay is not natively supported on macOS")
	}
	if _, err := exec.LookPath("scriptreplay"); err != nil {
		return fmt.Errorf("'scriptreplay' command not found")
	}
	return nil
}

func EnsureLogDir() (string, error) {
	dir, err := config.GetLogsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create log dir: %w", err)
	}
	return dir, nil
}
