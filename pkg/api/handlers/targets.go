package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

func TargetRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleTargetsList)
	r.Post("/", handleTargetCreate)
	r.Put("/{name}/switch", handleTargetSwitch)
	r.Delete("/{name}", handleTargetDelete)
	r.Delete("/clear", handleTargetClear)
	return r
}

func handleTargetsList(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	targets, err := mgr.LoadTargets()
	if err != nil {
		http.Error(w, `{"error":"Failed to load targets"}`, http.StatusInternalServerError)
		return
	}

	ctx, _ := mgr.LoadContext()

	var result []map[string]interface{}
	var current *map[string]interface{}
	for _, t := range targets.Targets {
		item := map[string]interface{}{
			"name":       t.Name,
			"ip":         t.IP,
			"is_current": false,
		}
		if ctx != nil && t.Name == ctx.Target {
			item["is_current"] = true
			current = &item
		}
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i]["name"].(string)) < strings.ToLower(result[j]["name"].(string))
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"targets": result,
		"current": current,
	})
}

func handleTargetCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.IP = strings.TrimSpace(req.IP)
	if req.Name == "" {
		http.Error(w, `{"error":"Name is required"}`, http.StatusBadRequest)
		return
	}

	mgr := config.Manager()
	targets, err := mgr.LoadTargets()
	if err != nil {
		http.Error(w, `{"error":"Failed to load targets"}`, http.StatusInternalServerError)
		return
	}

	for _, t := range targets.Targets {
		if strings.EqualFold(t.Name, req.Name) {
			http.Error(w, `{"error":"Target already exists"}`, http.StatusBadRequest)
			return
		}
	}

	targets.Targets = append(targets.Targets, config.Target{Name: req.Name, IP: req.IP})
	if err := mgr.SaveTargets(targets); err != nil {
		http.Error(w, `{"error":"Failed to save target"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Target created",
		"target":  map[string]string{"name": req.Name, "ip": req.IP},
	})
}

func handleTargetSwitch(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	targets, err := mgr.LoadTargets()
	if err != nil {
		http.Error(w, `{"error":"Failed to load targets"}`, http.StatusInternalServerError)
		return
	}

	var found *config.Target
	for _, t := range targets.Targets {
		if strings.EqualFold(t.Name, name) {
			t := t
			found = &t
			break
		}
	}
	if found == nil {
		http.Error(w, `{"error":"Target not found"}`, http.StatusNotFound)
		return
	}

	ctx.Target = found.Name
	ctx.TargetIP = found.IP
	ctx.Timestamp = time.Now().Format(time.RFC3339)

	if err := mgr.SaveContext(ctx); err != nil {
		http.Error(w, `{"error":"Failed to save context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Switched to target",
		"target":  map[string]string{"name": found.Name, "ip": found.IP},
	})
}

func handleTargetDelete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	mgr := config.Manager()
	targets, err := mgr.LoadTargets()
	if err != nil {
		http.Error(w, `{"error":"Failed to load targets"}`, http.StatusInternalServerError)
		return
	}

	var remaining []config.Target
	found := false
	for _, t := range targets.Targets {
		if strings.EqualFold(t.Name, name) {
			found = true
			continue
		}
		remaining = append(remaining, t)
	}
	if !found {
		http.Error(w, `{"error":"Target not found"}`, http.StatusNotFound)
		return
	}

	targets.Targets = remaining
	if err := mgr.SaveTargets(targets); err != nil {
		http.Error(w, `{"error":"Failed to save targets"}`, http.StatusInternalServerError)
		return
	}

	ctx, _ := mgr.LoadContext()
	if ctx != nil && strings.EqualFold(ctx.Target, name) {
		ctx.Target = ""
		ctx.TargetIP = ""
		ctx.Timestamp = time.Now().Format(time.RFC3339)
		_ = mgr.SaveContext(ctx)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Target deleted"})
}

func handleTargetClear(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	ctx.Target = ""
	ctx.TargetIP = ""
	ctx.Timestamp = time.Now().Format(time.RFC3339)
	if err := mgr.SaveContext(ctx); err != nil {
		http.Error(w, `{"error":"Failed to update context"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Cleared active target"})
}
