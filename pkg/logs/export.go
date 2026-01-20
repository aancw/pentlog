package logs

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/templates"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"
	"sort"
	"strings"
)

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
	if len(sessions) == 0 {
		return "", fmt.Errorf("no sessions to report")
	}

	grouped := groupSessions(sessions)

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
			sort.Slice(sessList, func(i, j int) bool {
				return sessList[i].ID < sessList[j].ID
			})

			for _, s := range sessList {
				f, err := os.Open(s.Path)
				if err != nil {
					continue
				}

				var r io.Reader = f
				if strings.HasSuffix(s.Path, ".tty") {
					r = NewTtyReader(f)
				}

				rawData, err := io.ReadAll(r)
				f.Close()

				if err != nil {
					continue
				}

				cleanData := utils.CleanTuiMarkers(rawData)
				lines := strings.Split(string(cleanData), "\n")

				builder.WriteString(fmt.Sprintf("#### Session %d (%s)\n", s.ID, s.ModTime))
				builder.WriteString("```bash\n")
				for _, line := range lines {
					builder.WriteString(utils.RenderPlain(line) + "\n")
				}
				builder.WriteString("\n```\n\n")
			}
		}
	}

	return builder.String(), nil
}

func groupSessions(sessions []Session) map[string]map[string][]Session {
	grouped := make(map[string]map[string][]Session)
	for _, s := range sessions {
		if grouped[s.Metadata.Engagement] == nil {
			grouped[s.Metadata.Engagement] = make(map[string][]Session)
		}
		grouped[s.Metadata.Engagement][s.Metadata.Phase] = append(grouped[s.Metadata.Engagement][s.Metadata.Phase], s)
	}
	return grouped
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

	return GenerateHTMLReport(filtered, client, nil, "")
}

func GenerateHTMLReport(sessions []Session, client string, findings []vulns.Vuln, aiAnalysis string) (string, error) {
	if len(sessions) == 0 {
		return "", fmt.Errorf("no sessions to report")
	}
	grouped := groupSessions(sessions)

	// Prepare data for template
	reportData := templates.ReportTemplateData{
		Client:     client,
		Findings:   findings,
		AIAnalysis: template.HTML(aiAnalysis),
	}

	var engKeys []string
	for k := range grouped {
		engKeys = append(engKeys, k)
	}
	sort.Strings(engKeys)

	for _, eng := range engKeys {
		eData := templates.EngagementTemplateData{Name: eng}
		phases := grouped[eng]
		var phaseKeys []string
		for k := range phases {
			phaseKeys = append(phaseKeys, k)
		}
		sort.Strings(phaseKeys)

		for _, p := range phaseKeys {
			pData := templates.PhaseTemplateData{Name: p}
			sessList := phases[p]
			sort.Slice(sessList, func(i, j int) bool {
				return sessList[i].ID < sessList[j].ID
			})

			for _, s := range sessList {
				f, err := os.Open(s.Path)
				if err != nil {
					continue
				}

				var r io.Reader = f
				if strings.HasSuffix(s.Path, ".tty") {
					r = NewTtyReader(f)
				}

				rawData, err := io.ReadAll(r)
				f.Close()
				if err != nil {
					continue
				}
				cleanData := utils.CleanTuiMarkers(rawData)

				// Convert ANSI to HTML
				var htmlContentBuilder strings.Builder
				lines := strings.Split(string(cleanData), "\n")
				for _, line := range lines {
					htmlContent := utils.RenderAnsiHTML(line)
					htmlContentBuilder.WriteString(htmlContent + "\n")
				}

				sData := templates.SessionTemplateData{
					ID:      s.ID,
					ModTime: s.ModTime,
					Content: template.HTML(htmlContentBuilder.String()),
				}
				pData.Sessions = append(pData.Sessions, sData)
			}
			if len(pData.Sessions) > 0 {
				eData.Phases = append(eData.Phases, pData)
			}
		}
		if len(eData.Phases) > 0 {
			reportData.Engagements = append(reportData.Engagements, eData)
		}
	}

	// Load templates from disk
	templatesDir, err := config.GetTemplatesDir()
	if err != nil {
		return "", fmt.Errorf("failed to get templates dir: %w", err)
	}

	htmlPath := filepath.Join(templatesDir, "report.html")
	cssPath := filepath.Join(templatesDir, "report.css")

	// Check if files exist
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

	tmpl, err := template.New("report").Parse(string(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, reportData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func ListClientReports(client string) ([]string, error) {
	reportsDir, err := config.GetReportsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get reports directory: %w", err)
	}

	clientDir := filepath.Join(reportsDir, utils.Slugify(client))
	entries, err := os.ReadDir(clientDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read client report directory: %w", err)
	}

	var reports []string
	for _, entry := range entries {
		if !entry.IsDir() {
			reports = append(reports, entry.Name())
		}
	}
	return reports, nil
}
