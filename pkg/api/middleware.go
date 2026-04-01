package api

import (
	"net/http"
	"time"

	"pentlog/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrw, r)

			logger.Info("api_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrw.Status(),
				"duration", time.Since(start).Round(time.Millisecond),
				"ip", r.RemoteAddr,
			)
		})
	}
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Pentlog-Token")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logger.Error("panic_recovered", "error", rvr, "path", r.URL.Path)
				InternalError(w, "Internal server error", "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
