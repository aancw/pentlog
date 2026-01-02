package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"pentlog/pkg/utils"
	"sort"
	"strings"
)

func ExportCommands(client, engagement, phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	grouped := make(map[string]map[string][]Session)
	hasData := false

	for _, s := range sessions {
		if client != "" && s.Metadata.Client != client {
			continue
		}
		if engagement != "" && s.Metadata.Engagement != engagement {
			continue
		}
		if phase != "" && strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(phase)) {
			continue
		}

		if grouped[s.Metadata.Engagement] == nil {
			grouped[s.Metadata.Engagement] = make(map[string][]Session)
		}
		grouped[s.Metadata.Engagement][s.Metadata.Phase] = append(grouped[s.Metadata.Engagement][s.Metadata.Phase], s)
		hasData = true
	}

	if !hasData {
		return "", fmt.Errorf("no sessions found matching criteria")
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Report for Client %s\n\n", client))

	var engKeys []string
	for k := range grouped {
		engKeys = append(engKeys, k)
	}
	sort.Strings(engKeys)

	for _, eng := range engKeys {
		builder.WriteString(fmt.Sprintf("## Engagement: %s\n", eng))
		builder.WriteString("---------------------------------------------------\n\n")

		phases := grouped[eng]
		var phaseKeys []string
		for k := range phases {
			phaseKeys = append(phaseKeys, k)
		}
		sort.Strings(phaseKeys)

		for _, p := range phaseKeys {
			builder.WriteString(fmt.Sprintf("### Phase: %s\n", p))
			builder.WriteString("--------------------\n\n")

			sessList := phases[p]
			// Sort sessions by ID
			sort.Slice(sessList, func(i, j int) bool {
				return sessList[i].ID < sessList[j].ID
			})

			for _, s := range sessList {
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

				builder.WriteString(fmt.Sprintf("#### Session %d (%s)\n", s.ID, s.ModTime))
				builder.WriteString("```bash\n")
				builder.Write(data)
				builder.WriteString("\n```\n\n")
			}
		}
	}

	return builder.String(), nil
}

func ExportCommandsHTML(client, engagement, phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	grouped := make(map[string]map[string][]Session)
	hasData := false

	for _, s := range sessions {
		if client != "" && s.Metadata.Client != client {
			continue
		}
		if engagement != "" && s.Metadata.Engagement != engagement {
			continue
		}
		if phase != "" && strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(phase)) {
			continue
		}

		if grouped[s.Metadata.Engagement] == nil {
			grouped[s.Metadata.Engagement] = make(map[string][]Session)
		}
		grouped[s.Metadata.Engagement][s.Metadata.Phase] = append(grouped[s.Metadata.Engagement][s.Metadata.Phase], s)
		hasData = true
	}

	if !hasData {
		return "", fmt.Errorf("no sessions found matching criteria")
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
        h1 { color: #569cd6; border-bottom: 2px solid #569cd6; padding-bottom: 10px; }
        h2 { color: #4ec9b0; margin-top: 40px; border-bottom: 1px solid #444; padding-bottom: 5px; }
        h3 { color: #dcdcaa; margin-top: 30px; }
        h4 { color: #9cdcfe; margin-top: 20px; font-size: 1.1em; }
        
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

	builder.WriteString(fmt.Sprintf("    <h1>Report for Client: %s</h1>\n", client))

	var engKeys []string
	for k := range grouped {
		engKeys = append(engKeys, k)
	}
	sort.Strings(engKeys)

	for _, eng := range engKeys {
		builder.WriteString(fmt.Sprintf("    <h2>Engagement: %s</h2>\n", eng))

		phases := grouped[eng]
		var phaseKeys []string
		for k := range phases {
			phaseKeys = append(phaseKeys, k)
		}
		sort.Strings(phaseKeys)

		for _, p := range phaseKeys {
			builder.WriteString(fmt.Sprintf("    <h3>Phase: %s</h3>\n", p))

			sessList := phases[p]
			sort.Slice(sessList, func(i, j int) bool {
				return sessList[i].ID < sessList[j].ID
			})

			for _, s := range sessList {
				f, err := os.Open(s.Path)
				if err != nil {
					continue
				}

				builder.WriteString("    <div class=\"session\">\n")
				builder.WriteString(fmt.Sprintf("        <h4>Session %d (%s)</h4>\n", s.ID, s.ModTime))
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
			}
		}
	}

	builder.WriteString("</body>\n</html>")

	return builder.String(), nil
}
