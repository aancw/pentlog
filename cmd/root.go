package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/db"
	"pentlog/pkg/logs"
	"pentlog/pkg/system"
	"time"

	"github.com/spf13/cobra"
)

var fullReport bool

const Version = "v0.12.0"

var rootCmd = &cobra.Command{
	Use:   "pentlog",
	Short: "Evidence-First Pentest Logging Tool",
	Long: `pentlog is a CLI tool designed to orchestrate ttyrec for pentest and exam use cases.
It ensures that all terminal activity is recorded, context-aware, and replayable.
Features include automated hashing (integrity), markdown export, and full shell replay capability.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Commands that don't require setup
		allowed := map[string]bool{
			"setup":      true,
			"version":    true,
			"update":     true,
			"completion": true,
			"help":       true,
		}

		if allowed[cmd.Name()] {
			return
		}

		// Check if setup has run
		if run, _ := system.IsSetupRun(); !run {
			fmt.Println("Error: pentlog is not initialized.")
			fmt.Println("Please run 'pentlog setup' first to initialize the environment.")
			os.Exit(1)
		}

		// Initialize database and run migrations early
		if _, err := db.GetDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize database: %v\n", err)
		}

		checkForCrashedSessions(cmd.Name())
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of pentlog",
	Long:  `All software has versions. This is pentlog's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pentlog %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().BoolVar(&fullReport, "full-report", false, "Perform a full analysis without summarization")
}

func checkForCrashedSessions(cmdName string) {
	if cmdName == "recover" || cmdName == "shell" {
		return
	}

	logs.MarkStaleSessions(5 * time.Minute)

	crashed, err := logs.GetCrashedSessions()
	if err != nil || len(crashed) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\n⚠️  Warning: %d crashed session(s) detected.\n", len(crashed))
	fmt.Fprintln(os.Stderr, "   Run 'pentlog recover' to review and recover them.")
}
