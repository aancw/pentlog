package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

var outputFile string

var timelineCmd = &cobra.Command{
	Use:   "timeline [id]",
	Short: "Extract command timeline from a session",
	Long: `Analyze a terminal session recording and extract a chronological timeline
of commands executed and their outputs.

If no session ID is provided, an interactive session selector will be displayed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		var err error

		if len(args) > 0 {
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Invalid session ID: %s\n", args[0])
				os.Exit(1)
			}
		} else {
			sessions, err := logs.ListSessions()
			if err != nil {
				fmt.Printf("Error listing sessions: %v\n", err)
				os.Exit(1)
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			startIdx := 0
			if len(sessions) > 15 {
				startIdx = len(sessions) - 15
			}
			displaySessions := sessions[startIdx:]

			var items []string
			for _, s := range displaySessions {
				items = append(items, fmt.Sprintf("ID %d | %s | %s", s.ID, s.ModTime, s.DisplayPath))
			}

			fmt.Println("Recent Sessions:")
			idx := utils.SelectItem("Select Session to Analyze:", items)
			if idx == -1 {
				fmt.Println("Selection canceled.")
				return
			}
			id = displaySessions[idx].ID
		}

		session, err := logs.GetSession(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if session.Path == "" {
			fmt.Println("Error: session file missing; cannot analyze.")
			os.Exit(1)
		}

		spin := utils.NewSpinner(fmt.Sprintf("Analyzing session %d...", id))
		spin.Start()

		timeline, err := logs.ParseTimeline(session.Path)
		spin.Stop()

		if err != nil {
			fmt.Printf("Error parsing timeline: %v\n", err)
			os.Exit(1)
		}

		if len(timeline.Commands) == 0 {
			fmt.Println("No commands found in this session")
			return
		}

		if outputFile != "" {
			jsonOutput, err := timeline.ToJSON()
			if err != nil {
				fmt.Printf("Error generating JSON: %v\n", err)
				os.Exit(1)
			}

			err = os.WriteFile(outputFile, []byte(jsonOutput), 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Timeline saved to %s\n", outputFile)
			return
		}

		displayInteractiveTimeline(timeline, session)
	},
}

func init() {
	timelineCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path for JSON export")
	rootCmd.AddCommand(timelineCmd)
}

type timelineItem struct {
	Label       string
	Index       int
	Command     logs.CommandExecution
	FullDetails string
	IsControl   bool
}

func displayInteractiveTimeline(timeline *logs.Timeline, session *logs.Session) {
	width := utils.GetTerminalWidth()
	safeWidth := width - 2
	if safeWidth < 40 {
		safeWidth = 40
	}

	var items []timelineItem

	for i, cmd := range timeline.Commands {
		cmdPreview := cmd.Command
		if len(cmdPreview) > 60 {
			cmdPreview = cmdPreview[:57] + "..."
		}

		label := fmt.Sprintf("[%d] %s - %s", i+1, cmd.Timestamp, cmdPreview)

		details := buildCommandDetails(cmd, i+1, safeWidth)

		items = append(items, timelineItem{
			Label:       label,
			Index:       i,
			Command:     cmd,
			FullDetails: details,
			IsControl:   false,
		})
	}

	items = append(items, timelineItem{
		Label:     "Export Timeline as JSON",
		IsControl: true,
	})

	items = append(items, timelineItem{
		Label:     "Exit",
		IsControl: true,
	})

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000025B6 {{ .Label | cyan }}",
		Inactive: "  {{ .Label }}",
		Selected: "\U000025B6 Command: {{ .Label | cyan }}",
		Details: `
{{ if .FullDetails }}
{{ .FullDetails }}
{{ end }}`,
	}

	for {
		prompt := promptui.Select{
			Label:     fmt.Sprintf("Timeline: %d commands (Use arrow keys, / to search, Esc to exit)", len(timeline.Commands)),
			Items:     items,
			Templates: templates,
			Size:      15,
			Searcher: func(input string, index int) bool {
				return strings.Contains(strings.ToLower(items[index].Label), strings.ToLower(input))
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return
			}
			continue
		}

		selectedItem := items[idx]

		if selectedItem.Label == "Exit" {
			return
		}

		if selectedItem.Label == "Export Timeline as JSON" {
			exportTimeline(timeline)
			continue
		}

		showCommandActions(selectedItem.Command, session)
	}
}

func buildCommandDetails(cmd logs.CommandExecution, cmdNum int, width int) string {
	boxInnerWidth := width - 4
	if boxInnerWidth < 20 {
		boxInnerWidth = 20
	}

	makeLine := func(label, value string) string {
		labelDecor := label + ":"
		labelWidth := runewidth.StringWidth(labelDecor)
		targetLabelWidth := 12

		paddedLabel := labelDecor
		if labelWidth < targetLabelWidth {
			paddedLabel += strings.Repeat(" ", targetLabelWidth-labelWidth)
		}

		contentLimit := boxInnerWidth - targetLabelWidth - 1
		valWidth := runewidth.StringWidth(value)
		displayValue := value
		if valWidth > contentLimit {
			displayValue = runewidth.Truncate(value, contentLimit, "...")
		}

		fullContent := fmt.Sprintf("%s %s", paddedLabel, displayValue)
		return "│ " + runewidth.FillRight(fullContent, boxInnerWidth) + " │"
	}

	msgWidth := boxInnerWidth + 2
	headerLabel := fmt.Sprintf(" Command #%d ", cmdNum)
	topBorder := "┌─" + headerLabel + strings.Repeat("─", msgWidth-len(headerLabel)-1) + "┐"
	sepBorder := "├─ Output " + strings.Repeat("─", msgWidth-9) + "┤"
	botBorder := "└" + strings.Repeat("─", msgWidth) + "┘"

	var result strings.Builder
	result.WriteString(topBorder + "\n")
	result.WriteString(makeLine("Timestamp", cmd.Timestamp) + "\n")
	result.WriteString(makeLine("Command", cmd.Command) + "\n")
	result.WriteString(sepBorder + "\n")

	if cmd.Output != "" {
		outputLines := strings.Split(cmd.Output, "\n")
		displayLines := outputLines
		if len(displayLines) > 10 {
			displayLines = outputLines[:10]
		}

		for _, line := range displayLines {
			cleanL := strings.ReplaceAll(line, "\t", "    ")
			if runewidth.StringWidth(cleanL) > boxInnerWidth {
				cleanL = runewidth.Truncate(cleanL, boxInnerWidth, "...")
			}
			result.WriteString("│ " + runewidth.FillRight(cleanL, boxInnerWidth) + " │\n")
		}

		if len(outputLines) > 10 {
			moreMsg := fmt.Sprintf("... (%d more lines)", len(outputLines)-10)
			result.WriteString("│ " + runewidth.FillRight(moreMsg, boxInnerWidth) + " │\n")
		}
	} else {
		result.WriteString("│ " + runewidth.FillRight("(no output)", boxInnerWidth) + " │\n")
	}

	result.WriteString(botBorder)

	return result.String()
}

func showCommandActions(cmd logs.CommandExecution, session *logs.Session) {
	options := []string{
		"View full output in pager",
		"Back to command list",
	}

	choice := utils.SelectItem("What would you like to do?", options)
	if choice < 0 || choice == 1 {
		return
	}

	if choice == 0 {
		viewCommandInPager(cmd, session)
	}
}

func viewCommandInPager(cmd logs.CommandExecution, session *logs.Session) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	var execCmd *exec.Cmd
	if strings.Contains(pager, "less") {
		execCmd = exec.Command("less", "-R")
	} else {
		execCmd = exec.Command(pager)
	}

	r, w := io.Pipe()
	execCmd.Stdin = r
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	go func() {
		defer w.Close()
		fmt.Fprintf(w, "Command: %s\n", cmd.Command)
		fmt.Fprintf(w, "Timestamp: %s\n", cmd.Timestamp)
		fmt.Fprintf(w, "%s\n\n", strings.Repeat("=", 80))
		if cmd.Output != "" {
			fmt.Fprintf(w, "%s\n", cmd.Output)
		} else {
			fmt.Fprintf(w, "(no output)\n")
		}
	}()

	if err := execCmd.Run(); err != nil {
		fmt.Printf("Error opening pager: %v\n", err)
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func exportTimeline(timeline *logs.Timeline) {
	fmt.Print("Enter output filename (default: timeline.json): ")
	var filename string
	fmt.Scanln(&filename)

	if filename == "" {
		filename = "timeline.json"
	}

	jsonOutput, err := timeline.ToJSON()
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		return
	}

	err = os.WriteFile(filename, []byte(jsonOutput), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Timeline exported to %s\n\n", filename)
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}
