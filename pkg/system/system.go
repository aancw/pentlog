package system

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"pentlog/pkg/config"
)

func IsRoot() bool {
	return os.Geteuid() == 0
}

func CheckDependencies() error {
	deps := []string{"tlog-rec-session", "tlog-play", "jq"}
	for _, dep := range deps {
		_, err := exec.LookPath(dep)
		if err != nil {
			return fmt.Errorf("dependency missing: %s. Please install it using your package manager (e.g., apt, dnf, pacman)", dep)
		}
	}
	return nil
}

func DetectLocalPamFile() (string, error) {
	candidates := []string{
		"/etc/pam.d/common-session",
		"/etc/pam.d/system-auth",
		"/etc/pam.d/login",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("could not detect a standard PAM session file (checked: %v). Please configure PAM manually", candidates)
}

func backupFile(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	timestamp := time.Now().Format("20060102150405")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)

	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return os.Chmod(backupPath, info.Mode())
}

func EnsureTlogConfig() error {
	configDir := filepath.Dir(config.TlogConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create tlog config dir: %w", err)
	}

	if err := backupFile(config.TlogConfigPath); err != nil {
		return fmt.Errorf("failed to backup existing tlog config: %w", err)
	}

	content := `{
    "write_mode": "immediate",
    "log_path": "/var/log/tlog"
}`
	if err := os.WriteFile(config.TlogConfigPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write tlog config: %w", err)
	}

	if err := os.MkdirAll(config.TlogLogPath, 0700); err != nil {
		return fmt.Errorf("failed to create tlog log dir: %w", err)
	}
	
	return nil
}

func EnablePamTlog(pamFile string) (bool, error) {
	f, err := os.Open(pamFile)
	if err != nil {
		return false, fmt.Errorf("failed to open PAM file %s: %w", pamFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := []string{}
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if strings.Contains(line, "pam_tlog.so") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			found = true
		}
	}

	if found {
		return false, nil
	}

	if err := backupFile(pamFile); err != nil {
		return false, fmt.Errorf("failed to backup PAM file %s: %w", pamFile, err)
	}

	lines = append(lines, config.PamTlogLine)
	output := strings.Join(lines, "\n") + "\n"
	
	if err := os.WriteFile(pamFile, []byte(output), 0644); err != nil {
		return false, fmt.Errorf("failed to update PAM file %s: %w", pamFile, err)
	}

	return true, nil
}

func DisablePamTlog(pamFile string) (bool, error) {
    input, err := os.ReadFile(pamFile)
    if err != nil {
        return false, err
    }

    lines := strings.Split(string(input), "\n")
    newLines := []string{}
    changed := false

    for _, line := range lines {
        if strings.Contains(line, "pam_tlog.so") {
            changed = true
            continue
        }
        newLines = append(newLines, line)
    }

    if !changed {
        return false, nil
    }

    if err := backupFile(pamFile); err != nil {
        return false, fmt.Errorf("failed to backup PAM file %s: %w", pamFile, err)
    }

    output := strings.Join(newLines, "\n")
    if err := os.WriteFile(pamFile, []byte(output), 0644); err != nil {
        return false, err
    }
    return true, nil
}
