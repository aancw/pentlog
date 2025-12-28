package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"pentlog/pkg/metadata"
	"strings"
	"time"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type TlogHeader struct {
	Ver     string `json:"ver"`
	Host    string `json:"host"`
	Rec     string `json:"rec"`
	User    string `json:"user"`
	Term    string `json:"term"`
	Session int    `json:"session"`
}

func ExtractCommands(phase string) (string, error) {
	history, err := metadata.LoadHistory()
	if err != nil {
		return "", err
	}

	var ranges []TimeRange

	for i, entry := range history {
		if entry.Phase == phase {
			start, err := time.Parse(time.RFC3339, entry.Timestamp)
			if err != nil {
				continue
			}

			end := time.Now()
			if i+1 < len(history) {
				nextStart, err := time.Parse(time.RFC3339, history[i+1].Timestamp)
				if err == nil {
					end = nextStart
				}
			}
			ranges = append(ranges, TimeRange{Start: start, End: end})
		}
	}

	if len(ranges) == 0 {
		return "", fmt.Errorf("no history found for phase: %s", phase)
	}

	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Extraction Report: %s\n\n", phase))

	for _, s := range sessions {
		f, err := os.Open(s.Path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		if !scanner.Scan() {
			continue
		}
		var header TlogHeader
		if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
			continue
		}

		recStart, err := time.Parse(time.RFC3339, header.Rec)
		if err != nil {
			continue
		}

		inRange := false
		for _, r := range ranges {
			if (recStart.Equal(r.Start) || recStart.After(r.Start)) && recStart.Before(r.End) {
				inRange = true
				break
			}
		}

		if !inRange {
			continue
		}

		builder.WriteString(fmt.Sprintf("## Session %d (%s)\n", s.ID, header.Rec))
		builder.WriteString("```bash\n")

		var currentCmd strings.Builder

		for scanner.Scan() {
			var msg map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}

			if val, ok := msg["in_txt"]; ok {
				if str, ok := val.(string); ok {
					for _, r := range str {
						if r == '\r' || r == '\n' {
							cmdStr := currentCmd.String()
							if strings.TrimSpace(cmdStr) != "" {
								builder.WriteString(cmdStr + "\n")
							}
							currentCmd.Reset()
						} else if r == 127 || r == 8 {
							s := currentCmd.String()
							if len(s) > 0 {
								rs := []rune(s)
								currentCmd.Reset()
								currentCmd.WriteString(string(rs[:len(rs)-1]))
							}
						} else {
							if r >= 32 || r == '\t' {
								currentCmd.WriteRune(r)
							}
						}
					}
				}
			}
		}
		if currentCmd.Len() > 0 {
			builder.WriteString(currentCmd.String() + "\n")
		}

		builder.WriteString("```\n\n")
	}

	return builder.String(), nil
}
