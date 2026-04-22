package vulns

import (
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"testing"
	"time"
)

func TestList_Aggregation(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	// Setup temporary directory structure
	tmpDir, err := os.MkdirTemp("", "pentlog_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock config via environment variable
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	client := "TestClient"
	eng1 := "Eng1"
	eng2 := "Eng2"

	// Create vulns for Eng1
	mgr1 := NewManager(client, eng1)
	v1 := Vuln{
		ID:        "vuln-001",
		Title:     "Title 1",
		Severity:  SeverityHigh,
		Status:    StatusOpen,
		CreatedAt: time.Now(),
	}
	if err := mgr1.Save(v1); err != nil {
		t.Fatalf("Failed to save v1: %v", err)
	}

	// Create vulns for Eng2
	mgr2 := NewManager(client, eng2)
	v2 := Vuln{
		ID:        "vuln-002",
		Title:     "Title 2",
		Severity:  SeverityLow,
		Status:    StatusClosed,
		CreatedAt: time.Now().Add(time.Hour),
	}
	if err := mgr2.Save(v2); err != nil {
		t.Fatalf("Failed to save v2: %v", err)
	}

	// Test Aggregation (Empty Engagement)
	mgrAgg := NewManager(client, "")
	vulns, err := mgrAgg.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(vulns) != 2 {
		t.Errorf("Expected 2 aggregated vulns, got %d", len(vulns))
	}

	foundV1 := false
	foundV2 := false
	for _, v := range vulns {
		if v.ID == "vuln-001" {
			foundV1 = true
		}
		if v.ID == "vuln-002" {
			foundV2 = true
		}
	}

	if !foundV1 || !foundV2 {
		t.Errorf("Failed to find all vulns. FoundV1: %v, FoundV2: %v", foundV1, foundV2)
	}
}

func TestSavePermissions(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	tmpDir, err := os.MkdirTemp("", "pentlog_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := NewManager("ClientA", "EngA")
	if err := mgr.Save(Vuln{
		ID:        "vuln-001",
		Title:     "Sensitive file disclosure",
		Severity:  SeverityHigh,
		Status:    StatusOpen,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	vulnFile, err := mgr.GetVulnsFile()
	if err != nil {
		t.Fatalf("GetVulnsFile() failed: %v", err)
	}

	info, err := os.Stat(vulnFile)
	if err != nil {
		t.Fatalf("Failed to stat vuln file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected vuln file permissions 0600, got %#o", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(vulnFile))
	if err != nil {
		t.Fatalf("Failed to stat vuln dir: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Fatalf("expected vuln dir permissions 0700, got %#o", got)
	}
}
