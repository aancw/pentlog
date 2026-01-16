package logs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
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

func ArchiveSessions(clientName, engagement, phase string, olderThan time.Duration, deleteOriginals bool) (int, error) {
	toArchive, err := GetSessionsToArchive(clientName, engagement, phase, olderThan)
	if err != nil {
		return 0, err
	}
	return ArchiveSessionsFromList(toArchive, clientName, deleteOriginals, nil)
}

func ArchiveSessionsFromList(toArchive []Session, clientName string, deleteOriginals bool, extraFiles []string) (int, error) {
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

			if err := addFileToTar(tw, fPath, targetPath); err != nil {
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
			targetPath = filepath.Join("reports", utils.Slugify(clientName), filepath.Base(extraFile))
		}

		if err := addFileToTar(tw, extraFile, targetPath); err != nil {
			os.Remove(archivePath)
			return 0, fmt.Errorf("failed to add extra file %s to archive: %w", extraFile, err)
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

func addFileToTar(tw *tar.Writer, path string, targetName string) error {
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

	if targetName != "" {
		header.Name = targetName
	} else {
		logsDir, _ := config.GetLogsDir()
		if rel, err := filepath.Rel(logsDir, path); err == nil && !strings.HasPrefix(rel, "..") {
			header.Name = rel
		} else {
			header.Name = filepath.Base(path)
		}
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

func LoadSessionsFromArchive(archivePath string) ([]Session, string, error) {
	tempDir, err := os.MkdirTemp("", "pentlog_export_*")
	if err != nil {
		return nil, "", err
	}

	if err := ExtractTarGz(archivePath, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", err
	}

	sessions, err := ScanSessionsFromDir(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", err
	}

	return sessions, tempDir, nil
}

func ExtractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
