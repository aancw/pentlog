package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func SessionRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleSessionsList)
	r.Get("/tags", handleTagsList)
	r.Get("/{id}", handleSessionByID)
	r.Delete("/{id}", handleDeleteSessionByID)
	mountSessionContentRoutes(r)
	return r
}

func handleSessionsList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	client := strings.TrimSpace(r.URL.Query().Get("client"))
	phase := strings.TrimSpace(r.URL.Query().Get("phase"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	includeArchived := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_archived")), "true")

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	sessions, err := logs.ListSessionsWithOptions(logs.SessionListOptions{
		IncludeArchived: includeArchived || strings.EqualFold(state, string(logs.SessionStateArchived)),
		OnlyArchived:    strings.EqualFold(state, string(logs.SessionStateArchived)),
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	filtered := make([]logs.Session, 0, len(sessions))
	for _, s := range sessions {
		if tag != "" {
			tags, _ := logs.GetSessionTags(s.ID)
			matched := false
			for _, existing := range tags {
				if strings.EqualFold(existing, tag) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		if client != "" && !strings.Contains(strings.ToLower(s.Metadata.Client), strings.ToLower(client)) {
			continue
		}
		if phase != "" && !strings.EqualFold(s.Metadata.Phase, phase) {
			continue
		}
		if state != "" && !strings.EqualFold(string(s.State), state) {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{
				s.Filename,
				s.DisplayPath,
				s.Metadata.Client,
				s.Metadata.Engagement,
				s.Metadata.Operator,
				s.Metadata.Phase,
				s.Metadata.Target,
			}, " "))
			if !strings.Contains(haystack, strings.ToLower(query)) {
				continue
			}
		}

		filtered = append(filtered, s)
	}

	totalCount := len(filtered)
	if offset > totalCount {
		offset = totalCount
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}
	pageSessions := filtered[offset:end]
	hasMore := end < totalCount

	var respSessions []map[string]interface{}
	for _, s := range pageSessions {
		respSessions = append(respSessions, sessionToMap(s))
	}

	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": respSessions,
		"total":    totalCount,
		"page":     page,
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
