package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pentlog/pkg/logs"
	"pentlog/pkg/search"
	"pentlog/pkg/utils"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

var (
	flagAfter  string
	flagBefore string
	flagRegex  bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search command history across all sessions (supports Regex)",
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		var scope []logs.Session

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

		// Interactive Wizard for Filters
		opts := search.SearchOptions{
			IsRegex: flagRegex,
		}

		// If no flags were provided, ask interactively
		if flagAfter == "" && flagBefore == "" && !flagRegex {
			configureIdx := utils.SelectItem("Filter by Date Range?", []string{"No", "Yes"})
			if configureIdx == 1 {
				flagAfter = utils.PromptString("Start Date (DDMMYYYY)", "")
				flagBefore = utils.PromptString("End Date (DDMMYYYY)", "")
			}
		}

		if flagAfter != "" {
			t, err := parseDate(flagAfter)
			if err == nil {
				opts.After = t
			} else {
				fmt.Printf("Warning: Invalid After date format: %v\n", err)
			}
		}
		if flagBefore != "" {
			t, err := parseDate(flagBefore)
			if err == nil {
				opts.Before = t
			} else {
				fmt.Printf("Warning: Invalid Before date format: %v\n", err)
			}
		}

		if len(args) > 0 {
			query = strings.Join(args, " ")
		} else {
			query = utils.PromptString("Search Query", "")
		}

		if query == "" {
			fmt.Println("Error: Search query cannot be empty.")
			os.Exit(1)
		}

		// Pagination setup
		offset := 0
		limit := 10

		type item struct {
			Label          string
			CleanContent   string
			CleanContext   string
			DisplaySession string
			DisplayFile    string
			DisplayTime    string
			Match          search.Match
			IsControl      bool
			FullDetails    string
		}

		var allContentItems []item
		moreAvailable := true

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U000025B6 {{ .Label | cyan }}",
			Inactive: "  {{ .Label }}",
			Selected: "\U000025B6 Match: {{ .Label | cyan }}",
			Details: `
{{ if .FullDetails }}
{{ .FullDetails }}
{{ end }}`,
		}

		// Initial Fetch
		for {
			opts.Limit = limit
			opts.Offset = offset

			if moreAvailable {
				fmt.Printf("Fetching results (Offset: %d)...\n", offset)
				results, err := search.Search(query, scope, opts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
					os.Exit(1)
				}

				if len(results) == 0 && offset == 0 {
					fmt.Println("No matches found.")
					return
				}

				if len(results) < limit {
					moreAvailable = false
				} else {
					moreAvailable = true
				}

				if len(results) > 0 {
					// Transform results to items and append
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

						displayTime := match.Session.Metadata.Timestamp
						if displayTime == "" {
							displayTime = match.Session.SortKey.Format("2006-01-02 15:04:05")
						}

						// Build Full Boxed Details
						// We need a fixed width for the box content to align right border
						// Using safeWidth as the outer width
						boxInnerWidth := safeWidth - 4 // │_..._│
						if boxInnerWidth < 10 {
							boxInnerWidth = 10
						}

						makeLine := func(label, value string) string {
							// "Match Details" box style
							// │ Session:    Value...   │
							// Label takes e.g. 12 chars
							// Value takes rest

							// Visual width check
							labelDecor := label + ":"
							labelWidth := runewidth.StringWidth(labelDecor)
							targetLabelWidth := 12

							paddedLabel := labelDecor
							if labelWidth < targetLabelWidth {
								paddedLabel += strings.Repeat(" ", targetLabelWidth-labelWidth)
							}

							// Full line content: "Label:      Value"
							contentLimit := boxInnerWidth - targetLabelWidth - 1 // -1 for space

							valWidth := runewidth.StringWidth(value)
							displayValue := value
							if valWidth > contentLimit {
								displayValue = runewidth.Truncate(value, contentLimit, "...")
							}

							fullContent := fmt.Sprintf("%s %s", paddedLabel, displayValue)
							return "│ " + runewidth.FillRight(fullContent, boxInnerWidth) + " │"
						}

						// Header
						// ┌─ Match Details ──────┐
						headerLabel := " Match Details "

						msgWidth := boxInnerWidth + 2 // The text area width including spaces inside borders

						topBorder := "┌─" + headerLabel + strings.Repeat("─", msgWidth-len(headerLabel)-1) + "┐"
						sepBorder := "├─ Context " + strings.Repeat("─", msgWidth-10) + "┤"
						botBorder := "└" + strings.Repeat("─", msgWidth) + "┘"

						fullDetails := topBorder + "\n"
						fullDetails += makeLine("Session", displaySession) + "\n"
						fullDetails += makeLine("Timestamp", displayTime) + "\n"
						fullDetails += makeLine("File", displayFile) + "\n"
						fullDetails += sepBorder + "\n"

						// Re-process context lines to fit box
						for _, line := range contextLines {
							cleanL := strings.ReplaceAll(line, "\t", "    ")
							// Use runewidth to truncate
							if runewidth.StringWidth(cleanL) > boxInnerWidth {
								cleanL = runewidth.Truncate(cleanL, boxInnerWidth, "...")
							}
							fullDetails += "│ " + runewidth.FillRight(cleanL, boxInnerWidth) + " │\n"
						}
						fullDetails += botBorder

						if match.IsNote {
							label = fmt.Sprintf("[%d] %s [NOTE]: %s", match.Session.ID, match.Session.DisplayPath, content)
						} else {
							label = fmt.Sprintf("[%d] %s:%d: %s", match.Session.ID, match.Session.DisplayPath, match.LineNum, content)
						}

						if len(label) > maxLabelWidth {
							label = utils.TruncateString(label, maxLabelWidth-3) + "..."
						}

						allContentItems = append(allContentItems, item{
							Label:          label,
							CleanContent:   content,
							CleanContext:   cleanContext,
							DisplaySession: displaySession,
							DisplayFile:    displayFile,
							DisplayTime:    displayTime,
							Match:          match,
							FullDetails:    fullDetails,
						})
					}
				} else if offset > 0 {
					fmt.Println("No more results.")
					moreAvailable = false
				}
			}

			// Build Display List
			var displayItems []item
			displayItems = append(displayItems, allContentItems...)

			if moreAvailable {
				displayItems = append(displayItems, item{
					Label:     "--- Load More Results ---",
					IsControl: true,
				})
			}

			displayItems = append(displayItems, item{
				Label:     "Exit Search",
				IsControl: true,
			})

			// Interaction Loop
			prompt := promptui.Select{
				Label:     fmt.Sprintf("Showing %d matches. Select to view context (Esc/Ctrl+C to exit):", len(allContentItems)),
				Items:     displayItems,
				Templates: templates,
				Size:      10,
				CursorPos: len(allContentItems) - limit, // Try to position cursor at start of new batch
				Searcher: func(input string, index int) bool {
					return strings.Contains(strings.ToLower(displayItems[index].Label), strings.ToLower(input))
				},
			}

			// Safety check for cursor pos
			if prompt.CursorPos < 0 {
				prompt.CursorPos = 0
			}

			i, _, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					break // Exit entire search
				}
				continue
			}

			selectedItem := displayItems[i]

			if selectedItem.Label == "Exit Search" {
				return
			}

			if selectedItem.Label == "--- Load More Results ---" {
				offset += limit
				continue // Loop back to fetch
			}

			// It's a match item
			viewInPager(selectedItem.Match)
			// Loop back to show list again (no fetch needed unless we add logic to optimized usage)
		}
	},
}

