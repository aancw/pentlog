package dashboard

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TimelineFilter struct {
	Client     string
	Engagement string
	All        bool
}

type TimelineItemType string

const (
	TimelineSession TimelineItemType = "session"
	TimelinePhase   TimelineItemType = "phase"
	TimelineNote    TimelineItemType = "note"
	TimelineVuln    TimelineItemType = "vuln"
	TimelineGroup   TimelineItemType = "group"
)

type TimelineItem struct {
	Time       time.Time
	Kind       TimelineItemType
	Label      string
	Detail     string
	Client     string
	Engagement string
	Phase      string
}

type timelineDataMsg struct {
	items []TimelineItem
	info  string
}

type TimelineModel struct {
	filter       TimelineFilter
	filterLabel  string
	items        []TimelineItem
	cursor       int
	scrollOffset int
	width        int
	height       int
	loaded       bool
	err          error
}

const (
	timelineBoxWidth  = 120
	timelineBoxHeight = 32
	timelineViewport  = 12
	detailMaxLines    = 4
)

func InitialTimelineModel(filter TimelineFilter) TimelineModel {
	return TimelineModel{
		filter: filter,
	}
}

func (m TimelineModel) Init() tea.Cmd {
	return loadTimelineCmd(m.filter)
}

func (m TimelineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.stepCursor(-1)
		case tea.KeyDown:
			m.stepCursor(1)
		case tea.KeyPgUp:
			m.jumpCursor(m.cursor-m.viewportSize(), -1)
		case tea.KeyPgDown:
			m.jumpCursor(m.cursor+m.viewportSize(), 1)
		case tea.KeyHome:
			m.jumpCursor(m.firstSelectable(), 1)
		case tea.KeyEnd:
			m.jumpCursor(m.lastSelectable(), -1)
		}

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "k", "up":
			m.stepCursor(-1)
		case "j", "down":
			m.stepCursor(1)
		case "pgup":
			m.jumpCursor(m.cursor-m.viewportSize(), -1)
		case "pgdown":
			m.jumpCursor(m.cursor+m.viewportSize(), 1)
		case "home":
			m.jumpCursor(m.firstSelectable(), 1)
		case "end":
			m.jumpCursor(m.lastSelectable(), -1)
		}
	case timelineDataMsg:
		m.items = msg.items
		m.filterLabel = msg.info
		m.loaded = true
		m.cursor = m.firstSelectable()
		m.scrollOffset = 0
		m.updateScrollOffset()
	case error:
		m.err = msg
		return m, tea.Quit
	}

	return m, nil
}

func (m TimelineModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if !m.loaded {
		return "Loading timeline..."
	}

	width := m.width
	if width <= 0 {
		width = utils.GetTerminalWidth()
	}

	boxWidth := timelineBoxWidth
	boxHeight := timelineBoxHeight
	if boxWidth < 60 {
		boxWidth = 60
	}
	if boxHeight < 20 {
		boxHeight = 20
	}
	if width > 0 && boxWidth > width-4 {
		boxWidth = width - 4
		if boxWidth < 60 {
			boxWidth = 60
		}
	}
	innerWidth := boxWidth - 4

	header := titleStyle.Render("Engagement Timeline")
	filterLine := lipgloss.NewStyle().Foreground(subtle).Render(fmt.Sprintf("Scope: %s", m.filterLabel))
	legend := lipgloss.NewStyle().Foreground(subtle).Render("[S] Session  [P] Phase  [N] Note  [V] Vuln")
	help := lipgloss.NewStyle().Foreground(subtle).Render("Move: j/k or ↑/↓  Page: PgUp/PgDn  Home/End  Quit: q")
	totalSelectable := countSelectable(m.items)
	position := selectablePosition(m.items, m.cursor)
	status := lipgloss.NewStyle().Foreground(subtle).Render(fmt.Sprintf("Items: %d  Position: %d/%d", totalSelectable, position, totalSelectable))

	var lines []string
	viewportSize := m.viewportSize()
	endIdx := m.scrollOffset + viewportSize
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	if m.scrollOffset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(subtle).Render("  ▲ ..."))
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		item := m.items[i]
		line := renderTimelineLine(item, innerWidth-2)
		if i == m.cursor && isSelectable(item) {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("62")).Render("  "+line))
		} else {
			lines = append(lines, "  "+line)
		}
	}

	if endIdx < len(m.items) {
		lines = append(lines, lipgloss.NewStyle().Foreground(subtle).Render("  ▼ ..."))
	}

	body := "(no timeline items found)"
	if len(lines) > 0 {
		body = strings.Join(lines, "\n")
	}

	detail := ""
	if len(m.items) > 0 {
		item := m.items[m.cursor]
		detail = renderDetail(item, innerWidth)
	}

	footer := lipgloss.NewStyle().Foreground(subtle).Render("Press q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		filterLine,
		legend,
		help,
		status,
		"",
		body,
		"",
		detail,
		footer,
	)

	boxed := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Width(boxWidth).
		Height(boxHeight).
		Padding(1, 2).
		Render(content)

	return docStyle.Render(boxed)
}

