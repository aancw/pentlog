package logs

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"
)

func TestGenerateReportCuratedByDefault(t *testing.T) {
	testSession := setupExportFixture(t)

	report, err := GenerateReport([]Session{testSession}, "Acme")
	if err != nil {
		t.Fatalf("GenerateReport() failed: %v", err)
	}

	requiredSections := []string{
		"## Executive Summary",
		"## Engagement Metadata",
		"## Findings",
		"## Evidence Snippets",
		"## Command Appendix",
		"## Integrity and Archive References",
	}
	for _, section := range requiredSections {
		if !strings.Contains(report, section) {
			t.Fatalf("expected report to contain %q", section)
		}
	}

	if strings.Contains(report, "## Full Transcript Appendix") {
		t.Fatalf("default report should not include full transcript appendix")
	}

	if !strings.Contains(report, "External admin panel exposed") {
		t.Fatalf("expected finding content in report")
	}
	if !strings.Contains(report, "Confirmed admin portal on 10.10.10.10") {
		t.Fatalf("expected note content in report")
	}
	if !strings.Contains(report, "`nmap -sn 10.10.10.10`") {
		t.Fatalf("expected command appendix entry in report")
	}
}

func TestGenerateReportIncludesTranscriptAppendixWhenRequested(t *testing.T) {
	testSession := setupExportFixture(t)

	report, err := GenerateReportWithOptions([]Session{testSession}, "Acme", nil, "", ReportOptions{
		AppendixMode: ReportAppendixModeFullTranscript,
	})
	if err != nil {
		t.Fatalf("GenerateReportWithOptions() failed: %v", err)
	}

	if !strings.Contains(report, "## Full Transcript Appendix") {
		t.Fatalf("expected transcript appendix heading")
	}
	if !strings.Contains(report, "scan report for 10.10.10.10") {
		t.Fatalf("expected transcript content in appendix")
	}
}

func TestGenerateHTMLReportRendersCuratedSections(t *testing.T) {
	testSession := setupExportFixture(t)
	copyTemplateFixtures(t)

	report, err := GenerateHTMLReportWithOptions([]Session{testSession}, "Acme", nil, "", nil, ReportOptions{
		AppendixMode: ReportAppendixModeCommands,
	})
	if err != nil {
		t.Fatalf("GenerateHTMLReportWithOptions() failed: %v", err)
	}

	requiredContent := []string{
		"Scope Snapshot",
		"Operational Scope",
		"Recorded Issues",
		"Curated Excerpts",
		"Command Trail",
		"Evidence Traceability",
	}
	for _, item := range requiredContent {
		if !strings.Contains(report, item) {
			t.Fatalf("expected html report to contain %q", item)
		}
	}

	if strings.Contains(report, "Explicit Transcript Mode") {
		t.Fatalf("default html report should not include transcript appendix section")
	}
}

func setupExportFixture(t *testing.T) Session {
	t.Helper()

	config.ResetManagerForTesting()
	db.CloseDB()
	t.Cleanup(func() {
		db.CloseDB()
		config.ResetManagerForTesting()
		os.Unsetenv("PENTLOG_TEST_HOME")
	})

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)

	mgr := config.Manager()
	paths := mgr.GetPaths()
	if err := os.MkdirAll(paths.LogsDir, 0700); err != nil {
		t.Fatalf("failed to create logs dir: %v", err)
	}
	if err := os.MkdirAll(paths.TemplatesDir, 0700); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	database, err := db.GetDB()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	relativePath := filepath.Join("acme", "internal", "recon", "session-1.tty")
	logPath := filepath.Join(paths.LogsDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	transcript := strings.Join([]string{
		"user@host:~$ nmap -sn 10.10.10.10",
		"scan report for 10.10.10.10",
		"user@host:~$ curl -I https://10.10.10.10/admin",
		"HTTP/1.1 200 OK",
	}, "\n") + "\n"
	if err := writeTtyrecFixture(logPath, transcript); err != nil {
		t.Fatalf("failed to write ttyrec fixture: %v", err)
	}

	result, err := database.Exec(`
		INSERT INTO sessions (
			client, engagement, scope, operator, phase, timestamp, filename, relative_path, size, target, target_ip
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "Acme", "internal", "corp", "pentester", "recon", "2026-06-06T10:15:00Z", "session-1.tty", relativePath, len(transcript), "web01", "10.10.10.10")
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get session id: %v", err)
	}

	if err := AddTag(int(sessionID), "initial-access"); err != nil {
		t.Fatalf("failed to add tag: %v", err)
	}
	if err := AddTag(int(sessionID), "web"); err != nil {
		t.Fatalf("failed to add tag: %v", err)
	}

	noteOffset := int64(fileSize(t, logPath))
	if err := AppendNote(strings.TrimSuffix(logPath, ".tty")+".notes.json", SessionNote{
		Timestamp:  "2026-06-06 10:18:00",
		Content:    "Confirmed admin portal on 10.10.10.10",
		ByteOffset: noteOffset,
	}); err != nil {
		t.Fatalf("failed to append note: %v", err)
	}

	vulnManager := vulns.NewManager("Acme", "internal")
	if err := vulnManager.Save(vulns.Vuln{
		ID:          "ACME-001",
		Title:       "External admin panel exposed",
		Severity:    vulns.SeverityHigh,
		Status:      vulns.StatusOpen,
		Phase:       "recon",
		Description: "Administrative interface responded without prior network restriction.",
		Remediation: "Restrict administrative access and require trusted-network controls.",
		Evidence:    []string{"`curl -I https://10.10.10.10/admin` returned `HTTP/1.1 200 OK`."},
		References:  []string{"Internal validation note"},
		CreatedAt:   time.Date(2026, 6, 6, 10, 20, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 6, 10, 20, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("failed to save vuln: %v", err)
	}

	session, err := GetSession(int(sessionID))
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}
	return *session
}

func copyTemplateFixtures(t *testing.T) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve current file path")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))
	templateDir := filepath.Join(root, "assets", "templates")

	mgr := config.Manager()
	for _, name := range []string{"report.html", "report.css"} {
		data, err := os.ReadFile(filepath.Join(templateDir, name))
		if err != nil {
			t.Fatalf("failed to read template fixture %s: %v", name, err)
		}
		if err := utils.WritePrivateFile(filepath.Join(mgr.GetPaths().TemplatesDir, name), data); err != nil {
			t.Fatalf("failed to copy template fixture %s: %v", name, err)
		}
	}
}

func writeTtyrecFixture(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	payload := []byte(content)
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:4], uint32(time.Now().Unix()))
	binary.LittleEndian.PutUint32(header[4:8], 0)
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(payload)))
	if _, err := file.Write(header); err != nil {
		return err
	}
	_, err = file.Write(payload)
	return err
}

func fileSize(t *testing.T, path string) int {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	return int(info.Size())
}
