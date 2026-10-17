package auth

import (
	"net/http"
	"time"

	"github.com/kenmoini/openshell-proxy/internal/config"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

const sanitySkew = 60 * time.Second

var defaultExempt = []string{"/healthz", "/readyz", "/openapi.json", "/openapi.yaml", "/docs", "/v1/health"}

// Options configures Middleware.
type Options struct {
	Mode             config.Mode
	LocalSanityCheck bool
	AllowQueryToken  bool
	Exempt           []string
	Now              func() time.Time
}

// Middleware extracts the caller's bearer token into the request context and, in
// bearer-passthrough mode, rejects missing or obviously invalid tokens.
func Middleware(o Options) func(http.Handler) http.Handler {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	exempt := map[string]bool{}
	for _, p := range defaultExempt {
		exempt[p] = true
	}
	for _, p := range o.Exempt {
		exempt[p] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			token, ok := ParseBearer(hdr)
			if !ok && o.AllowQueryToken {
				if q := r.URL.Query().Get("access_token"); q != "" {
					token, ok = q, true
				}
			}

			if exempt[r.URL.Path] || o.Mode == config.ModeNone {
				if ok {
					r = r.WithContext(WithToken(r.Context(), token))
				}
				next.ServeHTTP(w, r)
				return
			}

			if !ok {
				if hdr != "" {
					reject(w, r, "TOKEN_MALFORMED", "Authorization header is not a valid bearer token.", "Bearer")
				} else {
					reject(w, r, "TOKEN_MISSING", "Bearer token required.", "Bearer")
				}
				return
			}
			if o.LocalSanityCheck {
				switch err := SanityCheck(token, now(), sanitySkew); err {
				case nil:
				case ErrExpired:
					reject(w, r, "TOKEN_EXPIRED", "Bearer token has expired.", `Bearer error="invalid_token"`)
					return
				default:
					reject(w, r, "TOKEN_MALFORMED", "Bearer token is malformed.", "Bearer")
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(WithToken(r.Context(), token)))
		})
	}
}

func reject(w http.ResponseWriter, r *http.Request, reason, detail, challenge string) {
	p := httperr.New(http.StatusUnauthorized, detail).WithHeader("WWW-Authenticate", challenge)
	p.Reason = reason
	httperr.Write(w, r, p)
}
