package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pentlog/pkg/ai"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/spf13/cobra"
)

var analyze bool

var exportCmd = &cobra.Command{
	Use:   "export [phase]",
	Short: "Export commands for a specific phase (recon, exploit, etc.)",
	Run: func(cmd *cobra.Command, args []string) {
		sessions, err := logs.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return
		}

		clientMap := make(map[string]bool)
		for _, s := range sessions {
			clientMap[s.Metadata.Client] = true
		}
		var clients []string
		for c := range clientMap {
			clients = append(clients, c)
		}
		if len(clients) == 0 {
			fmt.Println("No clients found.")
			return
		}

		clientIdx := utils.SelectItem("Select Client", clients)
		if clientIdx == -1 {
			return
		}
		selectedClient := clients[clientIdx]

		engMap := make(map[string]bool)
		for _, s := range sessions {
			if s.Metadata.Client == selectedClient {
				engMap[s.Metadata.Engagement] = true
			}
		}
		var engagements []string
		for e := range engMap {
			engagements = append(engagements, e)
		}

		engagements = append([]string{"All Engagements", "[View Existing Reports]"}, engagements...)

		var selectedEngagement string
		for {
			engIdx := utils.SelectItem("Select Engagement", engagements)
			if engIdx == -1 {
				return
			}
			selectedEngagement = engagements[engIdx]

			if selectedEngagement == "[View Existing Reports]" {
				reports, err := logs.ListClientReports(selectedClient)
				if err != nil {
					fmt.Printf("Error listing reports: %v\n", err)
				} else if len(reports) == 0 {
					fmt.Println("No existing reports found for this client.")
				} else {
					for {
						// Create a separate list for the menu to include "Back"
						menuItems := append([]string{"[Back]"}, reports...)

						reportIdx := utils.SelectItem("Select Report to Open", menuItems)
						if reportIdx == -1 {
							break // Exit inner loop
						}

						selectedReport := menuItems[reportIdx]
						if selectedReport == "[Back]" {
							break
						}

						// Open the report
						reportsDir, err := config.GetReportsDir()
						if err != nil {
							fmt.Printf("Error getting reports dir: %v\n", err)
							continue
						}
						fullPath := filepath.Join(reportsDir, utils.Slugify(selectedClient), selectedReport)

						fmt.Printf("Opening %s...\n", selectedReport)
						if err := utils.OpenFile(fullPath); err != nil {
							fmt.Printf("Error opening file: %v\n", err)
						}
					}
				}
				continue
			}
			break
		}

		if selectedEngagement == "All Engagements" {
			selectedEngagement = ""
		}

		phaseMap := make(map[string]bool)
		for _, s := range sessions {
			if s.Metadata.Client == selectedClient {
				if selectedEngagement != "" && s.Metadata.Engagement != selectedEngagement {
					continue
				}
				phaseMap[s.Metadata.Phase] = true
			}
		}
		var phases []string
		for p := range phaseMap {
			phases = append(phases, p)
		}
		phases = append([]string{"All Phases"}, phases...)

		phaseIdx := utils.SelectItem("Select Phase to Export", phases)
		if phaseIdx == -1 {
			return
		}
		selectedPhase := phases[phaseIdx]
		if selectedPhase == "All Phases" {
			selectedPhase = ""
		}

		// --- Check for existing reports ---
		reportsDir, err := config.GetReportsDir()
		if err == nil {
			clientDir := filepath.Join(reportsDir, utils.Slugify(selectedClient))

			// Construct default filenames (logic duplicated from save steps to check existence)
			fileNameEng := selectedEngagement
			if fileNameEng == "" {
				fileNameEng = "all-engagements"
			}
			fileNamePhase := selectedPhase
			if fileNamePhase == "" {
				fileNamePhase = "all-phases"
			}

			baseName := fmt.Sprintf("%s_%s_%s_report", utils.Slugify(selectedClient), utils.Slugify(fileNameEng), utils.Slugify(fileNamePhase))
			mdName := baseName + ".md"
			htmlName := baseName + ".html"

			mdPath := filepath.Join(clientDir, mdName)
			htmlPath := filepath.Join(clientDir, htmlName)

			var existing []string
			if info, err := os.Stat(mdPath); err == nil {
				existing = append(existing, fmt.Sprintf("%s (Created: %s)", mdPath, info.ModTime().Format("2006-01-02 15:04:05")))
			}
			if info, err := os.Stat(htmlPath); err == nil {
				existing = append(existing, fmt.Sprintf("%s (Created: %s)", htmlPath, info.ModTime().Format("2006-01-02 15:04:05")))
			}

			if len(existing) > 0 {
				fmt.Println("\nExisting report(s) found for this scope:")
				for _, p := range existing {
					fmt.Printf("- %s\n", p)
				}
				fmt.Println("")

				regen := utils.SelectItem("Do you still want to generate the report?", []string{"No", "Yes"})
				if regen == 0 { // No
					return
				}
			}
		}
		// ----------------------------------

		fmt.Printf("Exporting logs for Client: %s, Engagement: %s, Phase: %s...\n", selectedClient, selectedEngagement, selectedPhase)

		var finalSessions []logs.Session
		for _, s := range sessions {
			if s.Metadata.Client != selectedClient {
				continue
			}
			if selectedEngagement != "" && s.Metadata.Engagement != selectedEngagement {
				continue
			}
			if selectedPhase != "" && strings.TrimSpace(strings.ToLower(s.Metadata.Phase)) != strings.TrimSpace(strings.ToLower(selectedPhase)) {
				continue
			}
			finalSessions = append(finalSessions, s)
		}

		report, err := logs.GenerateReport(finalSessions, selectedClient)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var analysisResult string
		if analyze {
			confDir, err := config.GetUserPentlogDir()
			if err != nil {
				fmt.Printf("Error getting pentlog directory: %v\n", err)
				os.Exit(1)
			}
			aiConfigPath := filepath.Join(confDir, "ai.yaml")

			if _, err := os.Stat(aiConfigPath); os.IsNotExist(err) {
				idx := utils.SelectItem("AI config not found. Create one?", []string{"Yes", "No"})
				if idx != 0 {
					return
				}

				providerIdx := utils.SelectItem("Select Provider", []string{"Gemini", "Ollama"})
				var content string
				if providerIdx == 0 { // Gemini
					apiKey := utils.PromptString("Enter Gemini API Key", "")
					content = fmt.Sprintf("provider: \"gemini\"\ngemini:\n  api_key: \"%s\"\n", apiKey)
				} else if providerIdx == 1 { // Ollama
					model := utils.PromptString("Enter Ollama Model", "llama3:8b")
					url := utils.PromptString("Enter Ollama URL", "http://localhost:11434")
					content = fmt.Sprintf("provider: \"ollama\"\nollama:\n  model: \"%s\"\n  url: \"%s\"\n", model, url)
				} else {
					return
				}

				if err := os.WriteFile(aiConfigPath, []byte(content), 0600); err != nil {
					fmt.Printf("Error creating config file: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("AI config created at %s\n", aiConfigPath)
			}

			spin := utils.NewSpinner("Summarizing report with AI...")
			spin.Start()
			cfg, err := ai.LoadConfig(aiConfigPath)
			if err != nil {
				spin.Stop()
				fmt.Printf("failed to load config: %v\n", err)
				os.Exit(1)
			}

			var analyzer ai.AIAnalyzer
			switch cfg.Provider {
			case "gemini":
				analyzer, err = ai.NewGeminiClient(cfg)
				if err != nil {
					fmt.Printf("failed to create gemini client: %v\n", err)
					os.Exit(1)
				}
			case "ollama":
				analyzer, err = ai.NewOllamaClient(cfg)
				if err != nil {
					fmt.Printf("failed to create ollama client: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("unknown AI provider: %s\n", cfg.Provider)
				os.Exit(1)
			}

			analysisResult, err = analyzer.Analyze(report, !fullReport)
			spin.Stop()
			if err != nil {
				fmt.Printf("failed to analyze report: %v\n", err)
				os.Exit(1)
			}

			// Clean up excessive newlines
			analysisResult = strings.TrimSpace(analysisResult)

			analysisBlock := "\n## AI Analysis\n\n" + analysisResult + "\n\n---\n"

			lines := strings.SplitN(report, "\n", 2)
			if len(lines) > 1 {
				report = lines[0] + "\n" + analysisBlock + lines[1]
			} else {
				report = analysisBlock + report
			}
		}

		// --- Findings (Top of Report) ---
		manager := vulns.NewManager(selectedClient, selectedEngagement)
		findingsList, err := manager.List()
		if err == nil && len(findingsList) > 0 {
			filteredFindings := []vulns.Vuln{}
			for _, f := range findingsList {
				if selectedPhase != "" && !strings.EqualFold(f.Phase, selectedPhase) {
					continue
				}
				filteredFindings = append(filteredFindings, f)
			}

			if len(filteredFindings) > 0 {
				var sb strings.Builder
				sb.WriteString("\n## Findings & Vulnerabilities\n\n")
				sb.WriteString("| ID | Severity | Title | Phase | Status |\n")
				sb.WriteString("|---|---|---|---|---|\n")
				for _, f := range filteredFindings {
					phaseDisplay := f.Phase
					if phaseDisplay == "" {
						phaseDisplay = "-"
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", f.ID, f.Severity, f.Title, phaseDisplay, f.Status))
				}
				sb.WriteString("\n---\n\n")

				lines := strings.SplitN(report, "\n", 2)
				if len(lines) > 1 {
					report = lines[0] + "\n" + sb.String() + lines[1]
				} else {
					report = sb.String() + report
				}
			}
		}
		// ----------------------------

		for {
			actions := []string{"Preview (pager)", "Save to File", "Save to HTML", "Exit"}
			actionIdx := utils.SelectItem("Action", actions)

			if actionIdx == -1 || actionIdx == 3 {
				break
			}

			switch actionIdx {
			case 0:
				pager := os.Getenv("PAGER")
				if pager == "" {
					pager = "less -R"
				}
				cmd := exec.Command("sh", "-c", fmt.Sprintf("%s", pager))
				cmd.Stdin = strings.NewReader(report)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("Error running pager: %v\n", err)
				}

			case 1:
				fileNamePhase := selectedPhase
				if fileNamePhase == "" {
					fileNamePhase = "all-phases"
				}
				fileNameEng := selectedEngagement
				if fileNameEng == "" {
					fileNameEng = "all-engagements"
				}
				reportsBaseDir, err := config.GetReportsDir()
				if err != nil {
					fmt.Printf("Error getting reports directory: %v\n", err)
					return
				}
				reportDir := filepath.Join(reportsBaseDir, utils.Slugify(selectedClient))
				if err := os.MkdirAll(reportDir, 0755); err != nil {
					fmt.Printf("Error creating report directory: %v\n", err)
					return
				}

				defaultName := fmt.Sprintf("%s_%s_%s_report.md", utils.Slugify(selectedClient), utils.Slugify(fileNameEng), utils.Slugify(fileNamePhase))
				filename := utils.PromptString("Enter filename", defaultName)
				if filename == "" {
					filename = defaultName
				}

				fullPath := filepath.Join(reportDir, filename)

				if err := os.WriteFile(fullPath, []byte(report), 0644); err != nil {
					fmt.Printf("Error saving file: %v\n", err)
				} else {
					fmt.Printf("Report saved to %s\n", fullPath)

					prompt := utils.PromptString("Do you want to open the file? (y/N)", "no")
					if strings.ToLower(prompt) == "y" || strings.ToLower(prompt) == "yes" {
						if err := utils.OpenFile(fullPath); err != nil {
							fmt.Printf("Error opening file: %v\n", err)
						}
					}
					return
				}

			case 2:
				fileNamePhase := selectedPhase
				if fileNamePhase == "" {
					fileNamePhase = "all-phases"
				}
				fileNameEng := selectedEngagement
				if fileNameEng == "" {
					fileNameEng = "all-engagements"
				}
				reportsBaseDir, err := config.GetReportsDir()
				if err != nil {
					fmt.Printf("Error getting reports directory: %v\n", err)
					continue
				}
				reportDir := filepath.Join(reportsBaseDir, utils.Slugify(selectedClient))
				if err := os.MkdirAll(reportDir, 0755); err != nil {
					fmt.Printf("Error creating report directory: %v\n", err)
					continue
				}

				defaultName := fmt.Sprintf("%s_%s_%s_report.html", utils.Slugify(selectedClient), utils.Slugify(fileNameEng), utils.Slugify(fileNamePhase))
				filename := utils.PromptString("Enter filename", defaultName)
				if filename == "" {
					filename = defaultName
				}

				fullPath := filepath.Join(reportDir, filename)

				// --- Findings and Analysis ---
				var filteredFindings []vulns.Vuln
				manager := vulns.NewManager(selectedClient, selectedEngagement)
				findingsList, err := manager.List()
				if err == nil {
					for _, f := range findingsList {
						if selectedPhase != "" && !strings.EqualFold(f.Phase, selectedPhase) {
							continue
						}
						filteredFindings = append(filteredFindings, f)
					}
				}

				var analysisHTML string
				if analyze && analysisResult != "" {
					analysisHTML = string(markdown.ToHTML([]byte(analysisResult), nil, nil))
				}

				htmlReport, err := logs.GenerateHTMLReport(finalSessions, selectedClient, filteredFindings, analysisHTML)
				if err != nil {
					fmt.Printf("Error generating HTML: %v\n", err)
					continue
				}
				// ----------------------------

				if err := os.WriteFile(fullPath, []byte(htmlReport), 0644); err != nil {
					fmt.Printf("Error saving file: %v\n", err)
				} else {
					fmt.Printf("HTML Report saved to %s\n", fullPath)

					prompt := utils.PromptString("Do you want to open the file? (y/N)", "no")
					if strings.ToLower(prompt) == "y" || strings.ToLower(prompt) == "yes" {
						if err := utils.OpenFile(fullPath); err != nil {
							fmt.Printf("Error opening file: %v\n", err)
						}
					}
					return
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().BoolVar(&analyze, "analyze", false, "Analyze the report with an AI provider")
}
