package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"

	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract <phase>",
	Short: "Extract commands for a specific phase (recon, exploit, etc.)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		phase := args[0]
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
