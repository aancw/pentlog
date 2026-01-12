package logs

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveSessions(t *testing.T) {
	// Setup tmp env
	tmpDir, err := os.MkdirTemp("", "pentlog-archive-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

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
	count, err := ArchiveSessions("testclient", "", "", 0, false)
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

	// Verify Originals Still Exist
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Original log file should still exist")
	}
	if _, err := os.Stat(timingFile); os.IsNotExist(err) {
		t.Error("Original timing file should still exist")
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

	// Archive ONLY the new one if possible... but currently filtering is by client/time.
	// The first session is still there. So archive will pick up BOTH sessions now if we run again.
	// That's fine.

	time.Sleep(1 * time.Second) // Ensure different timestamp for archive file

	count, err = ArchiveSessions("testclient", "", "", 0, true)
	if err != nil {
		t.Fatalf("Archive delete failed: %v", err)
	}
	// Should pick up both files now (original + new one)
	if count != 2 {
		t.Errorf("Expected 2 archived sessions, got %d", count)
	}

	// Verify Originals GONE
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Error("Original log file should be deleted")
	}
	if _, err := os.Stat(timingFile); !os.IsNotExist(err) {
		t.Error("Original timing file should be deleted")
	}
	if _, err := os.Stat(logFile2); !os.IsNotExist(err) {
		t.Error("Original log file 2 should be deleted")
	}
	if _, err := os.Stat(timingFile2); !os.IsNotExist(err) {
		t.Error("Original timing file 2 should be deleted")
	}

	// Verify Archive Content (check inside the latest tar.gz)
	archives, _ = ListArchives()
	if len(archives) != 2 {
		t.Errorf("Expected 2 archive files now, got %d", len(archives))
		// Dump fs
		filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			t.Logf("FS: %s", path)
			return nil
		})
		return
	}

	// Get the last one
	latest := archives[len(archives)-1]

	f, err := os.Open(latest.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	foundTiming := false
	foundLog := false
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(header.Name, ".tty") {
			foundLog = true
		}
		if strings.Contains(header.Name, ".timing") {
			foundTiming = true
		}
	}

	if !foundLog {
		t.Error("Archive should contain tty log")
	}
	if foundTiming {
		t.Error("Archive should NOT contain timing file")
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
	count, err := ArchiveSessions("filterclient", "", "recon", 0, true)
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
}
