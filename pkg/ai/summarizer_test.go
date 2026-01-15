package ai

import (
	"strings"
	"testing"
)

// TestChunkText verifies that chunkText splits based on length
// and preserves content.
func TestChunkText(t *testing.T) {
	// Create a long text with multiple lines
	line := "This is a test line for chunking." // 33 chars
	// We want to exceed 4000 chars.
	// 4000 / 33 ~= 121 lines.

	var sb strings.Builder
	for i := 0; i < 150; i++ {
		sb.WriteString(line + "\n")
	}
	longText := sb.String()

	chunks := chunkText(longText)

	if len(chunks) <= 1 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	// Verify total content is preserved
	reconstructed := strings.Join(chunks, "")
	if reconstructed != longText {
		t.Errorf("Content mismatch. \nExpected length: %d\nGot length: %d", len(longText), len(reconstructed))
	}

	// Verify max chunk size (approximate, since we split by line)
	const maxChunkSize = 4000
	for i, chunk := range chunks {
		if len(chunk) > maxChunkSize+len(line)+10 {
			// Allow some buffer because our logic allows going over if a single line pushes it over,
			// but here lines are short, so it should be very close to 4000.
			t.Errorf("Chunk %d is too large: %d characters", i, len(chunk))
		}
	}
}

// MockAnalyzer for testing Summarize
type MockAnalyzer struct{}

func (m *MockAnalyzer) Analyze(text string, summarize bool) (string, error) {
	return "Summary", nil
}

func TestSummarize(t *testing.T) {
	analyzer := &MockAnalyzer{}
	text := "Short text"
	summary, err := Summarize(text, analyzer)
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}
	if summary != "Summary" {
		t.Errorf("Expected 'Summary', got '%s'", summary)
	}
}
