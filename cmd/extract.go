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

var extractCmd = &cobra.Command{
	Use:   "extract [phase]",
	Short: "Extract commands for a specific phase (recon, exploit, etc.)",
	Run: func(cmd *cobra.Command, args []string) {
		phase := ""
		if len(args) > 0 {
			phase = args[0]
		}

		if phase == "" {
			phases := []string{"recon", "exploitation", "post-exploitation", "pivot", "cleanup", "Custom"}
			idx := utils.SelectItem("Select Phase", phases)
			if idx == -1 {
				fmt.Println("Selection canceled.")
				return
			}
			phase = phases[idx]

			if phase == "Custom" {
				phase = utils.PromptString("Enter Custom Phase", "")
			}
		}

		if phase == "" {
			fmt.Println("Error: Phase cannot be empty.")
			os.Exit(1)
		}
		fmt.Printf("Extracting logs for phase: %s...\n", phase)

		report, err := logs.ExtractCommands(phase)
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
				defaultName := fmt.Sprintf("%s_report.md", phase)
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
				defaultName := fmt.Sprintf("%s_report.html", phase)
				filename := utils.PromptString("Enter filename", defaultName)
				if filename == "" {
					filename = defaultName
				}

				htmlReport, err := logs.ExtractCommandsHTML(phase)
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
	rootCmd.AddCommand(extractCmd)
}
