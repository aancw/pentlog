package config

import (
	"os"
	"path/filepath"
)

const (
	PentlogDirName  = ".pentlog"
	ContextFileName = "context.json"
	HashesDirName   = "hashes"
	ExtractsDirName = "extracts"
	HashesFileName  = "sha256.txt"
	LogsDirName     = "logs"
)

func GetUserPentlogDir() (string, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		return filepath.Join("/home", sudoUser, PentlogDirName), nil
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
