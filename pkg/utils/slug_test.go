package utils

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal String", "Normal String", "normal_string"},
		{"Already Slugified", "valid-string_123", "valid-string_123"},
		{"Special Characters", "Invalid!@#Characters", "invalid_characters"},
		{"Empty String", "", "default"},
		{"Only Special Chars", "!@#$", "default"},
		{"Multiple Spaces", "Hello   World", "hello_world"},
		{"Leading Trailing Spaces", "  trim me  ", "trim_me"},
		{"Dots Allowed", "file.name", "file.name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.expected {
				t.Errorf("Slugify(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
