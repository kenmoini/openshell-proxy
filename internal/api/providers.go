package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

type providerOut struct{ Body dto.Provider }
type providerPageOut struct{ Body dto.Page[dto.Provider] }

type providerCreateIn struct {
	WorkspacePath
	Body dto.Provider
}

type providerListIn struct {
	WorkspacePath
	ListQuery
}

type ProviderPath struct {
	WorkspacePath
	Name string `path:"name" minLength:"1"`
}

type providerPutIn struct {
	ProviderPath
	Ensure bool `query:"ensure" doc:"Create the provider when it does not exist (Ensure) instead of failing."`
	Body   dto.Provider
}

type providerDeleteIn struct {
	ProviderPath
	AllowMissingQuery
}

// registerProviders registers provider CRUD operations.
func registerProviders(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Providers"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/providers"

	create := op("createProvider", http.MethodPost, base, "Create a provider")
	create.DefaultStatus = http.StatusCreated
	huma.Register(api, create, func(ctx context.Context, in *providerCreateIn) (*providerOut, error) {
		p, err := d.Client.Providers().Create(ctx, in.Workspace, convert.ProviderToSDK(in.Body))
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &providerOut{Body: convert.ProviderFromSDK(p)}, nil
	})

	huma.Register(api, op("listProviders", http.MethodGet, base, "List providers"),
		func(ctx context.Context, in *providerListIn) (*providerPageOut, error) {
			pager, err := d.Client.Providers().List(in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.ProviderFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &providerPageOut{Body: *page}, nil
		})

	huma.Register(api, op("getProvider", http.MethodGet, base+"/{name}", "Get a provider"),
		func(ctx context.Context, in *ProviderPath) (*providerOut, error) {
			p, err := d.Client.Providers().Get(ctx, in.Workspace, in.Name)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &providerOut{Body: convert.ProviderFromSDK(p)}, nil
		})

	huma.Register(api, op("updateProvider", http.MethodPut, base+"/{name}", "Update (or with ensure=true, create or update) a provider"),
		func(ctx context.Context, in *providerPutIn) (*providerOut, error) {
			if in.Body.Name != in.Name {
				return nil, httperr.New(http.StatusBadRequest, "body name must equal the path name")
			}
			sp := convert.ProviderToSDK(in.Body)
			p := d.Client.Providers()
			if in.Ensure {
				out, err := p.Ensure(ctx, in.Workspace, sp)
				if err != nil {
					return nil, sdkErr(err, false)
				}
				return &providerOut{Body: convert.ProviderFromSDK(out)}, nil
			}
			out, err := p.Update(ctx, in.Workspace, sp)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &providerOut{Body: convert.ProviderFromSDK(out)}, nil
		})

	huma.Register(api, op("deleteProvider", http.MethodDelete, base+"/{name}", "Delete a provider"),
		func(ctx context.Context, in *providerDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Providers().Delete(ctx, in.Workspace, in.Name, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
