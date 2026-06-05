package logs

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"strings"
	"testing"
	"time"
)

func TestArchiveSessions(t *testing.T) {
	// Reset config singleton for test isolation
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	// Setup tmp env
	tmpDir, err := os.MkdirTemp("", "pentlog-archive-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")
	defer db.CloseDB()

	// Setup data
	logsDir := filepath.Join(tmpDir, ".pentlog", "logs")
	clientDir := filepath.Join(logsDir, "testclient", "eng", "recon")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a session file
	baseName := "testsession"
	logFile := filepath.Join(clientDir, baseName+".tty")
	os.WriteFile(logFile, []byte("log data"), 0644)

	// Create metadata file (REQUIRED for ListSessions to identify client)
	meta := SessionMetadata{
		Client:    "testclient",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(clientDir, baseName+".json"), metaBytes, 0644)

	// Create timing file (should be ignored/deleted)
	timingFile := filepath.Join(clientDir, baseName+".timing")
	os.WriteFile(timingFile, []byte("timing data"), 0644)

	// Test 1: Archive with Keep (default)
	if err := SyncSessions(); err != nil {
		t.Fatalf("SyncSessions failed: %v", err)
	}

	count, err := ArchiveSessions("testclient", "", "", 0, false, "")
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 archived session, got %d", count)
	}

	// Verify archive exists
	archives, err := ListArchives()
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Errorf("Expected 1 archive file, got %d", len(archives))
	}
	if !strings.HasSuffix(archives[0].Filename, ".zip") {
		t.Errorf("Expected .zip archive, got %s", archives[0].Filename)
	}

	// Verify Originals Still Exist
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Original log file should still exist")
	}
	if _, err := os.Stat(timingFile); os.IsNotExist(err) {
		t.Error("Original timing file should still exist")
	}

	activeSessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(activeSessions) != 0 {
		t.Fatalf("expected archived sessions to be excluded from default listings, got %d", len(activeSessions))
	}

	allSessions, err := ListSessionsWithOptions(SessionListOptions{IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListSessionsWithOptions failed: %v", err)
	}
	if len(allSessions) != 1 {
		t.Fatalf("expected 1 session including archived, got %d", len(allSessions))
	}
	if allSessions[0].State != SessionStateArchived {
		t.Fatalf("expected session state %q, got %q", SessionStateArchived, allSessions[0].State)
	}
	if allSessions[0].ArchivePath == "" {
		t.Fatal("expected archive_path to be recorded")
	}
	if allSessions[0].ArchiveManifestSHA256 == "" {
		t.Fatal("expected archive_manifest_sha256 to be recorded")
	}
	if allSessions[0].ArchivedAt == "" {
		t.Fatal("expected archived_at to be recorded")
	}

	// Test 2: Archive with Delete
	// Wait a bit to ensure timestamp diff maybe? Or just run it.
	// Since we already have an archive, it will create a NEW timestamped archive.
	// But let's delete the old archive first to keep it simple, or just handle getting count 2 total.

	// Let's create a NEW session for the delete test
	baseName2 := "testsession2"
	logFile2 := filepath.Join(clientDir, baseName2+".tty")
	os.WriteFile(logFile2, []byte("log data 2"), 0644)

	// Create metadata for session 2
	os.WriteFile(filepath.Join(clientDir, baseName2+".json"), metaBytes, 0644)

	timingFile2 := filepath.Join(clientDir, baseName2+".timing")
	os.WriteFile(timingFile2, []byte("timing data 2"), 0644)

	// Force sync because we added files manually and ListSessions cache might be active
	if err := SyncSessions(); err != nil {
		t.Fatalf("SyncSessions failed: %v", err)
	}

	// Archive ONLY the new one if possible... but currently filtering is by client/time.
	// The first session is still there. So archive will pick up BOTH sessions now if we run again.
	// That's fine.

	time.Sleep(1 * time.Second) // Ensure different timestamp for archive file

	count, err = ArchiveSessions("testclient", "", "", 0, true, "")
	if err != nil {
		t.Fatalf("Archive delete failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 newly archived session, got %d", count)
	}

	// Verify Originals GONE
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("First archived log file should remain because it was archived with keep mode")
	}
	if _, err := os.Stat(timingFile); os.IsNotExist(err) {
		t.Error("First archived timing file should remain because it was archived with keep mode")
	}
	if _, err := os.Stat(logFile2); !os.IsNotExist(err) {
		t.Error("Original log file 2 should be deleted")
	}
	if _, err := os.Stat(timingFile2); !os.IsNotExist(err) {
		t.Error("Original timing file 2 should be deleted")
	}

	// Verify Archive Content (check inside the latest zip)
	archives, _ = ListArchives()
	if len(archives) != 2 {
		t.Errorf("Expected 2 archive files now, got %d", len(archives))
		return
	}

	// Get the last one
	latest := archives[len(archives)-1]

	zr, err := zip.OpenReader(latest.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	foundTiming := false
	foundLog := false
	foundManifest := false
	for _, f := range zr.File {
		if strings.Contains(f.Name, ".tty") {
			foundLog = true
		}
		if strings.Contains(f.Name, ".timing") {
			foundTiming = true
		}
		if f.Name == archiveManifestName {
			foundManifest = true
		}
	}

	if !foundLog {
		t.Error("Archive should contain tty log")
	}
	if foundTiming {
		t.Error("Archive should NOT contain timing file")
	}
	if !foundManifest {
		t.Error("Archive should contain manifest.json")
	}

	orphaned, err := GetOrphanedSessions()
	if err != nil {
		t.Fatalf("GetOrphanedSessions failed: %v", err)
	}
	if len(orphaned) != 0 {
		t.Fatalf("expected archived sessions deleted from disk to stay out of orphan cleanup, got %d", len(orphaned))
	}
}

