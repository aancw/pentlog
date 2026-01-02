package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"pentlog/pkg/utils"
	"strings"
)

func ExportCommands(phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Export Report: %s\n\n", phase))

	found := false
	for _, s := range sessions {
		if strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(phase)) {
			continue
		}

		f, err := os.Open(s.Path)
		if err != nil {
			continue
		}

		cleaner := utils.NewCleanReader(f)
		data, err := io.ReadAll(cleaner)
		f.Close()

		if err != nil {
			continue
		}

		found = true
		builder.WriteString(fmt.Sprintf("## Session %d (%s)\n", s.ID, s.ModTime))
		builder.WriteString("```bash\n")
		builder.Write(data)
		builder.WriteString("\n```\n\n")
	}

	if !found {
		return "", fmt.Errorf("no sessions found for phase: %s", phase)
	}

	return builder.String(), nil
}

func ExportCommandsHTML(phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pentlog Export Report</title>
    <style>
        body {
            background-color: #1e1e1e;
            color: #d4d4d4;
            font-family: 'Courier New', Courier, monospace;
            padding: 20px;
        }
        h1 {
            color: #569cd6;
        }
        h2 {
            color: #4ec9b0;
            border-bottom: 1px solid #444;
            padding-bottom: 5px;
            margin-top: 30px;
        }
        .session {
            background-color: #252526;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
        }
        .log-content {
            white-space: pre-wrap;
            word-wrap: break-word;
            font-size: 14px;
            line-height: 1.5;
        }

        /* ANSI Colors */
        .ansi-bold { font-weight: bold; }
        .ansi-black { color: #000000; }
        .ansi-red { color: #cd3131; }
        .ansi-green { color: #0dbc79; }
        .ansi-yellow { color: #e5e510; }
        .ansi-blue { color: #2472c8; }
        .ansi-magenta { color: #bc3fbc; }
        .ansi-cyan { color: #11a8cd; }
        .ansi-white { color: #e5e5e5; }
        
        .ansi-bright-black { color: #666666; }
        .ansi-bright-red { color: #f14c4c; }
        .ansi-bright-green { color: #23d18b; }
        .ansi-bright-yellow { color: #f5f543; }
        .ansi-bright-blue { color: #3b8eea; }
        .ansi-bright-magenta { color: #d670d6; }
        .ansi-bright-cyan { color: #29b8db; }
        .ansi-bright-white { color: #ffffff; }
    </style>
</head>
<body>
`)

	builder.WriteString(fmt.Sprintf("    <h1>Export Report: %s</h1>\n", phase))

	found := false
	for _, s := range sessions {
		if strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(phase)) {
			continue
		}

		f, err := os.Open(s.Path)
		if err != nil {
			continue
		}

		builder.WriteString("    <div class=\"session\">\n")
		builder.WriteString(fmt.Sprintf("        <h2>Session %d (%s)</h2>\n", s.ID, s.ModTime))
		builder.WriteString("        <div class=\"log-content\">\n")

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			htmlLine := utils.RenderAnsiHTML(line)
			builder.WriteString(htmlLine + "\n")
		}
		f.Close()

		builder.WriteString("\n        </div>\n")
		builder.WriteString("    </div>\n")
		found = true
	}

	builder.WriteString("</body>\n</html>")

	if !found {
		return "", fmt.Errorf("no sessions found for phase: %s", phase)
	}

	return builder.String(), nil
}
