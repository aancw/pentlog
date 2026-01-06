package dashboard

import (
	"fmt"
	"pentlog/pkg/logs"
	"pentlog/pkg/vulns"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#AAAAAA"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	url = lipgloss.NewStyle().Foreground(special).Render

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("63")).
			Padding(0, 1).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtle).
			Padding(1)

	listHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			Render

	listItem = lipgloss.NewStyle().PaddingLeft(2).Render
)

type Stats struct {
	TotalSessions     int
	TotalSize         int64
	UniqueClients     int
	UniqueEngagements int
	TotalNotes        int
	RecentSessions    []logs.Session
	PhaseCounts       map[string]int
	EngagementCounts  map[string]int
	ClientSizes       map[string]int64
	EngagementSizes   map[string]int64
	RecentVulns       []vulns.Vuln
}

type Model struct {
	Stats    Stats
	Loaded   bool
	Err      error
	Quitting bool
}

func InitialModel() Model {
	return Model{
		Stats: Stats{},
	}
}

func (m Model) Init() tea.Cmd {
	return loadStats
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	case Stats:
		m.Stats = msg
		m.Loaded = true
	case error:
		m.Err = msg
		return m, tea.Quit
	}
	return m, nil
}

func loadStats() tea.Msg {
	sessions, err := logs.ListSessions()
	if err != nil {
		return err
	}

	stats := Stats{
		TotalSessions:    len(sessions),
		PhaseCounts:      make(map[string]int),
		EngagementCounts: make(map[string]int),
		ClientSizes:      make(map[string]int64),
		EngagementSizes:  make(map[string]int64),
	}

	clients := make(map[string]bool)
	engagements := make(map[string]bool)
	noteCount := 0

	reversedSessions := make([]logs.Session, len(sessions))
	for i, j := 0, len(sessions)-1; i < len(sessions); i, j = i+1, j-1 {
		reversedSessions[i] = sessions[j]
	}

	for _, s := range reversedSessions {
		stats.TotalSize += s.Size
		if s.Metadata.Client != "" {
			clients[s.Metadata.Client] = true
			stats.ClientSizes[s.Metadata.Client] += s.Size
		}
		if s.Metadata.Engagement != "" {
			engagements[s.Metadata.Engagement] = true
			stats.EngagementCounts[s.Metadata.Engagement]++
			stats.EngagementSizes[s.Metadata.Engagement] += s.Size
		}
		if s.Metadata.Phase != "" {
			stats.PhaseCounts[s.Metadata.Phase]++
		}

		if s.NotesPath != "" {
			notes, err := logs.ReadNotes(s.NotesPath)
			if err == nil {
				noteCount += len(notes)
			}
		}
	}

	stats.UniqueClients = len(clients)
	stats.UniqueEngagements = len(engagements)
	stats.TotalNotes = noteCount

	var allVulns []vulns.Vuln

	uniqueContexts := make(map[string]bool)
	for _, s := range sessions {
		key := s.Metadata.Client + "|" + s.Metadata.Engagement
		if !uniqueContexts[key] {
			uniqueContexts[key] = true
			mgr := vulns.NewManager(s.Metadata.Client, s.Metadata.Engagement)
			if vList, err := mgr.List(); err == nil {
				allVulns = append(allVulns, vList...)
			}
		}
	}

	sort.Slice(allVulns, func(i, j int) bool {
		return allVulns[i].CreatedAt.After(allVulns[j].CreatedAt)
	})

	maxCount := 5
	if len(allVulns) < 5 {
		maxCount = len(allVulns)
	}
	stats.RecentVulns = allVulns[:maxCount]

	countSession := 5
	if len(reversedSessions) < 5 {
		countSession = len(reversedSessions)
	}
	stats.RecentSessions = reversedSessions[:countSession]

	return stats
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if !m.Loaded {
		return "Loading..."
	}
	if m.Quitting {
		return ""
	}

	header := titleStyle.Render("Pentlog Dashboard")

	statBox := func(label string, value string) string {
		return infoStyle.Render(fmt.Sprintf("%s\n%s", label, lipgloss.NewStyle().Foreground(special).Bold(true).Render(value)))
	}

	statsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statBox("Total Sessions", fmt.Sprintf("%d", m.Stats.TotalSessions)),
		statBox("Total Notes", fmt.Sprintf("%d", m.Stats.TotalNotes)),
		statBox("Evidence Size", formatSize(m.Stats.TotalSize)),
		statBox("Clients", fmt.Sprintf("%d", m.Stats.UniqueClients)),
		statBox("Engagements", fmt.Sprintf("%d", m.Stats.UniqueEngagements)),
	)

	var phaseStrs []string
	phaseStrs = append(phaseStrs, listHeader("Phase Distribution"))

	var phases []string
	for p := range m.Stats.PhaseCounts {
		phases = append(phases, p)
	}
	sort.Strings(phases)

	for _, p := range phases {
		count := m.Stats.PhaseCounts[p]
		bar := strings.Repeat("█", count)
		line := fmt.Sprintf("%-12s %s %d", p, bar, count)
		phaseStrs = append(phaseStrs, listItem(line))
	}
	phaseSection := lipgloss.JoinVertical(lipgloss.Left, phaseStrs...)

	var activityStrs []string
	activityStrs = append(activityStrs, listHeader("Recent Sessions"))

	for _, s := range m.Stats.RecentSessions {
		line := fmt.Sprintf("[%d] %s / %s (%s)", s.ID, s.Metadata.Client, s.Metadata.Phase, s.ModTime)
		activityStrs = append(activityStrs, listItem(line))
	}
	activitySection := lipgloss.JoinVertical(lipgloss.Left, activityStrs...)

	var engStrs []string
	engStrs = append(engStrs, listHeader("Engagement Logs"))
	var engs []string
	for e := range m.Stats.EngagementCounts {
		engs = append(engs, e)
	}
	sort.Strings(engs)
	for _, e := range engs {
		size := formatSize(m.Stats.EngagementSizes[e])
		line := fmt.Sprintf("%-20s : %d logs (%s)", e, m.Stats.EngagementCounts[e], size)
		engStrs = append(engStrs, listItem(line))
	}
	engSection := lipgloss.JoinVertical(lipgloss.Left, engStrs...)

	var clientStrs []string
	clientStrs = append(clientStrs, listHeader("Client Data"))
	var clients []string
	for c := range m.Stats.ClientSizes {
		clients = append(clients, c)
	}
	sort.Strings(clients)
	for _, c := range clients {
		size := formatSize(m.Stats.ClientSizes[c])
		line := fmt.Sprintf("%-20s : %s", c, size)
		clientStrs = append(clientStrs, listItem(line))
	}
	clientSection := lipgloss.JoinVertical(lipgloss.Left, clientStrs...)

	var noteStrs []string
	noteStrs = append(noteStrs, listHeader("Recent Findings"))
	if len(m.Stats.RecentVulns) == 0 {
		noteStrs = append(noteStrs, listItem("No vulnerabilities found."))
	} else {
		for _, v := range m.Stats.RecentVulns {
			title := v.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}
			line := fmt.Sprintf("[%s] %s (%s)", v.Severity, title, v.Status)
			noteStrs = append(noteStrs, listItem(line))
		}
	}
	noteSection := lipgloss.JoinVertical(lipgloss.Left, noteStrs...)

	middleRow := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(4).Render(phaseSection),
		lipgloss.NewStyle().MarginRight(4).Render(clientSection),
		engSection,
	)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(4).Render(activitySection),
		noteSection,
	)

	footer := lipgloss.NewStyle().Foreground(subtle).Render("\nPress 'q' to quit.")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		statsRow,
		"\n",
		middleRow,
		"\n",
		bottomRow,
		footer,
	)

	boxedContent := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(1, 2).
		Render(content)

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		boxedContent,
	)

	return docStyle.Render(ui)
}
