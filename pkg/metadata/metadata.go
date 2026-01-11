package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
)

type Context struct {
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"` // "Client" or "Exam/Lab"
}

func Save(ctx Context) error {
	path, err := config.GetContextFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	historyPath := filepath.Join(dir, "history.jsonl")
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	lineData, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(lineData, '\n')); err != nil {
		return err
	}

	return nil
}

func Load() (*Context, error) {
	path, err := config.GetContextFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("context file not found. Run 'pentlog create' first")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

func LoadHistory() ([]Context, error) {
	dir, err := config.GetUserPentlogDir()
	if err != nil {
		return nil, err
	}
	historyPath := filepath.Join(dir, "history.jsonl")

	f, err := os.Open(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Context{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var history []Context
	dec := json.NewDecoder(f)
	for dec.More() {
		var ctx Context
		if err := dec.Decode(&ctx); err == nil {
			history = append(history, ctx)
		}
	}
	return history, nil
}
