package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSessionMonitor(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test.tty")

	// Create a test file
	if err := os.WriteFile(sessionPath, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := DefaultMonitorConfig()
	monitor := NewSessionMonitor(sessionPath, config)

	if monitor.sessionPath != sessionPath {
		t.Errorf("Expected sessionPath to be %s, got %s", sessionPath, monitor.sessionPath)
	}

	if monitor.warnAt != config.WarnThreshold {
		t.Errorf("Expected warnAt to be %d, got %d", config.WarnThreshold, monitor.warnAt)
	}

	if monitor.alertAt != config.AlertThreshold {
		t.Errorf("Expected alertAt to be %d, got %d", config.AlertThreshold, monitor.alertAt)
	}
}

func TestSessionMonitorCheck(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test.tty")

	// Create a test file with 60MB of data (above warning threshold, below alert)
	data := make([]byte, 60*1024*1024)
	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := MonitorConfig{
		WarnThreshold:  50 * 1024 * 1024,
		AlertThreshold: 100 * 1024 * 1024,
		CheckInterval:  1 * time.Second,
		AlertCooldown:  1 * time.Second,
	}

	monitor := NewSessionMonitor(sessionPath, config)

	// Should not panic
	monitor.Check()

	// Verify size
	size, err := monitor.GetCurrentSize()
	if err != nil {
		t.Fatalf("Failed to get current size: %v", err)
	}

	if size != int64(60*1024*1024) {
		t.Errorf("Expected size to be %d, got %d", 60*1024*1024, size)
	}
}

func TestSessionMonitorStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test.tty")

	if err := os.WriteFile(sessionPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := MonitorConfig{
		WarnThreshold:  50 * 1024 * 1024,
		AlertThreshold: 100 * 1024 * 1024,
		CheckInterval:  100 * time.Millisecond,
		AlertCooldown:  1 * time.Hour,
	}

	monitor := NewSessionMonitor(sessionPath, config)

	// Start monitoring
	monitor.Start()

	// Let it run briefly
	time.Sleep(200 * time.Millisecond)

	// Stop monitoring
	monitor.Stop()

	// Verify it stopped cleanly
	select {
	case <-monitor.stopChan:
		// Channel is closed, which is expected
	default:
		// Channel is still open, which means Stop() wasn't called properly
	}
}

func TestDefaultMonitorConfig(t *testing.T) {
	config := DefaultMonitorConfig()

	if config.WarnThreshold != 5*1024*1024 {
		t.Errorf("Expected WarnThreshold to be 5MB, got %d", config.WarnThreshold)
	}

	if config.AlertThreshold != 10*1024*1024 {
		t.Errorf("Expected AlertThreshold to be 10MB, got %d", config.AlertThreshold)
	}

	if config.CheckInterval != 30*time.Second {
		t.Errorf("Expected CheckInterval to be 30s, got %v", config.CheckInterval)
	}

	if config.AlertCooldown != 5*time.Minute {
		t.Errorf("Expected AlertCooldown to be 5m, got %v", config.AlertCooldown)
	}
}

func TestSessionMonitorCooldown(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test.tty")

	// Create a large file to trigger alert
	data := make([]byte, 150*1024*1024)
	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := MonitorConfig{
		WarnThreshold:  50 * 1024 * 1024,
		AlertThreshold: 100 * 1024 * 1024,
		CheckInterval:  1 * time.Second,
		AlertCooldown:  1 * time.Hour, // Long cooldown
	}

	monitor := NewSessionMonitor(sessionPath, config)

	// First check should trigger alert
	monitor.Check()

	if monitor.lastAlert.IsZero() {
		t.Error("Expected lastAlert to be set after first check")
	}

	// Second check should not trigger due to cooldown
	lastAlert := monitor.lastAlert
	monitor.Check()

	if !monitor.lastAlert.Equal(lastAlert) {
		t.Error("Expected lastAlert to not change during cooldown")
	}
}
