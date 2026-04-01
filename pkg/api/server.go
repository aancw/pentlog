package api

import (
	"context"
	"fmt"
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
	s.Router.Use(middleware.StripSlashes)

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
		r.Mount("/system", handlers.SystemRoutes())
	})
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
