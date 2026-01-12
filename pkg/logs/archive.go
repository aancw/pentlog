package logs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"strings"
	"time"
)

type ArchiveItem struct {
	Client      string
	Filename    string
	Path        string
	DisplayPath string
	Size        int64
	ModTime     time.Time
}

func ArchiveSessions(clientName, engagement, phase string, olderThan time.Duration, deleteOriginals bool) (int, error) {
	allSessions, err := ListSessions()
	if err != nil {
		return 0, err
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
	archiveFilename := fmt.Sprintf("%s.tar.gz", timestamp)
	archivePath := filepath.Join(clientArchiveDir, archiveFilename)

	file, err := os.Create(archivePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	var filesToDelete []string

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

			if err := addFileToTar(tw, fPath, s.DisplayPath); err != nil {
				os.Remove(archivePath)
				return 0, fmt.Errorf("failed to add file %s to archive: %w", fPath, err)
			}

			if deleteOriginals {
				filesToDelete = append(filesToDelete, fPath)
			}
		}
	}

	tw.Close()
	gw.Close()
	file.Close()

	if deleteOriginals {
		for _, fPath := range filesToDelete {
			os.Remove(fPath)
		}
	}

	return len(toArchive), nil
}

func addFileToTar(tw *tar.Writer, path string, baseDirHints string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	logsDir, _ := config.GetLogsDir()
	if rel, err := filepath.Rel(logsDir, path); err == nil {
		header.Name = rel
	} else {
		header.Name = filepath.Base(path)
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
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
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".tar.gz") {
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
