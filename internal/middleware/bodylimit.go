package middleware

import (
	"net/http"

	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

// BodyLimit caps request bodies at n bytes. n <= 0 disables the limit.
func BodyLimit(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if n <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				httperr.Write(w, r, httperr.New(http.StatusRequestEntityTooLarge, "request body too large"))
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