func (m *TimelineModel) updateScrollOffset() {
	viewportSize := m.viewportSize()
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+viewportSize {
		m.scrollOffset = m.cursor - viewportSize + 1
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	maxScroll := len(m.items) - viewportSize
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

func (m TimelineModel) viewportSize() int {
	return timelineViewport
}

func (m *TimelineModel) stepCursor(dir int) {
	if len(m.items) == 0 {
		return
	}
	idx := m.cursor
	for {
		idx += dir
		if idx < 0 || idx >= len(m.items) {
			return
		}
		if isSelectable(m.items[idx]) {
			m.cursor = idx
			m.updateScrollOffset()
			return
		}
	}
}

func (m *TimelineModel) jumpCursor(target int, dir int) {
	if len(m.items) == 0 {
		return
	}
	if target < 0 {
		target = 0
	}
	if target >= len(m.items) {
		target = len(m.items) - 1
	}
	if dir == 0 {
		dir = 1
	}
	for i := target; i >= 0 && i < len(m.items); i += dir {
		if isSelectable(m.items[i]) {
			m.cursor = i
			m.updateScrollOffset()
			return
		}
	}
}

func (m TimelineModel) firstSelectable() int {
	for i, item := range m.items {
		if isSelectable(item) {
			return i
		}
	}
	return 0
}

func (m TimelineModel) lastSelectable() int {
	for i := len(m.items) - 1; i >= 0; i-- {
		if isSelectable(m.items[i]) {
			return i
		}
	}
	return 0
}

func renderTimelineLine(item TimelineItem, width int) string {
	if item.Kind == TimelineGroup {
		title := item.Label
		if title == "" {
			title = "Client"
		}
		line := fmt.Sprintf("=== %s ===", title)
		return truncate(line, width)
	}
	timeStr := formatTime(item.Time)
	kindLabel, kindStyle := itemKindStyle(item.Kind)
	marker := kindStyle.Render("●")
	label := fmt.Sprintf("%s %s %s", timeStr, marker, kindLabel+" "+item.Label)
	return truncate(label, width)
}

func renderDetail(item TimelineItem, width int) string {
	if width <= 0 {
		width = 80
	}

	labelStyle := lipgloss.NewStyle().Foreground(subtle)
	valueStyle := lipgloss.NewStyle().Foreground(subtle)

	detail := item.Detail
	if item.Kind == TimelineGroup {
		detail = "Client group"
	}
	if strings.TrimSpace(detail) == "" {
		detail = "-"
	}

	context := "-"
	if item.Client != "" || item.Engagement != "" {
		context = strings.TrimSpace(fmt.Sprintf("%s / %s", item.Client, item.Engagement))
		context = strings.Trim(context, " /")
		if context == "" {
			context = "-"
		}
	}

	phase := item.Phase
	if strings.TrimSpace(phase) == "" {
		phase = "-"
	}

	lines := []string{}
	lines = append(lines, labelStyle.Render("details:"))
	lines = append(lines, indentWrapped(detail, width, valueStyle)...)
	lines = append(lines, labelStyle.Render("context:"))
	lines = append(lines, indentWrapped(context, width, valueStyle)...)
	lines = append(lines, labelStyle.Render("phase:"))
	lines = append(lines, indentWrapped(phase, width, valueStyle)...)

	return strings.Join(lines, "\n")
}

func itemKindStyle(kind TimelineItemType) (string, lipgloss.Style) {
	switch kind {
	case TimelineSession:
		return "[S]", lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	case TimelinePhase:
		return "[P]", lipgloss.NewStyle().Foreground(lipgloss.Color("171"))
	case TimelineNote:
		return "[N]", lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case TimelineVuln:
		return "[V]", lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	case TimelineGroup:
		return "[C]", lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	default:
		return "[?]", lipgloss.NewStyle().Foreground(subtle)
	}
}

func truncate(s string, max int) string {
	if max <= 0 {
		return s
	}
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)
	if max < 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

func indentWrapped(value string, width int, style lipgloss.Style) []string {
	if width < 4 {
		return []string{style.Render(value)}
	}
	parts := wrapText(value, width-2)
	if len(parts) > detailMaxLines {
		parts = parts[:detailMaxLines]
		last := strings.TrimRight(parts[len(parts)-1], " .")
		parts[len(parts)-1] = last + "..."
	}
	lines := make([]string, 0, len(parts))
	for _, line := range parts {
		lines = append(lines, "  "+style.Render(line))
	}
	if len(lines) == 0 {
		lines = append(lines, "  "+style.Render("-"))
	}
	return lines
}

func wrapText(value string, width int) []string {
	if width <= 0 {
		return []string{value}
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	var line strings.Builder
	lineLen := 0

	flush := func() {
		if lineLen > 0 {
			lines = append(lines, line.String())
			line.Reset()
			lineLen = 0
		}
	}

	for _, word := range words {
		wordRunes := []rune(word)
		wordLen := len(wordRunes)

		if lineLen == 0 {
			if wordLen <= width {
				line.WriteString(word)
				lineLen = wordLen
				continue
			}
		}

		if lineLen > 0 && lineLen+1+wordLen <= width {
			line.WriteByte(' ')
			line.WriteString(word)
			lineLen += 1 + wordLen
			continue
		}

		flush()

		if wordLen <= width {
			line.WriteString(word)
			lineLen = wordLen
			continue
		}

		for len(wordRunes) > 0 {
			chunkSize := width
			if len(wordRunes) < chunkSize {
				chunkSize = len(wordRunes)
			}
			lines = append(lines, string(wordRunes[:chunkSize]))
			wordRunes = wordRunes[chunkSize:]
		}
	}

	flush()
	return lines
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}

func clampPos(pos, total int) int {
	if total == 0 {
		return 0
	}
	if pos < 1 {
		return 1
	}
	if pos > total {
		return total
	}
	return pos
}

func isSelectable(item TimelineItem) bool {
	return item.Kind != TimelineGroup
}

func countSelectable(items []TimelineItem) int {
	count := 0
	for _, item := range items {
		if isSelectable(item) {
			count++
		}
	}
	return count
}

func selectablePosition(items []TimelineItem, cursor int) int {
	if cursor < 0 || cursor >= len(items) {
		return 0
	}
	count := 0
	for i, item := range items {
		if isSelectable(item) {
			count++
		}
		if i == cursor {
			break
		}
	}
	if count == 0 {
		return 0
	}
	return count
}

func loadTimelineCmd(filter TimelineFilter) tea.Cmd {
	return func() tea.Msg {
		items, label, err := buildTimelineItems(filter)
		if err != nil {
			return err
		}
		return timelineDataMsg{items: items, info: label}
	}
}

func buildTimelineItems(filter TimelineFilter) ([]TimelineItem, string, error) {
	sessions, err := logs.ListSessions()
	if err != nil {
		return nil, "", err
	}

	label := buildFilterLabel(filter)
	items := []TimelineItem{}
	contexts := make(map[string]struct{})

	if !filter.All && filter.Client != "" {
		key := filter.Client + "|" + filter.Engagement
		contexts[key] = struct{}{}
	}

	for _, s := range sessions {
		if !matchFilter(filter, s.Metadata.Client, s.Metadata.Engagement) {
			continue
		}

		timestamp := parseSessionTime(s)
		items = append(items, TimelineItem{
			Time:       timestamp,
			Kind:       TimelineSession,
			Label:      fmt.Sprintf("Session #%d (%s)", s.ID, s.Metadata.Phase),
			Detail:     fmt.Sprintf("%s", s.DisplayPath),
			Client:     s.Metadata.Client,
			Engagement: s.Metadata.Engagement,
			Phase:      s.Metadata.Phase,
		})

		key := s.Metadata.Client + "|" + s.Metadata.Engagement
		contexts[key] = struct{}{}

		if s.NotesPath != "" {
			notes, err := logs.ReadNotes(s.NotesPath)
			if err == nil {
				for idx, n := range notes {
					noteTime := parseNoteTime(timestamp, n, idx)
					items = append(items, TimelineItem{
						Time:       noteTime,
						Kind:       TimelineNote,
						Label:      summarizeNote(n.Content, 60),
						Detail:     n.Content,
						Client:     s.Metadata.Client,
						Engagement: s.Metadata.Engagement,
						Phase:      s.Metadata.Phase,
					})
				}
			}
		}
	}

	// Phase history from context history
	mgr := config.Manager()
	if history, err := mgr.LoadContextHistory(); err == nil {
		for _, ctx := range history {
			if !matchFilter(filter, ctx.Client, ctx.Engagement) {
				continue
			}
			if ctx.Phase == "" {
				continue
			}

			ts, err := time.Parse(time.RFC3339, ctx.Timestamp)
			if err != nil {
				continue
			}

			items = append(items, TimelineItem{
				Time:       ts,
				Kind:       TimelinePhase,
				Label:      fmt.Sprintf("Phase set to %s", ctx.Phase),
				Detail:     fmt.Sprintf("Context type: %s", ctx.Type),
				Client:     ctx.Client,
				Engagement: ctx.Engagement,
				Phase:      ctx.Phase,
			})
		}
	}

	// Vulns
	for key := range contexts {
		parts := strings.SplitN(key, "|", 2)
		client := parts[0]
		engagement := ""
		if len(parts) > 1 {
			engagement = parts[1]
		}
		if client == "" {
			continue
		}

		vmgr := vulns.NewManager(client, engagement)
		vlist, err := vmgr.List()
		if err != nil {
			continue
		}
		for _, v := range vlist {
			items = append(items, TimelineItem{
				Time:       v.CreatedAt,
				Kind:       TimelineVuln,
				Label:      fmt.Sprintf("%s: %s", v.Severity, summarizeNote(v.Title, 50)),
				Detail:     fmt.Sprintf("Status: %s", v.Status),
				Client:     client,
				Engagement: engagement,
				Phase:      v.Phase,
			})
		}
	}

	if filter.All {
		sort.SliceStable(items, func(i, j int) bool {
			ci := clientKey(items[i])
			cj := clientKey(items[j])
			if ci != cj {
				return ci < cj
			}
			return compareTimeline(items[i], items[j])
		})

		grouped := []TimelineItem{}
		lastClient := ""
		for _, item := range items {
			client := clientKey(item)
			if client != lastClient {
				grouped = append(grouped, TimelineItem{
					Kind:   TimelineGroup,
					Label:  fmt.Sprintf("Client: %s", client),
					Client: client,
				})
				lastClient = client
			}
			grouped = append(grouped, item)
		}
		items = grouped
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			return compareTimeline(items[i], items[j])
		})
	}

	return items, label, nil
}

func compareTimeline(a, b TimelineItem) bool {
	ta := a.Time
	tb := b.Time
	if ta.IsZero() && tb.IsZero() {
		return a.Kind < b.Kind
	}
	if ta.IsZero() {
		return false
	}
	if tb.IsZero() {
		return true
	}
	if ta.Equal(tb) {
		return a.Kind < b.Kind
	}
	return ta.Before(tb)
}

func clientKey(item TimelineItem) string {
	client := strings.TrimSpace(item.Client)
	if client == "" {
		return "Unknown"
	}
	return client
}

func parseSessionTime(s logs.Session) time.Time {
	if s.Metadata.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339, s.Metadata.Timestamp); err == nil {
			return ts
		}
	}
	if !s.SortKey.IsZero() {
		return s.SortKey
	}
	return time.Time{}
}