func parseDate(d string) (time.Time, error) {
	formats := []string{
		"02012006",
		"02-01-2006",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, d); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown date format")
}

func viewInPager(m search.Match) {
	line := m.LineNum

	var args []string

	args = append(args, fmt.Sprintf("+%dG", line), "-j.5", "-N")

	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	var cmd *exec.Cmd

	if strings.Contains(pager, "less") {
		finalArgs := []string{"-R"}
		finalArgs = append(finalArgs, args...)
		cmd = exec.Command("less", finalArgs...)
	} else {
		cmd = exec.Command(pager, args...)
	}

	r, w := io.Pipe()
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		defer w.Close()
		f, err := os.Open(m.Session.Path)
		if err != nil {
			fmt.Fprintf(w, "Error opening file: %v\n", err)
			return
		}
		defer f.Close()

		var r io.Reader = f
		if strings.HasSuffix(m.Session.Path, ".tty") {
			r = logs.NewTtyReader(f)
		}

		cleaner := utils.NewCleanReader(r)
		if _, err := io.Copy(w, cleaner); err != nil {
			return
		}
	}()

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error opening pager: %v\n", err)
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&flagAfter, "after", "a", "", "Filter logs after date (DDMMYYYY)")
	searchCmd.Flags().StringVarP(&flagBefore, "before", "b", "", "Filter logs before date (DDMMYYYY)")
	searchCmd.Flags().BoolVarP(&flagRegex, "regex", "r", false, "Treat query as regex (default: boolean)")
}
