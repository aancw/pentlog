package search

import (
	"bufio"
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"strings"
)

type Match struct {
	Session logs.Session
	LineNum int
	Content string
}

func Search(query string) ([]Match, error) {
	sessions, err := logs.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var results []Match
	lowerQuery := strings.ToLower(query)

	for _, session := range sessions {
		if session.Path == "" {
			continue
		}

		f, err := os.Open(session.Path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if strings.Contains(strings.ToLower(line), lowerQuery) {
				results = append(results, Match{
					Session: session,
					LineNum: lineNum,
					Content: line,
				})
			}
		}
		f.Close()
	}

	return results, nil
}
