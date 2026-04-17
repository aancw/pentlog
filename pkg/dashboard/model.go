package dashboard

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/vulns"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	highlight = lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#2DD4BF"}
	special   = lipgloss.AdaptiveColor{Light: "#14532D", Dark: "#86EFAC"}
	warnColor = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}
	danger    = lipgloss.AdaptiveColor{Light: "#991B1B", Dark: "#F87171"}

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#0F766E")).
			Padding(0, 1).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(subtle)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E2E8F0"))

	statLabelStyle = lipgloss.NewStyle().
			Foreground(subtle)

	statValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(special)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(lipgloss.Color("#0F766E"))

	keyBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECFEFF")).
			Background(lipgloss.Color("#155E75")).
			Padding(0, 1)
)

type dashboardDataMsg struct {
	data DashboardData
}

type focusArea int

const (
	focusActions focusArea = iota
	focusSessions
)

type Action struct {
	Key         string
	Label       string
	Description string
	Command     []string
}

type SeveritySummary struct {
	Total    int
	Open     int
	Verified int
	Closed   int
	Critical int
	High     int
	Medium   int
	Low      int
	Info     int
}

// Stats is retained for API compatibility with the web dashboard handlers.
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

type PhaseStat struct {
	Name  string
	Count int
}

type DashboardData struct {
	Context              *config.ContextData
	HasContext           bool
	TotalSessions        int
	CurrentScopeSessions int
	CurrentPhaseSessions int
	TotalNotes           int
	CurrentScopeNotes    int
	TotalSize            int64
	LastActivity         *logs.Session
	RecentSessions       []logs.Session
	PhaseStats           []PhaseStat
	RecentFindings       []vulns.Vuln
	VulnSummary          SeveritySummary
	Actions              []Action
}

type Model struct {
	data          DashboardData
	loaded        bool
	err           error
	quitting      bool
	helpVisible   bool
	width         int
	height        int
	focus         focusArea
	actionCursor  int
	sessionCursor int
	nextAction    []string
}

func InitialModel() Model {
	return Model{
		focus: focusActions,
	}
}

func (m Model) Init() tea.Cmd {
	return loadDashboardData
}

