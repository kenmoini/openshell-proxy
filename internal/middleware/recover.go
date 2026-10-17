package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

// Recover turns handler panics into a 500 problem response and logs the stack.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Error("panic recovered",
					"panic", v,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
					"requestId", RequestIDFrom(r.Context()),
				)
				if rec.wrote {
					return
				}
				httperr.Write(w, r, httperr.New(http.StatusInternalServerError, "internal error"))
			}()
			next.ServeHTTP(rec, r)
		})
	}
}
