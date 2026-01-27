package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"strings"
	"time"

	"github.com/yeka/zip"
)

type ImportResult struct {
	TotalFiles       int
	ImportedFiles    int
	SkippedFiles     int
	Errors           []string
	ImportedSessions []Session
}

type ImportOptions struct {
	Password          string
	TargetClient      string
	TargetEngagement  string
	TargetPhase       string
	OverwriteExisting bool
}

type ArchiveType int

const (
	ArchiveTypePentlog ArchiveType = iota
	ArchiveTypeGeneric
)

func DetectArchiveType(archivePath, password string) (ArchiveType, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return ArchiveTypeGeneric, fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.IsEncrypted() && password == "" {
			return ArchiveTypeGeneric, fmt.Errorf("archive is password protected")
		}

		if strings.HasPrefix(f.Name, "logs/") {
			return ArchiveTypePentlog, nil
		}
	}

	return ArchiveTypeGeneric, nil
}

func ListArchiveContents(archivePath, password string) ([]string, bool, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	var files []string
	needsPassword := false

	for _, f := range r.File {
		if f.IsEncrypted() {
			needsPassword = true
			if password == "" {
				continue
			}
			f.SetPassword(password)
		}
		files = append(files, f.Name)
	}

	return files, needsPassword, nil
}

func ImportFromPentlogArchive(archivePath string, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	mgr := config.Manager()
	logsDir := mgr.GetPaths().LogsDir
	reportsDir := mgr.GetPaths().ReportsDir

	tempDir, err := os.MkdirTemp("", "pentlog-import-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	extractedFiles := make(map[string]string)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.IsEncrypted() {
			if opts.Password == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("file %s is encrypted, password required", f.Name))
				result.SkippedFiles++
				continue
			}
			f.SetPassword(opts.Password)
		}

		result.TotalFiles++

		rc, err := f.Open()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to open %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		tempPath := filepath.Join(tempDir, f.Name)
		if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
			rc.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create dir for %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		outFile, err := os.Create(tempPath)
		if err != nil {
			rc.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create temp file for %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to extract %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		extractedFiles[f.Name] = tempPath
	}

	for archivePath, tempPath := range extractedFiles {
		if !strings.HasSuffix(archivePath, ".tty") {
			continue
		}

		var destPath string
		if strings.HasPrefix(archivePath, "logs/") {
			relPath := strings.TrimPrefix(archivePath, "logs/")
			destPath = filepath.Join(logsDir, relPath)
		} else if strings.HasPrefix(archivePath, "reports/") {
			relPath := strings.TrimPrefix(archivePath, "reports/")
			destPath = filepath.Join(reportsDir, relPath)
			if err := copyFile(tempPath, destPath, opts.OverwriteExisting); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to copy report %s: %v", archivePath, err))
				result.SkippedFiles++
			} else {
				result.ImportedFiles++
			}
			continue
		}

		if !opts.OverwriteExisting {
			if _, err := os.Stat(destPath); err == nil {
				result.Errors = append(result.Errors, fmt.Sprintf("file already exists: %s", destPath))
				result.SkippedFiles++
				continue
			}
		}

		if err := copyFile(tempPath, destPath, opts.OverwriteExisting); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to copy %s: %v", archivePath, err))
			result.SkippedFiles++
			continue
		}

		jsonArchivePath := strings.TrimSuffix(archivePath, ".tty") + ".json"
		if jsonTempPath, exists := extractedFiles[jsonArchivePath]; exists {
			jsonDestPath := strings.TrimSuffix(destPath, ".tty") + ".json"
			if err := copyFile(jsonTempPath, jsonDestPath, opts.OverwriteExisting); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to copy metadata %s: %v", jsonArchivePath, err))
			}
		}

		notesArchivePath := strings.TrimSuffix(archivePath, ".tty") + ".notes.json"
		if notesTempPath, exists := extractedFiles[notesArchivePath]; exists {
			notesDestPath := strings.TrimSuffix(destPath, ".tty") + ".notes.json"
			if err := copyFile(notesTempPath, notesDestPath, opts.OverwriteExisting); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to copy notes %s: %v", notesArchivePath, err))
			}
		}

		meta := SessionMetadata{}
		metaPath := strings.TrimSuffix(destPath, ".tty") + ".json"
		if f, err := os.Open(metaPath); err == nil {
			json.NewDecoder(f).Decode(&meta)
			f.Close()
		}

		if meta.Client == "" {
			parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(archivePath, "logs/"), "/"), "/")
			if len(parts) >= 3 {
				meta.Client = parts[0]
				meta.Engagement = parts[1]
				meta.Phase = parts[2]
			} else if len(parts) >= 1 {
				meta.Client = parts[0]
			}
		}

		if meta.Timestamp == "" {
			info, _ := os.Stat(destPath)
			if info != nil {
				meta.Timestamp = info.ModTime().Format(time.RFC3339)
			} else {
				meta.Timestamp = time.Now().Format(time.RFC3339)
			}
		}

		sessionID, err := AddSessionToDBWithState(meta, destPath, SessionStateCompleted)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to add session to DB: %v", err))
		} else {
			session, _ := GetSession(int(sessionID))
			if session != nil {
				result.ImportedSessions = append(result.ImportedSessions, *session)
			}
		}

		result.ImportedFiles++
	}

	for archivePath, tempPath := range extractedFiles {
		if !strings.HasPrefix(archivePath, "reports/") {
			continue
		}

		relPath := strings.TrimPrefix(archivePath, "reports/")
		destPath := filepath.Join(reportsDir, relPath)

		if err := copyFile(tempPath, destPath, opts.OverwriteExisting); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to copy report %s: %v", archivePath, err))
			result.SkippedFiles++
		} else {
			result.ImportedFiles++
		}
	}

	return result, nil
}

