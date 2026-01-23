package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	Paths  PathsConfig
	Env    EnvConfig
	AI     AIConfig
	System SystemConfig
}

type PathsConfig struct {
	Home      string
	LogsDir   string
	ReportsDir string
	ArchiveDir string
	HashesDir  string
	ExtractsDir string
	TemplatesDir string
	ContextFile string
	HistoryFile string
	DatabaseFile string
	AIConfigFile string
}

type EnvConfig struct {
	LogLevel string
	TestHome string
	SudoUser string
}

type AIConfig struct {
	Provider string
	Gemini   struct {
		APIKey string
	}
	Ollama struct {
		Model string
		URL   string
	}
}

type SystemConfig struct {
	DBPath string
}

type ConfigManager struct {
	config *AppConfig
	mu     sync.RWMutex
}

var (
	manager *ConfigManager
	once    sync.Once
)

func Manager() *ConfigManager {
	once.Do(func() {
		manager = &ConfigManager{}
		manager.Load()
	})
	return manager
}

func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cfg := &AppConfig{}

	if err := cm.loadDefaults(cfg); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	if err := cm.loadFromEnv(cfg); err != nil {
		return fmt.Errorf("failed to load from environment: %w", err)
	}

	if _, err := os.Stat(cfg.Paths.AIConfigFile); err == nil {
		if err := cm.loadAIConfig(cfg); err != nil {
			return fmt.Errorf("failed to load AI config: %w", err)
		}
	}

	if err := cm.validate(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	cm.config = cfg
	return nil
}

func (cm *ConfigManager) loadDefaults(cfg *AppConfig) error {
	home, err := GetUserPentlogDir()
	if err != nil {
		return err
	}

	cfg.Paths = PathsConfig{
		Home:          home,
		LogsDir:       filepath.Join(home, LogsDirName),
		ReportsDir:    filepath.Join(home, ReportsDirName),
		ArchiveDir:    filepath.Join(home, ArchiveDirName),
		HashesDir:     filepath.Join(home, HashesDirName),
		ExtractsDir:   filepath.Join(home, ExtractsDirName),
		TemplatesDir:  filepath.Join(home, TemplatesDirName),
		ContextFile:   filepath.Join(home, ContextFileName),
		HistoryFile:   filepath.Join(home, "history.jsonl"),
		DatabaseFile:  filepath.Join(home, "pentlog.db"),
		AIConfigFile:  filepath.Join(home, "ai_config.yaml"),
	}

	cfg.Env = EnvConfig{
		LogLevel: os.Getenv("PENTLOG_LOG_LEVEL"),
		TestHome: os.Getenv("PENTLOG_TEST_HOME"),
		SudoUser: os.Getenv("SUDO_USER"),
	}

	cfg.AI = AIConfig{
		Provider: "ollama",
	}

	cfg.System = SystemConfig{
		DBPath: cfg.Paths.DatabaseFile,
	}

	return nil
}

func (cm *ConfigManager) loadFromEnv(cfg *AppConfig) error {
	if home := os.Getenv("PENTLOG_HOME"); home != "" {
		cfg.Paths.Home = home
		cfg.Paths.LogsDir = filepath.Join(home, LogsDirName)
		cfg.Paths.ReportsDir = filepath.Join(home, ReportsDirName)
		cfg.Paths.ArchiveDir = filepath.Join(home, ArchiveDirName)
		cfg.Paths.HashesDir = filepath.Join(home, HashesDirName)
		cfg.Paths.ExtractsDir = filepath.Join(home, ExtractsDirName)
		cfg.Paths.TemplatesDir = filepath.Join(home, TemplatesDirName)
		cfg.Paths.ContextFile = filepath.Join(home, ContextFileName)
		cfg.Paths.HistoryFile = filepath.Join(home, "history.jsonl")
		cfg.Paths.DatabaseFile = filepath.Join(home, "pentlog.db")
		cfg.Paths.AIConfigFile = filepath.Join(home, "ai_config.yaml")
	}

	if dbPath := os.Getenv("PENTLOG_DB_PATH"); dbPath != "" {
		cfg.Paths.DatabaseFile = dbPath
		cfg.System.DBPath = dbPath
	}
	if contextFile := os.Getenv("PENTLOG_CONTEXT_FILE"); contextFile != "" {
		cfg.Paths.ContextFile = contextFile
	}

	if provider := os.Getenv("PENTLOG_AI_PROVIDER"); provider != "" {
		cfg.AI.Provider = provider
	}
	if apiKey := os.Getenv("PENTLOG_GEMINI_API_KEY"); apiKey != "" {
		cfg.AI.Gemini.APIKey = apiKey
	}
	if ollamaURL := os.Getenv("PENTLOG_OLLAMA_URL"); ollamaURL != "" {
		cfg.AI.Ollama.URL = ollamaURL
	}
	if ollamaModel := os.Getenv("PENTLOG_OLLAMA_MODEL"); ollamaModel != "" {
		cfg.AI.Ollama.Model = ollamaModel
	}

	if logLevel := os.Getenv("PENTLOG_LOG_LEVEL"); logLevel != "" {
		cfg.Env.LogLevel = logLevel
	}

	return nil
}

