package cmd

import (
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"testing"
)

func TestFilterSessionsByScopePhaseInsensitive(t *testing.T) {
	sessions := []logs.Session{
		{Metadata: logs.SessionMetadata{Client: "ACME", Engagement: "Q1", Phase: "Recon"}},
		{Metadata: logs.SessionMetadata{Client: "ACME", Engagement: "Q1", Phase: "Exploit"}},
		{Metadata: logs.SessionMetadata{Client: "Other", Engagement: "Q1", Phase: "Recon"}},
	}

	filtered := filterSessionsByScope(sessions, sessionScope{
		Client:     "ACME",
		Engagement: "Q1",
		Phase:      "recon",
	})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered session, got %d", len(filtered))
	}
	if filtered[0].Metadata.Phase != "Recon" {
		t.Fatalf("unexpected phase %q", filtered[0].Metadata.Phase)
	}
}

func TestFormatScopeLabel(t *testing.T) {
	ctx := &config.ContextData{
		Client:     "ACME",
		Engagement: "Q1",
		Phase:      "recon",
	}

	if got := formatScopeLabel(ctx, false); got != "Current engagement (ACME / Q1)" {
		t.Fatalf("unexpected engagement label: %q", got)
	}
	if got := formatScopeLabel(ctx, true); got != "Current phase (ACME / Q1 / recon)" {
		t.Fatalf("unexpected phase label: %q", got)
	}
}
