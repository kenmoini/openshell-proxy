package api

import (
	"context"
	"net/http"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

type (
	draftPolicyOut     struct{ Body dto.DraftPolicy }
	approveOut         struct{ Body dto.ApproveResult }
	approveAllOut      struct{ Body dto.ApproveAllResult }
	undoOut            struct{ Body dto.UndoResult }
	clearOut           struct{ Body dto.ClearResult }
	draftHistoryOut    struct{ Body dto.DraftHistory }
	policyStatusOut    struct{ Body dto.PolicyStatusResult }
	policyRevisionsOut struct {
		Body dto.Page[dto.SandboxPolicyRevision]
	}
)

// PolicyPageQuery carries pagination for policy revision listings.
type PolicyPageQuery struct {
	PageSize  int    `query:"pageSize" minimum:"0" maximum:"1000"`
	PageToken string `query:"pageToken"`
}

func (q PolicyPageQuery) opts() []v1.ListPolicyOption {
	var o []v1.ListPolicyOption
	if q.PageSize > 0 {
		o = append(o, v1.WithPageSize(int32(q.PageSize)))
	}
	if q.PageToken != "" {
		o = append(o, v1.WithPageToken(q.PageToken))
	}
	return o
}

type ChunkPath struct {
	SandboxPath
	Chunk string `path:"chunk" minLength:"1"`
}

type draftGetIn struct {
	SandboxPath
	Status string `query:"status" doc:"Filter chunks by status, e.g. pending."`
}

type chunkApproveIn struct {
	ChunkPath
	Body dto.ApproveRequest
}

type chunkRejectIn struct {
	ChunkPath
	Body *dto.RejectRequest
}

type chunkEditIn struct {
	ChunkPath
	Body dto.NetworkPolicyRule
}

type approveAllIn struct {
	SandboxPath
	Body *dto.ApproveAllRequest
}

type policyStatusIn struct {
	SandboxPath
	Version uint32 `query:"version" minimum:"0" doc:"Specific policy version; 0 means latest."`
}

type policyRevisionsIn struct {
	SandboxPath
	PolicyPageQuery
}

type globalStatusIn struct {
	WorkspacePath
	Version uint32 `query:"version" minimum:"0" doc:"Specific policy version; 0 means latest."`
}

type globalRevisionsIn struct {
	WorkspacePath
	PolicyPageQuery
}

// registerPolicy registers policy draft review, status and revision operations.
func registerPolicy(api huma.API, d Deps) {
	op := func(id, method, path, summary string) huma.Operation {
		o := d.op(id, method, path, summary)
		o.Tags = []string{"Policy"}
		return o
	}
	const base = "/v1/workspaces/{workspace}/sandboxes/{sandbox}/policy"
	pol := func() v1.PolicyInterface { return d.Client.Policy() }

	registerPolicyDraft(api, d, op, base)

	huma.Register(api, op("getPolicyStatus", http.MethodGet, base+"/status", "Get a sandbox policy revision status"),
		func(ctx context.Context, in *policyStatusIn) (*policyStatusOut, error) {
			var opts []v1.GetStatusOption
			if in.Version > 0 {
				opts = append(opts, v1.WithVersion(in.Version))
			}
			r, err := pol().GetStatus(ctx, in.Workspace, in.Sandbox, opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &policyStatusOut{Body: convert.PolicyStatusFromSDK(r)}, nil
		})

	huma.Register(api, op("listPolicyRevisions", http.MethodGet, base+"/revisions", "List sandbox policy revisions"),
		func(ctx context.Context, in *policyRevisionsIn) (*policyRevisionsOut, error) {
			pager, err := pol().List(in.Workspace, in.Sandbox, in.opts()...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.PolicyRevisionFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &policyRevisionsOut{Body: *page}, nil
		})

	huma.Register(api, op("getGlobalPolicyStatus", http.MethodGet, "/v1/workspaces/{workspace}/policy/global/status", "Get a gateway-global policy revision status"),
		func(ctx context.Context, in *globalStatusIn) (*policyStatusOut, error) {
			opts := []v1.GetStatusOption{v1.WithStatusGlobal(true)}
			if in.Version > 0 {
				opts = append(opts, v1.WithVersion(in.Version))
			}
			r, err := pol().GetStatus(ctx, in.Workspace, "", opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &policyStatusOut{Body: convert.PolicyStatusFromSDK(r)}, nil
		})

	huma.Register(api, op("listGlobalPolicyRevisions", http.MethodGet, "/v1/workspaces/{workspace}/policy/global/revisions", "List gateway-global policy revisions"),
		func(ctx context.Context, in *globalRevisionsIn) (*policyRevisionsOut, error) {
			opts := append([]v1.ListPolicyOption{v1.WithListGlobal(true)}, in.opts()...)
			pager, err := pol().List(in.Workspace, "", opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			page, err := listPage(ctx, pager, convert.PolicyRevisionFromSDK)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &policyRevisionsOut{Body: *page}, nil
		})
}
