package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"pentlog/pkg/config"
	"strconv"
	"time"

	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func RecoveryRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", handleRecoveryStatus)
	r.Post("/mark-stale", handleRecoveryMarkStale)
	r.Post("/recover-all", handleRecoveryRecoverAll)
	r.Post("/recover/{id}", handleRecoveryRecoverOne)
	r.Delete("/orphans", handleRecoveryDeleteOrphans)
	return r
}

func handleRecoveryStatus(w http.ResponseWriter, r *http.Request) {
	timeout := configuredRecoveryTimeout()
	overview, err := logs.GetRecoveryOverview(timeout)
	if err != nil {
		http.Error(w, `{"error":"Failed to get recovery status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stale_timeout_minutes": int(timeout / time.Minute),
		"active":                recoveryCandidatesToList(overview.Active),
		"paused":                recoveryCandidatesToList(overview.Paused),
		"review_needed":         recoveryCandidatesToList(overview.ReviewNeeded),
		"stale":                 recoveryCandidatesToList(overview.Stale),
		"crashed":               recoveryCandidatesToList(overview.Crashed),
		"orphaned":              sessionsToList(overview.Orphaned),
	})
}

func handleRecoveryMarkStale(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TimeoutMinutes int `json:"timeout_minutes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	timeout := configuredRecoveryTimeout()
	if req.TimeoutMinutes > 0 {
		timeout = time.Duration(req.TimeoutMinutes) * time.Minute
	}

	count, err := logs.MarkStaleSessions(timeout)
	if err != nil {
		http.Error(w, `{"error":"Failed to mark stale sessions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"marked_count": count,
		"message":      "Marked stale sessions as crashed",
	})
}

func handleRecoveryRecoverAll(w http.ResponseWriter, r *http.Request) {
	overview, err := logs.GetRecoveryOverview(configuredRecoveryTimeout())
	if err != nil {
		http.Error(w, `{"error":"Failed to get crashed sessions"}`, http.StatusInternalServerError)
		return
	}

	recovered := 0
	failed := 0
	for _, candidate := range overview.Crashed {
		if err := logs.RecoverSession(candidate.Session.ID); err == nil {
			recovered++
		} else {
			failed++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recovered_count": recovered,
		"failed_count":    failed,
		"total_count":     len(overview.Crashed),
		"message":         "Recovered crashed sessions",
	})
}

func handleRecoveryRecoverOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	if err := logs.RecoverSession(id); err != nil {
		http.Error(w, `{"error":"Failed to recover session"}`, http.StatusInternalServerError)
		return
	}

	session, _ := logs.GetSession(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recovered session",
		"session": sessionToMap(*session),
	})
}

func handleRecoveryDeleteOrphans(w http.ResponseWriter, r *http.Request) {
	orphans, err := logs.GetOrphanedSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to get orphaned sessions"}`, http.StatusInternalServerError)
		return
	}

	deleted := 0
	for _, orphan := range orphans {
		if err := logs.DeleteSession(orphan.ID); err != nil {
			continue
		}
		_ = os.Remove(orphan.Path)
		_ = os.Remove(orphan.MetaPath)
		_ = os.Remove(orphan.NotesPath)
		deleted++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted_count": deleted,
		"message":       "Removed orphaned sessions",
	})
}

func sessionsToList(sessions []logs.Session) []map[string]interface{} {
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, sessionToMap(s))
	}
	return result
}

func recoveryCandidatesToList(candidates []logs.RecoveryCandidate) []map[string]interface{} {
	var result []map[string]interface{}
	for _, candidate := range candidates {
		result = append(result, map[string]interface{}{
			"session":        sessionToMap(candidate.Session),
			"disposition":    string(candidate.Disposition),
			"reason":         candidate.Reason,
			"last_seen_at":   candidate.LastSeenAt,
			"last_seen_age":  candidate.LastSeenAge,
			"recorder_alive": candidate.RecorderAlive,
		})
	}
	return result
}

func configuredRecoveryTimeout() time.Duration {
	cfg := config.Manager().GetMonitor()
	if cfg.StaleTimeoutMin <= 0 {
		return 30 * time.Minute
	}
	return time.Duration(cfg.StaleTimeoutMin) * time.Minute
}
