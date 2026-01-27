package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"pentlog/pkg/config"
	"time"

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

func BackupDB() (string, error) {
	mgr := config.Manager()
	dbPath := mgr.GetPaths().DatabaseFile

	if _, err := os.Stat(dbPath); err != nil {
		return "", nil
	}

	backupPath := dbPath + ".backup-" + time.Now().Format("20060102-150405")

	src, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open database for backup: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to backup database: %w", err)
	}

	if err := os.Chmod(backupPath, 0600); err != nil {
		return "", fmt.Errorf("failed to set backup permissions: %w", err)
	}

	return backupPath, nil
}
