package search

import (
	"bufio"
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"regexp"
)

type Match struct {
	Session logs.Session
	LineNum int
	Content string
}

func Search(query string) ([]Match, error) {
	regex, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex query: %w", err)
	}

	sessions, err := logs.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var results []Match

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

			if regex.MatchString(line) {
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
