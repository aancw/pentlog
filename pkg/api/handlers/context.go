package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

func ContextRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/create", handleContextCreate)
	r.Get("/current", handleContextCurrent)
	r.Get("/history", handleContextHistory)
	r.Put("/current", handleContextUpdate)
	r.Delete("/reset", handleContextReset)
	return r
}

func handleContextCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type       string `json:"type"`
		Client     string `json:"client"`
		Engagement string `json:"engagement"`
		Scope      string `json:"scope"`
		Operator   string `json:"operator"`
		Phase      string `json:"phase"`
		Target     string `json:"target"`
		TargetIP   string `json:"target_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Type = strings.TrimSpace(req.Type)
	req.Client = strings.TrimSpace(req.Client)
	req.Engagement = strings.TrimSpace(req.Engagement)
	req.Scope = strings.TrimSpace(req.Scope)
	req.Operator = strings.TrimSpace(req.Operator)
	req.Phase = strings.TrimSpace(req.Phase)
	req.Target = strings.TrimSpace(req.Target)
	req.TargetIP = strings.TrimSpace(req.TargetIP)

	if req.Type == "" {
		req.Type = "Client"
	}
	if req.Client == "" || req.Engagement == "" || req.Phase == "" {
		http.Error(w, `{"error":"client, engagement, and phase are required"}`, http.StatusBadRequest)
		return
	}

	ctx := &config.ContextData{
		Client:     req.Client,
		Engagement: req.Engagement,
		Scope:      req.Scope,
		Operator:   req.Operator,
		Phase:      req.Phase,
		Target:     req.Target,
		TargetIP:   req.TargetIP,
		Timestamp:  time.Now().Format(time.RFC3339),
		Type:       req.Type,
	}

	mgr := config.Manager()
	if err := mgr.SaveContext(ctx); err != nil {
		http.Error(w, `{"error":"Failed to save context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Context created",
		"context": ctx,
	})
}

func handleContextCurrent(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_context": false,
			"context":     nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_context": true,
		"context":     ctx,
	})
}

func handleContextHistory(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	history, err := mgr.LoadContextHistory()
	if err != nil {
		http.Error(w, `{"error":"Failed to load history"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"history": history})
}

func handleContextUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phase    string `json:"phase"`
		Target   string `json:"target"`
		TargetIP string `json:"target_ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	if phase := strings.TrimSpace(req.Phase); phase != "" {
		ctx.Phase = phase
	}
	if target := strings.TrimSpace(req.Target); target != "" {
		ctx.Target = target
	}
	if targetIP := strings.TrimSpace(req.TargetIP); targetIP != "" {
		ctx.TargetIP = targetIP
	}
	ctx.Timestamp = time.Now().Format(time.RFC3339)

	if err := mgr.SaveContext(ctx); err != nil {
		http.Error(w, `{"error":"Failed to save context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Context updated",
		"context": ctx,
	})
}

func handleContextReset(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	paths := mgr.GetPaths()

	if err := os.Remove(paths.ContextFile); err != nil && !os.IsNotExist(err) {
		http.Error(w, `{"error":"Failed to reset context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Context reset successfully"})
}
