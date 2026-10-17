package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type workspaceOut struct{ Body dto.Workspace }
type workspacePageOut struct{ Body dto.Page[dto.Workspace] }
type memberOut struct{ Body dto.WorkspaceMember }
type memberPageOut struct{ Body dto.Page[dto.WorkspaceMember] }

// deletionOut carries a deletion result; huma honors Status (200 or 202).
type deletionOut struct {
	Status int
	Body   dto.DeletionResult
}

type workspaceCreateIn struct {
	Body dto.WorkspaceCreate
}

type workspaceListIn struct{ ListQuery }

type workspaceDeleteIn struct {
	WorkspacePath
	AllowMissingQuery
}

type memberListIn struct {
	WorkspacePath
	ListQuery
}

type memberAddIn struct {
	WorkspacePath
	Body dto.WorkspaceMemberAdd
}

type memberDeleteIn struct {
	WorkspacePath
	AllowMissingQuery
	Subject string `path:"subject" minLength:"1"`
}

// registerWorkspaces registers workspace and membership operations.
func registerWorkspaces(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Workspaces"}
		return o
	}

	create := op("createWorkspace", http.MethodPost, "/v1/workspaces", "Create a workspace")
	create.DefaultStatus = http.StatusCreated
	huma.Register(api, create, func(ctx context.Context, in *workspaceCreateIn) (*workspaceOut, error) {
		w, err := d.Client.Workspaces().Create(ctx, in.Body.Name, in.Body.Labels)
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &workspaceOut{Body: convert.WorkspaceFromSDK(w)}, nil
	})

	huma.Register(api, op("listWorkspaces", http.MethodGet, "/v1/workspaces", "List workspaces"),
		func(ctx context.Context, in *workspaceListIn) (*workspacePageOut, error) {
			pager, err := d.Client.Workspaces().List(in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.WorkspaceFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &workspacePageOut{Body: *page}, nil
		})

	huma.Register(api, op("getWorkspace", http.MethodGet, "/v1/workspaces/{workspace}", "Get a workspace"),
		func(ctx context.Context, in *WorkspacePath) (*workspaceOut, error) {
			w, err := d.Client.Workspaces().Get(ctx, in.Workspace)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &workspaceOut{Body: convert.WorkspaceFromSDK(w)}, nil
		})

	huma.Register(api, op("deleteWorkspace", http.MethodDelete, "/v1/workspaces/{workspace}", "Delete a workspace"),
		func(ctx context.Context, in *workspaceDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Workspaces().Delete(ctx, in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})

	huma.Register(api, op("listWorkspaceMembers", http.MethodGet, "/v1/workspaces/{workspace}/members", "List workspace members"),
		func(ctx context.Context, in *memberListIn) (*memberPageOut, error) {
			pager, err := d.Client.Workspaces().ListMembers(in.Workspace, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.MemberFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &memberPageOut{Body: *page}, nil
		})

	add := op("addWorkspaceMember", http.MethodPost, "/v1/workspaces/{workspace}/members", "Add a workspace member")
	add.DefaultStatus = http.StatusCreated
	huma.Register(api, add, func(ctx context.Context, in *memberAddIn) (*memberOut, error) {
		m, err := d.Client.Workspaces().AddMember(ctx, in.Workspace, in.Body.PrincipalSubject, convert.RoleToSDK(in.Body.Role))
		if err != nil {
			return nil, sdkErr(err, false)
		}
		return &memberOut{Body: convert.MemberFromSDK(m)}, nil
	})

	huma.Register(api, op("removeWorkspaceMember", http.MethodDelete, "/v1/workspaces/{workspace}/members/{subject}", "Remove a workspace member"),
		func(ctx context.Context, in *memberDeleteIn) (*deletionOut, error) {
			res, err := d.Client.Workspaces().RemoveMember(ctx, in.Workspace, in.Subject, in.SDK())
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &deletionOut{Status: convert.DeletionStatus(res), Body: convert.Deletion(res)}, nil
		})
}
