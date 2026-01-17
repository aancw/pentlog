package logs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"pentlog/pkg/utils"
	"regexp"
	"strings"
	"time"
)

type CommandExecution struct {
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
	Output    string `json:"output"`
}

type Timeline struct {
	Commands []CommandExecution `json:"commands"`
}

type TtyrecFrame struct {
	Sec  uint32
	Usec uint32
	Len  uint32
	Data []byte
}

func readTtyrecFrame(r io.Reader) (*TtyrecFrame, error) {
	header := make([]byte, 12)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	frame := &TtyrecFrame{
		Sec:  binary.LittleEndian.Uint32(header[0:4]),
		Usec: binary.LittleEndian.Uint32(header[4:8]),
		Len:  binary.LittleEndian.Uint32(header[8:12]),
	}

	frame.Data = make([]byte, frame.Len)
	_, err = io.ReadFull(r, frame.Data)
	if err != nil {
		return nil, err
	}

	return frame, nil
}

type FrameWithTimestamp struct {
	Timestamp time.Time
	Data      []byte
}

func readAllFrames(ttyPath string) ([]FrameWithTimestamp, error) {
	f, err := os.Open(ttyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var frames []FrameWithTimestamp
	for {
		frame, err := readTtyrecFrame(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read ttyrec frame: %w", err)
		}

		timestamp := time.Unix(int64(frame.Sec), int64(frame.Usec)*1000)
		frames = append(frames, FrameWithTimestamp{
			Timestamp: timestamp,
			Data:      frame.Data,
		})
	}

	return frames, nil
}

func stripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\a]*\a|\x1b[=>]|\x1b\?[0-9]+[hl]`)
	return ansiRegex.ReplaceAllString(str, "")
}

func cleanControlChars(str string) string {
	var buffer []rune
	cursor := 0

	for _, r := range str {
		if r == '\b' { // Backspace
			if cursor > 0 {
				cursor--
			}
		} else if r == '\r' { // Carriage Return
			cursor = 0
		} else if r == '\n' {
			buffer = append(buffer, r)
			cursor = len(buffer)
		} else if r >= 0x20 || r == '\t' {
			if cursor < len(buffer) {
				buffer[cursor] = r
			} else {
				buffer = append(buffer, r)
			}
			cursor++
		}
	}

	return string(buffer)
}

func isPromptLine(line string) bool {
	patterns := []string{
		`└─\$`,            // Kali/custom prompt like "└─$" (ACTUAL command line, not header)
		`\$\s+$`,          // Simple $ at end
		`#\s+$`,           // Root # at end
		`[a-zA-Z0-9_-]+@`, // user@host pattern
		`:\~\$`,           // :~$ pattern
		`:\~#`,            // :~# pattern
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}

	return false
}

func extractCommand(line string) string {
	cleaned := stripANSI(line)

	escapeRegex := regexp.MustCompile(`\[\?[0-9]+[hl]`)
	cleaned = escapeRegex.ReplaceAllString(cleaned, "")

	// Look for common prompt patterns and extract what comes after
	// Pattern 1: "└─$ command" or similar
	if idx := strings.Index(cleaned, "└─$"); idx != -1 {
		after := cleaned[idx+len("└─$"):]
		after = strings.TrimSpace(after)

		// Remove context label like "(pentlog:HTB/recon)" from beginning and end
		contextRegex := regexp.MustCompile(`^\([^)]+\)\s*|\s*\([^)]+\)$`)
		after = contextRegex.ReplaceAllString(after, "")

		after = cleanCommandText(after)

		return strings.TrimSpace(after)
	}

	// Pattern 2: Simple $ or # prompt
	dollarIdx := strings.LastIndex(cleaned, "$")
	hashIdx := strings.LastIndex(cleaned, "#")

	promptIdx := -1
	if dollarIdx > hashIdx {
		promptIdx = dollarIdx
	} else if hashIdx != -1 {
		promptIdx = hashIdx
	}

	if promptIdx != -1 && promptIdx < len(cleaned)-1 {
		after := cleaned[promptIdx+1:]
		after = strings.TrimSpace(after)

		contextRegex := regexp.MustCompile(`^\([^)]+\)\s*|\s*\([^)]+\)$`)
		after = contextRegex.ReplaceAllString(after, "")

		after = cleanCommandText(after)

		return strings.TrimSpace(after)
	}

	return ""
}

