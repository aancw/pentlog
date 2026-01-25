package db

import (
	"database/sql"
	"fmt"
)

func createSchema(db *sql.DB) error {
	baseSchema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client TEXT NOT NULL,
		engagement TEXT NOT NULL,
		scope TEXT,
		operator TEXT,
		phase TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		filename TEXT NOT NULL,
		relative_path TEXT NOT NULL,
		notes_path TEXT,
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

	_, err := db.Exec(baseSchema)
	if err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	if err := migrateSchema(db); err != nil {
		return fmt.Errorf("error migrating schema: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_state ON sessions(state)`)
	if err != nil {
		return fmt.Errorf("error creating state index: %w", err)
	}

	return nil
}

func migrateSchema(db *sql.DB) error {
	columns := []struct {
		name         string
		definition   string
		defaultValue string
		feature      string
	}{
		{"state", "TEXT", "completed", "crash recovery"},
		{"last_sync_at", "TEXT", "", "crash recovery"},
	}

	migratedFeatures := make(map[string]bool)

	for _, col := range columns {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = ?`, col.name).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			alterSQL := fmt.Sprintf("ALTER TABLE sessions ADD COLUMN %s %s", col.name, col.definition)
			if col.defaultValue != "" {
				alterSQL += fmt.Sprintf(" DEFAULT '%s'", col.defaultValue)
			}
			if _, err := db.Exec(alterSQL); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			migratedFeatures[col.feature] = true
		}
	}

	for feature := range migratedFeatures {
		fmt.Printf("📦 Database migrated: %s feature enabled\n", feature)
	}

	return nil
}
