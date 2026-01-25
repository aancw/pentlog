package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"testing"
	"time"
)

// TestListSessionsPaginatedMissingFile verifies that sessions with missing files are still listed
func TestListSessionsPaginatedMissingFile(t *testing.T) {
	// Reset config singleton for test isolation
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	// Setup temporary home directory
	tmpDir, err := os.MkdirTemp("", "pentlog-missing-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to force config to use tmpDir
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")
	defer db.CloseDB()

	// Create expected structure
	logsDir := filepath.Join(tmpDir, ".pentlog", "logs")
	sessionDir := filepath.Join(logsDir, "testclient", "testeng", "recon")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create a session file and metadata
	ts := time.Now()
	baseName := fmt.Sprintf("session-tester-%s", ts.Format("20060102-150405"))
	sessionPath := filepath.Join(sessionDir, baseName+".tty")

	if err := os.WriteFile(sessionPath, []byte("log content"), 0600); err != nil {
		t.Fatal(err)
	}

	meta := SessionMetadata{
		Client:     "testclient",
		Engagement: "testeng",
		Phase:      "recon",
		Timestamp:  ts.Format(time.RFC3339),
	}

	// Add session to DB manually
	if _, err := AddSessionToDB(meta, sessionPath); err != nil {
		t.Fatal(err)
	}

	// Delete the .tty file to simulate missing file
	if err := os.Remove(sessionPath); err != nil {
		t.Fatal(err)
	}

	// List sessions - should succeed and include the session
	sessions, err := ListSessions()

	// Verify no error occurred
	if err != nil {
		t.Fatalf("ListSessions should not fail with missing file: %v", err)
	}

	// Verify session is still listed
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session in results despite missing file, got %d", len(sessions))
	}

	// Verify session metadata is correct
	if len(sessions) > 0 {
		s := sessions[0]
		if s.Metadata.Client != "testclient" {
			t.Errorf("Expected client 'testclient', got %s", s.Metadata.Client)
		}
		if s.Path != sessionPath {
			t.Errorf("Expected path %s, got %s", sessionPath, s.Path)
		}
	}

	// Verify GetOrphanedSessions detects the missing file
	orphaned, err := GetOrphanedSessions()
	if err != nil {
		t.Fatalf("GetOrphanedSessions failed: %v", err)
	}

	if len(orphaned) != 1 {
		t.Errorf("Expected 1 orphaned session, got %d", len(orphaned))
	}

	if len(orphaned) > 0 && orphaned[0].Path != sessionPath {
		t.Errorf("Expected orphaned session path %s, got %s", sessionPath, orphaned[0].Path)
	}
}
