package cmd

import (
	"testing"

	"github.com/google/go-github/v80/github"
)

func TestFindCompatibleAsset(t *testing.T) {
	nameLinux := "pentlog-linux-amd64"
	nameWindows := "pentlog-windows-amd64.exe"
	nameDarwin := "pentlog-darwin-arm64"
	nameChecksum := "pentlog-linux-amd64.sha256"
	nameOther := "other-file.txt"

	assets := []*github.ReleaseAsset{
		{Name: &nameLinux},
		{Name: &nameWindows},
		{Name: &nameDarwin},
		{Name: &nameChecksum},
		{Name: &nameOther},
	}

	tests := []struct {
		goos     string
		goarch   string
		expected *string
	}{
		{"linux", "amd64", &nameLinux},
		{"darwin", "arm64", &nameDarwin},
		{"windows", "amd64", nil}, // Windows not supported
		{"linux", "arm64", nil},   // Not in list
		{"darwin", "amd64", nil},  // Not in list
	}

	for _, tt := range tests {
		result := findCompatibleAsset(assets, tt.goos, tt.goarch)
		if tt.expected == nil {
			if result != nil {
				t.Errorf("expected nil for %s/%s, got %s", tt.goos, tt.goarch, result.GetName())
			}
		} else {
			if result == nil {
				t.Errorf("expected %s for %s/%s, got nil", *tt.expected, tt.goos, tt.goarch)
			} else if result.GetName() != *tt.expected {
				t.Errorf("expected %s for %s/%s, got %s", *tt.expected, tt.goos, tt.goarch, result.GetName())
			}
		}
	}
}
