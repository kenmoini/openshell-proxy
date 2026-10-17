package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
	"github.com/kenmoini/openshell-proxy/internal/httperr"
)

type templateOut struct{ Body dto.SandboxWorkloadTemplate }
type templatePageOut struct {
	Body dto.Page[dto.SandboxWorkloadTemplate]
}

type templateCreateIn struct {
	WorkspacePath
	Body dto.SandboxWorkloadTemplate
}

type templateListIn struct {
	WorkspacePath
	ListQuery
}

type templatePath struct {
	WorkspacePath
	Name string `path:"name" minLength:"1"`
}

type templateDeleteIn struct {
	WorkspacePath
	Name string `path:"name" minLength:"1"`
	AllowMissingQuery
}

// registerTemplates registers sandbox workload template operations.
func registerTemplates(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"SandboxTemplates"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/sandbox-templates"

	create := op("createSandboxTemplate", http.MethodPost, base, "Create a sandbox workload template")
	create.DefaultStatus = http.StatusCreated
	huma.Register(api, create, func(ctx context.Context, in *templateCreateIn) (*templateOut, error) {
		tpl, err := convert.TemplateToSDK(in.Body)
		if err != nil {
			return nil, httperr.New(http.StatusBadRequest, err.Error())
		}
		res, err := d.Client.SandboxTemplates().Create(ctx, in.Workspace, tpl)
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &templateOut{Body: convert.TemplateFromSDK(res)}, nil
	})

	huma.Register(api, op("listSandboxTemplates", http.MethodGet, base, "List sandbox workload templates"),
		func(ctx context.Context, in *templateListIn) (*templatePageOut, error) {
			pager, err := d.Client.SandboxTemplates().List(in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.TemplateFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &templatePageOut{Body: *page}, nil
		})

	huma.Register(api, op("getSandboxTemplate", http.MethodGet, base+"/{name}", "Get a sandbox workload template"),
		func(ctx context.Context, in *templatePath) (*templateOut, error) {
			res, err := d.Client.SandboxTemplates().Get(ctx, in.Workspace, in.Name)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &templateOut{Body: convert.TemplateFromSDK(res)}, nil
		})

	huma.Register(api, op("deleteSandboxTemplate", http.MethodDelete, base+"/{name}", "Delete a sandbox workload template"),
		func(ctx context.Context, in *templateDeleteIn) (*deletionOut, error) {
			res, err := d.Client.SandboxTemplates().Delete(ctx, in.Workspace, in.Name, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
