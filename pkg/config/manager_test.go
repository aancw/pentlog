package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestManagerSingleton(t *testing.T) {
	// Reset singleton
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Get manager first time
	mgr1 := Manager()
	if mgr1 == nil {
		t.Fatal("Manager should not be nil")
	}

	// Get manager second time - should be same instance
	mgr2 := Manager()
	if mgr1 != mgr2 {
		t.Fatal("Manager should be singleton")
	}
}

func TestLoadDefaults(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()
	cfg := mgr.Get()

	if cfg.Paths.Home == "" {
		t.Error("Home path should be set")
	}
	if cfg.Paths.LogsDir == "" {
		t.Error("LogsDir should be set")
	}
	if cfg.Paths.DatabaseFile == "" {
		t.Error("DatabaseFile should be set")
	}
	if cfg.AI.Provider != "ollama" {
		t.Errorf("Default AI provider should be 'ollama', got %q", cfg.AI.Provider)
	}
}

func TestLoadFromEnv(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	os.Setenv("PENTLOG_AI_PROVIDER", "gemini")
	os.Setenv("PENTLOG_GEMINI_API_KEY", "test-key-123")
	os.Setenv("PENTLOG_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("PENTLOG_TEST_HOME")
		os.Unsetenv("PENTLOG_AI_PROVIDER")
		os.Unsetenv("PENTLOG_GEMINI_API_KEY")
		os.Unsetenv("PENTLOG_LOG_LEVEL")
	}()

	mgr := Manager()
	cfg := mgr.Get()

	if cfg.AI.Provider != "gemini" {
		t.Errorf("AI provider should be 'gemini', got %q", cfg.AI.Provider)
	}
	if cfg.AI.Gemini.APIKey != "test-key-123" {
		t.Errorf("Gemini API key should be 'test-key-123', got %q", cfg.AI.Gemini.APIKey)
	}
	if cfg.Env.LogLevel != "debug" {
		t.Errorf("Log level should be 'debug', got %q", cfg.Env.LogLevel)
	}
}

func TestLoadContext(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	// Create a test context
	testCtx := &ContextData{
		Client:     "TestClient",
		Engagement: "TestEngagement",
		Scope:      "TestScope",
		Operator:   "TestOperator",
		Phase:      "TestPhase",
		Type:       "Client",
	}

	// Save context
	if err := mgr.SaveContext(testCtx); err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// Load context
	loadedCtx, err := mgr.LoadContext()
	if err != nil {
		t.Fatalf("Failed to load context: %v", err)
	}

	if loadedCtx.Client != testCtx.Client {
		t.Errorf("Client should be %q, got %q", testCtx.Client, loadedCtx.Client)
	}
	if loadedCtx.Engagement != testCtx.Engagement {
		t.Errorf("Engagement should be %q, got %q", testCtx.Engagement, loadedCtx.Engagement)
	}
}

func TestLoadContextHistory(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	// Save multiple contexts
	for i := 1; i <= 3; i++ {
		ctx := &ContextData{
			Client:     "Client" + string(rune('0'+i)),
			Engagement: "Engagement",
			Phase:      "Phase" + string(rune('0'+i)),
			Type:       "Client",
		}
		if err := mgr.SaveContext(ctx); err != nil {
			t.Fatalf("Failed to save context %d: %v", i, err)
		}
	}

	// Load history
	history, err := mgr.LoadContextHistory()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("History should have 3 entries, got %d", len(history))
	}

	if history[0].Client != "Client1" {
		t.Errorf("First context client should be 'Client1', got %q", history[0].Client)
	}
}

func TestLoadAIConfig(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Create AI config file
	aiConfigPath := filepath.Join(testHome, ".pentlog", "ai_config.yaml")
	os.MkdirAll(filepath.Dir(aiConfigPath), 0700)

	configContent := `
provider: gemini
gemini:
  api_key: sk-test-key
ollama:
  model: neural-chat
  url: http://localhost:11434
`
	if err := os.WriteFile(aiConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create AI config: %v", err)
	}

	mgr := Manager()
	cfg := mgr.Get()

	if cfg.AI.Provider != "gemini" {
		t.Errorf("AI provider should be 'gemini', got %q", cfg.AI.Provider)
	}
	if cfg.AI.Gemini.APIKey != "sk-test-key" {
		t.Errorf("Gemini API key should be 'sk-test-key', got %q", cfg.AI.Gemini.APIKey)
	}
	if cfg.AI.Ollama.Model != "neural-chat" {
		t.Errorf("Ollama model should be 'neural-chat', got %q", cfg.AI.Ollama.Model)
	}
}