func cleanCommandText(cmd string) string {
	// First pass: remove all excessive whitespace
	cmd = regexp.MustCompile(`\s+`).ReplaceAllString(cmd, " ")
	cmd = strings.TrimSpace(cmd)

	// Look for patterns where characters repeat with garbage in between
	// This handles: "pentlog shelpw pwd" -> should be "pwd"
	// The pattern is: previous typing followed by many spaces/chars then the actual command

	// Strategy: if we see the same word/command repeated, take the last occurrence
	// Split by spaces and look for command patterns
	words := strings.Fields(cmd)
	if len(words) == 0 {
		return ""
	}

	// Check if this looks like progressive typing (same command name appears multiple times)
	// For example: "nmap -sn ... nmap -Pn ..." should extract the last complete command
	firstWord := words[0]
	indices := []int{}
	for i, word := range words {
		if word == firstWord {
			indices = append(indices, i)
		}
	}

	// If first word appears multiple times, take from the LAST occurrence
	if len(indices) > 1 {
		lastIdx := indices[len(indices)-1]
		words = words[lastIdx:]
		cmd = strings.Join(words, " ")
	}

	// Look for obviously corrupted patterns - if command contains mixed fragments
	// Common pattern: "commandA fragmentsB commandC" where commandC is likely the real one
	// Heuristic: if length of "words" varies dramatically, take the consistent tail
	if len(words) > 3 {
		// Check for very short words mixed with normal ones (typing artifacts)
		var cleanWords []string
		inCleanSection := false

		for i := len(words) - 1; i >= 0; i-- {
			word := words[i]
			// If we're building from the end and hit a likely artifact, stop
			if !inCleanSection && (len(word) <= 2 && !isCommonShortCommand(word)) {
				continue
			}
			inCleanSection = true
			cleanWords = append([]string{word}, cleanWords...)
		}

		if len(cleanWords) > 0 && len(cleanWords) < len(words) {
			cmd = strings.Join(cleanWords, " ")
		}
	}

	return strings.TrimSpace(cmd)
}

func isCommonShortCommand(word string) bool {
	common := map[string]bool{
		"ls": true, "cd": true, "cp": true, "mv": true, "rm": true,
		"id": true, "ps": true, "su": true, "vi": true, "ip": true,
	}
	return common[word]
}

func ParseTimeline(ttyPath string) (*Timeline, error) {
	frames, err := readAllFrames(ttyPath)
	if err != nil {
		return nil, err
	}

	// Combine all frames into a single text stream with frame timestamps
	var accumulated strings.Builder
	var frameTimestamps []time.Time
	var framePositions []int

	for _, frame := range frames {
		framePositions = append(framePositions, accumulated.Len())
		frameTimestamps = append(frameTimestamps, frame.Timestamp)
		accumulated.Write(frame.Data)
	}

	fullText := accumulated.String()
	cleanedBytes := utils.CleanTuiMarkers([]byte(fullText))
	cleaned := string(cleanedBytes)

	rawLines := strings.Split(cleaned, "\n")
	var lines []string
	for _, line := range rawLines {
		// Process each line individually to properly handle cursor movements
		cleanedLine := utils.RenderPlain(line)
		lines = append(lines, cleanedLine)
	}

	// Parse commands and outputs
	var timeline Timeline
	var currentCommand *CommandExecution
	var outputLines []string

	// Estimate timestamp based on proximity to frames
	estimateTimestamp := func(lineNum int) time.Time {
		if len(frameTimestamps) > 0 {
			return frameTimestamps[0].Add(time.Duration(lineNum) * time.Second)
		}
		return time.Now()
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" && currentCommand == nil {
			continue
		}

		if isPromptLine(line) {
			if currentCommand != nil {
				currentCommand.Output = strings.TrimSpace(strings.Join(outputLines, "\n"))
				timeline.Commands = append(timeline.Commands, *currentCommand)
			}

			cmd := extractCommand(line)

			if cmd != "" {
				currentCommand = &CommandExecution{
					Timestamp: estimateTimestamp(i).Format("2006-01-02 15:04:05"),
					Command:   cmd,
					Output:    "",
				}
				outputLines = []string{}
			}
		} else if currentCommand != nil {
			outputLines = append(outputLines, line)
		}
	}

	if currentCommand != nil {
		currentCommand.Output = strings.TrimSpace(strings.Join(outputLines, "\n"))
		timeline.Commands = append(timeline.Commands, *currentCommand)
	}

	return &timeline, nil
}

func extractFinalCommand(rawData string) string {
	cleaned := utils.RenderPlain(rawData)

	lines := strings.Split(cleaned, "\n")

	var commandLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if isPromptLine(line) {
			commandLine = line
		}
	}

	if commandLine == "" {
		return ""
	}

	cmd := extractCommand(commandLine)

	return strings.TrimSpace(cmd)
}

func cleanOutput(rawData string) string {
	cleaned := utils.RenderPlain(rawData)

	cleanedBytes := utils.CleanTuiMarkers([]byte(cleaned))
	cleaned = string(cleanedBytes)

	lines := strings.Split(cleaned, "\n")
	var outputLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !isPromptLine(line) && line != "" {
			outputLines = append(outputLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(outputLines, "\n"))
}

func (t *Timeline) ToJSON() (string, error) {
	data, err := json.MarshalIndent(t.Commands, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal timeline: %w", err)
	}
	return string(data), nil
}
