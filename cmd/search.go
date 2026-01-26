package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/search"
	"pentlog/pkg/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	flagAfter  string
	flagBefore string
	flagRegex  bool
)

type searchModel struct {
	textInput        textinput.Model
	results          []search.Match
	filteredSessions []logs.Session
	searchOpts       search.SearchOptions
	cursor           int
	scrollOffset     int
	loading          bool
	err              error
	debounceTimer    *time.Timer
	lastQuery        string
	selectedMatch    *search.Match
	styles           struct {
		header   lipgloss.Style
		help     lipgloss.Style
		selected lipgloss.Style
		normal   lipgloss.Style
	}
}

type tickMsg time.Time

type searchResultsMsg struct {
	results []search.Match
	err     error
}

func newSearchModel(scope []logs.Session, opts search.SearchOptions) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	m := searchModel{
		textInput:        ti,
		results:          []search.Match{},
		filteredSessions: scope,
		searchOpts:       opts,
		cursor:           0,
		loading:          false,
	}

	m.styles.header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00"))

	m.styles.help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true)

	m.styles.selected = lipgloss.NewStyle().
		Background(lipgloss.Color("#0066cc")).
		Foreground(lipgloss.Color("#ffffff"))

	m.styles.normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc"))

	return m
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search command history across all sessions (supports Regex)",
	Run: func(cmd *cobra.Command, args []string) {
		var scope []logs.Session

		allSessions, err := logs.ListSessions()
		if err != nil {
			errors.DatabaseErr("list sessions", err).Fatal()
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

		opts := search.SearchOptions{
			IsRegex: flagRegex,
		}

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
			opts.Limit = 50
		}

		model := newSearchModel(scope, opts)
		if len(args) > 0 {
			initialQuery := strings.Join(args, " ")
			model.textInput.SetValue(initialQuery)
			model = model.handleSearch()
		}

		p := tea.NewProgram(model)
		finalModel, err := p.Run()
		if err != nil {
			errors.FromError(errors.Generic, "Error running search UI", err).Fatal()
		}

		if m, ok := finalModel.(searchModel); ok {
			if m.err != nil {
				errors.FromError(errors.Generic, "Search error", m.err).Fatal()
			}
			if m.selectedMatch != nil {
				viewInPager(*m.selectedMatch)
			}
		}
	},
}

func (m searchModel) Init() tea.Cmd {
	return nil
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.results) > 0 && m.cursor >= 0 && m.cursor < len(m.results) {
				m.selectedMatch = &m.results[m.cursor]
				return m, tea.Quit
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				(&m).updateScrollOffset()
			}

		case tea.KeyDown:
			if m.cursor < len(m.results)-1 {
				m.cursor++
				(&m).updateScrollOffset()
			}

		case tea.KeyHome:
			m.cursor = 0
			m.scrollOffset = 0

		case tea.KeyEnd:
			if len(m.results) > 0 {
				m.cursor = len(m.results) - 1
				(&m).updateScrollOffset()
			}
		}

		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if m.debounceTimer != nil {
			m.debounceTimer.Stop()
		}

		m.debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
		})

		query := m.textInput.Value()
		if query != m.lastQuery {
			m.lastQuery = query
			m.cursor = 0
			if query == "" {
				m.results = []search.Match{}
				return m, nil
			}
			return m, m.searchCmd(query)
		}

		return m, cmd

	case searchResultsMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.results = msg.results
		m.loading = false
		m.cursor = 0
		m.scrollOffset = 0
		if m.cursor >= len(m.results) && len(m.results) > 0 {
			m.cursor = len(m.results) - 1
		}
		return m, nil
	}

	return m, nil
}

func (m searchModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	width := utils.GetTerminalWidth()
	if width < 40 {
		width = 40
	}

	searchBox := fmt.Sprintf("Query: %s", m.textInput.View())

	statusMsg := ""
	if m.loading {
		statusMsg = "(searching...)"
	} else {
		if len(m.results) > 0 {
			statusMsg = fmt.Sprintf("Found %d matches | Result %d/%d", len(m.results), m.cursor+1, len(m.results))
		} else {
			statusMsg = fmt.Sprintf("Found %d matches", len(m.results))
		}
	}

	var resultLines []string
	const viewportSize = 10

	endIdx := m.scrollOffset + viewportSize
	if endIdx > len(m.results) {
		endIdx = len(m.results)
	}

	if m.scrollOffset > 0 {
		resultLines = append(resultLines, m.styles.help.Render("  ▲ ... (scroll up)"))
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		match := m.results[i]
		content := utils.StripANSI(match.Content)
		if len(content) > width-10 {
			content = content[:width-10] + "..."
		}

		label := ""
		if match.IsNote {
			label = fmt.Sprintf("[NOTE] %s: %s", match.Session.DisplayPath, content)
		} else {
			label = fmt.Sprintf("[%d] %s: %s", match.LineNum, match.Session.DisplayPath, content)
		}

		if i == m.cursor {
			resultLines = append(resultLines, m.styles.selected.Render("  > "+label))
		} else {
			resultLines = append(resultLines, "    "+label)
		}
	}

	if endIdx < len(m.results) {
		resultLines = append(resultLines, m.styles.help.Render("  ▼ ... (scroll down)"))
	}

	resultView := strings.Join(resultLines, "\n")
	if resultView == "" && m.lastQuery != "" && !m.loading {
		resultView = "  (no matches)"
	}

	helpText := m.styles.help.Render("↑↓: navigate  Enter: open  ESC/Ctrl+C: quit | Home/End: jump")

	output := fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		searchBox,
		statusMsg,
		resultView,
		helpText,
	)

	return output
}

func (m *searchModel) updateScrollOffset() {
	const viewportSize = 10

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}

	if m.cursor >= m.scrollOffset+viewportSize {
		m.scrollOffset = m.cursor - viewportSize + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	maxScroll := len(m.results) - viewportSize
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

func (m searchModel) handleSearch() searchModel {
	query := m.textInput.Value()
	if query == "" {
		m.results = []search.Match{}
		return m
	}

	m.loading = true
	results, err := search.Search(query, m.filteredSessions, m.searchOpts)
	m.loading = false

	if err != nil {
		m.err = err
		return m
	}

	m.results = results
	m.cursor = 0
	return m
}

func (m searchModel) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			return searchResultsMsg{results: []search.Match{}, err: nil}
		}

		results, err := search.Search(query, m.filteredSessions, m.searchOpts)
		return searchResultsMsg{results: results, err: err}
	}
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
