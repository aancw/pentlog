package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func SessionRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleSessionsList)
	r.Get("/tags", handleTagsList)
	r.Get("/{id}", handleSessionByID)
	r.Delete("/{id}", handleDeleteSessionByID)
	return r
}

func handleSessionsList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	tag := r.URL.Query().Get("tag")
	client := r.URL.Query().Get("client")
	phase := r.URL.Query().Get("phase")
	state := r.URL.Query().Get("state")

	if limit == 0 {
		limit = 20
	}

	var sessions []logs.Session
	var err error

	if tag != "" {
		sessions, err = logs.ListSessionsByTag(tag)
	} else {
		sessions, err = logs.ListSessionsPaginated(limit+1, offset)
	}

	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	if client != "" {
		filtered := []logs.Session{}
		for _, s := range sessions {
			if s.Metadata.Client == client {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	if phase != "" {
		filtered := []logs.Session{}
		for _, s := range sessions {
			if s.Metadata.Phase == phase {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	if state != "" {
		filtered := []logs.Session{}
		for _, s := range sessions {
			if string(s.State) == state {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	hasMore := len(sessions) > limit
	if hasMore {
		sessions = sessions[:limit]
	}

	totalCount := len(sessions)
	allSessions, _ := logs.ListSessions()
	if allSessions != nil {
		totalCount = len(allSessions)
	}

	var respSessions []map[string]interface{}
	for _, s := range sessions {
		respSessions = append(respSessions, sessionToMap(s))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": respSessions,
		"total":    totalCount,
		"page":     offset/limit + 1,
		"has_more": hasMore,
	})
}

func handleSessionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	session, err := logs.GetSession(id)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionToMap(*session))
}

func handleDeleteSessionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	session, err := logs.GetSession(id)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	filesToDelete := []string{
		session.Path,
		session.Path + ".json",
		session.Path + ".notes.json",
	}

	deletedFiles := 0
	for _, file := range filesToDelete {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err == nil {
				deletedFiles++
			}
		}
	}

	if err := logs.DeleteSession(id); err != nil {
		http.Error(w, `{"error":"Failed to delete session from database"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Session deleted successfully",
		"deleted_files": deletedFiles,
		"session_id":    id,
	})
}

func handleTagsList(w http.ResponseWriter, r *http.Request) {
	tags, err := logs.ListAllTags()
	if err != nil {
		http.Error(w, `{"error":"Failed to list tags"}`, http.StatusInternalServerError)
		return
	}

	tagCounts := make(map[string]int)
	for _, tag := range tags {
		sessions, _ := logs.ListSessionsByTag(tag)
		tagCounts[tag] = len(sessions)
	}

	var tagResponses []map[string]interface{}
	for name, count := range tagCounts {
		tagResponses = append(tagResponses, map[string]interface{}{
			"name":  name,
			"count": count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tags": tagResponses})
}
