package search

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"regexp"
	"strings"
	"time"
)

type Match struct {
	Session logs.Session
	LineNum int
	Content string
	Context []string
	IsNote  bool
}

type SearchOptions struct {
	After   time.Time
	Before  time.Time
	IsRegex bool
	Limit   int
	Offset  int
}

func Search(query string, scopeSessions []logs.Session, opts SearchOptions) ([]Match, error) {
	var results []Match

	var sessions []logs.Session
	if len(scopeSessions) > 0 {
		sessions = scopeSessions
	} else {
		all, err := logs.ListSessions()
		if err != nil {
			return nil, fmt.Errorf("failed to list sessions: %w", err)
		}
		sessions = all
	}

	filteredSessions := []logs.Session{}
	for _, s := range sessions {
		if s.Metadata.Timestamp == "" {

		}

		ts, err := time.Parse(time.RFC3339, s.Metadata.Timestamp)
		if err != nil {
			ts = s.SortKey
		}

		if !opts.After.IsZero() && ts.Before(opts.After) {
			continue
		}
		if !opts.Before.IsZero() && ts.After(opts.Before) {
			continue
		}
		filteredSessions = append(filteredSessions, s)
	}

	var matcher func(string) bool
	if opts.IsRegex {
		regex, err := regexp.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("invalid regex query: %w", err)
		}
		matcher = func(text string) bool {
			return regex.MatchString(text)
		}
	} else {

		matcher = createBooleanMatcher(query)
	}

	matchCount := 0
	for _, session := range filteredSessions {
		// Stop early if we have enough results (optimization)
		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}

		if session.Path != "" {
			f, err := os.Open(session.Path)
			if err == nil {

				var r io.Reader = f
				if strings.HasSuffix(session.Path, ".tty") {
					r = logs.NewTtyReader(f)
				}

				var lines []string
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					cleanText := utils.StripANSI(scanner.Text())
					lines = append(lines, cleanText)
				}
				f.Close()

				for i, line := range lines {
					if matcher(line) {
						if matchCount < opts.Offset {
							matchCount++
							continue
						}

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
						matchCount++

						if opts.Limit > 0 && len(results) >= opts.Limit {
							break
						}
					}
				}
			}
		}

		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}

		if session.NotesPath != "" {
			notes, err := logs.ReadNotes(session.NotesPath)
			if err == nil {
				for _, note := range notes {
					if matcher(note.Content) {
						if matchCount < opts.Offset {
							matchCount++
							continue
						}

						results = append(results, Match{
							Session: session,
							LineNum: int(note.ByteOffset),
							Content: fmt.Sprintf("[%s] %s", note.Timestamp, note.Content),
							IsNote:  true,
						})
						matchCount++

						if opts.Limit > 0 && len(results) >= opts.Limit {
							break
						}
					}
				}
			}
		}
	}

	return results, nil
}

func createBooleanMatcher(query string) func(string) bool {
	query = strings.Join(strings.Fields(query), " ")

	orGroups := strings.Split(query, " OR ")

	type term struct {
		val string
		not bool
	}

	var parsedGroups [][]term

	for _, group := range orGroups {
		parts := strings.Fields(group)
		var groupTerms []term
		for _, p := range parts {
			if strings.HasPrefix(p, "-") && len(p) > 1 {
				groupTerms = append(groupTerms, term{val: strings.ToLower(p[1:]), not: true})
			} else {
				groupTerms = append(groupTerms, term{val: strings.ToLower(p), not: false})
			}
		}
		parsedGroups = append(parsedGroups, groupTerms)
	}

	return func(text string) bool {
		lowerText := strings.ToLower(text)
		for _, group := range parsedGroups {
			matchGroup := true
			for _, t := range group {
				contains := strings.Contains(lowerText, t.val)
				if t.not && contains {
					matchGroup = false
					break
				}
				if !t.not && !contains {
					matchGroup = false
					break
				}
			}
			if matchGroup {
				return true
			}
		}
		return false
	}
}
