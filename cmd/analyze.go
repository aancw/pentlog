package cmd

import (
	"fmt"
	"io/ioutil"
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [report_file]",
	Short: "Analyze a report with an AI provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reportFile := args[0]

		mgr := config.Manager()
		aiConfigPath := mgr.GetPaths().AIConfigFile

		if err := ensureAIConfigInteractive(aiConfigPath); err != nil {
			if err == errAIConfigSetupCancelled {
				return nil
			}
			return err
		}

		spin := utils.NewSpinner("Analyzing report with AI...")
		spin.Start()

		reportData, err := ioutil.ReadFile(reportFile)
		if err != nil {
			spin.Stop()
			return fmt.Errorf("failed to read report file: %w", err)
		}

		analyzer, err := newAnalyzerFromConfig(aiConfigPath)
		if err != nil {
			spin.Stop()
			return err
		}

		analysis, err := analyzer.Analyze(string(reportData), !fullReport)
		spin.Stop()
		if err != nil {
			return fmt.Errorf("failed to analyze report: %w", err)
		}

		// Clean up excessive newlines
		analysis = strings.TrimSpace(analysis)

		fmt.Println("--- AI Analysis ---")
		fmt.Println(analysis)

		return nil
	},
}