func (m Model) LaunchArgs() []string {
	return append([]string(nil), m.nextAction...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case dashboardDataMsg:
		m.data = msg.data
		m.loaded = true
		m.err = nil
		m.actionCursor = 0
		m.sessionCursor = 0
		return m, nil
	case error:
		m.loaded = false
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.helpVisible = !m.helpVisible
		return m, nil
	case "r":
		m.loaded = false
		m.err = nil
		return m, loadDashboardData
	}

	if args, ok := m.shortcutAction(msg.String()); ok {
		m.nextAction = args
		m.quitting = true
		return m, tea.Quit
	}

	if m.err != nil || !m.loaded {
		return m, nil
	}

	switch msg.String() {
	case "tab", "shift+tab", "right", "left", "l", "h":
		m.cycleFocus(msg.String())
		return m, nil
	case "down", "j":
		m.moveSelection(1)
		return m, nil
	case "up", "k":
		m.moveSelection(-1)
		return m, nil
	case "enter":
		if args, ok := m.selectedAction(); ok {
			m.nextAction = args
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) cycleFocus(key string) {
	if len(m.data.RecentSessions) == 0 {
		m.focus = focusActions
		return
	}

	switch key {
	case "shift+tab", "left", "h":
		if m.focus == focusActions {
			m.focus = focusSessions
			return
		}
		m.focus = focusActions
	default:
		if m.focus == focusSessions {
			m.focus = focusActions
			return
		}
		m.focus = focusSessions
	}
}

func (m *Model) moveSelection(delta int) {
	switch m.focus {
	case focusActions:
		if len(m.data.Actions) == 0 {
			return
		}
		m.actionCursor = clampIndex(m.actionCursor+delta, len(m.data.Actions))
	case focusSessions:
		if len(m.data.RecentSessions) == 0 {
			return
		}
		m.sessionCursor = clampIndex(m.sessionCursor+delta, len(m.data.RecentSessions))
	}
}

func (m Model) shortcutAction(key string) ([]string, bool) {
	switch key {
	case "c":
		return []string{"create"}, true
	case "h":
		if !m.loaded || m.focus == focusActions {
			return []string{"shell"}, true
		}
	case "s":
		return []string{"sessions", "list"}, true
	case "t":
		if len(m.data.RecentSessions) > 0 && m.focus == focusSessions {
			return []string{"timeline", fmt.Sprintf("%d", m.data.RecentSessions[m.sessionCursor].ID)}, true
		}
		return []string{"dashboard", "timeline"}, true
	case "/":
		return []string{"search"}, true
	case "v":
		return []string{"vuln", "list"}, true
	case "e":
		return []string{"export"}, true
	case "a":
		return []string{"export", "--analyze"}, true
	}

	return nil, false
}

func (m Model) selectedAction() ([]string, bool) {
	switch m.focus {
	case focusActions:
		if len(m.data.Actions) == 0 {
			return nil, false
		}
		return m.data.Actions[m.actionCursor].Command, true
	case focusSessions:
		if len(m.data.RecentSessions) == 0 {
			return []string{"dashboard", "timeline"}, true
		}
		return []string{"timeline", fmt.Sprintf("%d", m.data.RecentSessions[m.sessionCursor].ID)}, true
	default:
		return nil, false
	}
}

func loadDashboardData() tea.Msg {
	mgr := config.Manager()

	var ctx *config.ContextData
	if loadedCtx, err := mgr.LoadContext(); err == nil {
		ctx = loadedCtx
	}

	sessions, err := logs.ListSessions()
	if err != nil {
		return err
	}

	data := buildDashboardData(ctx, sessions)
	return dashboardDataMsg{data: data}
}

func buildDashboardData(ctx *config.ContextData, sessions []logs.Session) DashboardData {
	data := DashboardData{
		Context:       ctx,
		HasContext:    ctx != nil && ctx.Client != "",
		TotalSessions: len(sessions),
		Actions:       defaultActions(),
	}

	phaseCounts := make(map[string]int)

	for _, session := range sessions {
		data.TotalSize += session.Size
		if phase := strings.TrimSpace(session.Metadata.Phase); phase != "" {
			phaseCounts[phase]++
		}

		notes, err := logs.ReadNotes(session.NotesPath)
		if err == nil {
			data.TotalNotes += len(notes)
		}
	}

	if len(sessions) > 0 {
		data.LastActivity = cloneSession(sessions[0])
	}

	limit := 6
	if len(sessions) < limit {
		limit = len(sessions)
	}
	for i := 0; i < limit; i++ {
		data.RecentSessions = append(data.RecentSessions, sessions[i])
	}

	data.PhaseStats = summarizePhases(phaseCounts)

	if !data.HasContext {
		return data
	}

	currentScopeSessions := filterSessionsForContext(sessions, ctx, false)
	currentPhaseSessions := filterSessionsForContext(sessions, ctx, true)
	data.CurrentScopeSessions = len(currentScopeSessions)
	data.CurrentPhaseSessions = len(currentPhaseSessions)

	for _, session := range currentScopeSessions {
		notes, err := logs.ReadNotes(session.NotesPath)
		if err == nil {
			data.CurrentScopeNotes += len(notes)
		}
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	if findings, err := vmgr.List(); err == nil {
		data.VulnSummary = summarizeVulns(findings)
		maxFindings := 5
		if len(findings) < maxFindings {
			maxFindings = len(findings)
		}
		data.RecentFindings = append(data.RecentFindings, findings[:maxFindings]...)
	}

	return data
}

func defaultActions() []Action {
	return []Action{
		{Key: "h", Label: "Start recorded shell", Description: "Resume work in a captured shell session.", Command: []string{"shell"}},
		{Key: "s", Label: "Review sessions", Description: "Inspect recorded sessions and metadata.", Command: []string{"sessions", "list"}},
		{Key: "t", Label: "Open timeline", Description: "Browse engagement timeline or selected session.", Command: []string{"dashboard", "timeline"}},
		{Key: "/", Label: "Search evidence", Description: "Run full-text search with current-context defaults.", Command: []string{"search"}},
		{Key: "v", Label: "Review vulnerabilities", Description: "Inspect findings for the active engagement.", Command: []string{"vuln", "list"}},
		{Key: "e", Label: "Export report", Description: "Generate report for the active scope or choose manually.", Command: []string{"export"}},
		{Key: "a", Label: "Export + AI summary", Description: "Generate a report and include analysis in one pass.", Command: []string{"export", "--analyze"}},
		{Key: "c", Label: "Create context", Description: "Initialize or reset the active engagement context.", Command: []string{"create"}},
	}
}

func filterSessionsForContext(sessions []logs.Session, ctx *config.ContextData, includePhase bool) []logs.Session {
	if ctx == nil {
		return nil
	}

	var filtered []logs.Session
	for _, session := range sessions {
		if session.Metadata.Client != ctx.Client {
			continue
		}
		if session.Metadata.Engagement != ctx.Engagement {
			continue
		}
		if includePhase && !strings.EqualFold(strings.TrimSpace(session.Metadata.Phase), strings.TrimSpace(ctx.Phase)) {
			continue
		}
		filtered = append(filtered, session)
	}

	return filtered
}

func summarizePhases(phaseCounts map[string]int) []PhaseStat {
	stats := make([]PhaseStat, 0, len(phaseCounts))
	for name, count := range phaseCounts {
		stats = append(stats, PhaseStat{Name: name, Count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return strings.ToLower(stats[i].Name) < strings.ToLower(stats[j].Name)
		}
		return stats[i].Count > stats[j].Count
	})

	if len(stats) > 5 {
		stats = stats[:5]
	}

	return stats
}

func summarizeVulns(findings []vulns.Vuln) SeveritySummary {
	var summary SeveritySummary
	for _, finding := range findings {
		summary.Total++
		switch finding.Status {
		case vulns.StatusOpen:
			summary.Open++
		case vulns.StatusVerified:
			summary.Verified++
		case vulns.StatusClosed:
			summary.Closed++
		}

		switch finding.Severity {
		case vulns.SeverityCritical:
			summary.Critical++
		case vulns.SeverityHigh:
			summary.High++
		case vulns.SeverityMedium:
			summary.Medium++
		case vulns.SeverityLow:
			summary.Low++
		case vulns.SeverityInfo:
			summary.Info++
		}
	}

	return summary
}

func cloneSession(session logs.Session) *logs.Session {
	cloned := session
	return &cloned
}

func clampIndex(index, length int) int {
	if length == 0 {
		return 0
	}
	if index < 0 {
		return length - 1
	}
	if index >= length {
		return 0
	}
	return index
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

func formatLastActivity(session *logs.Session) string {
	if session == nil {
		return "No recorded activity"
	}

	if session.SortKey.IsZero() {
		return session.ModTime
	}

	now := time.Now()
	delta := now.Sub(session.SortKey)
	switch {
	case delta < time.Minute:
		return "Just now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	default:
		return session.SortKey.Format("2006-01-02 15:04")
	}
}

func formatContextSummary(ctx *config.ContextData) string {
	if ctx == nil || ctx.Client == "" {
		return "No active engagement context"
	}

	if ctx.Engagement == "" {
		return ctx.Client
	}

	return fmt.Sprintf("%s / %s", ctx.Client, ctx.Engagement)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return docStyle.Render(m.renderErrorState())
	}

	if !m.loaded {
		return docStyle.Render(m.renderLoadingState())
	}

	return docStyle.Render(m.renderDashboard())
}

func (m Model) renderLoadingState() string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("PentLog Dashboard"),
		"",
		panelStyle(false, 80).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				panelTitleStyle.Render("Loading"),
				"",
				"Collecting session summary, context, and recent findings...",
				subtitleStyle.Render("Press q to quit."),
			),
		),
	)

	return body
}

