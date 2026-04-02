package handlers

import (
	"encoding/json"
	"net/http"
	"os"
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
	crashed, err := logs.GetCrashedSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to get crashed sessions"}`, http.StatusInternalServerError)
		return
	}

	active, err := logs.GetActiveSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to get active sessions"}`, http.StatusInternalServerError)
		return
	}

	orphaned, err := logs.GetOrphanedSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to get orphaned sessions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"crashed":  sessionsToList(crashed),
		"active":   sessionsToList(active),
		"orphaned": sessionsToList(orphaned),
	})
}

func handleRecoveryMarkStale(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TimeoutMinutes int `json:"timeout_minutes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	timeout := 5 * time.Minute
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
	crashed, err := logs.GetCrashedSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to get crashed sessions"}`, http.StatusInternalServerError)
		return
	}

	recovered := 0
	failed := 0
	for _, s := range crashed {
		if err := logs.RecoverSession(s.ID); err == nil {
			recovered++
		} else {
			failed++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recovered_count": recovered,
		"failed_count":    failed,
		"total_count":     len(crashed),
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
