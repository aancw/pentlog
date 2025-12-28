package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"pentlog/pkg/config"
)

type Session struct {
	ID       int
	Filename string
	Path     string
	ModTime  string
	Size     int64
}

func ListSessions() ([]Session, error) {
	files, err := os.ReadDir(config.TlogLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Session{}, nil
		}
		return nil, err
	}

	var fileInfos []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			fileInfos = append(fileInfos, f)
		}
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		iInfo, _ := fileInfos[i].Info()
		jInfo, _ := fileInfos[j].Info()
		return iInfo.ModTime().Before(jInfo.ModTime())
	})

	var sessions []Session
	for i, f := range fileInfos {
		info, _ := f.Info()
		sessions = append(sessions, Session{
			ID:       i + 1,
			Filename: f.Name(),
			Path:     filepath.Join(config.TlogLogPath, f.Name()),
			ModTime:  info.ModTime().Format("2006-01-02 15:04:05"),
			Size:     info.Size(),
		})
	}

	return sessions, nil
}

func GetSessionPath(id int) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}
	for _, s := range sessions {
		if s.ID == id {
			return s.Path, nil
		}
	}
	return "", fmt.Errorf("session ID %d not found", id)
}