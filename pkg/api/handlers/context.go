package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

func ContextRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/current", handleContextCurrent)
	r.Get("/history", handleContextHistory)
	r.Put("/current", handleContextUpdate)
	r.Delete("/reset", handleContextReset)
	return r
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
		"context": map[string]interface{}{
			"client":     ctx.Client,
			"engagement": ctx.Engagement,
			"scope":      ctx.Scope,
			"operator":   ctx.Operator,
			"phase":      ctx.Phase,
			"target":     ctx.Target,
			"target_ip":  ctx.TargetIP,
			"timestamp":  ctx.Timestamp,
			"type":       ctx.Type,
		},
	})
}

func handleContextHistory(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	history, err := mgr.LoadContextHistory()
	if err != nil {
		http.Error(w, `{"error":"Failed to load history"}`, http.StatusInternalServerError)
		return
	}

	var result []map[string]interface{}
	for _, ctx := range history {
		result = append(result, map[string]interface{}{
			"client":     ctx.Client,
			"engagement": ctx.Engagement,
			"scope":      ctx.Scope,
			"operator":   ctx.Operator,
			"phase":      ctx.Phase,
			"target":     ctx.Target,
			"target_ip":  ctx.TargetIP,
			"timestamp":  ctx.Timestamp,
			"type":       ctx.Type,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"history": result})
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

	if req.Phase != "" {
		ctx.Phase = req.Phase
	}
	if req.Target != "" {
		ctx.Target = req.Target
	}
	if req.TargetIP != "" {
		ctx.TargetIP = req.TargetIP
	}

	if err := mgr.SaveContext(ctx); err != nil {
		http.Error(w, `{"error":"Failed to save context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Context updated",
		"context": map[string]interface{}{
			"client":     ctx.Client,
			"engagement": ctx.Engagement,
			"phase":      ctx.Phase,
			"target":     ctx.Target,
			"target_ip":  ctx.TargetIP,
		},
	})
}

func handleContextReset(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	paths := mgr.GetPaths()

	os.Remove(paths.ContextFile)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Context reset successfully"})
}
