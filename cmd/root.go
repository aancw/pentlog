package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pentlog",
	Short: "Evidence-First Pentest Logging Tool",
	Long: `pentlog is a CLI tool designed to orchestrate tlog for pentest and exam use cases.
It ensures that all terminal activity is recorded, context-aware, and replayable.
Features include automated hashing (integrity), markdown export, and full shell replay capability.`,
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

func init() {
}
