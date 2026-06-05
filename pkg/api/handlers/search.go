package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pentlog/pkg/logs"
	"pentlog/pkg/search"
	"pentlog/pkg/utils"

	"github.com/go-chi/chi/v5"
)

func SearchRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleSearch)
	return r
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"Query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	isRegex := r.URL.Query().Get("regex") == "true"
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	includeArchived := r.URL.Query().Get("include_archived") == "true"

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 500 {
		limit = 500
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	sessions, err := logs.ListSessionsWithOptions(logs.SessionListOptions{
		IncludeArchived: includeArchived,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	opts := search.SearchOptions{
		IsRegex: isRegex,
		Limit:   limit,
		Offset:  offset,
	}

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			opts.After = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			opts.Before = t.Add(24*time.Hour - time.Nanosecond)
		}
	}

	page, err := search.SearchPage(query, sessions, opts)
	if err != nil {
		http.Error(w, `{"error":"Search failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for _, m := range page.Matches {
		results = append(results, map[string]interface{}{
			"session_id":         m.Session.ID,
			"session_path":       m.Session.DisplayPath,
			"line_num":           m.LineNum,
			"content":            utils.StripANSI(m.Content),
			"context":            m.Context,
			"context_start_line": m.ContextStartLine,
			"is_note":            m.IsNote,
			"note_timestamp":     m.NoteTimestamp,
		})
	}

	resp := map[string]interface{}{
		"results":       results,
		"total_matches": page.Total,
		"query":         query,
		"is_regex":      isRegex,
		"limit":         limit,
		"offset":        offset,
		"has_more":      offset+len(results) < page.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
