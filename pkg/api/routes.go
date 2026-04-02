package api

import (
	"net/http"

	"pentlog/pkg/api/handlers"

	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Mount("/dashboard", handlers.DashboardRoutes())
	r.Mount("/sessions", handlers.SessionRoutes())
	r.Mount("/session-content", handlers.SessionContentRoutes())
	r.Mount("/system", handlers.SystemRoutes())
	r.Mount("/vulns", handlers.VulnRoutes())
	r.Mount("/search", handlers.SearchRoutes())
	r.Mount("/reports", handlers.ReportRoutes())
	r.Mount("/archives", handlers.ArchiveRoutes())
	r.Mount("/context", handlers.ContextRoutes())
	r.Mount("/contexts", handlers.ContextRoutes())
	r.Mount("/targets", handlers.TargetRoutes())
	r.Mount("/recovery", handlers.RecoveryRoutes())
	r.Mount("/share", handlers.ShareRoutes())
}
