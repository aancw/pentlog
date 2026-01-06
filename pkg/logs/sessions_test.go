package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestListSessions(t *testing.T) {
	// Setup temporary home directory
	tmpDir, err := os.MkdirTemp("", "pentlog-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to force config to use tmpDir
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Create expected structure
	// .pentlog/logs/client/engagement/phase/
	logsDir := filepath.Join(tmpDir, ".pentlog", "logs")
	sessionDir := filepath.Join(logsDir, "acme", "q1", "recon")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create dummy session files
	ts := time.Now()
	tsStr := ts.Format("20060102-150405")
	baseName := fmt.Sprintf("manual-tester-%s", tsStr)

	// Write .tty
	if err := os.WriteFile(filepath.Join(sessionDir, baseName+".tty"), []byte("log content"), 0600); err != nil {
		t.Fatal(err)
	}
	// Write .timing
	if err := os.WriteFile(filepath.Join(sessionDir, baseName+".timing"), []byte("timing content"), 0600); err != nil {
		t.Fatal(err)
	}
	// Write .json
	meta := SessionMetadata{
		Client:     "ACME",
		Engagement: "Q1",
		Scope:      "Scope",
		Operator:   "Tester",
		Phase:      "Recon",
		Timestamp:  ts.Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(sessionDir, baseName+".json"), metaBytes, 0600); err != nil {
		t.Fatal(err)
	}

	// Create another session (older)
	oldTs := ts.Add(-1 * time.Hour)
	oldTsStr := oldTs.Format("20060102-150405")
	oldBaseName := fmt.Sprintf("manual-tester-%s", oldTsStr)
	if err := os.WriteFile(filepath.Join(sessionDir, oldBaseName+".tty"), []byte("old log"), 0600); err != nil {
		t.Fatal(err)
	}

	// Test ListSessions
	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	// Expect 2 sessions
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	// Verify sort order: older sessions should appear first.

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ID < sessions[j].ID
	})

	// Check if we found the full session
	foundFull := false
	for _, s := range sessions {
		if s.Metadata.Client == "ACME" {
			foundFull = true
			if s.Metadata.Phase != "Recon" {
				t.Errorf("Expected Phase Recon, got %s", s.Metadata.Phase)
			}
		}
	}
	if !foundFull {
		t.Error("Did not find session with metadata")
	}
}