func (m Model) renderErrorState() string {
	box := panelStyle(true, 90).BorderForeground(danger).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			panelTitleStyle.Copy().Foreground(danger).Render("Dashboard failed to load"),
			"",
			m.err.Error(),
			"",
			subtitleStyle.Render("Press r to retry or q to quit."),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("PentLog Dashboard"),
		"",
		box,
	)
}

func (m Model) renderDashboard() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("PentLog Dashboard"),
		subtitleStyle.Render(fmt.Sprintf("Context: %s", formatContextSummary(m.data.Context))),
	)

	overview := m.renderOverviewPanel()
	actions := m.renderActionsPanel()
	sessions := m.renderSessionsPanel()
	findings := m.renderFindingsPanel()
	phases := m.renderPhasesPanel()

	mainWidth := 112
	if m.width > 0 && m.width-8 < mainWidth {
		mainWidth = m.width - 8
	}
	if mainWidth < 72 {
		mainWidth = 72
	}

	sideWidth := (mainWidth - 4) / 2
	if sideWidth < 34 {
		sideWidth = 34
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, actions, sessions)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, findings, phases)

	content := []string{header, "", overview, "", topRow, "", bottomRow, "", m.renderFooter()}
	if m.helpVisible {
		content = append(content, "", m.renderHelpPanel(mainWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m Model) renderOverviewPanel() string {
	contextState := "Not set"
	scopeSessions := "-"
	phaseSessions := "-"
	scopeNotes := "-"
	vulnState := "Set a context to see findings"

	if m.data.HasContext {
		contextState = fmt.Sprintf("%s | phase %s", formatContextSummary(m.data.Context), m.data.Context.Phase)
		scopeSessions = fmt.Sprintf("%d", m.data.CurrentScopeSessions)
		phaseSessions = fmt.Sprintf("%d", m.data.CurrentPhaseSessions)
		scopeNotes = fmt.Sprintf("%d", m.data.CurrentScopeNotes)
		vulnState = fmt.Sprintf("%d open / %d total", m.data.VulnSummary.Open, m.data.VulnSummary.Total)
	}

	stats := []string{
		renderStatBox("Active context", contextState),
		renderStatBox("Last activity", formatLastActivity(m.data.LastActivity)),
		renderStatBox("All sessions", fmt.Sprintf("%d", m.data.TotalSessions)),
		renderStatBox("Current scope", scopeSessions+" sessions"),
		renderStatBox("Current phase", phaseSessions+" sessions"),
		renderStatBox("Notes", fmt.Sprintf("%d total / %s scope", m.data.TotalNotes, scopeNotes)),
		renderStatBox("Evidence size", formatSize(m.data.TotalSize)),
		renderStatBox("Vulns", vulnState),
	}

	grid := joinStatGrid(stats, 4)
	return panelStyle(false, 112).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			panelTitleStyle.Render("Overview"),
			"",
			grid,
		),
	)
}

