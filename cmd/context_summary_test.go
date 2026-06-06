package cmd

import (
	"pentlog/pkg/config"
	"strings"
	"testing"
	"time"
)

func TestSummarizeRecentContextChangesAtSkipsTimestampOnlySaves(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)

	history := []config.ContextData{
		{
			Client:     "ACME",
			Engagement: "Q2",
			Phase:      "recon",
			Timestamp:  now.Add(-4 * time.Hour).Format(time.RFC3339),
			Type:       "Client",
		},
		{
			Client:     "ACME",
			Engagement: "Q2",
			Phase:      "exploit",
			Timestamp:  now.Add(-2 * time.Hour).Format(time.RFC3339),
			Type:       "Client",
		},
		{
			Client:     "ACME",
			Engagement: "Q2",
			Phase:      "exploit",
			Target:     "dc01",
			TargetIP:   "10.10.10.10",
			Timestamp:  now.Add(-90 * time.Minute).Format(time.RFC3339),
			Type:       "Client",
		},
		{
			Client:     "ACME",
			Engagement: "Q2",
			Phase:      "exploit",
			Target:     "dc01",
			TargetIP:   "10.10.10.10",
			Timestamp:  now.Add(-30 * time.Minute).Format(time.RFC3339),
			Type:       "Client",
		},
	}

	changes := summarizeRecentContextChangesAt(history, 4, now)
	if len(changes) != 3 {
		t.Fatalf("expected 3 meaningful changes, got %d: %#v", len(changes), changes)
	}

	assertContains(t, changes[0], "target -> dc01 (10.10.10.10)")
	assertContains(t, changes[1], "phase recon -> exploit")
	assertContains(t, changes[2], "Context created for ACME / Q2 / recon")
}

func TestCollectShellPreflightWarningsAtFlagsKeyRisks(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)

	saved := config.ContextData{
		Client:     "ACME",
		Engagement: "Q2",
		Phase:      "recon",
		Timestamp:  now.Add(-10 * time.Hour).Format(time.RFC3339),
		Type:       "Client",
	}

	effective := saved
	effective.Phase = "exploit"

	targets := []config.Target{
		{Name: "dc01", IP: "10.10.10.10"},
		{Name: "web01", IP: "10.10.10.20"},
	}

	warnings := collectShellPreflightWarningsAt(saved, effective, targets, now)
	if len(warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d: %#v", len(warnings), warnings)
	}

	assertContains(t, warnings[0], "Saved context is")
	assertContains(t, warnings[1], "Multiple targets are configured")
	assertContains(t, warnings[2], "phase recon -> exploit")
}

func TestDescribePendingContextMutationSummarizesPhaseAndTarget(t *testing.T) {
	saved := config.ContextData{
		Phase:    "recon",
		Target:   "dc01",
		TargetIP: "10.10.10.10",
	}
	effective := config.ContextData{
		Phase:    "post",
		Target:   "web01",
		TargetIP: "10.10.10.20",
	}

	summary := describePendingContextMutation(saved, effective)
	assertContains(t, summary, "phase recon -> post")
	assertContains(t, summary, "target dc01 (10.10.10.10) -> web01 (10.10.10.20)")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}
