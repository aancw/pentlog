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
	
safeCheck := "pam_exec.so quiet /usr/bin/test -t 0"
	tlogLine := "pam_tlog.so"
	beginMarker := "# BEGIN PENTLOG MANAGED BLOCK"
	endMarker := "# END PENTLOG MANAGED BLOCK"

	hasSafeBlock := false
	needsFix := false

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	for i, line := range lines {
		if strings.Contains(line, tlogLine) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			if i > 0 && strings.Contains(lines[i-1], safeCheck) && !strings.HasPrefix(strings.TrimSpace(lines[i-1]), "#") {
				if i > 1 && strings.TrimSpace(lines[i-2]) == beginMarker {
					hasSafeBlock = true
				} else {
					needsFix = true
				}
			} else {
				needsFix = true
			}
		}
	}

	if hasSafeBlock && !needsFix {
		return false, nil
	}

	if err := backupFile(pamFile); err != nil {
		return false, fmt.Errorf("failed to backup PAM file %s: %w", pamFile, err)
	}

	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(line, tlogLine) && !strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(line, safeCheck) && !strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == beginMarker || trimmed == endMarker {
			continue
		}
		newLines = append(newLines, line)
	}

	newLines = append(newLines, beginMarker)
	newLines = append(newLines, "session [success=ignore default=1] pam_exec.so quiet /usr/bin/test -t 0")
	newLines = append(newLines, config.PamTlogLine)
	newLines = append(newLines, endMarker)

	output := strings.Join(newLines, "\n") + "\n"
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

    safeCheck := "pam_exec.so quiet /usr/bin/test -t 0"
	beginMarker := "# BEGIN PENTLOG MANAGED BLOCK"
	endMarker := "# END PENTLOG MANAGED BLOCK"

    for _, line := range lines {
		trimmed := strings.TrimSpace(line)
        if strings.Contains(line, "pam_tlog.so") {
            changed = true
            continue
        }
        if strings.Contains(line, safeCheck) {
            changed = true
            continue
        }
		if trimmed == beginMarker || trimmed == endMarker {
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

func RestartSSHService() error {
	services := []string{"ssh", "sshd"}
	for _, s := range services {
		cmd := exec.Command("systemctl", "restart", s)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to restart ssh service")
}