func parseNoteTime(sessionTime time.Time, note logs.SessionNote, idx int) time.Time {
	if note.Timestamp == "" {
		return sessionTime
	}

	// Try full timestamp first
	if ts, err := time.Parse(time.RFC3339, note.Timestamp); err == nil {
		return ts
	}

	// Most notes are HH:MM:SS
	if sessionTime.IsZero() {
		return time.Time{}
	}

	parsed, err := time.Parse("15:04:05", note.Timestamp)
	if err != nil {
		return sessionTime.Add(time.Duration(idx) * time.Second)
	}

	return time.Date(sessionTime.Year(), sessionTime.Month(), sessionTime.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, sessionTime.Location())
}

func summarizeNote(content string, maxLen int) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	if maxLen <= 0 {
		return trimmed
	}
	if len([]rune(trimmed)) <= maxLen {
		return trimmed
	}
	r := []rune(trimmed)
	if maxLen < 3 {
		return string(r[:maxLen])
	}
	return string(r[:maxLen-3]) + "..."
}

func matchFilter(filter TimelineFilter, client, engagement string) bool {
	if filter.All {
		return true
	}
	if filter.Client != "" && client != filter.Client {
		return false
	}
	if filter.Engagement != "" && engagement != filter.Engagement {
		return false
	}
	return true
}

func buildFilterLabel(filter TimelineFilter) string {
	if filter.All {
		return "All engagements"
	}
	parts := []string{}
	if filter.Client != "" {
		parts = append(parts, filter.Client)
	}
	if filter.Engagement != "" {
		parts = append(parts, filter.Engagement)
	}
	if len(parts) == 0 {
		return "Current context"
	}
	return strings.Join(parts, " / ")
}
