package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type sandboxConfigOut struct{ Body dto.SandboxConfig }
type gatewayConfigOut struct{ Body dto.GatewayConfig }
type configUpdateOut struct{ Body dto.ConfigUpdateResult }

type configPatchIn struct {
	SandboxPath
	IfMatchHeader
	Body dto.ConfigUpdateRequest
}

type gatewayConfigPatchIn struct {
	IfMatchHeader
	Workspace string `query:"workspace" doc:"Workspace passed to the gateway; empty for gateway-global scope."`
	Body      dto.ConfigUpdateRequest
}

// expectedVersion resolves the expected resource version: If-Match wins,
// else the body's expectedResourceVersion. set reports whether either was given.
func expectedVersion(h IfMatchHeader, body dto.ConfigUpdateRequest) (v uint64, set bool, err error) {
	v, set, err = h.Version()
	if err != nil || set {
		return v, set, err
	}
	if body.ExpectedResourceVersion != nil {
		return uint64(*body.ExpectedResourceVersion), true, nil
	}
	return 0, false, nil
}

// registerConfig registers sandbox and gateway configuration operations.
func registerConfig(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Config"}
		return o
	}
	const sbPath = "/v1/workspaces/{workspace}/sandboxes/{sandbox}/config"

	huma.Register(api, op("getSandboxConfig", http.MethodGet, sbPath, "Get effective sandbox configuration"),
		func(ctx context.Context, in *SandboxPath) (*sandboxConfigOut, error) {
			c, err := d.Client.Config().GetSandbox(ctx, in.Workspace, in.Sandbox)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &sandboxConfigOut{Body: convert.SandboxConfigFromSDK(c)}, nil
		})

	huma.Register(api, op("updateSandboxConfig", http.MethodPatch, sbPath, "Update sandbox configuration (If-Match optional)"),
		func(ctx context.Context, in *configPatchIn) (*configUpdateOut, error) {
			v, set, err := expectedVersion(in.IfMatchHeader, in.Body)
			if err != nil {
				return nil, ifMatchProblem(err)
			}
			res, err := d.Client.Config().Update(ctx, in.Workspace, convert.ConfigUpdateToSDK(in.Body, in.Sandbox, false, v))
			if err != nil {
				return nil, sdkErr(err, set)
			}
			return &configUpdateOut{Body: convert.ConfigUpdateResultFromSDK(res)}, nil
		})

	huma.Register(api, op("getGatewayConfig", http.MethodGet, "/v1/gateway/config", "Get gateway-global settings"),
		func(ctx context.Context, _ *struct{}) (*gatewayConfigOut, error) {
			c, err := d.Client.Config().GetGateway(ctx)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &gatewayConfigOut{Body: convert.GatewayConfigFromSDK(c)}, nil
		})

	huma.Register(api, op("updateGatewayConfig", http.MethodPatch, "/v1/gateway/config", "Update gateway-global settings or policy (If-Match optional)"),
		func(ctx context.Context, in *gatewayConfigPatchIn) (*configUpdateOut, error) {
			v, set, err := expectedVersion(in.IfMatchHeader, in.Body)
			if err != nil {
				return nil, ifMatchProblem(err)
			}
			res, err := d.Client.Config().Update(ctx, in.Workspace, convert.ConfigUpdateToSDK(in.Body, "", true, v))
			if err != nil {
				return nil, sdkErr(err, set)
			}
			return &configUpdateOut{Body: convert.ConfigUpdateResultFromSDK(res)}, nil
		})
}
