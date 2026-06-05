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
	Session          logs.Session
	LineNum          int
	Content          string
	Context          []string
	ContextStartLine int
	IsNote           bool
	NoteTimestamp    string
}

type SearchOptions struct {
	After   time.Time
	Before  time.Time
	IsRegex bool
	Limit   int
	Offset  int
}

type Page struct {
	Matches []Match
	Total   int
}

func Search(query string, scopeSessions []logs.Session, opts SearchOptions) ([]Match, error) {
	filteredSessions, matcher, err := prepareSearch(query, scopeSessions, opts)
	if err != nil {
		return nil, err
	}

	results := make([]Match, 0)
	matchCount := 0

	for _, session := range filteredSessions {
		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}

		sessionMatches := collectSessionMatches(session, matcher)
		for _, match := range sessionMatches {
			if matchCount < opts.Offset {
				matchCount++
				continue
			}

			results = append(results, match)
			matchCount++

			if opts.Limit > 0 && len(results) >= opts.Limit {
				break
			}
		}
	}

	return results, nil
}

func SearchPage(query string, scopeSessions []logs.Session, opts SearchOptions) (Page, error) {
	filteredSessions, matcher, err := prepareSearch(query, scopeSessions, opts)
	if err != nil {
		return Page{}, err
	}

	allMatches := make([]Match, 0)
	for _, session := range filteredSessions {
		allMatches = append(allMatches, collectSessionMatches(session, matcher)...)
	}

	total := len(allMatches)
	if opts.Offset > total {
		opts.Offset = total
	}

	end := total
	if opts.Limit > 0 && opts.Offset+opts.Limit < end {
		end = opts.Offset + opts.Limit
	}

	return Page{
		Matches: allMatches[opts.Offset:end],
		Total:   total,
	}, nil
}

func prepareSearch(query string, scopeSessions []logs.Session, opts SearchOptions) ([]logs.Session, func(string) bool, error) {
	var sessions []logs.Session
	if len(scopeSessions) > 0 {
		sessions = scopeSessions
	} else {
		all, err := logs.ListSessions()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list sessions: %w", err)
		}
		sessions = all
	}

	filteredSessions := make([]logs.Session, 0, len(sessions))
	for _, session := range sessions {
		ts, err := time.Parse(time.RFC3339, session.Metadata.Timestamp)
		if err != nil {
			ts = session.SortKey
		}

		if !opts.After.IsZero() && ts.Before(opts.After) {
			continue
		}
		if !opts.Before.IsZero() && ts.After(opts.Before) {
			continue
		}

		filteredSessions = append(filteredSessions, session)
	}

	matcher, err := buildMatcher(query, opts.IsRegex)
	if err != nil {
		return nil, nil, err
	}

	return filteredSessions, matcher, nil
}

func buildMatcher(query string, isRegex bool) (func(string) bool, error) {
	if isRegex {
		regex, err := regexp.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("invalid regex query: %w", err)
		}

		return func(text string) bool {
			return regex.MatchString(text)
		}, nil
	}

	return createBooleanMatcher(query), nil
}

func collectSessionMatches(session logs.Session, matcher func(string) bool) []Match {
	results := make([]Match, 0)

	if session.Path != "" {
		lines, err := readSearchLines(session.Path)
		if err == nil {
			for index, line := range lines {
				if !matcher(line) {
					continue
				}

				start := index - 2
				if start < 0 {
					start = 0
				}
				end := index + 3
				if end > len(lines) {
					end = len(lines)
				}

				results = append(results, Match{
					Session:          session,
					LineNum:          index + 1,
					Content:          line,
					Context:          lines[start:end],
					ContextStartLine: start + 1,
				})
			}
		}
	}

	if session.NotesPath != "" {
		notes, err := logs.ReadNotes(session.NotesPath)
		if err == nil {
			for _, note := range notes {
				if !matcher(note.Content) {
					continue
				}

				results = append(results, Match{
					Session:       session,
					LineNum:       int(note.ByteOffset),
					Content:       note.Content,
					IsNote:        true,
					NoteTimestamp: note.Timestamp,
				})
			}
		}
	}

	return results
}

func readSearchLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".tty") {
		reader = logs.NewTtyReader(f)
	}

	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, utils.StripANSI(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
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
