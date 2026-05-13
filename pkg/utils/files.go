package utils

import "os"

const PrivateFileMode = 0o600

func WritePrivateFile(path string, data []byte) error {
	return os.WriteFile(path, data, PrivateFileMode)
}

func OpenPrivateFile(path string, flags int) (*os.File, error) {
	f, err := os.OpenFile(path, flags, PrivateFileMode)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(path, PrivateFileMode); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}

func CreatePrivateFile(path string) (*os.File, error) {
	return OpenPrivateFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
}
