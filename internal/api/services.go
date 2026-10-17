package api

import (
	"context"
	"net/http"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type serviceOut struct{ Body dto.ServiceEndpoint }
type servicePageOut struct{ Body dto.Page[dto.ServiceEndpoint] }

type serviceExposeIn struct {
	SandboxPath
	Body dto.ServiceExpose
}

type serviceListIn struct {
	SandboxPath
	ListQuery
}

type serviceGetIn struct {
	SandboxPath
	Name string `path:"name" minLength:"1"`
}

type serviceDeleteIn struct {
	SandboxPath
	AllowMissingQuery
	Name string `path:"name" minLength:"1"`
}

// registerServices registers sandbox service exposure operations.
func registerServices(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Services"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/sandboxes/{sandbox}/services"

	expose := op("exposeSandboxService", http.MethodPost, base, "Expose a sandbox service")
	expose.DefaultStatus = http.StatusCreated
	huma.Register(api, expose, func(ctx context.Context, in *serviceExposeIn) (*serviceOut, error) {
		ep, err := d.Client.Services().Expose(ctx, in.Workspace, in.Sandbox, in.Body.Name, in.Body.TargetPort, in.Body.Domain,
			v1.ExposeServiceOptions{AuthorizationMode: convert.AuthModeToSDK(in.Body.AuthorizationMode)})
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &serviceOut{Body: convert.ServiceFromSDK(ep)}, nil
	})

	huma.Register(api, op("listSandboxServices", http.MethodGet, base, "List sandbox services"),
		func(ctx context.Context, in *serviceListIn) (*servicePageOut, error) {
			pager, err := d.Client.Services().List(in.Workspace, in.Sandbox, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.ServiceFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &servicePageOut{Body: *page}, nil
		})

	huma.Register(api, op("getSandboxService", http.MethodGet, base+"/{name}", "Get a sandbox service"),
		func(ctx context.Context, in *serviceGetIn) (*serviceOut, error) {
			ep, err := d.Client.Services().Get(ctx, in.Workspace, in.Sandbox, in.Name)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &serviceOut{Body: convert.ServiceFromSDK(ep)}, nil
		})

	huma.Register(api, op("deleteSandboxService", http.MethodDelete, base+"/{name}", "Delete a sandbox service"),
		func(ctx context.Context, in *serviceDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Services().Delete(ctx, in.Workspace, in.Sandbox, in.Name, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
