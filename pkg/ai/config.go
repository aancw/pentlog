package ai

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Provider string `yaml:"provider"`
	Gemini   struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"gemini"`
	Ollama struct {
		Model string `yaml:"model"`
		URL   string `yaml:"url"`
	} `yaml:"ollama"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
