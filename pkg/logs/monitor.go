package logs

import (
	"fmt"
	"os"
	"sync"
	"time"

	"pentlog/pkg/utils"
)

// SessionMonitor monitors session file size and alerts when thresholds are exceeded
type SessionMonitor struct {
	sessionPath   string
	alertAt       int64 // bytes - critical threshold
	warnAt        int64 // bytes - warning threshold
	checkInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	lastAlert     time.Time
	alertCooldown time.Duration
}

// MonitorConfig holds configuration for session monitoring
type MonitorConfig struct {
	WarnThreshold  int64         // Warning threshold in bytes (default: 50MB)
	AlertThreshold int64         // Alert threshold in bytes (default: 100MB)
	CheckInterval  time.Duration // How often to check (default: 30s)
	AlertCooldown  time.Duration // Minimum time between alerts (default: 5m)
}

// DefaultMonitorConfig returns default monitoring configuration
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		WarnThreshold:  5 * 1024 * 1024,  // 5MB
		AlertThreshold: 10 * 1024 * 1024, // 10MB
		CheckInterval:  30 * time.Second,
		AlertCooldown:  5 * time.Minute,
	}
}

// NewSessionMonitor creates a new session monitor
func NewSessionMonitor(sessionPath string, config MonitorConfig) *SessionMonitor {
	return &SessionMonitor{
		sessionPath:   sessionPath,
		warnAt:        config.WarnThreshold,
		alertAt:       config.AlertThreshold,
		checkInterval: config.CheckInterval,
		stopChan:      make(chan struct{}),
		alertCooldown: config.AlertCooldown,
	}
}

// Start begins monitoring the session file in a background goroutine
func (sm *SessionMonitor) Start() {
	sm.wg.Add(1)
	go sm.monitor()
}

// Stop stops the monitoring goroutine
func (sm *SessionMonitor) Stop() {
	close(sm.stopChan)
	sm.wg.Wait()
}

// monitor runs the monitoring loop
func (sm *SessionMonitor) monitor() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.check()
		case <-sm.stopChan:
			return
		}
	}
}

// check evaluates the current session file size and emits alerts if needed
func (sm *SessionMonitor) Check() {
	sm.check()
}

func (sm *SessionMonitor) check() {
	info, err := os.Stat(sm.sessionPath)
	if err != nil {
		return
	}

	size := info.Size()
	now := time.Now()

	// Check if we're still in cooldown period
	if now.Sub(sm.lastAlert) < sm.alertCooldown {
		return
	}

	if size > sm.alertAt {
		fmt.Fprintf(os.Stderr,
			"\n⚠️  Session size: %s - Consider splitting session (use 'exit' and start new)\n",
			utils.FormatBytes(size))
		sm.lastAlert = now
	} else if size > sm.warnAt {
		fmt.Fprintf(os.Stderr,
			"\n⚡ Session size: %s - Approaching limit\n",
			utils.FormatBytes(size))
		sm.lastAlert = now
	}
}

// GetCurrentSize returns the current size of the session file
func (sm *SessionMonitor) GetCurrentSize() (int64, error) {
	info, err := os.Stat(sm.sessionPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
