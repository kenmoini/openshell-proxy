package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type refreshOut struct{ Body dto.RefreshStatus }
type refreshListOut struct{ Body dto.RefreshStatusList }

type refreshListIn struct {
	WorkspacePath
	Provider      string `path:"provider" minLength:"1"`
	CredentialKey string `query:"credentialKey"`
}

type RefreshKeyPath struct {
	WorkspacePath
	Provider string `path:"provider" minLength:"1"`
	Key      string `path:"key" minLength:"1"`
}

type refreshConfigIn struct {
	RefreshKeyPath
	Body dto.RefreshConfigRequest
}

type refreshDeleteIn struct {
	RefreshKeyPath
	AllowMissingQuery
}

// registerRefresh registers credential refresh operations.
func registerRefresh(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"CredentialRefresh"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/providers/{provider}/refresh"

	huma.Register(api, op("getCredentialRefreshStatus", http.MethodGet, base, "Get credential refresh status"),
		func(ctx context.Context, in *refreshListIn) (*refreshListOut, error) {
			sts, err := d.Client.Providers().Refresh().GetStatus(ctx, in.Workspace, in.Provider, in.CredentialKey)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			out := dto.RefreshStatusList{Items: make([]dto.RefreshStatus, 0, len(sts))}
			for _, s := range sts {
				out.Items = append(out.Items, convert.RefreshStatusFromSDK(s))
			}
			return &refreshListOut{Body: out}, nil
		})

	huma.Register(api, op("configureCredentialRefresh", http.MethodPut, base+"/{key}", "Configure credential refresh"),
		func(ctx context.Context, in *refreshConfigIn) (*refreshOut, error) {
			s, err := d.Client.Providers().Refresh().Configure(ctx, in.Workspace, convert.RefreshConfigToSDK(in.Provider, in.Key, in.Body))
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &refreshOut{Body: convert.RefreshStatusFromSDK(s)}, nil
		})

	huma.Register(api, op("rotateCredential", http.MethodPost, base+"/{key}/rotate", "Rotate a credential now"),
		func(ctx context.Context, in *RefreshKeyPath) (*refreshOut, error) {
			s, err := d.Client.Providers().Refresh().Rotate(ctx, in.Workspace, in.Provider, in.Key)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &refreshOut{Body: convert.RefreshStatusFromSDK(s)}, nil
		})

	huma.Register(api, op("deleteCredentialRefresh", http.MethodDelete, base+"/{key}", "Delete credential refresh configuration"),
		func(ctx context.Context, in *refreshDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Providers().Refresh().Delete(ctx, in.Workspace, in.Provider, in.Key, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
