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
	"pentlog/pkg/logger"
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
	Target     string `json:"target,omitempty"`
	TargetIP   string `json:"target_ip,omitempty"`
	Timestamp  string `json:"timestamp"`
}

type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStateCompleted SessionState = "completed"
	SessionStateCrashed   SessionState = "crashed"
	SessionStatePaused    SessionState = "paused"
	SessionStateArchived  SessionState = "archived"
)

type SessionNote struct {
	Timestamp  string `json:"timestamp"`
	Content    string `json:"content"`
	ByteOffset int64  `json:"byte_offset"`
}

type Session struct {
	ID                    int
	Filename              string
	Path                  string
	DisplayPath           string
	MetaPath              string
	NotesPath             string
	ModTime               string
	Size                  int64
	Metadata              SessionMetadata
	SortKey               time.Time
	State                 SessionState
	LastSyncAt            string
	RecorderPID           int
	HostFingerprint       string
	Hostname              string
	StartedAt             string
	EndedAt               string
	ResumeCount           int
	ArchivedAt            string
	ArchivePath           string
	ArchiveManifestSHA256 string
}

func ListSessions() ([]Session, error) {
	return ListSessionsWithOptions(SessionListOptions{})
}

type SessionListOptions struct {
	IncludeArchived bool
	OnlyArchived    bool
}

type sessionScanner interface {
	Scan(dest ...interface{}) error
}

type sessionRow struct {
	ID                    int
	Client                string
	Engagement            string
	Scope                 sql.NullString
	Operator              sql.NullString
	Phase                 string
	Timestamp             string
	Filename              string
	RelativePath          string
	Size                  int64
	State                 sql.NullString
	LastSyncAt            sql.NullString
	RecorderPID           sql.NullInt64
	HostFingerprint       sql.NullString
	Hostname              sql.NullString
	StartedAt             sql.NullString
	EndedAt               sql.NullString
	ResumeCount           sql.NullInt64
	Target                sql.NullString
	TargetIP              sql.NullString
	ArchivedAt            sql.NullString
	ArchivePath           sql.NullString
	ArchiveManifestSHA256 sql.NullString
}

func ListSessionsWithOptions(opts SessionListOptions) ([]Session, error) {
	return listSessions(-1, 0, opts)
}

func ListSessionsPaginated(limit, offset int) ([]Session, error) {
	return ListSessionsPaginatedWithOptions(limit, offset, SessionListOptions{})
}

func ListSessionsPaginatedWithOptions(limit, offset int, opts SessionListOptions) ([]Session, error) {
	return listSessions(limit, offset, opts)
}

func listSessions(limit, offset int, opts SessionListOptions) ([]Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT %s FROM sessions", sessionSelectColumns(""))
	if filter := archiveFilterClause(opts); filter != "" {
		query += " WHERE " + filter
	}
	query += " ORDER BY timestamp DESC"
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

	for rows.Next() {
		row, err := scanSessionRow(rows)
		if err != nil {
			continue
		}
		sessions = append(sessions, buildSession(row))
	}

	return sessions, nil
}

func NeedsSessionSync() (bool, error) {
	database, err := db.GetDB()
	if err != nil {
		return false, err
	}

	var migrationDone string
	err = database.QueryRow("SELECT value FROM schema_info WHERE key = 'legacy_import_complete'").Scan(&migrationDone)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}

	return migrationDone != "true", nil
}

func GetSession(id int) (*Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	row := database.QueryRow(fmt.Sprintf("SELECT %s FROM sessions WHERE id = ?", sessionSelectColumns("")), id)
	sessionRow, err := scanSessionRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session ID %d not found", id)
		}
		return nil, err
	}

	s := buildSession(sessionRow)
	return &s, nil
}

func SyncSessions() error {
	mgr := config.Manager()
	rootDir := mgr.GetPaths().LogsDir

	database, err := db.GetDB()
	if err != nil {
		logger.Error("failed to get database for session sync", "error", err)
		return err
	}

	backupPath, err := db.BackupDB()
	if err != nil {
		logger.Warn("could not backup database", "error", err)
	} else if backupPath != "" {
		logger.Info("database backed up before migration", "path", backupPath)
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

	logger.Info("detected legacy session storage, starting migration")

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
				logger.Info("migrating context", "context", contextKey)
				seenContexts[contextKey] = true
			}
		}

		return err
	})

	if err != nil {
		logger.Error("session migration failed", "error", err)
		return err
	}

	logger.Info("session migration complete")

	_, err = database.Exec("INSERT OR REPLACE INTO schema_info (key, value) VALUES ('legacy_import_complete', 'true')")
	return err
}

