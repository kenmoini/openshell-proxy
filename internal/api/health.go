package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

type healthOut struct{ Body dto.Health }
type whoamiOut struct{ Body dto.CurrentUser }
type gatewayInfoOut struct{ Body dto.GatewayInfo }

// registerProbes adds the raw liveness and readiness endpoints.
func registerProbes(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := withTimeout(r.Context(), d.Cfg.Timeouts.Request)
		defer cancel()
		res, err := d.Client.Health().Check(ctx)
		if err != nil {
			httperr.Write(w, r, httperr.FromSDK(err, httperr.Options{}))
			return
		}
		if !res.Healthy {
			httperr.Write(w, r, httperr.New(http.StatusServiceUnavailable, "gateway unhealthy"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "version": res.Version})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// registerHealth registers the platform operations: health, whoami, gateway info.
func registerHealth(api huma.API, d Deps) {
	health := d.op("getHealth", http.MethodGet, "/v1/health", "Gateway health check")
	health.Security = nil
	huma.Register(api, health, func(ctx context.Context, _ *struct{}) (*healthOut, error) {
		res, err := d.Client.Health().Check(ctx)
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &healthOut{Body: convert.Health(res)}, nil
	})

	huma.Register(api, d.op("whoami", http.MethodGet, "/v1/whoami", "Identity of the calling token"),
		func(ctx context.Context, _ *struct{}) (*whoamiOut, error) {
			u, err := d.Client.Health().GetCurrentUser(ctx)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &whoamiOut{Body: convert.CurrentUser(u)}, nil
		})

	huma.Register(api, d.op("getGatewayInfo", http.MethodGet, "/v1/gateway/info", "Gateway version, drivers and extensions"),
		func(ctx context.Context, _ *struct{}) (*gatewayInfoOut, error) {
			g, err := d.Client.Health().GetGatewayInfo(ctx)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &gatewayInfoOut{Body: convert.GatewayInfo(g)}, nil
		})
}
