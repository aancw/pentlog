package config

import (
	"os"
	"path/filepath"
)

const (
	// System Paths (Linux)
	TlogLogPath        = "/var/log/tlog"
	TlogConfigPath     = "/etc/tlog/tlog-rec-session.conf"
	PamSSHD            = "/etc/pam.d/sshd"
	PamTlogLine        = "session required pam_tlog.so"

	// User Paths
	PentlogDirName     = ".pentlog"
	ContextFileName    = "context.json"
	HashesDirName      = "hashes"
	ExtractsDirName    = "extracts"
	HashesFileName     = "sha256.txt"
)

func GetUserPentlogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, PentlogDirName), nil
}

func GetContextFilePath() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ContextFileName), nil
}

func GetHashesDir() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, HashesDirName), nil
}

func GetExtractsDir() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ExtractsDirName), nil
}
