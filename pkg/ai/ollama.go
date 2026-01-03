package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OllamaClient is a client for the Ollama API.
type OllamaClient struct {
	Model string
	URL   string
}

// NewOllamaClient creates a new OllamaClient.
func NewOllamaClient(cfg *Config) (*OllamaClient, error) {
	if cfg.Ollama.Model == "" {
		return nil, fmt.Errorf("Ollama model not found in config")
	}
	if cfg.Ollama.URL == "" {
		return nil, fmt.Errorf("Ollama URL not found in config")
	}
	return &OllamaClient{Model: cfg.Ollama.Model, URL: cfg.Ollama.URL}, nil
}

// Analyze analyzes the report using the Ollama API.
func (c *OllamaClient) Analyze(report string, summarize bool) (string, error) {
	if summarize {
		return Summarize(report, c)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  c.Model,
		"prompt": "Analyze the following pentest report and provide a summary of the findings:\n\n" + report,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.URL+"/api/generate", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("no response found in result")
	}

	return response, nil
}
