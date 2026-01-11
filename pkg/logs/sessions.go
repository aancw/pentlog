package logs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"sort"
	"strings"
	"time"
)

type SessionMetadata struct {
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Timestamp  string `json:"timestamp"`
}

type SessionNote struct {
	Timestamp  string `json:"timestamp"`
	Content    string `json:"content"`
	ByteOffset int64  `json:"byte_offset"`
}

type Session struct {
	ID          int
	Filename    string
	Path        string
	DisplayPath string
	MetaPath    string
	NotesPath   string
	ModTime     string
	Size        int64
	Metadata    SessionMetadata
	SortKey     time.Time
}

func ListSessions() ([]Session, error) {
	rootDir, err := config.GetLogsDir()
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Session{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return []Session{}, nil
	}

	sMap := map[string]*Session{}

	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if strings.HasSuffix(path, ".notes.json") {
			base := strings.TrimSuffix(path, ".notes.json")
			session := sMap[base]
			if session == nil {
				session = &Session{}
				sMap[base] = session
			}
			session.NotesPath = path
			return nil
		}

		base := strings.TrimSuffix(path, ext)
		session := sMap[base]
		if session == nil {
			session = &Session{}
			sMap[base] = session
		}

		switch ext {
		case ".tty": // New ttyrec format
			session.Filename = filepath.Base(path)
			session.Path = path
			if rel, err := filepath.Rel(rootDir, path); err == nil {
				session.DisplayPath = rel
			} else {
				session.DisplayPath = session.Filename
			}
			if info, err := d.Info(); err == nil {
				session.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
				session.Size = info.Size()
				session.SortKey = info.ModTime()
			}

		case ".json":
			session.MetaPath = path
			if meta, err := loadMetadata(path); err == nil {
				session.Metadata = meta
				if ts, err := time.Parse(time.RFC3339, meta.Timestamp); err == nil {
					session.ModTime = ts.Format("2006-01-02 15:04:05")
					session.SortKey = ts
				}
			}
		}

		return nil
	})

	var sessions []Session
	for base, s := range sMap {
		if s.Path == "" {
			continue
		}
		if s.Filename == "" {
			s.Filename = filepath.Base(base) + ".tty"
		}
		if s.DisplayPath == "" {
			if rel, err := filepath.Rel(rootDir, s.Path); err == nil {
				s.DisplayPath = rel
			} else {
				s.DisplayPath = s.Filename
			}
		}
		sessions = append(sessions, *s)
		delete(sMap, base)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SortKey.Before(sessions[j].SortKey)
	})

	for i := range sessions {
		sessions[i].ID = i + 1
	}

	return sessions, nil
}

func GetSession(id int) (*Session, error) {
	sessions, err := ListSessions()
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		if sessions[i].ID == id {
			return &sessions[i], nil
		}
	}
	return nil, fmt.Errorf("session ID %d not found", id)
}

func loadMetadata(path string) (SessionMetadata, error) {
	var meta SessionMetadata
	f, err := os.Open(path)
	if err != nil {
		return meta, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return meta, err
	}
	return meta, nil
}

func AppendNote(notesPath string, note SessionNote) error {
	var notes []SessionNote

	if _, err := os.Stat(notesPath); err == nil {
		data, err := os.ReadFile(notesPath)
		if err == nil {
			json.Unmarshal(data, &notes)
		}
	}

	notes = append(notes, note)

	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(notesPath, data, 0644)
}

func ReadNotes(notesPath string) ([]SessionNote, error) {
	var notes []SessionNote

	data, err := os.ReadFile(notesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionNote{}, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, err
	}

	return notes, nil
}
