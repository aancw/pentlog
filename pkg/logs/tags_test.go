package logs

import (
	"os"
	"testing"

	"pentlog/pkg/config"
	"pentlog/pkg/db"
)

func TestTagFunctionality(t *testing.T) {
	// Reset config and DB singletons for testing
	config.ResetManagerForTesting()
	db.CloseDB()

	testHome := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", testHome)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	// Initialize config
	mgr := config.Manager()
	mgr.GetPaths()

	// Initialize DB
	database, _ := db.GetDB()
	defer db.CloseDB()

	// Create a test session
	_, err := database.Exec(`
		INSERT INTO sessions (client, engagement, phase, timestamp, filename, relative_path, size)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "TestClient", "TestEngagement", "TestPhase", "2024-01-01T00:00:00Z", "test.tty", "test/test.tty", 1024)

	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	sessionID := 1

	// Test AddTag
	if err := AddTag(sessionID, "important"); err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	if err := AddTag(sessionID, "dc-01"); err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	// Test GetSessionTags
	tags, err := GetSessionTags(sessionID)
	if err != nil {
		t.Errorf("GetSessionTags failed: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d. Tags: %v", len(tags), tags)
	}

	// Test ListAllTags
	allTags, err := ListAllTags()
	if err != nil {
		t.Errorf("ListAllTags failed: %v", err)
	}

	if len(allTags) != 2 {
		t.Errorf("Expected 2 total tags, got %d. Tags: %v", len(allTags), allTags)
	}

	// Test ListSessionsByTag
	sessions, err := ListSessionsByTag("important")
	if err != nil {
		t.Errorf("ListSessionsByTag failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session with tag, got %d", len(sessions))
	}

	// Test duplicate tag (should be ignored)
	if err := AddTag(sessionID, "important"); err != nil {
		t.Errorf("AddTag duplicate failed: %v", err)
	}

	tags, _ = GetSessionTags(sessionID)
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags after duplicate, got %d", len(tags))
	}

	// Test RemoveTag
	if err := RemoveTag(sessionID, "dc-01"); err != nil {
		t.Errorf("RemoveTag failed: %v", err)
	}

	tags, _ = GetSessionTags(sessionID)
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", len(tags))
	}

	if tags[0] != "important" {
		t.Errorf("Expected remaining tag to be 'important', got '%s'", tags[0])
	}
}