func (m Model) renderActionsPanel() string {
	lines := make([]string, 0, len(m.data.Actions)+2)
	lines = append(lines, panelTitleStyle.Render("Action Center"))
	lines = append(lines, subtitleStyle.Render("Enter opens the selected action. Direct hotkeys work anywhere."))
	lines = append(lines, "")

	for idx, action := range m.data.Actions {
		line := fmt.Sprintf("%s %s", keyBadgeStyle.Render(action.Key), action.Label)
		detail := subtitleStyle.Render("  " + action.Description)
		block := lipgloss.JoinVertical(lipgloss.Left, line, detail)
		if m.focus == focusActions && idx == m.actionCursor {
			block = selectedItemStyle.Padding(0, 1).Render(block)
		}
		lines = append(lines, block)
	}

	return panelStyle(m.focus == focusActions, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderSessionsPanel() string {
	lines := []string{
		panelTitleStyle.Render("Recent Sessions"),
		subtitleStyle.Render("Enter opens the selected session timeline."),
		"",
	}

	if len(m.data.RecentSessions) == 0 {
		empty := []string{
			"No recorded sessions yet.",
			subtitleStyle.Render("Start with `pentlog shell` after creating or confirming a context."),
		}
		return panelStyle(m.focus == focusSessions, 54).Render(lipgloss.JoinVertical(lipgloss.Left, append(lines, empty...)...))
	}

	for idx, session := range m.data.RecentSessions {
		line := fmt.Sprintf("[%d] %s / %s", session.ID, session.Metadata.Client, session.Metadata.Phase)
		detail := subtitleStyle.Render(fmt.Sprintf("  %s • %s", session.ModTime, session.DisplayPath))
		block := lipgloss.JoinVertical(lipgloss.Left, line, detail)
		if m.focus == focusSessions && idx == m.sessionCursor {
			block = selectedItemStyle.Padding(0, 1).Render(block)
		}
		lines = append(lines, block)
	}

	return panelStyle(m.focus == focusSessions, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderFindingsPanel() string {
	lines := []string{panelTitleStyle.Render("Finding Posture"), ""}

	if !m.data.HasContext {
		lines = append(lines, "No active context.")
		lines = append(lines, subtitleStyle.Render("Use `create` to set a client, engagement, and phase."))
		return panelStyle(false, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	summary := m.data.VulnSummary
	severity := fmt.Sprintf(
		"Open %d | Critical %d | High %d | Medium %d | Low %d | Info %d",
		summary.Open,
		summary.Critical,
		summary.High,
		summary.Medium,
		summary.Low,
		summary.Info,
	)
	lines = append(lines, severity)
	lines = append(lines, subtitleStyle.Render(fmt.Sprintf("Verified %d | Closed %d | Total %d", summary.Verified, summary.Closed, summary.Total)))
	lines = append(lines, "")

	if len(m.data.RecentFindings) == 0 {
		lines = append(lines, subtitleStyle.Render("No findings recorded for the active engagement."))
		return panelStyle(false, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	for _, finding := range m.data.RecentFindings {
		title := finding.Title
		if len(title) > 42 {
			title = title[:39] + "..."
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", finding.Severity, title))
		lines = append(lines, subtitleStyle.Render("  "+string(finding.Status)))
	}

	return panelStyle(false, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderPhasesPanel() string {
	lines := []string{panelTitleStyle.Render("Phase Distribution"), ""}

	if len(m.data.PhaseStats) == 0 {
		lines = append(lines, subtitleStyle.Render("No sessions recorded yet."))
		return panelStyle(false, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	maxCount := m.data.PhaseStats[0].Count
	if maxCount < 1 {
		maxCount = 1
	}

	for _, phase := range m.data.PhaseStats {
		barLen := int((float64(phase.Count) / float64(maxCount)) * 14)
		if barLen < 1 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)
		lines = append(lines, fmt.Sprintf("%-14s %s %d", phase.Name, bar, phase.Count))
	}

	return panelStyle(false, 54).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderFooter() string {
	focusLabel := "Actions"
	if m.focus == focusSessions && len(m.data.RecentSessions) > 0 {
		focusLabel = "Recent sessions"
	}

	return subtitleStyle.Render(
		fmt.Sprintf(
			"Focus: %s | Tab switch | ↑↓ move | Enter open | s sessions | t timeline | / search | v vulns | e export | a analyze | r refresh | ? help | q quit",
			focusLabel,
		),
	)
}

func (m Model) renderHelpPanel(width int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		panelTitleStyle.Render("Keyboard Help"),
		"",
		"Tab / Shift+Tab: switch focus between action center and recent sessions",
		"Enter: open selected action or session timeline",
		"r: refresh dashboard data",
		"c: create context",
		"h: start recorded shell",
		"s: open sessions list",
		"t: open engagement timeline or selected session timeline",
		"/: search evidence",
		"v: review vulnerabilities",
		"e: export report",
		"a: export with AI analysis",
		"q / Esc: quit dashboard",
	)

	return panelStyle(true, width).BorderForeground(warnColor).Render(content)
}

func renderStatBox(label, value string) string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(0, 1).
		Width(24).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				statLabelStyle.Render(label),
				statValueStyle.Render(value),
			),
		)
}

func joinStatGrid(items []string, columns int) string {
	if columns < 1 {
		columns = 1
	}

	var rows []string
	for start := 0; start < len(items); start += columns {
		end := start + columns
		if end > len(items) {
			end = len(items)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, items[start:end]...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func panelStyle(focused bool, width int) lipgloss.Style {
	borderColor := subtle
	if focused {
		borderColor = highlight
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	if width > 0 {
		style = style.Width(width)
	}

	return style
}
