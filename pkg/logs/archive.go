package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
	"strings"
	"time"

	"github.com/yeka/zip"
)

type ArchiveItem struct {
	Client      string
	Filename    string
	Path        string
	DisplayPath string
	Size        int64
	ModTime     time.Time
}

func GetSessionsToArchive(clientName, engagement, phase string, olderThan time.Duration) ([]Session, error) {
	allSessions, err := ListSessions()
	if err != nil {
		return nil, err
	}

	var toArchive []Session
	now := time.Now()

	for _, s := range allSessions {
		if s.Metadata.Client != clientName {
			continue
		}
		if engagement != "" && s.Metadata.Engagement != engagement {
			continue
		}
		if phase != "" && s.Metadata.Phase != phase {
			continue
		}

		if olderThan > 0 {
			var t time.Time
			if s.Metadata.Timestamp != "" {
				parsed, err := time.Parse(time.RFC3339, s.Metadata.Timestamp)
				if err == nil {
					t = parsed
				}
			}
			if t.IsZero() {
				t = s.SortKey
			}

			if now.Sub(t) < olderThan {
				continue
			}
		}
		toArchive = append(toArchive, s)
	}
	return toArchive, nil
}

func ArchiveSessions(clientName, engagement, phase string, olderThan time.Duration, deleteOriginals bool, password string) (int, error) {
	toArchive, err := GetSessionsToArchive(clientName, engagement, phase, olderThan)
	if err != nil {
		return 0, err
	}
	return ArchiveSessionsFromList(toArchive, clientName, deleteOriginals, nil, password)
}

func ArchiveSessionsFromList(toArchive []Session, clientName string, deleteOriginals bool, extraFiles []string, password string) (int, error) {
	if len(toArchive) == 0 {
		return 0, nil
	}

	archiveDir, err := config.GetArchiveDir()
	if err != nil {
		return 0, err
	}
	clientArchiveDir := filepath.Join(archiveDir, clientName)
	if err := os.MkdirAll(clientArchiveDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create archive dir: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")

	return archiveZip(toArchive, extraFiles, clientArchiveDir, timestamp, deleteOriginals, password)
}

func archiveZip(toArchive []Session, extraFiles []string, clientArchiveDir, timestamp string, deleteOriginals bool, password string) (int, error) {
	archiveFilename := fmt.Sprintf("%s.zip", timestamp)
	archivePath := filepath.Join(clientArchiveDir, archiveFilename)

	file, err := os.Create(archivePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	var filesToDelete []string

	// Helper to add file to zip
	addFile := func(path, targetPath string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		var w io.Writer
		if password != "" {
			w, err = zw.Encrypt(targetPath, password, zip.AES256Encryption)
			if err != nil {
				return err
			}
		} else {
			// Without password, use standard Create for compatibility
			w, err = zw.Create(targetPath)
			if err != nil {
				return err
			}
		}

		_, err = io.Copy(w, f)
		return err
	}

	for _, s := range toArchive {
		files := []string{s.Path}
		if s.MetaPath != "" {
			files = append(files, s.MetaPath)
		}
		if s.NotesPath != "" {
			files = append(files, s.NotesPath)
		}

		base := strings.TrimSuffix(s.Path, filepath.Ext(s.Path))
		timingPath := base + ".timing"
		if _, err := os.Stat(timingPath); err == nil {
			if deleteOriginals {
				filesToDelete = append(filesToDelete, timingPath)
			}
		}

		for _, fPath := range files {
			if fPath == "" {
				continue
			}

			logsDir, _ := config.GetLogsDir()
			var targetPath string
			if rel, err := filepath.Rel(logsDir, fPath); err == nil && !strings.HasPrefix(rel, "..") {
				targetPath = filepath.Join("logs", rel)
			} else {
				targetPath = filepath.Join("logs", filepath.Base(fPath))
			}

			if _, err := os.Stat(fPath); os.IsNotExist(err) {
				continue
			}

			if err := addFile(fPath, targetPath); err != nil {
				zw.Close()   // Flush and close ZIP writer
				file.Close() // Close file handle
				os.Remove(archivePath)
				return 0, fmt.Errorf("failed to add file %s to archive: %w", fPath, err)
			}

			if deleteOriginals {
				filesToDelete = append(filesToDelete, fPath)
			}
		}
	}

	for _, extraFile := range extraFiles {
		reportsDir, _ := config.GetReportsDir()
		var targetPath string
		if rel, err := filepath.Rel(reportsDir, extraFile); err == nil && !strings.HasPrefix(rel, "..") {
			targetPath = filepath.Join("reports", rel)
		} else {
			targetPath = filepath.Join("reports", utils.Slugify(filepath.Base(filepath.Dir(extraFile))), filepath.Base(extraFile))
		}

		if err := addFile(extraFile, targetPath); err != nil {
			zw.Close()   // Flush and close ZIP writer
			file.Close() // Close file handle
			os.Remove(archivePath)
			return 0, fmt.Errorf("failed to add extra file %s to archive: %w", extraFile, err)
		}
	}

	if err := zw.Flush(); err != nil {
		return 0, err
	}

	if deleteOriginals {
		for _, fPath := range filesToDelete {
			os.Remove(fPath)
		}
	}

	return len(toArchive), nil
}

func ListArchives() ([]ArchiveItem, error) {
	archiveDir, err := config.GetArchiveDir()
	if err != nil {
		return nil, err
	}

	var items []ArchiveItem

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ArchiveItem{}, nil
		}
		return nil, err
	}

	for _, clientEntry := range entries {
		if !clientEntry.IsDir() {
			continue
		}
		clientName := clientEntry.Name()
		clientPath := filepath.Join(archiveDir, clientName)

		files, err := os.ReadDir(clientPath)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() || (!strings.HasSuffix(f.Name(), ".tar.gz") && !strings.HasSuffix(f.Name(), ".zip")) {
				continue
			}

			info, err := f.Info()
			if err != nil {
				continue
			}

			items = append(items, ArchiveItem{
				Client:      clientName,
				Filename:    f.Name(),
				Path:        filepath.Join(clientPath, f.Name()),
				DisplayPath: fmt.Sprintf("%s/%s", clientName, f.Name()),
				Size:        info.Size(),
				ModTime:     info.ModTime(),
			})
		}
	}

	return items, nil
}

func LoadSessionsFromArchive(archivePath string) ([]Session, string, error) {
	// Not implemented for ZIP yet as part of this task
	return nil, "", fmt.Errorf("feature not available for this archive format yet")
}

// Deprecated: ExtractTarGz functionality removed in favor of Zip default
func ExtractTarGz(archivePath, destDir string) error {
	return fmt.Errorf("function deprecated and removed")
}
