package logs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
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
	// Wrapper for backward compatibility, currently fetching all (limit=0 means no limit in our logic, or we can pass -1)
	// Passing -1 as limit to fetch all
	return ListSessionsPaginated(-1, 0)
}

func ListSessionsPaginated(limit, offset int) ([]Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var migrationDone string
	err = database.QueryRow("SELECT value FROM schema_info WHERE key = 'legacy_import_complete'").Scan(&migrationDone)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if migrationDone != "true" {
		if err := SyncSessions(); err != nil {
			return nil, fmt.Errorf("sync failed: %w", err)
		}
	}

	query := "SELECT id, client, engagement, scope, operator, phase, timestamp, filename, relative_path, size FROM sessions ORDER BY timestamp DESC"
	var args []interface{}

	if limit >= 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	rootDir, _ := config.GetLogsDir()

	for rows.Next() {
		var s Session
		var client, engagement, scope, operator, phase, timestamp, filename, relPath string
		var size int64
		var id int

		if err := rows.Scan(&id, &client, &engagement, &scope, &operator, &phase, &timestamp, &filename, &relPath, &size); err != nil {
			continue
		}

		s.ID = id
		s.Filename = filename
		s.Path = filepath.Join(rootDir, relPath)
		s.DisplayPath = relPath
		s.Size = size
		s.Metadata = SessionMetadata{
			Client:     client,
			Engagement: engagement,
			Scope:      scope,
			Operator:   operator,
			Phase:      phase,
			Timestamp:  timestamp,
		}

		s.MetaPath = strings.Replace(s.Path, ".tty", ".json", 1)
		s.NotesPath = strings.Replace(s.Path, ".tty", ".notes.json", 1)

		if ts, err := time.Parse(time.RFC3339, timestamp); err == nil {
			s.ModTime = ts.Format("2006-01-02 15:04:05")
			s.SortKey = ts
		} else {
			s.ModTime = timestamp
		}

		// Verify session file exists on disk for evidence integrity
		if _, err := os.Stat(s.Path); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "WARNING: Session %d references missing file: %s\n", id, s.Path)
			// Continue to list session even if file is missing for visibility
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

func GetSession(id int) (*Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var s Session
	var client, engagement, scope, operator, phase, timestamp, filename, relPath string
	var size int64

	row := database.QueryRow("SELECT client, engagement, scope, operator, phase, timestamp, filename, relative_path, size FROM sessions WHERE id = ?", id)

	if err := row.Scan(&client, &engagement, &scope, &operator, &phase, &timestamp, &filename, &relPath, &size); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session ID %d not found", id)
		}
		return nil, err
	}

	rootDir, _ := config.GetLogsDir()

	s.ID = id
	s.Filename = filename
	s.Path = filepath.Join(rootDir, relPath)
	s.DisplayPath = relPath
	s.Size = size
	s.Metadata = SessionMetadata{
		Client:     client,
		Engagement: engagement,
		Scope:      scope,
		Operator:   operator,
		Phase:      phase,
		Timestamp:  timestamp,
	}
	s.MetaPath = strings.Replace(s.Path, ".tty", ".json", 1)
	s.NotesPath = strings.Replace(s.Path, ".tty", ".notes.json", 1)

	if ts, err := time.Parse(time.RFC3339, timestamp); err == nil {
		s.ModTime = ts.Format("2006-01-02 15:04:05")
		s.SortKey = ts
	} else {
		s.ModTime = timestamp
	}

	return &s, nil
}

func SyncSessions() error {
	rootDir, err := config.GetLogsDir()
	if err != nil {
		return err
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	existsStmt, err := database.Prepare("SELECT id FROM sessions WHERE relative_path = ?")
	if err != nil {
		return err
	}
	defer existsStmt.Close()

	insertStmt, err := database.Prepare(`
		INSERT INTO sessions (client, engagement, scope, operator, phase, timestamp, filename, relative_path, size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	fmt.Println("Detected legacy session storage (JSON).")
	fmt.Println("Migrating session metadata to the new database...")

	seenContexts := make(map[string]bool)

	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tty") {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		var dummyId int
		if err := existsStmt.QueryRow(relPath).Scan(&dummyId); err == nil {
			return nil
		}

		metaPath := strings.Replace(path, ".tty", ".json", 1)
		meta, err := loadMetadata(metaPath)
		if err != nil {
			meta = SessionMetadata{
				Client:    "Unknown",
				Phase:     "Unknown",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			if info, err := d.Info(); err == nil {
				meta.Timestamp = info.ModTime().Format(time.RFC3339)
			}
		}

		info, _ := d.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}

		_, err = insertStmt.Exec(
			meta.Client,
			meta.Engagement,
			meta.Scope,
			meta.Operator,
			meta.Phase,
			meta.Timestamp,
			filepath.Base(path),
			relPath,
			size,
		)

		if err == nil {
			contextKey := fmt.Sprintf("%s/%s/%s", meta.Client, meta.Engagement, meta.Phase)
			if !seenContexts[contextKey] {
				fmt.Printf(" [+] Migrating context: %s\n", contextKey)
				seenContexts[contextKey] = true
			}
		}

		return err
	})

	if err != nil {
		return err
	}

	fmt.Println(" [✓] Migration complete.")
	fmt.Println("--------------------------------------------------")

	_, err = database.Exec("INSERT OR REPLACE INTO schema_info (key, value) VALUES ('legacy_import_complete', 'true')")
	return err
}

func AddSessionToDB(meta SessionMetadata, absLogPath string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	rootDir, err := config.GetLogsDir()
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(rootDir, absLogPath)
	if err != nil {
		return err
	}

	_, err = database.Exec(`
		INSERT INTO sessions (client, engagement, scope, operator, phase, timestamp, filename, relative_path, size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, meta.Client, meta.Engagement, meta.Scope, meta.Operator, meta.Phase, meta.Timestamp, filepath.Base(absLogPath), relPath, 0)

	return err
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

	if err := os.MkdirAll(filepath.Dir(notesPath), 0700); err != nil {
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

func ScanSessionsFromDir(rootDir string) ([]Session, error) {

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
		case ".tty":
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
