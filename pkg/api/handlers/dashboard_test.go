package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pentlog/pkg/config"
	"pentlog/pkg/db"
	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func TestDashboardOverviewAggregatesStatsContextAndArtifacts(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()
	defer db.CloseDB()

	tmpDir := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := config.Manager()
	if err := mgr.EnsureDirectories(); err != nil {
		t.Fatalf("ensure directories: %v", err)
	}

	ctx := &config.ContextData{
		Client:     "Acme",
		Engagement: "Q2",
		Phase:      "recon",
		Operator:   "tester",
		Timestamp:  time.Now().Format(time.RFC3339),
		Type:       "Client",
	}
	if err := mgr.SaveContext(ctx); err != nil {
		t.Fatalf("save context: %v", err)
	}

	sessionPath := filepath.Join(mgr.GetPaths().LogsDir, "acme", "q2", "recon", "manual-tester-20260422-010101.tty")
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		t.Fatalf("mkdir logs path: %v", err)
	}
	if err := os.WriteFile(sessionPath, []byte("tty"), 0600); err != nil {
		t.Fatalf("write tty file: %v", err)
	}

	_, err := logs.AddSessionToDBWithState(logs.SessionMetadata{
		Client:     "Acme",
		Engagement: "Q2",
		Phase:      "recon",
		Target:     "dc01",
		TargetIP:   "10.10.10.10",
		Timestamp:  time.Now().Format(time.RFC3339),
	}, sessionPath, logs.SessionStateActive)
	if err != nil {
		t.Fatalf("add session: %v", err)
	}

	reportDir := filepath.Join(mgr.GetPaths().ReportsDir, "Acme")
	if err := os.MkdirAll(reportDir, 0700); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reportDir, "q2-recon.html"), []byte("<html></html>"), 0600); err != nil {
		t.Fatalf("write report: %v", err)
	}

	archiveDir := filepath.Join(mgr.GetPaths().ArchiveDir, "Acme")
	if err := os.MkdirAll(archiveDir, 0700); err != nil {
		t.Fatalf("mkdir archive dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveDir, "q2-recon.zip"), []byte("zip"), 0600); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	router := chi.NewRouter()
	router.Mount("/dashboard", DashboardRoutes())

	req := httptest.NewRequest(http.MethodGet, "/dashboard/overview", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Stats struct {
			TotalSessions int            `json:"total_sessions"`
			StateCounts   map[string]int `json:"state_counts"`
		} `json:"stats"`
		Context struct {
			HasContext bool                   `json:"has_context"`
			Context    map[string]interface{} `json:"context"`
		} `json:"context"`
		Artifacts struct {
			ReportsTotal  int                    `json:"reports_total"`
			ArchivesTotal int                    `json:"archives_total"`
			LatestReport  map[string]interface{} `json:"latest_report"`
			LatestArchive map[string]interface{} `json:"latest_archive"`
		} `json:"artifacts"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Stats.TotalSessions != 1 {
		t.Fatalf("expected total sessions 1, got %d", payload.Stats.TotalSessions)
	}
	if payload.Stats.StateCounts["active"] != 1 {
		t.Fatalf("expected active state count 1, got %+v", payload.Stats.StateCounts)
	}
	if !payload.Context.HasContext || payload.Context.Context == nil {
		t.Fatalf("expected active context in overview")
	}
	if payload.Artifacts.ReportsTotal != 1 || payload.Artifacts.ArchivesTotal != 1 {
		t.Fatalf("expected one report and one archive, got reports=%d archives=%d", payload.Artifacts.ReportsTotal, payload.Artifacts.ArchivesTotal)
	}
	if payload.Artifacts.LatestReport == nil || payload.Artifacts.LatestArchive == nil {
		t.Fatalf("expected latest report and archive metadata")
	}
}
