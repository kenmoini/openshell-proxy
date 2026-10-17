package api

import (
	"context"
	"net/http"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/dto/convert"
)

// registerPolicyDraft registers the draft policy review operations.
func registerPolicyDraft(api huma.API, d Deps, op func(id, method, path, summary string) huma.Operation, base string) {
	draft := base + "/draft"
	pol := func() v1.PolicyInterface { return d.Client.Policy() }
	noContent := func(o huma.Operation) huma.Operation {
		o.DefaultStatus = http.StatusNoContent
		return o
	}

	huma.Register(api, op("getPolicyDraft", http.MethodGet, draft, "Get the draft policy"),
		func(ctx context.Context, in *draftGetIn) (*draftPolicyOut, error) {
			var opts []v1.GetDraftOption
			if in.Status != "" {
				opts = append(opts, v1.WithStatusFilter(in.Status))
			}
			p, err := pol().GetDraft(ctx, in.Workspace, in.Sandbox, opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &draftPolicyOut{Body: convert.DraftPolicyFromSDK(p)}, nil
		})

	huma.Register(api, op("approvePolicyDraftChunk", http.MethodPost, draft+"/chunks/{chunk}/approve", "Approve a draft chunk"),
		func(ctx context.Context, in *chunkApproveIn) (*approveOut, error) {
			r, err := pol().ApproveDraftChunk(ctx, in.Workspace, in.Sandbox, in.Chunk, in.Body.ReviewToken)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &approveOut{Body: convert.ApproveResultFromSDK(r)}, nil
		})

	huma.Register(api, noContent(op("rejectPolicyDraftChunk", http.MethodPost, draft+"/chunks/{chunk}/reject", "Reject a draft chunk")),
		func(ctx context.Context, in *chunkRejectIn) (*struct{}, error) {
			if err := pol().RejectDraftChunk(ctx, in.Workspace, in.Sandbox, in.Chunk, in.Body.Reason); err != nil {
				return nil, sdkErr(err, false)
			}
			return nil, nil
		})

	huma.Register(api, noContent(op("editPolicyDraftChunk", http.MethodPut, draft+"/chunks/{chunk}", "Edit a draft chunk's proposed rule")),
		func(ctx context.Context, in *chunkEditIn) (*struct{}, error) {
			if err := pol().EditDraftChunk(ctx, in.Workspace, in.Sandbox, in.Chunk, convert.NetworkRuleToSDK(in.Body)); err != nil {
				return nil, sdkErr(err, false)
			}
			return nil, nil
		})

	huma.Register(api, op("undoPolicyDraftChunk", http.MethodPost, draft+"/chunks/{chunk}/undo", "Undo a draft chunk approval"),
		func(ctx context.Context, in *ChunkPath) (*undoOut, error) {
			r, err := pol().UndoDraftChunk(ctx, in.Workspace, in.Sandbox, in.Chunk)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &undoOut{Body: convert.UndoResultFromSDK(r)}, nil
		})

	huma.Register(api, op("approveAllPolicyDraftChunks", http.MethodPost, draft+"/approve-all", "Approve all draft chunks"),
		func(ctx context.Context, in *approveAllIn) (*approveAllOut, error) {
			var opts []v1.ApproveAllOption
			if in.Body.IncludeSecurityFlagged {
				opts = append(opts, v1.WithIncludeSecurityFlagged())
			}
			if len(in.Body.Approvals) > 0 {
				opts = append(opts, v1.WithDraftApprovals(convert.DraftApprovalsToSDK(in.Body.Approvals)...))
			}
			r, err := pol().ApproveAllDraftChunks(ctx, in.Workspace, in.Sandbox, opts...)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &approveAllOut{Body: convert.ApproveAllResultFromSDK(r)}, nil
		})

	huma.Register(api, op("clearPolicyDraftChunks", http.MethodPost, draft+"/clear", "Clear draft chunks"),
		func(ctx context.Context, in *SandboxPath) (*clearOut, error) {
			r, err := pol().ClearDraftChunks(ctx, in.Workspace, in.Sandbox)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &clearOut{Body: convert.ClearResultFromSDK(r)}, nil
		})

	huma.Register(api, op("getPolicyDraftHistory", http.MethodGet, draft+"/history", "Get draft policy history"),
		func(ctx context.Context, in *SandboxPath) (*draftHistoryOut, error) {
			es, err := pol().GetDraftHistory(ctx, in.Workspace, in.Sandbox)
			if err != nil {
				return nil, sdkErr(err, false)
			}
			return &draftHistoryOut{Body: convert.DraftHistoryFromSDK(es)}, nil
		})
}
