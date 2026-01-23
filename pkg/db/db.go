package db

import (
	"database/sql"
	"fmt"
	"os"
	"pentlog/pkg/config"

	_ "modernc.org/sqlite"
)

var dbInstance *sql.DB

func GetDB() (*sql.DB, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	if err := InitDB(); err != nil {
		return nil, err
	}

	return dbInstance, nil
}

func InitDB() error {
	mgr := config.Manager()
	dir := mgr.GetPaths().Home

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create pentlog dir: %w", err)
	}

	dbPath := mgr.GetPaths().DatabaseFile
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Ensure only owner can read/write the DB file
	if err := os.Chmod(dbPath, 0600); err != nil {
		return fmt.Errorf("failed to restrict database permissions: %w", err)
	}

	if err := createSchema(db); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	dbInstance = db
	return nil
}

func CloseDB() {
	if dbInstance != nil {
		dbInstance.Close()
		dbInstance = nil
	}
}
