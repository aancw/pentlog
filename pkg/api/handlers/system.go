package handlers

import (
	"encoding/json"
	"net/http"

	"pentlog/pkg/config"
	"pentlog/pkg/logs"

	"github.com/go-chi/chi/v5"
)

var Version = "dev"

func SystemRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", handleSystemStatus)
	r.Get("/info", handleSystemInfo)
	return r
}

func handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	ctx, err := mgr.LoadContext()

	hasContext := err == nil

	var contextResp map[string]interface{}
	if hasContext {
		contextResp = map[string]interface{}{
			"client":     ctx.Client,
			"engagement": ctx.Engagement,
			"scope":      ctx.Scope,
			"operator":   ctx.Operator,
			"phase":      ctx.Phase,
			"target":     ctx.Target,
			"target_ip":  ctx.TargetIP,
			"timestamp":  ctx.Timestamp,
			"type":       ctx.Type,
		}
	}

	sessions, _ := logs.ListSessions()

	resp := map[string]interface{}{
		"has_context":    hasContext,
		"context":        contextResp,
		"version":        Version,
		"db_path":        mgr.GetPaths().DatabaseFile,
		"total_sessions": len(sessions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	mgr := config.Manager()
	paths := mgr.GetPaths()

	resp := map[string]interface{}{
		"version": Version,
		"paths": map[string]string{
			"home":          paths.Home,
			"logs_dir":      paths.LogsDir,
			"reports_dir":   paths.ReportsDir,
			"archive_dir":   paths.ArchiveDir,
			"database_file": paths.DatabaseFile,
		},
		"uptime": "0s",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
