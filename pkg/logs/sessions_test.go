package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"sort"
	"testing"
	"time"
)

func TestListSessions(t *testing.T) {
	// Reset config singleton for test isolation
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	// Setup temporary home directory
	tmpDir, err := os.MkdirTemp("", "pentlog-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to force config to use tmpDir
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")
	defer db.CloseDB()

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
	baseName := fmt.Sprintf("session-tester-%s", tsStr)

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
	oldBaseName := fmt.Sprintf("session-tester-%s", oldTsStr)
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

func TestAppendNotePermissions(t *testing.T) {
	notePath := filepath.Join(t.TempDir(), "notes", "session.notes.json")

	if err := AppendNote(notePath, SessionNote{
		Timestamp:  "10:00:00",
		Content:    "check permissions",
		ByteOffset: 12,
	}); err != nil {
		t.Fatalf("AppendNote failed: %v", err)
	}

	info, err := os.Stat(notePath)
	if err != nil {
		t.Fatalf("Failed to stat note file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected note permissions 0600, got %#o", got)
	}
}

func TestListAndGetSessionHydratesStateAndTarget(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()
	defer db.CloseDB()

	tmpDir := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := config.Manager()
	logsDir := mgr.GetPaths().LogsDir
	sessionDir := filepath.Join(logsDir, "acme", "q2", "exploit")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}

	ttyPath := filepath.Join(sessionDir, "manual-tester-20260422-120000.tty")
	if err := os.WriteFile(ttyPath, []byte("session"), 0600); err != nil {
		t.Fatalf("write tty file: %v", err)
	}

	meta := SessionMetadata{
		Client:     "ACME",
		Engagement: "Q2",
		Scope:      "Internal",
		Operator:   "tester",
		Phase:      "exploit",
		Target:     "dc01",
		TargetIP:   "10.10.10.10",
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	id, err := AddSessionToDBWithState(meta, ttyPath, SessionStatePaused)
	if err != nil {
		t.Fatalf("add session to db: %v", err)
	}

	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	got := sessions[0]
	if got.State != SessionStatePaused {
		t.Fatalf("expected state %q, got %q", SessionStatePaused, got.State)
	}
	if got.Metadata.Target != meta.Target {
		t.Fatalf("expected target %q, got %q", meta.Target, got.Metadata.Target)
	}
	if got.Metadata.TargetIP != meta.TargetIP {
		t.Fatalf("expected target_ip %q, got %q", meta.TargetIP, got.Metadata.TargetIP)
	}

	full, err := GetSession(int(id))
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if full.State != SessionStatePaused {
		t.Fatalf("expected state %q from GetSession, got %q", SessionStatePaused, full.State)
	}
	if full.Metadata.Target != meta.Target || full.Metadata.TargetIP != meta.TargetIP {
		t.Fatalf("expected target fields to round-trip, got target=%q target_ip=%q", full.Metadata.Target, full.Metadata.TargetIP)
	}
}
