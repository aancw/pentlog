package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fullReport bool

const Version = "v0.3"

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
