package db

import (
	"os"
	"path/filepath"
	"testing"
	"pentlog/pkg/config"
)

func TestBackupDB(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Reset db instance for test
	dbInstance = nil

	mgr := config.Manager()
	paths := mgr.GetPaths()

	// Create test database
	if err := os.MkdirAll(paths.Home, 0700); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Initialize database
	if err := InitDB(); err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer CloseDB()

	// Create a backup
	backupPath, err := BackupDB()
	if err != nil {
		t.Errorf("BackupDB() failed: %v", err)
		return
	}

	if backupPath == "" {
		t.Errorf("BackupDB() returned empty path")
		return
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("Backup file not found: %v", err)
		return
	}

	// Verify backup file has correct permissions (0600)
	fileInfo, _ := os.Stat(backupPath)
	perms := fileInfo.Mode().Perm()
	if perms != 0600 {
		t.Errorf("Backup file permissions incorrect: got %o, want 0600", perms)
	}

	// Verify backup filename format
	expectedPrefix := paths.DatabaseFile + ".backup-"
	if !filepath.HasPrefix(backupPath, expectedPrefix) {
		t.Errorf("Backup filename incorrect: got %s, expected to start with %s", backupPath, expectedPrefix)
	}

	// Cleanup
	os.Remove(backupPath)
}

func TestBackupDBNonexistent(t *testing.T) {
	// Setup test environment with no DB
	tmpDir := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Reset db instance
	dbInstance = nil

	// Call backup without initializing DB
	backupPath, err := BackupDB()
	if err != nil {
		t.Errorf("BackupDB() failed for non-existent DB: %v", err)
	}

	// Should return empty string for non-existent DB
	if backupPath != "" {
		t.Errorf("BackupDB() should return empty string for non-existent DB, got: %s", backupPath)
	}
}
