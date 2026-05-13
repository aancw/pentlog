package logs

import (
	stdzip "archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"strings"
	"testing"
	"time"
)

func TestImportRejectsManifestHashMismatch(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()
	defer db.CloseDB()

	tmpDir, err := os.MkdirTemp("", "pentlog-import-manifest-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0700); err != nil {
		t.Fatal(err)
	}

	ttySource := filepath.Join(sourceDir, "session.tty")
	metaSource := filepath.Join(sourceDir, "session.json")

	if err := os.WriteFile(ttySource, []byte("original log data"), 0600); err != nil {
		t.Fatal(err)
	}

	meta := SessionMetadata{
		Client:     "client-a",
		Engagement: "eng-1",
		Phase:      "recon",
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metaSource, metaBytes, 0600); err != nil {
		t.Fatal(err)
	}

	ttyHash, ttySize, err := hashFileSHA256(ttySource)
	if err != nil {
		t.Fatal(err)
	}
	metaHash, metaSize, err := hashFileSHA256(metaSource)
	if err != nil {
		t.Fatal(err)
	}

	manifestData, err := manifestJSON(buildArchiveManifest("client-a", false, false, []ArchiveManifestFile{
		{
			ArchivePath: "logs/client-a/eng-1/recon/session.tty",
			Role:        "session_log",
			Size:        ttySize,
			SHA256:      ttyHash,
		},
		{
			ArchivePath: "logs/client-a/eng-1/recon/session.json",
			Role:        "session_metadata",
			Size:        metaSize,
			SHA256:      metaHash,
		},
	}))
	if err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(tmpDir, "tampered.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := stdzip.NewWriter(archiveFile)
	writeZipEntry := func(name string, data []byte) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}

	writeZipEntry("logs/client-a/eng-1/recon/session.tty", []byte("tampered log data"))
	writeZipEntry("logs/client-a/eng-1/recon/session.json", metaBytes)
	writeZipEntry(archiveManifestName, manifestData)

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = ImportFromPentlogArchive(archivePath, ImportOptions{})
	if err == nil {
		t.Fatal("expected manifest verification failure")
	}
	if !strings.Contains(err.Error(), "manifest verification failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
