package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

func errBadPoll(s string) error {
	return fmt.Errorf("invalid pollInterval %q: must be a positive Go duration", s)
}
func errBadSince(s string) error { return fmt.Errorf("invalid since %q: must be RFC 3339", s) }

type providerSummaryPageOut struct{ Body dto.Page[dto.ProviderSummary] }
type attachOut struct{ Body dto.ProviderAttachment }
type detachOut struct{ Body dto.ProviderDetachment }

type sandboxProvidersIn struct {
	SandboxPath
	ListQuery
}

type sandboxProviderIn struct {
	SandboxPath
	IfMatchHeader
	Provider string `path:"provider" minLength:"1"`
}

func registerSandboxProviders(api huma.API, d Deps, op func(id, method, path, summary string) huma.Operation) {
	const base = "/v1/workspaces/{workspace}/sandboxes/{sandbox}/providers"

	huma.Register(api, op("listSandboxProviders", http.MethodGet, base, "List providers attached to a sandbox"),
		func(ctx context.Context, in *sandboxProvidersIn) (*providerSummaryPageOut, error) {
			pager, err := d.Client.Sandboxes().ListProviders(in.Workspace, in.Sandbox, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.ProviderSummaryFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &providerSummaryPageOut{Body: *page}, nil
		})

	huma.Register(api, op("attachSandboxProvider", http.MethodPut, base+"/{provider}", "Attach a provider to a sandbox"),
		func(ctx context.Context, in *sandboxProviderIn) (*attachOut, error) {
			v, set, err := in.Version()
			if err != nil {
				return nil, ifMatchProblem(err)
			}
			res, err := d.Client.Sandboxes().AttachProvider(ctx, in.Workspace, in.Sandbox, in.Provider, v)
			if err != nil {
				return nil, sdkErr(err, set)
			}
			return &attachOut{Body: convert.AttachFromSDK(res)}, nil
		})

	huma.Register(api, op("detachSandboxProvider", http.MethodDelete, base+"/{provider}", "Detach a provider from a sandbox"),
		func(ctx context.Context, in *sandboxProviderIn) (*detachOut, error) {
			v, set, err := in.Version()
			if err != nil {
				return nil, ifMatchProblem(err)
			}
			res, err := d.Client.Sandboxes().DetachProvider(ctx, in.Workspace, in.Sandbox, in.Provider, v)
			if err != nil {
				return nil, sdkErr(err, set)
			}
			return &detachOut{Body: convert.DetachFromSDK(res)}, nil
		})
}