func (cm *ConfigManager) loadAIConfig(cfg *AppConfig) error {
	data, err := os.ReadFile(cfg.Paths.AIConfigFile)
	if err != nil {
		return err
	}

	var rawCfg map[string]interface{}
	if err := yaml.Unmarshal(data, &rawCfg); err != nil {
		return err
	}

	if provider, ok := rawCfg["provider"].(string); ok && provider != "" && cfg.AI.Provider == "ollama" {
		cfg.AI.Provider = provider
	}

	if gemini, ok := rawCfg["gemini"].(map[interface{}]interface{}); ok {
		if apiKey, ok := gemini["api_key"].(string); ok && apiKey != "" && cfg.AI.Gemini.APIKey == "" {
			cfg.AI.Gemini.APIKey = apiKey
		}
	}

	if ollama, ok := rawCfg["ollama"].(map[interface{}]interface{}); ok {
		if model, ok := ollama["model"].(string); ok && model != "" && cfg.AI.Ollama.Model == "" {
			cfg.AI.Ollama.Model = model
		}
		if url, ok := ollama["url"].(string); ok && url != "" && cfg.AI.Ollama.URL == "" {
			cfg.AI.Ollama.URL = url
		}
	}

	return nil
}

func (cm *ConfigManager) validate(cfg *AppConfig) error {
	if cfg.Paths.Home == "" {
		return fmt.Errorf("pentlog home directory not set")
	}
	if cfg.Paths.DatabaseFile == "" {
		return fmt.Errorf("database file path not set")
	}
	return nil
}

func (cm *ConfigManager) Get() *AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

func (cm *ConfigManager) GetPaths() PathsConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Paths
}

func (cm *ConfigManager) GetAI() AIConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.AI
}

func (cm *ConfigManager) GetEnv() EnvConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Env
}

func (cm *ConfigManager) LoadContext() (*ContextData, error) {
	cm.mu.RLock()
	contextPath := cm.config.Paths.ContextFile
	cm.mu.RUnlock()

	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("context file not found. Run 'pentlog create' first")
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, err
	}

	var ctx ContextData
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

func (cm *ConfigManager) SaveContext(ctx *ContextData) error {
	cm.mu.RLock()
	contextPath := cm.config.Paths.ContextFile
	historyPath := cm.config.Paths.HistoryFile
	cm.mu.RUnlock()

	dir := filepath.Dir(contextPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Save current context
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return err
	}

	// Append to history
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

func (cm *ConfigManager) LoadContextHistory() ([]ContextData, error) {
	cm.mu.RLock()
	historyPath := cm.config.Paths.HistoryFile
	cm.mu.RUnlock()

	f, err := os.Open(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ContextData{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var history []ContextData
	dec := json.NewDecoder(f)
	for dec.More() {
		var ctx ContextData
		if err := dec.Decode(&ctx); err == nil {
			history = append(history, ctx)
		}
	}
	return history, nil
}

type ContextData struct {
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"` // "Client" or "Exam/Lab"
}

func (cm *ConfigManager) EnsureDirectories() error {
	cm.mu.RLock()
	paths := cm.config.Paths
	cm.mu.RUnlock()

	dirs := []string{
		paths.Home,
		paths.LogsDir,
		paths.ReportsDir,
		paths.ArchiveDir,
		paths.HashesDir,
		paths.ExtractsDir,
		paths.TemplatesDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (cm *ConfigManager) Refresh() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Reset the singleton to force reload
	once = sync.Once{}
	manager = nil
	
	return Manager().Load()
}

func ResetManagerForTesting() {
	once = sync.Once{}
	manager = nil
}
