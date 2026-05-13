package logs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const archiveManifestName = "manifest.json"

type ArchiveManifest struct {
	FormatVersion   int                   `json:"format_version"`
	GeneratedAt     string                `json:"generated_at"`
	Client          string                `json:"client"`
	Encrypted       bool                  `json:"encrypted"`
	DeleteOriginals bool                  `json:"delete_originals"`
	Files           []ArchiveManifestFile `json:"files"`
}

type ArchiveManifestFile struct {
	ArchivePath string `json:"archive_path"`
	Role        string `json:"role"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
}

func buildArchiveManifest(clientName string, deleteOriginals bool, encrypted bool, files []ArchiveManifestFile) ArchiveManifest {
	return ArchiveManifest{
		FormatVersion:   1,
		GeneratedAt:     time.Now().Format(time.RFC3339),
		Client:          clientName,
		Encrypted:       encrypted,
		DeleteOriginals: deleteOriginals,
		Files:           files,
	}
}

func manifestJSON(manifest ArchiveManifest) ([]byte, error) {
	return json.MarshalIndent(manifest, "", "  ")
}

func hashFileSHA256(path string) (string, int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), info.Size(), nil
}

func loadArchiveManifest(path string) (*ArchiveManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest ArchiveManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	if manifest.FormatVersion <= 0 {
		return nil, fmt.Errorf("invalid manifest format version")
	}

	return &manifest, nil
}
