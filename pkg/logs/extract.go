package logs

import (
	"fmt"
	"os"
	"strings"
)

func ExtractCommands(phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Extraction Report: %s\n\n", phase))

	found := false
	for _, s := range sessions {
		if strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(phase)) {
			continue
		}

		data, err := os.ReadFile(s.Path)
		if err != nil {
			continue
		}

		found = true
		builder.WriteString(fmt.Sprintf("## Session %d (%s)\n", s.ID, s.ModTime))
		builder.WriteString("```text\n")
		builder.Write(data)
		builder.WriteString("\n```\n\n")
	}

	if !found {
		return "", fmt.Errorf("no sessions found for phase: %s", phase)
	}

	return builder.String(), nil
}