func AddSessionToDB(meta SessionMetadata, absLogPath string) (int64, error) {
	return AddSessionToDBWithState(meta, absLogPath, SessionStateActive)
}

func AddSessionToDBWithState(meta SessionMetadata, absLogPath string, state SessionState) (int64, error) {
	database, err := db.GetDB()
	if err != nil {
		return 0, err
	}

	mgr := config.Manager()
	rootDir := mgr.GetPaths().LogsDir

	relPath, err := filepath.Rel(rootDir, absLogPath)
	if err != nil {
		return 0, err
	}

	now := time.Now().Format(time.RFC3339)
	host := defaultLifecycleHost(state)
	startedAt, endedAt := defaultLifecycleTimestamps(state, meta.Timestamp, now)
	result, err := database.Exec(`
		INSERT INTO sessions (
			client, engagement, scope, operator, phase, timestamp, filename, relative_path, size,
			state, last_sync_at, recorder_pid, host_fingerprint, hostname, started_at, ended_at,
			resume_count, target, target_ip
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		meta.Client,
		meta.Engagement,
		meta.Scope,
		meta.Operator,
		meta.Phase,
		meta.Timestamp,
		filepath.Base(absLogPath),
		relPath,
		0,
		string(state),
		now,
		0,
		host.Fingerprint,
		host.Hostname,
		startedAt,
		endedAt,
		0,
		meta.Target,
		meta.TargetIP,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateSessionState(sessionID int64, state SessionState) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	endedAt := ""
	recorderPID := interface{}(nil)
	if state == SessionStateCompleted || state == SessionStateCrashed || state == SessionStateArchived {
		endedAt = now
		recorderPID = 0
	}

	_, err = database.Exec(`
		UPDATE sessions
		SET state = ?, last_sync_at = ?, ended_at = ?, recorder_pid = COALESCE(?, recorder_pid)
		WHERE id = ?
	`, string(state), now, endedAt, recorderPID, sessionID)
	return err
}

func AttachSessionRecorder(sessionID int64, recorderPID int) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	host := currentHostIdentity()
	_, err = database.Exec(`
		UPDATE sessions
		SET recorder_pid = ?, host_fingerprint = ?, hostname = ?, state = ?, last_sync_at = ?, started_at = COALESCE(NULLIF(started_at, ''), ?), ended_at = ''
		WHERE id = ?
	`, recorderPID, host.Fingerprint, host.Hostname, string(SessionStateActive), now, now, sessionID)
	return err
}

func PauseSession(sessionID int64) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	_, err = database.Exec(`UPDATE sessions SET state = ?, last_sync_at = ?, ended_at = '' WHERE id = ?`, string(SessionStatePaused), now, sessionID)
	return err
}

func UpdateSessionHeartbeat(sessionID int64) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	_, err = database.Exec(`UPDATE sessions SET last_sync_at = ? WHERE id = ?`, now, sessionID)
	return err
}

func UpdateSessionSize(sessionID int64, size int64) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	_, err = database.Exec(`UPDATE sessions SET size = ?, last_sync_at = ? WHERE id = ?`, size, now, sessionID)
	return err
}

func GetActiveSessions() ([]Session, error) {
	return GetSessionsByState(SessionStateActive)
}

func GetCrashedSessions() ([]Session, error) {
	return GetSessionsByState(SessionStateCrashed)
}

func GetPausedSessions() ([]Session, error) {
	return GetSessionsByState(SessionStatePaused)
}

func GetSessionsByState(state SessionState) ([]Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		fmt.Sprintf("SELECT %s FROM sessions WHERE state = ? ORDER BY timestamp DESC", sessionSelectColumns("")),
		string(state),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session

	for rows.Next() {
		row, err := scanSessionRow(rows)
		if err != nil {
			continue
		}
		sessions = append(sessions, buildSession(row))
	}

	return sessions, nil
}

func GetCrashedSessionsForContext(client, engagement, phase string) ([]Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		fmt.Sprintf(
			"SELECT %s FROM sessions WHERE state = ? AND client = ? AND engagement = ? AND phase = ? ORDER BY timestamp DESC",
			sessionSelectColumns(""),
		),
		string(SessionStateCrashed), client, engagement, phase,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session

	for rows.Next() {
		row, err := scanSessionRow(rows)
		if err != nil {
			continue
		}
		sessions = append(sessions, buildSession(row))
	}

	return sessions, nil
}

func ResumeSession(sessionID int64) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	_, err = database.Exec(`
		UPDATE sessions
		SET state = ?, last_sync_at = ?, ended_at = '', resume_count = COALESCE(resume_count, 0) + 1
		WHERE id = ?
	`, string(SessionStateActive), now, sessionID)
	return err
}

func MarkStaleSessions(timeout time.Duration) (int, error) {
	overview, err := GetRecoveryOverview(timeout)
	if err != nil {
		return 0, err
	}

	marked := 0
	for _, candidate := range overview.Stale {
		if err := UpdateSessionState(int64(candidate.Session.ID), SessionStateCrashed); err != nil {
			return marked, err
		}
		marked++
	}

	return marked, nil
}

func RecoverSession(sessionID int) error {
	session, err := GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.State == SessionStateArchived {
		return fmt.Errorf("session %d is archived and cannot be recovered as a local crashed session", sessionID)
	}
	if session.State != SessionStateCrashed {
		return fmt.Errorf("session %d is %q, not crashed", sessionID, session.State)
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(`UPDATE sessions SET state = ? WHERE id = ?`, string(SessionStateCompleted), sessionID)
	if err != nil {
		return err
	}

	session, err = GetSession(sessionID)
	if err != nil {
		return nil
	}

	if info, err := os.Stat(session.Path); err == nil {
		UpdateSessionSize(int64(sessionID), info.Size())
	}

	return nil
}

func GetOrphanedSessions() ([]Session, error) {
	sessions, err := ListSessions()
	if err != nil {
		return nil, err
	}

	var orphaned []Session
	for _, s := range sessions {
		if _, err := os.Stat(s.Path); os.IsNotExist(err) {
			orphaned = append(orphaned, s)
		}
	}

	return orphaned, nil
}

func DeleteSession(sessionID int) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

func MarkSessionsArchived(sessionIDs []int, archivePath, manifestSHA256 string) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	database, err := db.GetDB()
	if err != nil {
		return err
	}

	tx, err := database.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		UPDATE sessions
		SET state = ?, archived_at = ?, archive_path = ?, archive_manifest_sha256 = ?, last_sync_at = ?, ended_at = ?, recorder_pid = 0
		WHERE id = ?
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	for _, sessionID := range sessionIDs {
		if _, err := stmt.Exec(string(SessionStateArchived), now, archivePath, manifestSHA256, now, now, sessionID); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// AddTag adds a tag to a session
func AddTag(sessionID int, tag string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		`INSERT OR IGNORE INTO session_tags (session_id, tag) VALUES (?, ?)`,
		sessionID, tag,
	)
	return err
}

// RemoveTag removes a tag from a session
func RemoveTag(sessionID int, tag string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec(
		`DELETE FROM session_tags WHERE session_id = ? AND tag = ?`,
		sessionID, tag,
	)
	return err
}

// GetSessionTags returns all tags for a session
func GetSessionTags(sessionID int) ([]string, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		`SELECT tag FROM session_tags WHERE session_id = ? ORDER BY tag`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// ListSessionsByTag returns all sessions with a specific tag
func ListSessionsByTag(tag string) ([]Session, error) {
	return ListSessionsByTagWithOptions(tag, SessionListOptions{})
}

func ListSessionsByTagWithOptions(tag string, opts SessionListOptions) ([]Session, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM sessions s
		INNER JOIN session_tags st ON s.id = st.session_id
		WHERE st.tag = ?
	`, sessionSelectColumns("s"))
	if filter := archiveFilterClause(opts); filter != "" {
		query += " AND " + strings.ReplaceAll(filter, "state", "s.state")
	}
	query += " ORDER BY s.timestamp DESC"

	rows, err := database.Query(query, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session

	for rows.Next() {
		row, err := scanSessionRow(rows)
		if err != nil {
			continue
		}
		sessions = append(sessions, buildSession(row))
	}

	return sessions, nil
}

func archiveFilterClause(opts SessionListOptions) string {
	switch {
	case opts.OnlyArchived:
		return "TRIM(COALESCE(state, '')) = 'archived'"
	case opts.IncludeArchived:
		return ""
	default:
		return "TRIM(COALESCE(state, '')) != 'archived'"
	}
}

func sessionSelectColumns(prefix string) string {
	if prefix != "" {
		prefix += "."
	}

	return strings.Join([]string{
		prefix + "id",
		prefix + "client",
		prefix + "engagement",
		prefix + "scope",
		prefix + "operator",
		prefix + "phase",
		prefix + "timestamp",
		prefix + "filename",
		prefix + "relative_path",
		prefix + "size",
		prefix + "state",
		prefix + "last_sync_at",
		prefix + "recorder_pid",
		prefix + "host_fingerprint",
		prefix + "hostname",
		prefix + "started_at",
		prefix + "ended_at",
		prefix + "resume_count",
		prefix + "target",
		prefix + "target_ip",
		prefix + "archived_at",
		prefix + "archive_path",
		prefix + "archive_manifest_sha256",
	}, ", ")
}

func scanSessionRow(scanner sessionScanner) (sessionRow, error) {
	var row sessionRow
	err := scanner.Scan(
		&row.ID,
		&row.Client,
		&row.Engagement,
		&row.Scope,
		&row.Operator,
		&row.Phase,
		&row.Timestamp,
		&row.Filename,
		&row.RelativePath,
		&row.Size,
		&row.State,
		&row.LastSyncAt,
		&row.RecorderPID,
		&row.HostFingerprint,
		&row.Hostname,
		&row.StartedAt,
		&row.EndedAt,
		&row.ResumeCount,
		&row.Target,
		&row.TargetIP,
		&row.ArchivedAt,
		&row.ArchivePath,
		&row.ArchiveManifestSHA256,
	)
	return row, err
}

func buildSession(row sessionRow) Session {
	mgr := config.Manager()
	path := filepath.Join(mgr.GetPaths().LogsDir, row.RelativePath)
	session := Session{
		ID:          row.ID,
		Filename:    row.Filename,
		Path:        path,
		DisplayPath: row.RelativePath,
		MetaPath:    strings.Replace(path, ".tty", ".json", 1),
		NotesPath:   strings.Replace(path, ".tty", ".notes.json", 1),
		Size:        row.Size,
		Metadata: SessionMetadata{
			Client:     row.Client,
			Engagement: row.Engagement,
			Scope:      row.Scope.String,
			Operator:   row.Operator.String,
			Phase:      row.Phase,
			Target:     row.Target.String,
			TargetIP:   row.TargetIP.String,
			Timestamp:  row.Timestamp,
		},
		State:                 normalizeSessionState(row.State),
		LastSyncAt:            row.LastSyncAt.String,
		RecorderPID:           int(row.RecorderPID.Int64),
		HostFingerprint:       row.HostFingerprint.String,
		Hostname:              row.Hostname.String,
		StartedAt:             row.StartedAt.String,
		EndedAt:               row.EndedAt.String,
		ResumeCount:           int(row.ResumeCount.Int64),
		ArchivedAt:            row.ArchivedAt.String,
		ArchivePath:           row.ArchivePath.String,
		ArchiveManifestSHA256: row.ArchiveManifestSHA256.String,
	}

	if ts, err := time.Parse(time.RFC3339, row.Timestamp); err == nil {
		session.ModTime = ts.Format("2006-01-02 15:04:05")
		session.SortKey = ts
	} else {
		session.ModTime = row.Timestamp
	}

	return session
}

func normalizeSessionState(state sql.NullString) SessionState {
	if state.Valid {
		if cleaned := strings.TrimSpace(state.String); cleaned != "" {
			return SessionState(cleaned)
		}
	}
	return SessionStateCompleted
}

// ListAllTags returns all tags used in the system
func ListAllTags() ([]string, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query(
		`SELECT DISTINCT tag FROM session_tags ORDER BY tag`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
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

	return os.WriteFile(notesPath, data, 0600)
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
