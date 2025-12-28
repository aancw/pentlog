package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/config"
	"pentlog/pkg/metadata"
	"pentlog/pkg/system"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tool and engagement status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Pentlog Status ===")

		ctx, err := metadata.Load()
		if err != nil {
			fmt.Println("Context: No active engagement found.")
		} else {
			fmt.Println("Context: ACTIVE")
			fmt.Printf("  Client:   %s\n", ctx.Client)
			fmt.Printf("  Operator: %s\n", ctx.Operator)
		}

		logDir, err := config.GetLogsDir()
		if err != nil {
			fmt.Printf("Log directory: ERROR (%v)\n", err)
		} else {
			if _, err := os.Stat(logDir); err != nil {
				fmt.Printf("Log directory: missing (%s)\n", logDir)
			} else {
				fmt.Printf("Log directory: %s\n", logDir)
			}
		}

		if err := system.CheckDependencies(); err != nil {
			fmt.Printf("Recorder dependencies: MISSING (%v)\n", err)
		} else {
			fmt.Println("Recorder dependencies: OK (script, scriptreplay)")
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
