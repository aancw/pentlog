package api

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pentlog/pkg/api/handlers"
	"pentlog/pkg/logger"

	"github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

var StaticFS embed.FS
var hasStaticFiles bool
var distFS fs.FS

func SetStaticFS(fsys embed.FS) {
	StaticFS = fsys
	var err error
	distFS, err = fs.Sub(StaticFS, "dist")
	if err == nil {
		hasStaticFiles = true
		logger.Info("static_files_loaded", "source", "embedded")
	} else {
		logger.Warn("static_files_not_loaded", "error", err)
	}
}

type Server struct {
	Router *chi.Mux
	Port   int
	server *http.Server
}

func NewServer(port int) *Server {
	r := chi.NewRouter()

	s := &Server{
		Router: r,
		Port:   port,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(Recoverer)
	s.Router.Use(LoggerMiddleware())
	s.Router.Use(CORS([]string{"*"}))
	s.Router.Use(middleware.Compress(5))
}

func (s *Server) setupRoutes() {
	s.Router.Route("/api", func(r chi.Router) {
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
		r.Mount("/targets", handlers.TargetRoutes())
		r.Mount("/recovery", handlers.RecoveryRoutes())
	})

	if hasStaticFiles {
		s.setupStaticRoutes()
	}
}

func (s *Server) setupStaticRoutes() {
	s.Router.Get("/", s.serveIndex)
	s.Router.Get("/index.html", s.serveIndex)
	s.Router.Get("/favicon.svg", s.serveStatic)
	s.Router.Get("/icons.svg", s.serveStatic)
	s.Router.Get("/assets/{file}", s.serveStatic)

	s.Router.NotFound(s.serveSPA)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	s.serveFile(w, r, "index.html", "text/html; charset=utf-8")
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	s.serveFile(w, r, path[1:], "")
}

func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	s.serveFile(w, r, "index.html", "text/html; charset=utf-8")
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, name string, contentType string) {
	file, err := distFS.Open(name)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "File error", http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		http.Error(w, "Not a file", http.StatusNotFound)
		return
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.Header().Set("Cache-Control", "no-cache")
	io.Copy(w, file)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		logger.Info("server_starting", "port", s.Port)
		fmt.Printf("\n PentLog Web Dashboard running at http://localhost:%d\n\n", s.Port)
		fmt.Println(" API endpoints:")
		fmt.Println("   GET  /api/health")
		fmt.Println("   GET  /api/dashboard/stats")
		fmt.Println("   GET  /api/dashboard/activity")
		fmt.Println("   GET  /api/sessions")
		fmt.Println("   GET  /api/sessions/:id")
		fmt.Println("   GET  /api/system/status")
		fmt.Println("   GET  /api/system/info")
		if hasStaticFiles {
			fmt.Println("\n Web UI: http://localhost:" + fmt.Sprintf("%d", s.Port))
		}
		fmt.Println("\n Press Ctrl+C to stop\n")

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case sig := <-stop:
		logger.Info("server_shutdown", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
