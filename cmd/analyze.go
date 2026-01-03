package cmd

import (
	"fmt"
	"io/ioutil"
	"pentlog/pkg/ai"

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

		cfg, err := ai.LoadConfig("setting-ai.yaml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reportData, err := ioutil.ReadFile(reportFile)
		if err != nil {
			return fmt.Errorf("failed to read report file: %w", err)
		}

		var analyzer ai.AIAnalyzer
		switch cfg.Provider {
		case "gemini":
			analyzer, err = ai.NewGeminiClient(cfg)
			if err != nil {
				return fmt.Errorf("failed to create gemini client: %w", err)
			}
		case "ollama":
			analyzer, err = ai.NewOllamaClient(cfg)
			if err != nil {
				return fmt.Errorf("failed to create ollama client: %w", err)
			}
		default:
			return fmt.Errorf("unknown AI provider: %s", cfg.Provider)
		}

		analysis, err := analyzer.Analyze(string(reportData), !fullReport)
		if err != nil {
			return fmt.Errorf("failed to analyze report: %w", err)
		}

		fmt.Println("--- AI Analysis ---")
		fmt.Println(analysis)

		return nil
	},
}
