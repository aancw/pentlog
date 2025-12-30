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
	Context []string
	IsNote  bool
}

func Search(query string, scopeSessions []logs.Session) ([]Match, error) {
	regex, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex query: %w", err)
	}

	var sessions []logs.Session
	var errSession error

	if len(scopeSessions) > 0 {
		sessions = scopeSessions
	} else {
		sessions, errSession = logs.ListSessions()
		if errSession != nil {
			return nil, fmt.Errorf("failed to list sessions: %w", errSession)
		}
	}

	var results []Match

	for _, session := range sessions {
		if session.Path != "" {
			f, err := os.Open(session.Path)
			if err == nil {

				var lines []string
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				f.Close()

				for i, line := range lines {
					if regex.MatchString(line) {
						start := i - 2
						if start < 0 {
							start = 0
						}
						end := i + 3
						if end > len(lines) {
							end = len(lines)
						}

						results = append(results, Match{
							Session: session,
							LineNum: i + 1,
							Content: line,
							Context: lines[start:end],
							IsNote:  false,
						})
					}
				}
			}
		}

		if session.NotesPath != "" {
			notes, err := logs.ReadNotes(session.NotesPath)
			if err == nil {
				for _, note := range notes {
					if regex.MatchString(note.Content) {
						results = append(results, Match{
							Session: session,
							LineNum: int(note.ByteOffset),
							Content: fmt.Sprintf("[%s] %s", note.Timestamp, note.Content),
							IsNote:  true,
						})
					}
				}
			}
		}
	}

	return results, nil
}
