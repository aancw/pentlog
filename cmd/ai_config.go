package cmd

import (
	"errors"
	"fmt"
	"os"
	"pentlog/pkg/ai"
	"pentlog/pkg/utils"
)

var errAIConfigSetupCancelled = errors.New("ai configuration setup cancelled")

func ensureAIConfigInteractive(aiConfigPath string) error {
	if _, err := os.Stat(aiConfigPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	idx := utils.SelectItem("AI config not found. Create one?", []string{"Yes", "No"})
	if idx != 0 {
		return errAIConfigSetupCancelled
	}

	providerIdx := utils.SelectItem("Select Provider", []string{"Gemini", "Ollama"})
	var content string
	switch providerIdx {
	case 0:
		apiKey := utils.PromptString("Enter Gemini API Key", "")
		content = fmt.Sprintf("provider: \"gemini\"\ngemini:\n  api_key: \"%s\"\n", apiKey)
	case 1:
		model := utils.PromptString("Enter Ollama Model", "llama3:8b")
		url := utils.PromptString("Enter Ollama URL", "http://localhost:11434")
		content = fmt.Sprintf("provider: \"ollama\"\nollama:\n  model: \"%s\"\n  url: \"%s\"\n", model, url)
	default:
		return fmt.Errorf("unsupported AI provider selection")
	}

	if err := utils.WritePrivateFile(aiConfigPath, []byte(content)); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("AI config created at %s\n", aiConfigPath)
	return nil
}

func newAnalyzerFromConfig(aiConfigPath string) (ai.AIAnalyzer, error) {
	cfg, err := ai.LoadConfig(aiConfigPath)
	if err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case "gemini":
		return ai.NewGeminiClient(cfg)
	case "ollama":
		return ai.NewOllamaClient(cfg)
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
}
