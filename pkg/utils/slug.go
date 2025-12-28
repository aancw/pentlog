package utils

import (
	"regexp"
	"strings"
)

var sanitizePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func Slugify(input string) string {
	s := strings.TrimSpace(strings.ToLower(input))
	if s == "" {
		return "default"
	}
	s = sanitizePattern.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "default"
	}
	return s
}
