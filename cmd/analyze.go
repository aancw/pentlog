package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"pentlog/pkg/ai"
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

		confDir, err := config.GetUserPentlogDir()
		if err != nil {
			return fmt.Errorf("failed to get pentlog directory: %w", err)
		}
		aiConfigPath := filepath.Join(confDir, "ai.yaml")

		if _, err := os.Stat(aiConfigPath); os.IsNotExist(err) {
			idx := utils.SelectItem("AI config not found. Create one?", []string{"Yes", "No"})
			if idx != 0 {
				return nil
			}

			providerIdx := utils.SelectItem("Select Provider", []string{"Gemini", "Ollama"})
			var content string
			if providerIdx == 0 { // Gemini
				apiKey := utils.PromptString("Enter Gemini API Key", "")
				content = fmt.Sprintf("provider: \"gemini\"\ngemini:\n  api_key: \"%s\"\n", apiKey)
			} else if providerIdx == 1 { // Ollama
				model := utils.PromptString("Enter Ollama Model", "llama3:8b")
				url := utils.PromptString("Enter Ollama URL", "http://localhost:11434")
				content = fmt.Sprintf("provider: \"ollama\"\nollama:\n  model: \"%s\"\n  url: \"%s\"\n", model, url)
			} else {
				return nil
			}

			if err := os.WriteFile(aiConfigPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			fmt.Printf("AI config created at %s\n", aiConfigPath)
		}

		spin := utils.NewSpinner("Analyzing report with AI...")
		spin.Start()
		cfg, err := ai.LoadConfig(aiConfigPath)
		if err != nil {
			spin.Stop()
			return fmt.Errorf("failed to load config: %w", err)
		}

		reportData, err := ioutil.ReadFile(reportFile)
		if err != nil {
			spin.Stop()
			return fmt.Errorf("failed to read report file: %w", err)
		}

		var analyzer ai.AIAnalyzer
		switch cfg.Provider {
		case "gemini":
			analyzer, err = ai.NewGeminiClient(cfg)
			if err != nil {
				spin.Stop()
				return fmt.Errorf("failed to create gemini client: %w", err)
			}
		case "ollama":
			analyzer, err = ai.NewOllamaClient(cfg)
			if err != nil {
				spin.Stop()
				return fmt.Errorf("failed to create ollama client: %w", err)
			}
		default:
			spin.Stop()
			return fmt.Errorf("unknown AI provider: %s", cfg.Provider)
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
