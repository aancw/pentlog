package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"

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
			phase = utils.PromptString("Phase", "recon")
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

		fmt.Println(report)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
}
