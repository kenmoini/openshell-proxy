// Package api assembles the HTTP layer: huma on a stdlib mux behind the
// middleware chain. Resource routes live in their own files and are
// registered from New.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/kenmoini/openshell-proxy/internal/auth"
	"github.com/kenmoini/openshell-proxy/internal/config"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
	"github.com/kenmoini/openshell-proxy/internal/middleware"
	"github.com/kenmoini/openshell-proxy/internal/version"
)

// Deps are the collaborators the API needs.
type Deps struct {
	Client v1.ClientInterface
	// Dedicated yields a per-call client for long-lived streams (may be nil in tests).
	Dedicated func(ctx context.Context) (v1.ClientInterface, func(), error)
	Cfg       *config.Config
	Log       *slog.Logger
}

// errorStatuses is the shared error list documented on every operation.
var errorStatuses = []int{400, 401, 403, 404, 409, 412, 413, 422, 428, 429, 500, 502, 503, 504}

// metaLongRunning marks operations exempt from the per-request deadline.
const metaLongRunning = "longRunning"

// op builds a bearer-secured operation with the shared errors and body limit.
func (d Deps) op(id, method, path, summary string) huma.Operation {
	return huma.Operation{
		OperationID:  id,
		Method:       method,
		Path:         path,
		Summary:      summary,
		Tags:         []string{"Platform"},
		Security:     []map[string][]string{{"bearer": {}}},
		Errors:       errorStatuses,
		MaxBodyBytes: d.Cfg.Limits.MaxBody,
	}
}

// New builds the handler and the huma API (returned so callers can register more operations).
func New(d Deps) (http.Handler, huma.API) {
	httperr.Install()

	mux := http.NewServeMux()
	api := humago.New(mux, humaConfig())

	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		if lr, _ := ctx.Operation().Metadata[metaLongRunning].(bool); lr {
			next(ctx)
			return
		}
		c, cancel := withTimeout(ctx.Context(), d.Cfg.Timeouts.Request)
		defer cancel()
		next(huma.WithContext(ctx, c))
	})

	registerProbes(mux, d)
	registerHealth(api, d)
	registerWorkspaces(api, d)
	registerSandboxes(api, d)
	registerTemplates(api, d)
	registerExec(api, d)
	registerServices(api, d)

	cfg := d.Cfg
	h := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Recover(d.Log),
		middleware.Logging(d.Log),
		middleware.CORS(cfg.CORS),
		middleware.BodyLimit(cfg.Limits.MaxExecBody),
		auth.Middleware(auth.Options{
			Mode:             cfg.Auth.Mode,
			LocalSanityCheck: cfg.Auth.LocalSanityCheck,
			AllowQueryToken:  cfg.Auth.AllowQueryToken,
		}),
	)
	return h, api
}

// OpenAPI returns the JSON OpenAPI document of the registered API.
func OpenAPI(d Deps) ([]byte, error) {
	_, api := New(d)
	return json.Marshal(api.OpenAPI())
}

func humaConfig() huma.Config {
	cfg := huma.DefaultConfig("OpenShell REST Proxy", version.Version)
	if cfg.Components == nil {
		cfg.Components = &huma.Components{}
	}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}
	return cfg
}
