package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"pentlog/pkg/logs"
	"pentlog/pkg/utils"

	"github.com/go-chi/chi/v5"
)

func SessionContentRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/timeline", handleSessionTimeline)
	r.Get("/{id}/notes", handleSessionNotes)
	r.Post("/{id}/notes", handleSessionNoteAdd)
	r.Get("/{id}/content", handleSessionContent)
	r.Get("/{id}/metadata", handleSessionMetadata)
	return r
}

func mountSessionContentRoutes(r chi.Router) {
	r.Get("/{id}/timeline", handleSessionTimeline)
	r.Get("/{id}/notes", handleSessionNotes)
	r.Post("/{id}/notes", handleSessionNoteAdd)
	r.Get("/{id}/content", handleSessionContent)
	r.Get("/{id}/metadata", handleSessionMetadata)
}

func parseSessionID(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func loadSessionByRequest(r *http.Request) (*logs.Session, error) {
	id, err := parseSessionID(r)
	if err != nil {
		return nil, err
	}
	return logs.GetSession(id)
}

func handleSessionTimeline(w http.ResponseWriter, r *http.Request) {
	session, err := loadSessionByRequest(r)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	timeline, err := logs.ParseTimeline(session.Path)
	if err != nil {
		http.Error(w, `{"error":"Failed to parse session timeline"}`, http.StatusInternalServerError)
		return
	}

	for index := range timeline.Commands {
		timeline.Commands[index].Command = normalizeTerminalText(timeline.Commands[index].Command)
		timeline.Commands[index].Output = normalizeTerminalText(timeline.Commands[index].Output)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"commands":   timeline.Commands,
	})
}

func handleSessionNotes(w http.ResponseWriter, r *http.Request) {
	session, err := loadSessionByRequest(r)
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
	session, err := loadSessionByRequest(r)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, `{"error":"Content is required"}`, http.StatusBadRequest)
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

func handleSessionContent(w http.ResponseWriter, r *http.Request) {
	session, err := loadSessionByRequest(r)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	content, err := readSessionContent(session.Path)
	if err != nil {
		http.Error(w, `{"error":"Failed to read session content"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content":     content,
		"total_bytes": len(content),
		"has_more":    false,
	})
}

func handleSessionMetadata(w http.ResponseWriter, r *http.Request) {
	session, err := loadSessionByRequest(r)
	if err != nil {
		http.Error(w, `{"error":"Session not found"}`, http.StatusNotFound)
		return
	}

	if session.MetaPath == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.Metadata)
		return
	}

	data, err := os.ReadFile(session.MetaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(session.Metadata)
			return
		}
		http.Error(w, `{"error":"Failed to read metadata"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func readSessionContent(path string) (string, error) {
	if strings.HasSuffix(path, ".tty") {
		if timeline, err := logs.ParseTimeline(path); err == nil && len(timeline.Commands) > 0 {
			return renderTimelineTranscript(timeline), nil
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".tty") {
		reader = logs.NewTtyReader(f)
	}

	rawData, err := io.ReadAll(reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return normalizeTerminalText(string(rawData)), nil
}

func normalizeTerminalText(raw string) string {
	cleanData := utils.CleanTuiMarkers([]byte(raw))
	lines := strings.Split(string(cleanData), "\n")

	var builder strings.Builder
	for index, line := range lines {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(utils.RenderPlain(line))
	}

	return strings.TrimSpace(builder.String())
}

func renderTimelineTranscript(timeline *logs.Timeline) string {
	var builder strings.Builder

	for index, entry := range timeline.Commands {
		if index > 0 {
			builder.WriteString("\n\n")
		}

		command := normalizeTerminalText(entry.Command)
		output := normalizeTerminalText(entry.Output)

		if command != "" {
			builder.WriteString("$ ")
			builder.WriteString(command)
		}

		if output != "" {
			if command != "" {
				builder.WriteByte('\n')
			}
			builder.WriteString(output)
		}
	}

	return strings.TrimSpace(builder.String())
}
