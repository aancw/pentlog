package ai

import (
	"strings"
)

// Summarize takes a long text and summarizes it by chunking it and sending each chunk to an AI analyzer.
func Summarize(text string, analyzer AIAnalyzer) (string, error) {
	chunks := chunkText(text)
	var summaries []string

	for _, chunk := range chunks {
		summary, err := analyzer.Analyze(chunk, false)
		if err != nil {
			return "", err
		}
		summaries = append(summaries, summary)
	}

	finalSummary, err := analyzer.Analyze(strings.Join(summaries, "\n\n"), false)
	if err != nil {
		return "", err
	}

	return finalSummary, nil
}

// chunkText splits a log file into chunks, where each chunk is a command and its output.
func chunkText(text string) []string {
	var chunks []string
	var currentChunk strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if isCommandPrompt(line) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}
		}
		currentChunk.WriteString(line + "\n")
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// isCommandPrompt checks if a line is a command prompt.
// This is a simple implementation and can be improved.
func isCommandPrompt(line string) bool {
	return strings.HasPrefix(line, "$ ")
}