func TestArchiveSessionFiltering(t *testing.T) {
	// Setup tmp env
	tmpDir, err := os.MkdirTemp("", "pentlog-archive-filter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")
	defer db.CloseDB()

	clientDir := filepath.Join(tmpDir, ".pentlog", "logs", "filterclient", "eng", "recon")
	os.MkdirAll(clientDir, 0755)

	// Create Session 1: Recon
	base1 := "recon-sess"
	meta1 := SessionMetadata{Client: "filterclient", Engagement: "eng", Phase: "recon", Timestamp: time.Now().Format(time.RFC3339)}
	m1, _ := json.Marshal(meta1)
	os.WriteFile(filepath.Join(clientDir, base1+".json"), m1, 0644)
	os.WriteFile(filepath.Join(clientDir, base1+".tty"), []byte("log"), 0644)

	// Create Session 2: Exploit
	clientDir2 := filepath.Join(tmpDir, ".pentlog", "logs", "filterclient", "eng", "exploit")
	os.MkdirAll(clientDir2, 0755)
	base2 := "exploit-sess"
	meta2 := SessionMetadata{Client: "filterclient", Engagement: "eng", Phase: "exploit", Timestamp: time.Now().Format(time.RFC3339)}
	m2, _ := json.Marshal(meta2)
	os.WriteFile(filepath.Join(clientDir2, base2+".json"), m2, 0644)
	os.WriteFile(filepath.Join(clientDir2, base2+".tty"), []byte("log"), 0644)

	// Archive ONLY Recon
	if err := SyncSessions(); err != nil {
		t.Fatalf("SyncSessions failed: %v", err)
	}

	count, err := ArchiveSessions("filterclient", "", "recon", 0, true, "")
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 recon session archived, got %d", count)
	}

	// Verify Recon gone, Exploit remains
	if _, err := os.Stat(filepath.Join(clientDir, base1+".tty")); !os.IsNotExist(err) {
		t.Error("Recon session should be deleted")
	}
	if _, err := os.Stat(filepath.Join(clientDir2, base2+".tty")); os.IsNotExist(err) {
		t.Error("Exploit session should still exist")
	}

	activeSessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(activeSessions) != 1 {
		t.Fatalf("expected only the non-archived session to remain in default listings, got %d", len(activeSessions))
	}
	if activeSessions[0].Metadata.Phase != "exploit" {
		t.Fatalf("expected remaining active session to be exploit, got %s", activeSessions[0].Metadata.Phase)
	}
}

func TestArchiveSessionsEncrypted(t *testing.T) {
	// Reset config singleton for test isolation
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	// Setup tmp env
	tmpDir, err := os.MkdirTemp("", "pentlog-archive-enc-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")
	defer db.CloseDB()

	// Setup data
	logsDir := filepath.Join(tmpDir, ".pentlog", "logs")
	clientDir := filepath.Join(logsDir, "encclient", "eng", "recon")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatal(err)
	}

	baseName := "testsession"
	logFile := filepath.Join(clientDir, baseName+".tty")
	os.WriteFile(logFile, []byte("secret log data"), 0644)

	meta := SessionMetadata{
		Client:    "encclient",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(clientDir, baseName+".json"), metaBytes, 0644)

	// Archive with PASSWORD
	if err := SyncSessions(); err != nil {
		t.Fatalf("SyncSessions failed: %v", err)
	}

	password := "supersecret"
	count, err := ArchiveSessions("encclient", "", "", 0, false, password)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 archived session, got %d", count)
	}

	archives, err := ListArchives()
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Errorf("Expected 1 archive file, got %d", len(archives))
	}

	archive := archives[0]
	if !strings.HasSuffix(archive.Filename, ".zip") {
		t.Errorf("Expected .zip archive, got %s", archive.Filename)
	}

	sessions, err := ListSessionsWithOptions(SessionListOptions{IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListSessionsWithOptions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 archived session in database, got %d", len(sessions))
	}
	if sessions[0].State != SessionStateArchived {
		t.Fatalf("expected archived state after encryption flow, got %q", sessions[0].State)
	}

	// We can't easily test decryption without the same library logic or unzipping command
	// But simply verifying it is a zip and was created successfully is a good init check.
	// Also could try to open it with zip reader.

	// Optional: verify zip structure if needed, or just rely on manual verification for decryption
}
