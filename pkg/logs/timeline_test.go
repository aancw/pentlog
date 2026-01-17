package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTimeline(t *testing.T) {
	// Create a test ttyrec file with known content

	// This is a simplified ttyrec format simulation
	// In real scenarios, we would use actual ttyrec binary data
	// For now, we'll skip this test and rely on manual testing
	t.Skip("Skipping integration test - requires real ttyrec file")
}

func TestIsPromptLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Kali prompt", "└─$", true},
		{"Kali prompt start", "┌──(", true},
		{"Simple dollar", "user@host:~$ ", true},
		{"Root prompt", "root@host:~# ", true},
		{"Not a prompt", "some output text", false},
		{"Empty line", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPromptLine(tt.line)
			if result != tt.expected {
				t.Errorf("isPromptLine(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestExtractCommand(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Kali prompt with context",
			line:     "└─$ (pentlog:HTB/recon) pwd",
			expected: "pwd",
		},
		{
			name:     "Simple command",
			line:     "└─$ ls -la",
			expected: "ls -la",
		},
		{
			name:     "With ANSI codes",
			line:     "└─$ \x1b[1mpwd\x1b[0m",
			expected: "pwd",
		},
		{
			name:     "With escape sequences",
			line:     "└─$ [?1h[?2004hpwd[?1l[?2004l",
			expected: "pwd",
		},
		{
			name:     "No prompt",
			line:     "some output",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCommand(tt.line)
			if result != tt.expected {
				t.Errorf("extractCommand(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic ANSI color codes",
			input:    "\x1b[32mgreen text\x1b[0m",
			expected: "green text",
		},
		{
			name:     "Window title escape",
			input:    "\x1b]0;terminal title\x07text",
			expected: "text",
		},
		{
			name:     "No ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Backspace handling",
			input:    "helo\blo",
			expected: "hello",
		},
		{
			name:     "Carriage return handling",
			input:    "first\rsecond",
			expected: "second",
		},
		{
			name:     "No control chars",
			input:    "normal text",
			expected: "normal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanControlChars(tt.input)
			if result != tt.expected {
				t.Errorf("cleanControlChars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTimelineToJSON(t *testing.T) {
	timeline := &Timeline{
		Commands: []CommandExecution{
			{
				Timestamp: "2026-01-15 13:46:58",
				Command:   "pwd",
				Output:    "/home/kali",
			},
			{
				Timestamp: "2026-01-15 13:47:02",
				Command:   "id",
				Output:    "uid=1000(kali)",
			},
		},
	}

	jsonStr, err := timeline.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	if !strings.Contains(jsonStr, "pwd") {
		t.Errorf("JSON output missing expected command 'pwd': %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "/home/kali") {
		t.Errorf("JSON output missing expected output '/home/kali': %s", jsonStr)
	}
}

// TestSampleDataFile tests with actual sample data if it exists
func TestSampleDataFile(t *testing.T) {
	samplePath := filepath.Join("..", "..", "sample_data", "session-kali-20260115-134657.tty")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("Sample data file not found, skipping integration test")
		return
	}

	timeline, err := ParseTimeline(samplePath)
	if err != nil {
		t.Fatalf("ParseTimeline() failed: %v", err)
	}

	if len(timeline.Commands) == 0 {
		t.Error("Expected at least one command in timeline")
	}

	// Verify the first command is reasonable
	if len(timeline.Commands) > 0 {
		firstCmd := timeline.Commands[0]
		if firstCmd.Command == "" {
			t.Error("First command is empty")
		}
		if firstCmd.Timestamp == "" {
			t.Error("First command has no timestamp")
		}
	}
}
