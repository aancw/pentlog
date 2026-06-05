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

func TestSearchRouteReturnsPaginationAndContext(t *testing.T) {
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

	sessionPath := filepath.Join(mgr.GetPaths().LogsDir, "acme", "q2", "recon", "evidence.log")
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		t.Fatalf("mkdir logs path: %v", err)
	}
	if err := os.WriteFile(sessionPath, []byte("alpha one\nalpha two\nbravo\nalpha three\n"), 0600); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	_, err := logs.AddSessionToDBWithState(logs.SessionMetadata{
		Client:     "Acme",
		Engagement: "Q2",
		Phase:      "recon",
		Timestamp:  time.Date(2026, 4, 22, 15, 4, 5, 0, time.UTC).Format(time.RFC3339),
	}, sessionPath, logs.SessionStateCompleted)
	if err != nil {
		t.Fatalf("add session: %v", err)
	}

	router := chi.NewRouter()
	router.Mount("/search", SearchRoutes())

	req := httptest.NewRequest(http.MethodGet, "/search?q=alpha&limit=1&offset=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		TotalMatches int  `json:"total_matches"`
		Limit        int  `json:"limit"`
		Offset       int  `json:"offset"`
		HasMore      bool `json:"has_more"`
		Results      []struct {
			LineNum          int      `json:"line_num"`
			Context          []string `json:"context"`
			ContextStartLine int      `json:"context_start_line"`
			Content          string   `json:"content"`
			IsNote           bool     `json:"is_note"`
		} `json:"results"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.TotalMatches != 3 {
		t.Fatalf("expected 3 total matches, got %d", payload.TotalMatches)
	}
	if payload.Limit != 1 || payload.Offset != 1 {
		t.Fatalf("expected limit=1 offset=1, got limit=%d offset=%d", payload.Limit, payload.Offset)
	}
	if !payload.HasMore {
		t.Fatalf("expected has_more=true")
	}
	if len(payload.Results) != 1 {
		t.Fatalf("expected 1 paged result, got %d", len(payload.Results))
	}
	if payload.Results[0].LineNum != 2 {
		t.Fatalf("expected second match on line 2, got %d", payload.Results[0].LineNum)
	}
	if payload.Results[0].Content != "alpha two" {
		t.Fatalf("expected matched content alpha two, got %q", payload.Results[0].Content)
	}
	if payload.Results[0].IsNote {
		t.Fatalf("expected content match, not note match")
	}
	if payload.Results[0].ContextStartLine != 1 {
		t.Fatalf("expected context to begin at line 1, got %d", payload.Results[0].ContextStartLine)
	}
	if len(payload.Results[0].Context) != 4 {
		t.Fatalf("expected 4 context lines, got %d", len(payload.Results[0].Context))
	}
}
