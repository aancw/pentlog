package db

import (
	"database/sql"
	"fmt"
)

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client TEXT NOT NULL,
		engagement TEXT NOT NULL,
		scope TEXT,
		operator TEXT,
		phase TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		filename TEXT NOT NULL,
		relative_path TEXT NOT NULL, -- Path relative to logs dir, e.g. client/eng/phase/file.tty
		notes_path TEXT,            -- Path relative to logs dir, e.g. client/eng/phase/file.notes.json
		size INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_client ON sessions(client);
	CREATE INDEX IF NOT EXISTS idx_sessions_engagement ON sessions(engagement);
	CREATE INDEX IF NOT EXISTS idx_sessions_timestamp ON sessions(timestamp);

	CREATE TABLE IF NOT EXISTS schema_info (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	return nil
}
