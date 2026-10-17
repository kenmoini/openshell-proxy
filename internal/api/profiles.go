package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

type profileOut struct{ Body dto.ProviderProfile }
type profilePageOut struct{ Body dto.Page[dto.ProviderProfile] }
type profileImportOut struct{ Body dto.ProfileImportResult }
type profileUpdateOut struct{ Body dto.ProfileUpdateResult }
type profileLintOut struct{ Body dto.ProfileLintResult }

type profileListIn struct {
	WorkspacePath
	ListQuery
}

type ProfilePath struct {
	WorkspacePath
	ID string `path:"id" minLength:"1"`
}

type profilePutIn struct {
	ProfilePath
	IfMatchHeader
	Body dto.ProfileImportItem
}

type profileDeleteIn struct {
	ProfilePath
	AllowMissingQuery
}

type profileImportIn struct {
	WorkspacePath
	Body dto.ProfileImportRequest
}

// registerProfiles registers provider profile operations.
func registerProfiles(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"ProviderProfiles"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/provider-profiles"

	huma.Register(api, op("listProviderProfiles", http.MethodGet, base, "List provider profiles"),
		func(ctx context.Context, in *profileListIn) (*profilePageOut, error) {
			pager, err := d.Client.Providers().Profiles().List(in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.ProfileFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &profilePageOut{Body: *page}, nil
		})

	huma.Register(api, op("importProviderProfiles", http.MethodPost, base+"/import", "Import provider profiles"),
		func(ctx context.Context, in *profileImportIn) (*profileImportOut, error) {
			res, err := d.Client.Providers().Profiles().Import(ctx, in.Workspace, convert.ImportItemsToSDK(in.Body.Items))
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &profileImportOut{Body: convert.ImportResultFromSDK(res)}, nil
		})

	huma.Register(api, op("lintProviderProfiles", http.MethodPost, base+"/lint", "Lint provider profiles"),
		func(ctx context.Context, in *profileImportIn) (*profileLintOut, error) {
			res, err := d.Client.Providers().Profiles().Lint(ctx, in.Workspace, convert.ImportItemsToSDK(in.Body.Items))
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &profileLintOut{Body: convert.LintResultFromSDK(res)}, nil
		})

	huma.Register(api, op("getProviderProfile", http.MethodGet, base+"/{id}", "Get a provider profile"),
		func(ctx context.Context, in *ProfilePath) (*profileOut, error) {
			p, err := d.Client.Providers().Profiles().Get(ctx, in.Workspace, in.ID)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &profileOut{Body: convert.ProfileFromSDK(p)}, nil
		})

	huma.Register(api, op("updateProviderProfile", http.MethodPut, base+"/{id}", "Update a provider profile (If-Match required)"),
		func(ctx context.Context, in *profilePutIn) (*profileUpdateOut, error) {
			v, set, err := in.Version()
			if err != nil {
				return nil, ifMatchProblem(err)
			}
			if !set {
				return nil, httperr.New(http.StatusPreconditionRequired, "If-Match header is required")
			}
			res, err := d.Client.Providers().Profiles().Update(ctx, in.Workspace, in.ID, v, convert.ImportItemToSDK(in.Body))
			if err != nil {
				return nil, sdkErr(err, true)
			}
			return &profileUpdateOut{Body: convert.UpdateResultFromSDK(res)}, nil
		})

	huma.Register(api, op("deleteProviderProfile", http.MethodDelete, base+"/{id}", "Delete a provider profile"),
		func(ctx context.Context, in *profileDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Providers().Profiles().Delete(ctx, in.Workspace, in.ID, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
