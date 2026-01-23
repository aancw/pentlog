package logs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"strings"
	"testing"
	"time"
)

// TestListSessionsPaginatedMissingFile verifies that missing session files generate warnings
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
	if err := AddSessionToDB(meta, sessionPath); err != nil {
		t.Fatal(err)
	}

	// Delete the .tty file to simulate missing file
	if err := os.Remove(sessionPath); err != nil {
		t.Fatal(err)
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// List sessions - should succeed but warn
	sessions, err := ListSessions()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Verify no error occurred
	if err != nil {
		t.Fatalf("ListSessions should not fail with missing file: %v", err)
	}

	// Verify session is still listed
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session in results despite missing file, got %d", len(sessions))
	}

	// Verify warning was printed to stderr
	if !strings.Contains(stderrOutput, "WARNING: Session") {
		t.Errorf("Expected WARNING in stderr, got: %s", stderrOutput)
	}

	if !strings.Contains(stderrOutput, sessionPath) {
		t.Errorf("Expected warning to mention file path %s, got: %s", sessionPath, stderrOutput)
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
}
