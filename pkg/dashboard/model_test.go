package dashboard

import (
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/vulns"
	"strings"
	"testing"
	"time"
)

func TestBuildDashboardDataUsesCurrentContext(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	testHome := t.TempDir()
	if err := os.Setenv("PENTLOG_TEST_HOME", testHome); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	ctx := &config.ContextData{
		Client:     "ACME",
		Engagement: "Q1",
		Phase:      "recon",
	}

	notesPath := filepath.Join(testHome, "session.notes.json")
	if err := logs.AppendNote(notesPath, logs.SessionNote{Timestamp: "10:00:00", Content: "found something", ByteOffset: 10}); err != nil {
		t.Fatalf("AppendNote failed: %v", err)
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	if err := vmgr.Save(vulns.Vuln{
		ID:        "vuln-001",
		Title:     "Open redirect",
		Severity:  vulns.SeverityMedium,
		Status:    vulns.StatusOpen,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("Save vuln failed: %v", err)
	}

	sessions := []logs.Session{
		{
			ID:          10,
			DisplayPath: "acme/q1/recon/session.tty",
			ModTime:     "2026-04-17 10:00:00",
			SortKey:     time.Now(),
			Size:        1024,
			NotesPath:   notesPath,
			Metadata: logs.SessionMetadata{
				Client:     "ACME",
				Engagement: "Q1",
				Phase:      "recon",
			},
		},
		{
			ID:          9,
			DisplayPath: "acme/q1/exploit/session.tty",
			ModTime:     "2026-04-17 09:00:00",
			SortKey:     time.Now().Add(-1 * time.Hour),
			Size:        2048,
			Metadata: logs.SessionMetadata{
				Client:     "ACME",
				Engagement: "Q1",
				Phase:      "exploit",
			},
		},
		{
			ID:          8,
			DisplayPath: "other/client/recon/session.tty",
			ModTime:     "2026-04-16 09:00:00",
			SortKey:     time.Now().Add(-24 * time.Hour),
			Size:        512,
			Metadata: logs.SessionMetadata{
				Client:     "Other",
				Engagement: "Ext",
				Phase:      "recon",
			},
		},
	}

	data := buildDashboardData(ctx, sessions)

	if !data.HasContext {
		t.Fatal("expected active context to be detected")
	}
	if data.CurrentScopeSessions != 2 {
		t.Fatalf("expected 2 scope sessions, got %d", data.CurrentScopeSessions)
	}
	if data.CurrentPhaseSessions != 1 {
		t.Fatalf("expected 1 current-phase session, got %d", data.CurrentPhaseSessions)
	}
	if data.CurrentScopeNotes != 1 {
		t.Fatalf("expected 1 scope note, got %d", data.CurrentScopeNotes)
	}
	if data.VulnSummary.Open != 1 || data.VulnSummary.Total != 1 {
		t.Fatalf("unexpected vuln summary: %+v", data.VulnSummary)
	}
	if len(data.RecentFindings) != 1 {
		t.Fatalf("expected 1 recent finding, got %d", len(data.RecentFindings))
	}
}

func TestViewShowsEmptyStateGuidance(t *testing.T) {
	model := Model{
		loaded: true,
		data: DashboardData{
			Actions: defaultActions(),
		},
	}

	view := model.View()
	if !strings.Contains(view, "No recorded sessions yet.") {
		t.Fatalf("expected empty-state guidance in view, got %q", view)
	}
	if !strings.Contains(view, "Action Center") {
		t.Fatalf("expected action center in view, got %q", view)
	}
}
