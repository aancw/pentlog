package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [phase]",
	Short: "Export commands for a specific phase (recon, exploit, etc.)",
	Run: func(cmd *cobra.Command, args []string) {
		sessions, err := logs.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
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

		engagements = append([]string{"All Engagements"}, engagements...)

		engIdx := utils.SelectItem("Select Engagement", engagements)
		if engIdx == -1 {
			return
		}
		selectedEngagement := engagements[engIdx]
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

		fmt.Printf("Exporting logs for Client: %s, Engagement: %s, Phase: %s...\n", selectedClient, selectedEngagement, selectedPhase)

		report, err := logs.ExportCommands(selectedClient, selectedEngagement, selectedPhase)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

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
				defaultName := fmt.Sprintf("%s_%s_%s_report.md", utils.Slugify(selectedClient), utils.Slugify(fileNameEng), utils.Slugify(fileNamePhase))
				filename := utils.PromptString("Enter filename", defaultName)
				if filename == "" {
					filename = defaultName
				}

				if err := os.WriteFile(filename, []byte(report), 0644); err != nil {
					fmt.Printf("Error saving file: %v\n", err)
				} else {
					fmt.Printf("Report saved to %s\n", filename)
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
				defaultName := fmt.Sprintf("%s_%s_%s_report.html", utils.Slugify(selectedClient), utils.Slugify(fileNameEng), utils.Slugify(fileNamePhase))
				filename := utils.PromptString("Enter filename", defaultName)
				if filename == "" {
					filename = defaultName
				}

				htmlReport, err := logs.ExportCommandsHTML(selectedClient, selectedEngagement, selectedPhase)
				if err != nil {
					fmt.Printf("Error generating HTML: %v\n", err)
					continue
				}

				if err := os.WriteFile(filename, []byte(htmlReport), 0644); err != nil {
					fmt.Printf("Error saving file: %v\n", err)
				} else {
					fmt.Printf("HTML Report saved to %s\n", filename)
					return
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
