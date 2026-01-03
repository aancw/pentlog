package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GeminiClient is a client for the Gemini API.
type GeminiClient struct {
	APIKey string
}

// NewGeminiClient creates a new GeminiClient.
func NewGeminiClient(cfg *Config) (*GeminiClient, error) {
	if cfg.Gemini.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key not found in config")
	}
	return &GeminiClient{APIKey: cfg.Gemini.APIKey}, nil
}

// Analyze analyzes the report using the Gemini API.
func (c *GeminiClient) Analyze(report string, summarize bool) (string, error) {
	if summarize {
		return Summarize(report, c)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro-latest:generateContent"

	requestBody, err := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": "Analyze the following pentest report and provide a summary of the findings:\n\n" + report,
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-goog-api-key", c.APIKey)
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

	// Safely access the nested text field
	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("no candidates found in response")
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid candidate format")
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no content found in candidate")
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return "", fmt.Errorf("no parts found in content")
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid part format")
	}

	text, ok := part["text"].(string)
	if !ok {
		return "", fmt.Errorf("no text found in part")
	}

	return text, nil
}