func ImportFromGenericArchive(archivePath string, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	if opts.TargetClient == "" {
		return nil, fmt.Errorf("target client is required for generic archives")
	}
	if opts.TargetPhase == "" {
		opts.TargetPhase = "imported"
	}
	if opts.TargetEngagement == "" {
		opts.TargetEngagement = "imported"
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	mgr := config.Manager()
	logsDir := mgr.GetPaths().LogsDir

	tempDir, err := os.MkdirTemp("", "pentlog-import-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(logsDir, opts.TargetClient, opts.TargetEngagement, opts.TargetPhase)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	extractedFiles := make(map[string]string)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.IsEncrypted() {
			if opts.Password == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("file %s is encrypted, password required", f.Name))
				result.SkippedFiles++
				continue
			}
			f.SetPassword(opts.Password)
		}

		result.TotalFiles++

		rc, err := f.Open()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to open %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		tempPath := filepath.Join(tempDir, f.Name)
		if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
			rc.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create dir for %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		outFile, err := os.Create(tempPath)
		if err != nil {
			rc.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create temp file for %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to extract %s: %v", f.Name, err))
			result.SkippedFiles++
			continue
		}

		extractedFiles[f.Name] = tempPath
	}

	for archivePath, tempPath := range extractedFiles {
		if !strings.HasSuffix(archivePath, ".tty") {
			continue
		}

		destPath := filepath.Join(targetDir, filepath.Base(archivePath))

		if !opts.OverwriteExisting {
			if _, err := os.Stat(destPath); err == nil {
				result.Errors = append(result.Errors, fmt.Sprintf("file already exists: %s", destPath))
				result.SkippedFiles++
				continue
			}
		}

		if err := copyFile(tempPath, destPath, opts.OverwriteExisting); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to copy %s: %v", archivePath, err))
			result.SkippedFiles++
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(archivePath), ".tty")
		meta := SessionMetadata{
			Client:     opts.TargetClient,
			Engagement: opts.TargetEngagement,
			Phase:      opts.TargetPhase,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		for jsonArchive, jsonTemp := range extractedFiles {
			if strings.TrimSuffix(filepath.Base(jsonArchive), ".json") == baseName && strings.HasSuffix(jsonArchive, ".json") && !strings.HasSuffix(jsonArchive, ".notes.json") {
				jsonDestPath := strings.TrimSuffix(destPath, ".tty") + ".json"
				if err := copyFile(jsonTemp, jsonDestPath, opts.OverwriteExisting); err == nil {
					// Load metadata from json
					if f, err := os.Open(jsonDestPath); err == nil {
						var loadedMeta SessionMetadata
						if json.NewDecoder(f).Decode(&loadedMeta) == nil {
							if loadedMeta.Timestamp != "" {
								meta.Timestamp = loadedMeta.Timestamp
							}
							if loadedMeta.Operator != "" {
								meta.Operator = loadedMeta.Operator
							}
							if loadedMeta.Scope != "" {
								meta.Scope = loadedMeta.Scope
							}
						}
						f.Close()
					}
				}
				break
			}
		}

		for notesArchive, notesTemp := range extractedFiles {
			if strings.TrimSuffix(filepath.Base(notesArchive), ".notes.json") == baseName && strings.HasSuffix(notesArchive, ".notes.json") {
				notesDestPath := strings.TrimSuffix(destPath, ".tty") + ".notes.json"
				copyFile(notesTemp, notesDestPath, opts.OverwriteExisting)
				break
			}
		}

		metaPath := strings.TrimSuffix(destPath, ".tty") + ".json"
		if _, err := os.Stat(metaPath); os.IsNotExist(err) {
			if metaData, err := json.MarshalIndent(meta, "", "  "); err == nil {
				os.WriteFile(metaPath, metaData, 0644)
			}
		}

		sessionID, err := AddSessionToDBWithState(meta, destPath, SessionStateCompleted)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to add session to DB: %v", err))
		} else {
			session, _ := GetSession(int(sessionID))
			if session != nil {
				result.ImportedSessions = append(result.ImportedSessions, *session)
			}
		}

		result.ImportedFiles++
	}

	return result, nil
}

func copyFile(src, dst string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("file exists")
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func CheckArchivePassword(archivePath, password string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.IsEncrypted() {
			f.SetPassword(password)
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("incorrect password or corrupted archive")
			}
			buf := make([]byte, 1)
			_, err = rc.Read(buf)
			rc.Close()
			if err != nil && err != io.EOF {
				return fmt.Errorf("incorrect password")
			}
			return nil
		}
	}

	return nil
}
