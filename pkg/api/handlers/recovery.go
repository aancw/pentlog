package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func RecoveryRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", handleRecoveryStatus)
	r.Post("/mark-stale", handleRecoveryMarkStale)
	r.Post("/recover-all", handleRecoveryRecoverAll)
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
	count, err := logs.MarkStaleSessions(5 * time.Minute)
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
	for _, s := range crashed {
		if err := logs.RecoverSession(s.ID); err == nil {
			recovered++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recovered_count": recovered,
		"total_count":     len(crashed),
		"message":         "Recovered crashed sessions",
	})
}

func sessionsToList(sessions []logs.Session) []map[string]interface{} {
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, sessionToMap(s))
	}
	return result
}
