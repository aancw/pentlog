package logs

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	reporttpl "pentlog/pkg/templates"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"
	"sort"
	"strings"
	"time"
)

type ReportAppendixMode string

const (
	ReportAppendixModeCommands       ReportAppendixMode = "commands"
	ReportAppendixModeFullTranscript ReportAppendixMode = "full-transcript"
)

type ReportOptions struct {
	AppendixMode ReportAppendixMode
}

type reportSessionBundle struct {
	Session        Session
	Tags           []string
	Notes          []SessionNote
	Commands       []CommandExecution
	TranscriptText string
}

type reportSummary struct {
	SessionCount    int
	EngagementCount int
	PhaseCount      int
	FindingCount    int
	EvidenceCount   int
	CommandCount    int
	ArchiveRefCount int
	FirstObserved   string
	LastObserved    string
	CriticalCount   int
	HighCount       int
	MediumCount     int
	LowCount        int
	InfoCount       int
	OpenCount       int
	VerifiedCount   int
	ClosedCount     int
	Targets         []string
	Tags            []string
	Operators       []string
}

type reportEngagementMetadata struct {
	Name         string
	DateRange    string
	SessionCount int
	Operators    []string
	Targets      []string
	Tags         []string
	Phases       []reportPhaseMetadata
}

type reportPhaseMetadata struct {
	Name         string
	SessionCount int
	Targets      []string
	Tags         []string
}

type reportEvidenceSnippet struct {
	Title      string
	Source     string
	Summary    string
	Snippet    string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	SessionID  int
	FindingID  string
}

type reportCommandAppendix struct {
	SessionID  int
	Label      string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	Commands   []CommandExecution
}

type reportIntegrityReference struct {
	SessionID      int
	State          string
	Engagement     string
	Phase          string
	Target         string
	TranscriptPath string
	ArchivePath    string
	ManifestSHA256 string
	ArchivedAt     string
}

type reportTranscriptAppendix struct {
	SessionID  int
	Label      string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	Transcript string
}

type reportBundle struct {
	Client                string
	GeneratedAt           string
	Summary               reportSummary
	EngagementMetadata    []reportEngagementMetadata
	Findings              []vulns.Vuln
	EvidenceSnippets      []reportEvidenceSnippet
	CommandAppendix       []reportCommandAppendix
	IntegrityReferences   []reportIntegrityReference
	TranscriptAppendix    []reportTranscriptAppendix
	HasTranscriptAppendix bool
}

func DefaultReportOptions() ReportOptions {
	return ReportOptions{
		AppendixMode: ReportAppendixModeCommands,
	}
}

func (o ReportOptions) normalize() ReportOptions {
	switch o.AppendixMode {
	case "", ReportAppendixModeCommands:
		o.AppendixMode = ReportAppendixModeCommands
	case ReportAppendixModeFullTranscript:
	default:
		o.AppendixMode = ReportAppendixModeCommands
	}
	return o
}

func ExportCommands(client, engagement, phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	filtered := filterSessions(sessions, client, engagement, phase)
	if len(filtered) == 0 {
		return "", fmt.Errorf("no sessions found matching criteria")
	}

	return GenerateReport(filtered, client)
}

func filterSessions(sessions []Session, client, engagement, phase string) []Session {
	var filtered []Session
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
		filtered = append(filtered, s)
	}
	return filtered
}

func GenerateReport(sessions []Session, client string) (string, error) {
	return GenerateReportWithOptions(sessions, client, nil, "", DefaultReportOptions())
}

func GenerateReportWithOptions(sessions []Session, client string, findings []vulns.Vuln, aiAnalysis string, opts ReportOptions) (string, error) {
	bundle, err := buildReportBundle(sessions, client, findings, opts)
	if err != nil {
		return "", err
	}

	return renderMarkdownReport(bundle, aiAnalysis), nil
}

func ExportCommandsHTML(client, engagement, phase string) (string, error) {
	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}

	filtered := filterSessions(sessions, client, engagement, phase)
	if len(filtered) == 0 {
		return "", fmt.Errorf("no sessions found matching criteria")
	}

	return GenerateHTMLReport(filtered, client, nil, "", nil)
}

func ListClientReports(client string) ([]string, error) {
	mgr := config.Manager()
	reportDir := filepath.Join(mgr.GetPaths().ReportsDir, utils.Slugify(client))

	entries, err := os.ReadDir(reportDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var reports []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		switch strings.ToLower(filepath.Ext(name)) {
		case ".md", ".html":
			reports = append(reports, name)
		}
	}

	sort.Strings(reports)
	return reports, nil
}

func GenerateHTMLReport(sessions []Session, client string, findings []vulns.Vuln, aiAnalysis string, gifPaths map[int]string) (string, error) {
	return GenerateHTMLReportWithOptions(sessions, client, findings, aiAnalysis, gifPaths, DefaultReportOptions())
}

