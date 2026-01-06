package system

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/config"
)

func CheckDependencies() error {
	deps := []string{"ttyrec"}
	deps = append(deps, "ttyplay")

	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH. Please install it (e.g., 'brew install ttyrec' or 'apt install ttyrec')", dep)
		}
	}
	return nil
}

func CheckReplayDependencies() error {
	if _, err := exec.LookPath("ttyplay"); err != nil {
		return fmt.Errorf("'ttyplay' command not found")
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

func IsSetupRun() (bool, error) {
	dir, err := config.GetLogsDir()
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}
