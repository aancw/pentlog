package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"pentlog/pkg/config"
	"pentlog/pkg/logger"

	"github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

var hasStaticFiles bool
var localStaticDir string

func SetStaticDir(dir string) {
	indexPath := filepath.Join(dir, "index.html")
	if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
		localStaticDir = dir
		hasStaticFiles = true
		logger.Info("static_files_loaded", "source", "disk", "dir", dir)
		return
	}

	logger.Warn("static_files_not_loaded_from_disk", "dir", dir)
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
		MountRoutes(r)
	})

	s.setupArtifactRoutes()

	if hasStaticFiles {
		s.setupStaticRoutes()
	}
}

func (s *Server) setupArtifactRoutes() {
	s.Router.Get("/files/reports/*", s.serveReportFile)
	s.Router.Get("/files/archives/*", s.serveArchiveFile)
}

func (s *Server) serveReportFile(w http.ResponseWriter, r *http.Request) {
	s.serveManagedFile(w, r, config.Manager().GetPaths().ReportsDir, "/files/reports/")
}

func (s *Server) serveArchiveFile(w http.ResponseWriter, r *http.Request) {
	s.serveManagedFile(w, r, config.Manager().GetPaths().ArchiveDir, "/files/archives/")
}

func (s *Server) serveManagedFile(w http.ResponseWriter, r *http.Request, root string, prefix string) {
	rel := strings.TrimPrefix(r.URL.Path, prefix)
	rel = strings.TrimPrefix(filepath.Clean("/"+rel), "/")
	if rel == "" || rel == "." {
		http.NotFound(w, r)
		return
	}

	fullPath := filepath.Join(root, filepath.FromSlash(rel))
	cleanRoot := filepath.Clean(root)
	cleanPath := filepath.Clean(fullPath)
	if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(os.PathSeparator)) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(cleanPath); err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, cleanPath)
}

func (s *Server) setupStaticRoutes() {
	s.Router.Get("/", s.serveIndex)
	s.Router.Get("/index.html", s.serveIndex)
	s.Router.Get("/favicon.svg", s.serveStatic)
	s.Router.Get("/icons.svg", s.serveStatic)
	s.Router.Get("/assets/*", s.serveStatic)

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
	if localStaticDir != "" {
		fullPath := filepath.Join(localStaticDir, filepath.FromSlash(name))
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			if contentType == "" {
				contentType = getContentType(name)
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, fullPath)
			return
		}
	}

	http.NotFound(w, r)
}

func getContentType(name string) string {
	switch {
	case strings.HasSuffix(name, ".js"):
		return "application/javascript"
	case strings.HasSuffix(name, ".css"):
		return "text/css"
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".json"):
		return "application/json"
	default:
		return "application/octet-stream"
	}
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
		fmt.Println("   GET  /api/dashboard/overview")
		fmt.Println("   GET  /api/dashboard/stats")
		fmt.Println("   GET  /api/dashboard/activity")
		fmt.Println("   GET  /api/sessions")
		fmt.Println("   GET  /api/sessions/:id")
		fmt.Println("   GET  /api/system/status")
		fmt.Println("   GET  /api/system/info")
		if hasStaticFiles {
			fmt.Println("\n Web UI: http://localhost:" + fmt.Sprintf("%d", s.Port))
		}
		fmt.Println("\n Press Ctrl+C to stop")

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
