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

// chunkText splits a log file into chunks based on a character limit,
// ensuring that splits happen at line boundaries.
// chunkText splits a log file into chunks based on a character limit,
// ensuring that splits happen at line boundaries.
func chunkText(text string) []string {
	const maxChunkSize = 4000
	var chunks []string
	var currentChunk strings.Builder

	// SplitAfter keeps the delimiter in the substring, which helps preserve exact content
	lines := strings.SplitAfter(text, "\n")

	for _, line := range lines {
		// Skip empty strings that result from splitting strings ending with the delimiter
		if len(line) == 0 {
			continue
		}

		lineLength := len(line)

		// If adding this line would exceed the limit, and we have content, finalize the current chunk
		if currentChunk.Len()+lineLength > maxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		currentChunk.WriteString(line)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
