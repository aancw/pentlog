package cmd

import (
	"bufio"
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/search"
	"pentlog/pkg/utils"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search command history across all sessions (supports Regex)",
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		var scope []logs.Session

		if len(args) > 0 {
			query = args[0]
		} else {
			allSessions, err := logs.ListSessions()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
				os.Exit(1)
			}

			clientMap := make(map[string]bool)
			var clients []string
			for _, s := range allSessions {
				if s.Metadata.Client != "" && !clientMap[s.Metadata.Client] {
					clientMap[s.Metadata.Client] = true
					clients = append(clients, s.Metadata.Client)
				}
			}

			if len(clients) == 0 {
				fmt.Println("No clients found in sessions.")
				return
			}

			clientIdx := utils.SelectItem("Select Client", clients)
			if clientIdx == -1 {
				return
			}
			selectedClient := clients[clientIdx]

			var scopedSessions []logs.Session
			engagementMap := make(map[string]bool)
			var engagements []string

			for _, s := range allSessions {
				if s.Metadata.Client == selectedClient {
					scopedSessions = append(scopedSessions, s)
					if s.Metadata.Engagement != "" && !engagementMap[s.Metadata.Engagement] {
						engagementMap[s.Metadata.Engagement] = true
						engagements = append(engagements, s.Metadata.Engagement)
					}
				}
			}

			if len(engagements) == 0 {
				fmt.Println("No engagements found for this client.")
				return
			}

			engIdx := utils.SelectItem("Select Engagement", engagements)
			if engIdx == -1 {
				return
			}
			selectedEngagement := engagements[engIdx]

			var finalScope []logs.Session
			for _, s := range scopedSessions {
				if s.Metadata.Engagement == selectedEngagement {
					finalScope = append(finalScope, s)
				}
			}
			scope = finalScope

			query = utils.PromptString("Search Query (Regex)", "")
		}

		if query == "" {
			fmt.Println("Error: Search query cannot be empty.")
			os.Exit(1)
		}

		fmt.Printf("Searching for %q...\n", query)
		results, err := search.Search(query, scope)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No matches found.")
			return
		}

		type item struct {
			Label          string
			CleanContent   string
			CleanContext   string
			DisplaySession string
			DisplayFile    string
			Match          search.Match
		}

		var items []item
		width := utils.GetTerminalWidth()
		safeWidth := width - 2
		if safeWidth < 20 {
			safeWidth = 20
		}

		maxLabelWidth := width - 10
		if maxLabelWidth < 20 {
			maxLabelWidth = 20
		}

		for _, match := range results {
			label := ""
			content := utils.StripANSI(match.Content)

			var contextLines []string
			if len(match.Context) > 0 {
				for _, line := range match.Context {
					stripped := utils.StripANSI(line)
					truncated := utils.TruncateString(stripped, safeWidth)
					contextLines = append(contextLines, truncated)
				}
			} else {
				stripped := utils.StripANSI(content)
				contextLines = []string{utils.TruncateString(stripped, safeWidth)}
			}

			const forcedHeight = 5
			if len(contextLines) < forcedHeight {
				for i := len(contextLines); i < forcedHeight; i++ {
					contextLines = append(contextLines, "")
				}
			} else if len(contextLines) > forcedHeight {
				contextLines = contextLines[:forcedHeight]
			}

			cleanContext := strings.Join(contextLines, "\n")

			sessionStr := fmt.Sprintf("%s / %s", match.Session.Metadata.Client, match.Session.Metadata.Engagement)
			displaySession := utils.TruncateString(sessionStr, safeWidth)
			displayFile := utils.TruncateString(match.Session.DisplayPath, safeWidth)

			if match.IsNote {
				label = fmt.Sprintf("[%d] %s [NOTE]: %s", match.Session.ID, match.Session.DisplayPath, content)
			} else {
				label = fmt.Sprintf("[%d] %s:%d: %s", match.Session.ID, match.Session.DisplayPath, match.LineNum, content)
			}

			if len(label) > maxLabelWidth {
				label = utils.TruncateString(label, maxLabelWidth-3) + "..."
			}

			items = append(items, item{
				Label:          label,
				CleanContent:   content,
				CleanContext:   cleanContext,
				DisplaySession: displaySession,
				DisplayFile:    displayFile,
				Match:          match,
			})
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U000025B6 {{ .Label | cyan }}",
			Inactive: "  {{ .Label }}",
			Selected: "\U000025B6 Content: {{ .CleanContent | cyan }}",
			Details: `
--------- Match Details ----------
{{ "Session:" | faint }}	{{ .DisplaySession }}
{{ "File:" | faint }}	{{ .DisplayFile }}
{{ "Context (5 lines):" | faint }}
{{ .CleanContext }}`,
		}

		prompt := promptui.Select{
			Label:     fmt.Sprintf("Found %d matches. Select to view context:", len(results)),
			Items:     items,
			Templates: templates,
			Size:      10,
		}

		i, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			return
		}

		selected := items[i].Match
		viewContext(selected)
	},
}

func viewContext(m search.Match) {
	f, err := os.Open(m.Session.Path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Printf("\n--- Context View (%s) ---\n", m.Session.DisplayPath)

	scanner := bufio.NewScanner(f)
	currentLine := 0

	if m.IsNote {
		offset := int64(m.LineNum)
		if offset >= 0 {
			f.Seek(offset, 0)
			scanner = bufio.NewScanner(f)
			for i := 0; i < 10 && scanner.Scan(); i++ {
				fmt.Println(scanner.Text())
			}
		} else {
			fmt.Println("No offset information for this note.")
		}
	} else {
		targetLine := m.LineNum
		startLine := targetLine - 2
		if startLine < 1 {
			startLine = 1
		}

		for scanner.Scan() {
			currentLine++
			if currentLine >= startLine {
				if currentLine == targetLine {
					fmt.Printf("\033[1;32m> %s\033[0m\n", scanner.Text())
				} else {
					fmt.Println(scanner.Text())
				}
				if currentLine >= targetLine+5 {
					break
				}
			}
		}
	}
	fmt.Println("-----------------------------------")
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
