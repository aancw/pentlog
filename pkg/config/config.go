package config

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	PentlogDirName  = ".pentlog"
	ContextFileName = "context.json"
	HashesDirName   = "hashes"
	ExtractsDirName = "extracts"
	HashesFileName  = "sha256.txt"
	LogsDirName     = "logs"
	ReportsDirName  = "reports"
	ArchiveDirName  = "archive"
)

func GetUserPentlogDir() (string, error) {
	if testHome := os.Getenv("PENTLOG_TEST_HOME"); testHome != "" {
		return filepath.Join(testHome, PentlogDirName), nil
	}

	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		u, err := user.Lookup(sudoUser)
		if err == nil {
			return filepath.Join(u.HomeDir, PentlogDirName), nil
		}
	}

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

func GetLogsDir() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, LogsDirName), nil
}

func GetReportsDir() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ReportsDirName), nil
}

func GetArchiveDir() (string, error) {
	dir, err := GetUserPentlogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ArchiveDirName), nil
}