func GenerateHTMLReportWithOptions(sessions []Session, client string, findings []vulns.Vuln, aiAnalysis string, gifPaths map[int]string, opts ReportOptions) (string, error) {
	bundle, err := buildReportBundle(sessions, client, findings, opts)
	if err != nil {
		return "", err
	}

	reportData, err := buildHTMLTemplateData(bundle, aiAnalysis, gifPaths)
	if err != nil {
		return "", err
	}

	mgr := config.Manager()
	htmlPath := filepath.Join(mgr.GetPaths().TemplatesDir, "report.html")
	cssPath := filepath.Join(mgr.GetPaths().TemplatesDir, "report.css")

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s (run 'pentlog setup')", htmlPath)
	}
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		return "", fmt.Errorf("css file not found: %s (run 'pentlog setup')", cssPath)
	}

	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read html template: %w", err)
	}

	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		return "", fmt.Errorf("failed to read css file: %w", err)
	}

	reportData.CSS = template.CSS(cssContent)

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(string(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, reportData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func buildReportBundle(sessions []Session, client string, findings []vulns.Vuln, opts ReportOptions) (*reportBundle, error) {
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions to report")
	}

	opts = opts.normalize()
	sortedSessions := sortSessionsForReport(sessions)
	includeTranscript := opts.AppendixMode == ReportAppendixModeFullTranscript

	sessionBundles := make([]reportSessionBundle, 0, len(sortedSessions))
	for _, session := range sortedSessions {
		bundle, err := buildSessionBundle(session, includeTranscript)
		if err != nil {
			return nil, err
		}
		sessionBundles = append(sessionBundles, bundle)
	}

	if findings == nil {
		findings = collectFindingsForSessions(client, sortedSessions)
	}
	findings = sortFindings(findings)

	evidenceSnippets := buildEvidenceSnippets(sessionBundles, findings)
	commandAppendix := buildCommandAppendix(sessionBundles)
	integrityReferences := buildIntegrityReferences(sessionBundles)
	transcriptAppendix := buildTranscriptAppendix(sessionBundles, includeTranscript)

	bundle := &reportBundle{
		Client:                client,
		GeneratedAt:           time.Now().Format("2006-01-02 15:04:05 MST"),
		Findings:              findings,
		EvidenceSnippets:      evidenceSnippets,
		CommandAppendix:       commandAppendix,
		IntegrityReferences:   integrityReferences,
		TranscriptAppendix:    transcriptAppendix,
		HasTranscriptAppendix: includeTranscript,
	}

	bundle.EngagementMetadata = buildEngagementMetadata(sessionBundles)
	bundle.Summary = buildReportSummary(sessionBundles, findings, evidenceSnippets, commandAppendix, integrityReferences)

	return bundle, nil
}

func sortSessionsForReport(sessions []Session) []Session {
	sorted := append([]Session(nil), sessions...)
	sort.Slice(sorted, func(i, j int) bool {
		if !sorted[i].SortKey.IsZero() && !sorted[j].SortKey.IsZero() && !sorted[i].SortKey.Equal(sorted[j].SortKey) {
			return sorted[i].SortKey.Before(sorted[j].SortKey)
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func buildSessionBundle(session Session, includeTranscript bool) (reportSessionBundle, error) {
	tags, err := GetSessionTags(session.ID)
	if err != nil {
		return reportSessionBundle{}, err
	}

	notes, err := ReadNotes(session.NotesPath)
	if err != nil {
		return reportSessionBundle{}, err
	}

	commands, err := extractSessionCommands(session)
	if err != nil {
		return reportSessionBundle{}, err
	}

	bundle := reportSessionBundle{
		Session:  session,
		Tags:     tags,
		Notes:    notes,
		Commands: commands,
	}

	if includeTranscript {
		transcript, err := loadSessionTranscriptText(session.Path)
		if err != nil {
			return reportSessionBundle{}, err
		}
		bundle.TranscriptText = transcript
	}

	return bundle, nil
}

func collectFindingsForSessions(client string, sessions []Session) []vulns.Vuln {
	if len(sessions) == 0 {
		return nil
	}

	engagements := make(map[string]struct{})
	for _, session := range sessions {
		if session.Metadata.Engagement != "" {
			engagements[session.Metadata.Engagement] = struct{}{}
		}
	}

	managerEngagement := ""
	if len(engagements) == 1 {
		for engagement := range engagements {
			managerEngagement = engagement
		}
	}

	manager := vulns.NewManager(client, managerEngagement)
	findings, err := manager.List()
	if err != nil {
		return nil
	}

	phaseSet := make(map[string]struct{})
	for _, session := range sessions {
		phase := strings.TrimSpace(strings.ToLower(session.Metadata.Phase))
		if phase != "" {
			phaseSet[phase] = struct{}{}
		}
	}

	if len(phaseSet) == 0 {
		return findings
	}

	filtered := make([]vulns.Vuln, 0, len(findings))
	for _, finding := range findings {
		phase := strings.TrimSpace(strings.ToLower(finding.Phase))
		if phase == "" {
			filtered = append(filtered, finding)
			continue
		}
		if _, ok := phaseSet[phase]; ok {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}

func sortFindings(findings []vulns.Vuln) []vulns.Vuln {
	sorted := append([]vulns.Vuln(nil), findings...)
	severityRank := map[vulns.Severity]int{
		vulns.SeverityCritical: 0,
		vulns.SeverityHigh:     1,
		vulns.SeverityMedium:   2,
		vulns.SeverityLow:      3,
		vulns.SeverityInfo:     4,
	}
	sort.Slice(sorted, func(i, j int) bool {
		leftRank, ok := severityRank[sorted[i].Severity]
		if !ok {
			leftRank = len(severityRank)
		}
		rightRank, ok := severityRank[sorted[j].Severity]
		if !ok {
			rightRank = len(severityRank)
		}
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if !sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
			return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func buildEvidenceSnippets(sessions []reportSessionBundle, findings []vulns.Vuln) []reportEvidenceSnippet {
	snippets := make([]reportEvidenceSnippet, 0)

	for _, session := range sessions {
		for _, note := range session.Notes {
			snippet := extractNoteSnippet(session.Session.Path, note.ByteOffset)
			if snippet == "" {
				snippet = note.Content
			}

			snippets = append(snippets, reportEvidenceSnippet{
				Title:      fmt.Sprintf("Session %d note", session.Session.ID),
				Source:     "Session note",
				Summary:    note.Content,
				Snippet:    snippet,
				Timestamp:  firstNonEmpty(note.Timestamp, session.Session.ModTime),
				Engagement: session.Session.Metadata.Engagement,
				Phase:      session.Session.Metadata.Phase,
				Target:     formatSessionTarget(session.Session),
				Tags:       append([]string(nil), session.Tags...),
				SessionID:  session.Session.ID,
			})
		}
	}

	for _, finding := range findings {
		if len(finding.Evidence) == 0 {
			continue
		}
		for _, evidence := range finding.Evidence {
			evidence = strings.TrimSpace(evidence)
			if evidence == "" {
				continue
			}

			snippets = append(snippets, reportEvidenceSnippet{
				Title:     fmt.Sprintf("Finding evidence: %s", finding.Title),
				Source:    "Finding evidence",
				Summary:   shortenText(finding.Description, 160),
				Snippet:   evidence,
				Timestamp: formatFindingTimestamp(finding),
				Phase:     finding.Phase,
				FindingID: finding.ID,
			})
		}
	}

	return snippets
}

func buildCommandAppendix(sessions []reportSessionBundle) []reportCommandAppendix {
	appendix := make([]reportCommandAppendix, 0, len(sessions))
	for _, session := range sessions {
		if len(session.Commands) == 0 {
			continue
		}

		appendix = append(appendix, reportCommandAppendix{
			SessionID:  session.Session.ID,
			Label:      fmt.Sprintf("Session %d", session.Session.ID),
			Timestamp:  session.Session.ModTime,
			Engagement: session.Session.Metadata.Engagement,
			Phase:      session.Session.Metadata.Phase,
			Target:     formatSessionTarget(session.Session),
			Tags:       append([]string(nil), session.Tags...),
			Commands:   session.Commands,
		})
	}
	return appendix
}

func buildIntegrityReferences(sessions []reportSessionBundle) []reportIntegrityReference {
	references := make([]reportIntegrityReference, 0, len(sessions))
	for _, session := range sessions {
		references = append(references, reportIntegrityReference{
			SessionID:      session.Session.ID,
			State:          string(session.Session.State),
			Engagement:     session.Session.Metadata.Engagement,
			Phase:          session.Session.Metadata.Phase,
			Target:         formatSessionTarget(session.Session),
			TranscriptPath: session.Session.DisplayPath,
			ArchivePath:    session.Session.ArchivePath,
			ManifestSHA256: session.Session.ArchiveManifestSHA256,
			ArchivedAt:     session.Session.ArchivedAt,
		})
	}
	return references
}

func buildTranscriptAppendix(sessions []reportSessionBundle, includeTranscript bool) []reportTranscriptAppendix {
	if !includeTranscript {
		return nil
	}

	appendix := make([]reportTranscriptAppendix, 0, len(sessions))
	for _, session := range sessions {
		appendix = append(appendix, reportTranscriptAppendix{
			SessionID:  session.Session.ID,
			Label:      fmt.Sprintf("Session %d", session.Session.ID),
			Timestamp:  session.Session.ModTime,
			Engagement: session.Session.Metadata.Engagement,
			Phase:      session.Session.Metadata.Phase,
			Target:     formatSessionTarget(session.Session),
			Tags:       append([]string(nil), session.Tags...),
			Transcript: session.TranscriptText,
		})
	}
	return appendix
}

func buildEngagementMetadata(sessions []reportSessionBundle) []reportEngagementMetadata {
	type phaseAccumulator struct {
		sessionCount int
		targets      map[string]struct{}
		tags         map[string]struct{}
	}
	type engagementAccumulator struct {
		sessionCount int
		start        time.Time
		end          time.Time
		operators    map[string]struct{}
		targets      map[string]struct{}
		tags         map[string]struct{}
		phases       map[string]*phaseAccumulator
	}

	accumulators := make(map[string]*engagementAccumulator)
	for _, session := range sessions {
		name := firstNonEmpty(session.Session.Metadata.Engagement, "Unknown engagement")
		acc, ok := accumulators[name]
		if !ok {
			acc = &engagementAccumulator{
				operators: make(map[string]struct{}),
				targets:   make(map[string]struct{}),
				tags:      make(map[string]struct{}),
				phases:    make(map[string]*phaseAccumulator),
			}
			accumulators[name] = acc
		}

		acc.sessionCount++
		updateRange(&acc.start, &acc.end, session.Session.SortKey)
		if session.Session.Metadata.Operator != "" {
			acc.operators[session.Session.Metadata.Operator] = struct{}{}
		}
		if target := formatSessionTarget(session.Session); target != "" {
			acc.targets[target] = struct{}{}
		}
		for _, tag := range session.Tags {
			acc.tags[tag] = struct{}{}
		}

		phaseName := firstNonEmpty(session.Session.Metadata.Phase, "Unspecified")
		phaseAcc, ok := acc.phases[phaseName]
		if !ok {
			phaseAcc = &phaseAccumulator{
				targets: make(map[string]struct{}),
				tags:    make(map[string]struct{}),
			}
			acc.phases[phaseName] = phaseAcc
		}
		phaseAcc.sessionCount++
		if target := formatSessionTarget(session.Session); target != "" {
			phaseAcc.targets[target] = struct{}{}
		}
		for _, tag := range session.Tags {
			phaseAcc.tags[tag] = struct{}{}
		}
	}

	names := make([]string, 0, len(accumulators))
	for name := range accumulators {
		names = append(names, name)
	}
	sort.Strings(names)

	metadata := make([]reportEngagementMetadata, 0, len(names))
	for _, name := range names {
		acc := accumulators[name]
		phaseNames := make([]string, 0, len(acc.phases))
		for phase := range acc.phases {
			phaseNames = append(phaseNames, phase)
		}
		sort.Strings(phaseNames)

		phases := make([]reportPhaseMetadata, 0, len(phaseNames))
		for _, phaseName := range phaseNames {
			phaseAcc := acc.phases[phaseName]
			phases = append(phases, reportPhaseMetadata{
				Name:         phaseName,
				SessionCount: phaseAcc.sessionCount,
				Targets:      sortedKeys(phaseAcc.targets),
				Tags:         sortedKeys(phaseAcc.tags),
			})
		}

		metadata = append(metadata, reportEngagementMetadata{
			Name:         name,
			DateRange:    formatDateRange(acc.start, acc.end),
			SessionCount: acc.sessionCount,
			Operators:    sortedKeys(acc.operators),
			Targets:      sortedKeys(acc.targets),
			Tags:         sortedKeys(acc.tags),
			Phases:       phases,
		})
	}

	return metadata
}

func buildReportSummary(
	sessions []reportSessionBundle,
	findings []vulns.Vuln,
	evidenceSnippets []reportEvidenceSnippet,
	commandAppendix []reportCommandAppendix,
	integrityReferences []reportIntegrityReference,
) reportSummary {
	summary := reportSummary{
		SessionCount:    len(sessions),
		EngagementCount: len(uniqueValuesFromSessions(sessions, func(session reportSessionBundle) string { return session.Session.Metadata.Engagement })),
		PhaseCount:      len(uniqueValuesFromSessions(sessions, func(session reportSessionBundle) string { return session.Session.Metadata.Phase })),
		FindingCount:    len(findings),
		EvidenceCount:   len(evidenceSnippets),
		ArchiveRefCount: len(integrityReferences),
		Targets:         uniqueSortedSessionTargets(sessions),
		Tags:            uniqueSortedSessionTags(sessions),
		Operators:       uniqueSortedSessionOperators(sessions),
	}

	var first time.Time
	var last time.Time
	for _, session := range sessions {
		updateRange(&first, &last, session.Session.SortKey)
	}
	summary.FirstObserved = formatTimeOrDash(first)
	summary.LastObserved = formatTimeOrDash(last)

	for _, finding := range findings {
		switch finding.Severity {
		case vulns.SeverityCritical:
			summary.CriticalCount++
		case vulns.SeverityHigh:
			summary.HighCount++
		case vulns.SeverityMedium:
			summary.MediumCount++
		case vulns.SeverityLow:
			summary.LowCount++
		case vulns.SeverityInfo:
			summary.InfoCount++
		}

		switch finding.Status {
		case vulns.StatusOpen:
			summary.OpenCount++
		case vulns.StatusVerified:
			summary.VerifiedCount++
		case vulns.StatusClosed:
			summary.ClosedCount++
		}
	}

	for _, session := range commandAppendix {
		summary.CommandCount += len(session.Commands)
	}

	return summary
}

func renderMarkdownReport(bundle *reportBundle, aiAnalysis string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# PentLog Evidence Report: %s\n\n", bundle.Client))
	builder.WriteString(fmt.Sprintf("_Generated: %s_\n\n", bundle.GeneratedAt))

	builder.WriteString("## Executive Summary\n\n")
	builder.WriteString(fmt.Sprintf("- Scope covers %d engagement(s), %d phase(s), and %d recorded session(s).\n", bundle.Summary.EngagementCount, bundle.Summary.PhaseCount, bundle.Summary.SessionCount))
	builder.WriteString(fmt.Sprintf("- Observed activity window: %s to %s.\n", bundle.Summary.FirstObserved, bundle.Summary.LastObserved))
	builder.WriteString(fmt.Sprintf("- Findings captured: %d total (Critical: %d, High: %d, Medium: %d, Low: %d, Info: %d).\n", bundle.Summary.FindingCount, bundle.Summary.CriticalCount, bundle.Summary.HighCount, bundle.Summary.MediumCount, bundle.Summary.LowCount, bundle.Summary.InfoCount))
	builder.WriteString(fmt.Sprintf("- Curated evidence snippets: %d. Command appendix entries: %d. Integrity references: %d.\n", bundle.Summary.EvidenceCount, bundle.Summary.CommandCount, bundle.Summary.ArchiveRefCount))
	if len(bundle.Summary.Targets) > 0 {
		builder.WriteString(fmt.Sprintf("- Targets covered: %s.\n", strings.Join(bundle.Summary.Targets, ", ")))
	}
	if len(bundle.Summary.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("- Session tags in scope: %s.\n", strings.Join(bundle.Summary.Tags, ", ")))
	}
	if len(bundle.Summary.Operators) > 0 {
		builder.WriteString(fmt.Sprintf("- Operators observed: %s.\n", strings.Join(bundle.Summary.Operators, ", ")))
	}
	builder.WriteString("\n")

	if strings.TrimSpace(aiAnalysis) != "" {
		builder.WriteString("### AI Summary\n\n")
		builder.WriteString(strings.TrimSpace(aiAnalysis))
		builder.WriteString("\n\n")
	}

	builder.WriteString("## Engagement Metadata\n\n")
	for _, engagement := range bundle.EngagementMetadata {
		builder.WriteString(fmt.Sprintf("### %s\n\n", engagement.Name))
		builder.WriteString(fmt.Sprintf("- Sessions: %d\n", engagement.SessionCount))
		builder.WriteString(fmt.Sprintf("- Date range: %s\n", firstNonEmpty(engagement.DateRange, "-")))
		builder.WriteString(fmt.Sprintf("- Operators: %s\n", joinOrDash(engagement.Operators)))
		builder.WriteString(fmt.Sprintf("- Targets: %s\n", joinOrDash(engagement.Targets)))
		builder.WriteString(fmt.Sprintf("- Tags: %s\n", joinOrDash(engagement.Tags)))
		if len(engagement.Phases) > 0 {
			builder.WriteString("- Phases:\n")
			for _, phase := range engagement.Phases {
				builder.WriteString(fmt.Sprintf("  - %s: %d session(s); targets=%s; tags=%s\n", phase.Name, phase.SessionCount, joinOrDash(phase.Targets), joinOrDash(phase.Tags)))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## Findings\n\n")
	if len(bundle.Findings) == 0 {
		builder.WriteString("_No findings recorded for this report scope._\n\n")
	} else {
		for _, finding := range bundle.Findings {
			builder.WriteString(fmt.Sprintf("### [%s] %s (%s)\n\n", finding.Severity, finding.Title, finding.ID))
			builder.WriteString(fmt.Sprintf("- Status: %s\n", finding.Status))
			builder.WriteString(fmt.Sprintf("- Phase: %s\n", firstNonEmpty(finding.Phase, "-")))
			if strings.TrimSpace(finding.Description) != "" {
				builder.WriteString("\n")
				builder.WriteString(strings.TrimSpace(finding.Description))
				builder.WriteString("\n")
			}
			if len(finding.Evidence) > 0 {
				builder.WriteString("\nEvidence:\n")
				for _, evidence := range finding.Evidence {
					if strings.TrimSpace(evidence) == "" {
						continue
					}
					builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(evidence)))
				}
			}
			if strings.TrimSpace(finding.Remediation) != "" {
				builder.WriteString("\nRemediation:\n")
				builder.WriteString(strings.TrimSpace(finding.Remediation))
				builder.WriteString("\n")
			}
			if len(finding.References) > 0 {
				builder.WriteString("\nReferences:\n")
				for _, ref := range finding.References {
					if strings.TrimSpace(ref) == "" {
						continue
					}
					builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(ref)))
				}
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("## Evidence Snippets\n\n")
	if len(bundle.EvidenceSnippets) == 0 {
		builder.WriteString("_No curated evidence snippets recorded for this report scope._\n\n")
	} else {
		for _, snippet := range bundle.EvidenceSnippets {
			builder.WriteString(fmt.Sprintf("### %s\n\n", snippet.Title))
			builder.WriteString(fmt.Sprintf("- Source: %s\n", snippet.Source))
			if snippet.SessionID > 0 {
				builder.WriteString(fmt.Sprintf("- Session: %d\n", snippet.SessionID))
			}
			if snippet.FindingID != "" {
				builder.WriteString(fmt.Sprintf("- Finding: %s\n", snippet.FindingID))
			}
			builder.WriteString(fmt.Sprintf("- Engagement: %s\n", firstNonEmpty(snippet.Engagement, "-")))
			builder.WriteString(fmt.Sprintf("- Phase: %s\n", firstNonEmpty(snippet.Phase, "-")))
			builder.WriteString(fmt.Sprintf("- Target: %s\n", firstNonEmpty(snippet.Target, "-")))
			builder.WriteString(fmt.Sprintf("- Tags: %s\n", joinOrDash(snippet.Tags)))
			builder.WriteString(fmt.Sprintf("- Timestamp: %s\n", firstNonEmpty(snippet.Timestamp, "-")))
			if snippet.Summary != "" {
				builder.WriteString(fmt.Sprintf("- Summary: %s\n", snippet.Summary))
			}
			builder.WriteString("\n```text\n")
			builder.WriteString(strings.TrimSpace(snippet.Snippet))
			builder.WriteString("\n```\n\n")
		}
	}

	builder.WriteString("## Command Appendix\n\n")
	if len(bundle.CommandAppendix) == 0 {
		builder.WriteString("_No commands extracted for this report scope._\n\n")
	} else {
		for _, session := range bundle.CommandAppendix {
			builder.WriteString(fmt.Sprintf("### %s\n\n", session.Label))
			builder.WriteString(fmt.Sprintf("- Engagement: %s\n", firstNonEmpty(session.Engagement, "-")))
			builder.WriteString(fmt.Sprintf("- Phase: %s\n", firstNonEmpty(session.Phase, "-")))
			builder.WriteString(fmt.Sprintf("- Target: %s\n", firstNonEmpty(session.Target, "-")))
			builder.WriteString(fmt.Sprintf("- Tags: %s\n", joinOrDash(session.Tags)))
			builder.WriteString("\n")
			for _, command := range session.Commands {
				builder.WriteString(fmt.Sprintf("- `%s` `%s`\n", firstNonEmpty(command.Timestamp, session.Timestamp), command.Command))
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("## Integrity and Archive References\n\n")
	builder.WriteString("| Session | State | Engagement | Phase | Transcript | Archive | Manifest SHA256 | Archived At |\n")
	builder.WriteString("|---|---|---|---|---|---|---|---|\n")
	for _, ref := range bundle.IntegrityReferences {
		builder.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s | %s | %s |\n",
			ref.SessionID,
			firstNonEmpty(ref.State, "-"),
			firstNonEmpty(ref.Engagement, "-"),
			firstNonEmpty(ref.Phase, "-"),
			firstNonEmpty(ref.TranscriptPath, "-"),
			firstNonEmpty(ref.ArchivePath, "-"),
			firstNonEmpty(ref.ManifestSHA256, "-"),
			firstNonEmpty(ref.ArchivedAt, "-"),
		))
	}
	builder.WriteString("\n")

	if bundle.HasTranscriptAppendix {
		builder.WriteString("## Full Transcript Appendix\n\n")
		for _, transcript := range bundle.TranscriptAppendix {
			builder.WriteString(fmt.Sprintf("### %s\n\n", transcript.Label))
			builder.WriteString(fmt.Sprintf("- Engagement: %s\n", firstNonEmpty(transcript.Engagement, "-")))
			builder.WriteString(fmt.Sprintf("- Phase: %s\n", firstNonEmpty(transcript.Phase, "-")))
			builder.WriteString(fmt.Sprintf("- Target: %s\n", firstNonEmpty(transcript.Target, "-")))
			builder.WriteString(fmt.Sprintf("- Tags: %s\n", joinOrDash(transcript.Tags)))
			builder.WriteString("\n```bash\n")
			builder.WriteString(strings.TrimSpace(transcript.Transcript))
			builder.WriteString("\n```\n\n")
		}
	}

	return builder.String()
}

func buildHTMLTemplateData(bundle *reportBundle, aiAnalysis string, gifPaths map[int]string) (reporttpl.ReportTemplateData, error) {
	data := reporttpl.ReportTemplateData{
		Client:                bundle.Client,
		GeneratedAt:           bundle.GeneratedAt,
		AIAnalysis:            template.HTML(aiAnalysis),
		Summary:               buildSummaryTemplateData(bundle.Summary),
		EngagementMetadata:    buildEngagementTemplateData(bundle.EngagementMetadata),
		Findings:              bundle.Findings,
		EvidenceSnippets:      buildEvidenceTemplateData(bundle.EvidenceSnippets),
		CommandAppendix:       buildCommandTemplateData(bundle.CommandAppendix, gifPaths),
		IntegrityReferences:   buildIntegrityTemplateData(bundle.IntegrityReferences),
		HasTranscriptAppendix: bundle.HasTranscriptAppendix,
	}

	if bundle.HasTranscriptAppendix {
		appendix, err := buildTranscriptTemplateData(bundle.TranscriptAppendix, gifPaths)
		if err != nil {
			return reporttpl.ReportTemplateData{}, err
		}
		data.TranscriptAppendix = appendix
	}

	return data, nil
}

func buildSummaryTemplateData(summary reportSummary) reporttpl.ReportSummaryTemplateData {
	return reporttpl.ReportSummaryTemplateData{
		SessionCount:    summary.SessionCount,
		EngagementCount: summary.EngagementCount,
		PhaseCount:      summary.PhaseCount,
		FindingCount:    summary.FindingCount,
		EvidenceCount:   summary.EvidenceCount,
		CommandCount:    summary.CommandCount,
		ArchiveRefCount: summary.ArchiveRefCount,
		FirstObserved:   summary.FirstObserved,
		LastObserved:    summary.LastObserved,
		CriticalCount:   summary.CriticalCount,
		HighCount:       summary.HighCount,
		MediumCount:     summary.MediumCount,
		LowCount:        summary.LowCount,
		InfoCount:       summary.InfoCount,
		OpenCount:       summary.OpenCount,
		VerifiedCount:   summary.VerifiedCount,
		ClosedCount:     summary.ClosedCount,
		Targets:         summary.Targets,
		Tags:            summary.Tags,
		Operators:       summary.Operators,
	}
}

func buildEngagementTemplateData(metadata []reportEngagementMetadata) []reporttpl.EngagementMetadataTemplateData {
	items := make([]reporttpl.EngagementMetadataTemplateData, 0, len(metadata))
	for _, engagement := range metadata {
		phaseItems := make([]reporttpl.PhaseMetadataTemplateData, 0, len(engagement.Phases))
		for _, phase := range engagement.Phases {
			phaseItems = append(phaseItems, reporttpl.PhaseMetadataTemplateData{
				Name:         phase.Name,
				SessionCount: phase.SessionCount,
				Targets:      phase.Targets,
				Tags:         phase.Tags,
			})
		}
		items = append(items, reporttpl.EngagementMetadataTemplateData{
			Name:         engagement.Name,
			DateRange:    engagement.DateRange,
			SessionCount: engagement.SessionCount,
			Operators:    engagement.Operators,
			Targets:      engagement.Targets,
			Tags:         engagement.Tags,
			Phases:       phaseItems,
		})
	}
	return items
}

func buildEvidenceTemplateData(snippets []reportEvidenceSnippet) []reporttpl.EvidenceSnippetTemplateData {
	items := make([]reporttpl.EvidenceSnippetTemplateData, 0, len(snippets))
	for _, snippet := range snippets {
		items = append(items, reporttpl.EvidenceSnippetTemplateData{
			Title:      snippet.Title,
			Source:     snippet.Source,
			Summary:    snippet.Summary,
			Snippet:    snippet.Snippet,
			Timestamp:  snippet.Timestamp,
			Engagement: snippet.Engagement,
			Phase:      snippet.Phase,
			Target:     snippet.Target,
			Tags:       snippet.Tags,
			SessionID:  snippet.SessionID,
			FindingID:  snippet.FindingID,
		})
	}
	return items
}

func buildCommandTemplateData(appendix []reportCommandAppendix, gifPaths map[int]string) []reporttpl.CommandAppendixTemplateData {
	items := make([]reporttpl.CommandAppendixTemplateData, 0, len(appendix))
	for _, session := range appendix {
		commands := make([]reporttpl.CommandTemplateData, 0, len(session.Commands))
		for _, command := range session.Commands {
			commands = append(commands, reporttpl.CommandTemplateData{
				Timestamp: command.Timestamp,
				Command:   command.Command,
				Output:    shortenText(command.Output, 200),
			})
		}
		items = append(items, reporttpl.CommandAppendixTemplateData{
			SessionID:  session.SessionID,
			Label:      session.Label,
			Timestamp:  session.Timestamp,
			Engagement: session.Engagement,
			Phase:      session.Phase,
			Target:     session.Target,
			Tags:       session.Tags,
			GIFPath:    gifPaths[session.SessionID],
			Commands:   commands,
		})
	}
	return items
}

func buildIntegrityTemplateData(references []reportIntegrityReference) []reporttpl.IntegrityReferenceTemplateData {
	items := make([]reporttpl.IntegrityReferenceTemplateData, 0, len(references))
	for _, ref := range references {
		items = append(items, reporttpl.IntegrityReferenceTemplateData{
			SessionID:      ref.SessionID,
			State:          ref.State,
			Engagement:     ref.Engagement,
			Phase:          ref.Phase,
			Target:         ref.Target,
			TranscriptPath: ref.TranscriptPath,
			ArchivePath:    ref.ArchivePath,
			ManifestSHA256: ref.ManifestSHA256,
			ArchivedAt:     ref.ArchivedAt,
		})
	}
	return items
}

func buildTranscriptTemplateData(appendix []reportTranscriptAppendix, gifPaths map[int]string) ([]reporttpl.TranscriptAppendixTemplateData, error) {
	items := make([]reporttpl.TranscriptAppendixTemplateData, 0, len(appendix))
	for _, transcript := range appendix {
		item := reporttpl.TranscriptAppendixTemplateData{
			SessionID:  transcript.SessionID,
			Label:      transcript.Label,
			Timestamp:  transcript.Timestamp,
			Engagement: transcript.Engagement,
			Phase:      transcript.Phase,
			Target:     transcript.Target,
			Tags:       transcript.Tags,
			Content:    template.HTML(template.HTMLEscapeString(transcript.Transcript)),
		}
		if gifPaths != nil {
			item.GIFPath = gifPaths[transcript.SessionID]
		}
		items = append(items, item)
	}
	return items, nil
}

func extractSessionCommands(session Session) ([]CommandExecution, error) {
	if strings.HasSuffix(strings.ToLower(session.Path), ".tty") {
		timeline, err := ParseTimeline(session.Path)
		if err == nil && len(timeline.Commands) > 0 {
			return timeline.Commands, nil
		}
	}

	transcript, err := loadSessionTranscriptText(session.Path)
	if err != nil {
		return nil, err
	}
	return extractCommandsFromText(transcript, session.ModTime), nil
}

func extractCommandsFromText(transcript, fallbackTimestamp string) []CommandExecution {
	lines := strings.Split(transcript, "\n")
	commands := make([]CommandExecution, 0)
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		if cleaned == "" || !isPromptLine(cleaned) {
			continue
		}
		command := strings.TrimSpace(extractCommand(cleaned))
		if command == "" {
			continue
		}
		commands = append(commands, CommandExecution{
			Timestamp: fallbackTimestamp,
			Command:   command,
		})
	}
	return commands
}

func loadSessionTranscriptText(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(strings.ToLower(path), ".tty") {
		reader = NewTtyReader(f)
	}

	rawData, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	cleanData := utils.CleanTuiMarkers(rawData)
	lines := strings.Split(string(cleanData), "\n")

	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(utils.RenderPlain(line))
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String()), nil
}

func renderTranscriptHTML(path string) (template.HTML, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(strings.ToLower(path), ".tty") {
		reader = NewTtyReader(f)
	}

	rawData, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	cleanData := utils.CleanTuiMarkers(rawData)
	lines := strings.Split(string(cleanData), "\n")

	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(utils.RenderAnsiHTML(line))
		builder.WriteString("\n")
	}
	return template.HTML(builder.String()), nil
}

func extractNoteSnippet(logPath string, byteOffset int64) string {
	if logPath == "" || byteOffset <= 0 {
		return ""
	}

	f, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	limitReader := io.LimitReader(f, byteOffset)
	var reader io.Reader = limitReader
	if strings.HasSuffix(strings.ToLower(logPath), ".tty") {
		reader = NewTtyReader(limitReader)
	}

	cleaner := utils.NewCleanReader(reader)
	data, err := io.ReadAll(cleaner)
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		if cleaned == "" {
			continue
		}
		filtered = append(filtered, cleaned)
	}
	if len(filtered) == 0 {
		return ""
	}
	if len(filtered) > 6 {
		filtered = filtered[len(filtered)-6:]
	}
	return strings.Join(filtered, "\n")
}

func uniqueValuesFromSessions(sessions []reportSessionBundle, getter func(reportSessionBundle) string) []string {
	set := make(map[string]struct{})
	for _, session := range sessions {
		value := strings.TrimSpace(getter(session))
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}
	return sortedKeys(set)
}

func uniqueSortedSessionTargets(sessions []reportSessionBundle) []string {
	set := make(map[string]struct{})
	for _, session := range sessions {
		if target := formatSessionTarget(session.Session); target != "" {
			set[target] = struct{}{}
		}
	}
	return sortedKeys(set)
}

func uniqueSortedSessionTags(sessions []reportSessionBundle) []string {
	set := make(map[string]struct{})
	for _, session := range sessions {
		for _, tag := range session.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			set[tag] = struct{}{}
		}
	}
	return sortedKeys(set)
}

func uniqueSortedSessionOperators(sessions []reportSessionBundle) []string {
	set := make(map[string]struct{})
	for _, session := range sessions {
		operator := strings.TrimSpace(session.Session.Metadata.Operator)
		if operator == "" {
			continue
		}
		set[operator] = struct{}{}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func updateRange(start, end *time.Time, value time.Time) {
	if value.IsZero() {
		return
	}
	if start.IsZero() || value.Before(*start) {
		*start = value
	}
	if end.IsZero() || value.After(*end) {
		*end = value
	}
}

func formatDateRange(start, end time.Time) string {
	if start.IsZero() && end.IsZero() {
		return ""
	}
	if start.IsZero() {
		return end.Format("2006-01-02 15:04:05")
	}
	if end.IsZero() {
		return start.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("%s -> %s", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))
}

func formatTimeOrDash(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatSessionTarget(session Session) string {
	target := strings.TrimSpace(session.Metadata.Target)
	targetIP := strings.TrimSpace(session.Metadata.TargetIP)
	switch {
	case target != "" && targetIP != "":
		return fmt.Sprintf("%s (%s)", target, targetIP)
	case target != "":
		return target
	case targetIP != "":
		return targetIP
	default:
		return ""
	}
}

func formatFindingTimestamp(finding vulns.Vuln) string {
	if !finding.UpdatedAt.IsZero() {
		return finding.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	if !finding.CreatedAt.IsZero() {
		return finding.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return ""
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func shortenText(input string, maxRunes int) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	runes := []rune(input)
	if len(runes) <= maxRunes {
		return input
	}
	return string(runes[:maxRunes]) + "..."
}
