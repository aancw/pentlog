package system

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/config"
)

func CheckDependencies() error {
	deps := []string{"script", "scriptreplay"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH", dep)
		}
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