func TestEnsureDirectories(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	if err := mgr.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to ensure directories: %v", err)
	}

	cfg := mgr.Get()
	dirs := []string{
		cfg.Paths.Home,
		cfg.Paths.LogsDir,
		cfg.Paths.ReportsDir,
		cfg.Paths.ArchiveDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory should exist: %s", dir)
		}
	}
}

func TestEnvVarOverrides(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	customDBPath := filepath.Join(testHome, "custom.db")
	
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	os.Setenv("PENTLOG_DB_PATH", customDBPath)
	os.Setenv("PENTLOG_OLLAMA_URL", "http://custom:11434")
	defer func() {
		os.Unsetenv("PENTLOG_TEST_HOME")
		os.Unsetenv("PENTLOG_DB_PATH")
		os.Unsetenv("PENTLOG_OLLAMA_URL")
	}()

	mgr := Manager()
	cfg := mgr.Get()

	if cfg.Paths.DatabaseFile != customDBPath {
		t.Errorf("DatabaseFile should be %q, got %q", customDBPath, cfg.Paths.DatabaseFile)
	}
	if cfg.AI.Ollama.URL != "http://custom:11434" {
		t.Errorf("Ollama URL should be 'http://custom:11434', got %q", cfg.AI.Ollama.URL)
	}
}

func TestGetPaths(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()
	paths := mgr.GetPaths()

	if paths.Home == "" {
		t.Error("GetPaths should return populated paths")
	}
	if paths.DatabaseFile == "" {
		t.Error("GetPaths should include DatabaseFile")
	}
}

func TestGetAI(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	os.Setenv("PENTLOG_AI_PROVIDER", "gemini")
	defer func() {
		os.Unsetenv("PENTLOG_TEST_HOME")
		os.Unsetenv("PENTLOG_AI_PROVIDER")
	}()

	mgr := Manager()
	aiCfg := mgr.GetAI()

	if aiCfg.Provider != "gemini" {
		t.Errorf("GetAI should return correct provider, got %q", aiCfg.Provider)
	}
}

func TestSaveAndLoadContextJSON(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	// Ensure directories exist
	if err := mgr.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to ensure directories: %v", err)
	}

	// Test context with all fields
	testCtx := &ContextData{
		Client:     "Acme Corp",
		Engagement: "2024-Q1-Security",
		Scope:      "Internal Networks",
		Operator:   "Alice",
		Phase:      "Reconnaissance",
		Timestamp:  "2024-01-15T10:30:00Z",
		Type:       "Client",
	}

	// Save context
	if err := mgr.SaveContext(testCtx); err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// Verify JSON is valid
	cfg := mgr.Get()
	data, err := os.ReadFile(cfg.Paths.ContextFile)
	if err != nil {
		t.Fatalf("Failed to read context file: %v", err)
	}

	var savedCtx ContextData
	if err := json.Unmarshal(data, &savedCtx); err != nil {
		t.Fatalf("Invalid JSON in context file: %v", err)
	}

	if savedCtx.Client != testCtx.Client {
		t.Errorf("Saved client should be %q, got %q", testCtx.Client, savedCtx.Client)
	}
}

func TestContextFileNotFound(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	// Try to load context that doesn't exist
	_, err := mgr.LoadContext()
	if err == nil {
		t.Error("Loading non-existent context should return error")
	}
}

func TestLoadEmptyHistory(t *testing.T) {
	once = sync.Once{}
	manager = nil

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := Manager()

	// Load history when no history file exists
	history, err := mgr.LoadContextHistory()
	if err != nil {
		t.Fatalf("Loading empty history should not error: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Empty history should have 0 entries, got %d", len(history))
	}
}
