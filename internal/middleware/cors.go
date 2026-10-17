package middleware

import (
	"net/http"

	"github.com/rs/cors"

	"github.com/kenmoini/openshell-proxy/internal/config"
)

// CORS applies cross-origin policy; it is a no-op when no origins are configured.
func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowCredentials: cfg.AllowCredentials,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "If-Match", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "Retry-After"},
	})
	return c.Handler
}
