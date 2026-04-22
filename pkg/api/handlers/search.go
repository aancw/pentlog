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

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	opts := search.SearchOptions{
		IsRegex: isRegex,
		Limit:   limit,
	}

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			opts.After = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			opts.Before = t
		}
	}

	matches, err := search.Search(query, sessions, opts)
	if err != nil {
		http.Error(w, `{"error":"Search failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for _, m := range matches {
		results = append(results, map[string]interface{}{
			"session_id":   m.Session.ID,
			"session_path": m.Session.DisplayPath,
			"line_num":     m.LineNum,
			"content":      utils.StripANSI(m.Content),
			"is_note":      m.IsNote,
		})
	}

	resp := map[string]interface{}{
		"results":       results,
		"total_matches": len(results),
		"query":         query,
		"is_regex":      isRegex,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
