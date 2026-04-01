package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

func SessionContentRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/notes", handleSessionNotes)
	r.Post("/{id}/notes", handleSessionNoteAdd)
	return r
}

func handleSessionNotes(w http.ResponseWriter, r *http.Request) {
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

	if session.NotesPath == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"notes": []interface{}{}})
		return
	}

	notes, err := logs.ReadNotes(session.NotesPath)
	if err != nil {
		http.Error(w, `{"error":"Failed to read notes"}`, http.StatusInternalServerError)
		return
	}

	var result []map[string]interface{}
	for _, n := range notes {
		result = append(result, map[string]interface{}{
			"timestamp":   n.Timestamp,
			"content":     n.Content,
			"byte_offset": n.ByteOffset,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"notes": result})
}

func handleSessionNoteAdd(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error":"Content is required"}`, http.StatusBadRequest)
		return
	}

	session, err := logs.GetSession(id)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	note := logs.SessionNote{
		Timestamp:  session.ModTime,
		Content:    req.Content,
		ByteOffset: 0,
	}

	if err := logs.AppendNote(session.NotesPath, note); err != nil {
		http.Error(w, `{"error":"Failed to add note"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Note added",
		"note": map[string]interface{}{
			"timestamp": note.Timestamp,
			"content":   note.Content,
		},
	})
}
