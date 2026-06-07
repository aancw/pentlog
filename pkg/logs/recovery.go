package logs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"pentlog/pkg/config"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type RecoveryDisposition string

const (
	RecoveryDispositionLikelyLive  RecoveryDisposition = "likely_live"
	RecoveryDispositionNeedsReview RecoveryDisposition = "needs_review"
	RecoveryDispositionStale       RecoveryDisposition = "stale"
	RecoveryDispositionCrashed     RecoveryDisposition = "crashed"
)

type RecoveryCandidate struct {
	Session       Session
	Disposition   RecoveryDisposition
	Reason        string
	LastSeenAt    string
	LastSeenAge   string
	RecorderAlive bool
}

type RecoveryOverview struct {
	Timeout      time.Duration
	Active       []RecoveryCandidate
	Paused       []RecoveryCandidate
	ReviewNeeded []RecoveryCandidate
	Stale        []RecoveryCandidate
	Crashed      []RecoveryCandidate
	Orphaned     []Session
}

type hostIdentity struct {
	Hostname    string
	Fingerprint string
}

func GetRecoveryOverview(timeout time.Duration) (RecoveryOverview, error) {
	overview := RecoveryOverview{Timeout: timeout}

	active, err := GetActiveSessions()
	if err != nil {
		return overview, err
	}

	paused, err := GetPausedSessions()
	if err != nil {
		return overview, err
	}

	crashed, err := GetCrashedSessions()
	if err != nil {
		return overview, err
	}

	orphans, err := GetOrphanedSessions()
	if err != nil {
		return overview, err
	}

	for _, session := range crashed {
		overview.Crashed = append(overview.Crashed, buildRecoveryCandidate(session, RecoveryDispositionCrashed, "Marked crashed in session state", false))
	}

	for _, session := range active {
		candidate := classifyLiveSession(session, timeout)
		switch candidate.Disposition {
		case RecoveryDispositionLikelyLive:
			overview.Active = append(overview.Active, candidate)
		case RecoveryDispositionNeedsReview:
			overview.ReviewNeeded = append(overview.ReviewNeeded, candidate)
		case RecoveryDispositionStale:
			overview.Stale = append(overview.Stale, candidate)
		}
	}

	for _, session := range paused {
		candidate := classifyLiveSession(session, timeout)
		switch candidate.Disposition {
		case RecoveryDispositionLikelyLive:
			overview.Paused = append(overview.Paused, candidate)
		case RecoveryDispositionNeedsReview:
			overview.ReviewNeeded = append(overview.ReviewNeeded, candidate)
		case RecoveryDispositionStale:
			overview.Stale = append(overview.Stale, candidate)
		}
	}

	overview.Orphaned = orphans
	return overview, nil
}

func buildRecoveryCandidate(session Session, disposition RecoveryDisposition, reason string, recorderAlive bool) RecoveryCandidate {
	candidate := RecoveryCandidate{
		Session:       session,
		Disposition:   disposition,
		Reason:        reason,
		LastSeenAt:    session.LastSyncAt,
		RecorderAlive: recorderAlive,
	}
	if parsed, err := time.Parse(time.RFC3339, session.LastSyncAt); err == nil {
		candidate.LastSeenAge = time.Since(parsed).Round(time.Second).String()
	}
	return candidate
}

func classifyLiveSession(session Session, timeout time.Duration) RecoveryCandidate {
	sameHost := session.HostFingerprint != "" && session.HostFingerprint == currentHostIdentity().Fingerprint
	recorderKnown := session.RecorderPID > 0
	recorderAlive := false
	if sameHost && recorderKnown {
		recorderAlive = processExists(session.RecorderPID)
		if recorderAlive {
			return buildRecoveryCandidate(session, RecoveryDispositionLikelyLive, fmt.Sprintf("Recorder PID %d is still running on this host", session.RecorderPID), true)
		}
		return buildRecoveryCandidate(session, RecoveryDispositionStale, fmt.Sprintf("Recorder PID %d is no longer running on this host", session.RecorderPID), false)
	}

	if lastSeen, err := time.Parse(time.RFC3339, session.LastSyncAt); err == nil {
		age := time.Since(lastSeen)
		if age <= timeout {
			return buildRecoveryCandidate(session, RecoveryDispositionLikelyLive, fmt.Sprintf("Last heartbeat was %s ago", age.Round(time.Second)), false)
		}
	}

	switch {
	case session.HostFingerprint != "" && !sameHost:
		hostLabel := session.Hostname
		if strings.TrimSpace(hostLabel) == "" {
			hostLabel = "another host"
		}
		return buildRecoveryCandidate(session, RecoveryDispositionNeedsReview, fmt.Sprintf("Heartbeat is stale, but the session was recorded on %s so the recorder PID cannot be verified locally", hostLabel), false)
	case !recorderKnown:
		return buildRecoveryCandidate(session, RecoveryDispositionNeedsReview, "Heartbeat is stale, but this session has no recorder PID to verify", false)
	default:
		return buildRecoveryCandidate(session, RecoveryDispositionNeedsReview, "Heartbeat is stale and the recorder cannot be verified safely", false)
	}
}

func currentHostIdentity() hostIdentity {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "unknown-host"
	}

	seed := strings.Join([]string{
		hostname,
		runtime.GOOS,
		os.Getenv("USER"),
		config.Manager().GetPaths().Home,
	}, "|")
	sum := sha256.Sum256([]byte(seed))

	return hostIdentity{
		Hostname:    hostname,
		Fingerprint: hex.EncodeToString(sum[:]),
	}
}

func defaultLifecycleHost(state SessionState) hostIdentity {
	switch state {
	case SessionStateActive, SessionStatePaused:
		return currentHostIdentity()
	default:
		return hostIdentity{}
	}
}

func defaultLifecycleTimestamps(state SessionState, sessionTimestamp, now string) (string, string) {
	startedAt := strings.TrimSpace(sessionTimestamp)
	if startedAt == "" {
		startedAt = now
	}

	switch state {
	case SessionStateCompleted, SessionStateCrashed, SessionStateArchived:
		return startedAt, startedAt
	default:
		return startedAt, ""
	}
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	return errors.Is(err, syscall.EPERM)
}
