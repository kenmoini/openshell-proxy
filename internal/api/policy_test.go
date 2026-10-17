package api

import (
	"context"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const polBase = "/v1/workspaces/ws1/sandboxes/sb1/policy"

func TestPolicy_Revisions(t *testing.T) {
	c := gatewaytest.New()
	c.AddRevision("ws1", "sb1", v1.SandboxPolicyRevision{Version: 1, PolicyHash: "h1", Status: types.PolicyLoadStatusSuperseded})
	c.AddRevision("ws1", "sb1", v1.SandboxPolicyRevision{Version: 2, PolicyHash: "h2", Status: types.PolicyLoadStatusLoaded, Policy: &v1.SandboxPolicy{Version: 2}})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", polBase+"/revisions", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	items := decode(t, rec.Body.Bytes())["items"].([]any)
	require.Len(t, items, 2)
	require.Equal(t, "Superseded", items[0].(map[string]any)["status"])
	require.Equal(t, "Loaded", items[1].(map[string]any)["status"])
	require.Equal(t, "h2", items[1].(map[string]any)["policyHash"])
}

func TestPolicy_GlobalRevisions(t *testing.T) {
	c := gatewaytest.New()
	c.AddGlobalRevision(v1.SandboxPolicyRevision{Version: 4, Status: types.PolicyLoadStatusLoaded})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", "/v1/workspaces/ws1/policy/global/revisions", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	items := decode(t, rec.Body.Bytes())["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, float64(4), items[0].(map[string]any)["version"])

	rec = do(t, h, "GET", "/v1/workspaces/ws1/policy/global/status", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, float64(4), decode(t, rec.Body.Bytes())["activeVersion"])
}

func TestPolicy_Status(t *testing.T) {
	c := gatewaytest.New()
	c.AddRevision("ws1", "sb1", v1.SandboxPolicyRevision{Version: 1, Status: types.PolicyLoadStatusSuperseded})
	c.AddRevision("ws1", "sb1", v1.SandboxPolicyRevision{Version: 2, Status: types.PolicyLoadStatusLoaded})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", polBase+"/status?version=1", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, float64(1), m["revision"].(map[string]any)["version"])
	require.Equal(t, float64(2), m["activeVersion"])
	rec = do(t, h, "GET", polBase+"/status", "", authHeader())
	require.Equal(t, float64(2), decode(t, rec.Body.Bytes())["revision"].(map[string]any)["version"])
	rec = do(t, h, "GET", polBase+"/status?version=9", "", authHeader())
	require.Equal(t, 404, rec.Code, rec.Body.String())
}

func TestPolicy_DraftFlow(t *testing.T) {
	var (
		filter    string
		token     string
		reason    string
		approveAl []types.DraftChunkApproval
		flagged   bool
		editRule  *v1.NetworkPolicyRule
		editChunk string
		undone    string
	)
	stub := &gatewaytest.PolicyStub{
		GetDraftFn: func(_ context.Context, _, _ string, opts ...v1.GetDraftOption) (*v1.DraftPolicy, error) {
			cfg := types.ApplyGetDraftOptions(opts)
			filter = cfg.StatusFilter()
			return &v1.DraftPolicy{DraftVersion: 7, RollingSummary: "sum", Chunks: []v1.PolicyChunk{{
				ID: "c1", Status: "pending", RuleName: "r", HitCount: 3, Confidence: 0.5,
				ProposedRule:             &v1.NetworkPolicyRule{Name: "r"},
				CandidateEffectivePolicy: &v1.SandboxPolicy{Version: 2},
			}}}, nil
		},
		ApproveDraftChunkFn: func(_ context.Context, ws, sb, chunk, tok string) (*v1.ApproveResult, error) {
			require.Equal(t, "c1", chunk)
			token = tok
			return &v1.ApproveResult{PolicyVersion: 5, PolicyHash: "ph"}, nil
		},
		RejectDraftChunkFn: func(_ context.Context, _, _, chunk, r string) error {
			reason = r
			return nil
		},
		ApproveAllDraftChunksFn: func(_ context.Context, _, _ string, opts ...v1.ApproveAllOption) (*v1.ApproveAllResult, error) {
			cfg := types.ApplyApproveAllOptions(opts)
			approveAl, flagged = cfg.Approvals(), cfg.IncludeSecurityFlagged()
			return &v1.ApproveAllResult{PolicyVersion: 6, ChunksApproved: 2, ChunksSkipped: 1}, nil
		},
		ClearDraftChunksFn: func(context.Context, string, string) (*v1.ClearResult, error) {
			return &v1.ClearResult{ChunksCleared: 4}, nil
		},
		GetDraftHistoryFn: func(context.Context, string, string) ([]v1.DraftHistoryEntry, error) {
			return []v1.DraftHistoryEntry{{EventType: "a", ChunkID: "c1"}, {EventType: "b"}}, nil
		},
		EditDraftChunkFn: func(_ context.Context, _, _, chunk string, r *v1.NetworkPolicyRule) error {
			editChunk, editRule = chunk, r
			return nil
		},
		UndoDraftChunkFn: func(_ context.Context, _, _, chunk string) (*v1.UndoResult, error) {
			undone = chunk
			return &v1.UndoResult{PolicyVersion: 8, PolicyHash: "uh"}, nil
		},
	}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithPolicy(stub)), nil)

	rec := do(t, h, "GET", polBase+"/draft?status=pending", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "pending", filter)
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "7", m["draftVersion"])
	ch := m["chunks"].([]any)[0].(map[string]any)
	require.Equal(t, "c1", ch["id"])
	require.Equal(t, float64(2), ch["candidateEffectivePolicy"].(map[string]any)["version"])

	rec = do(t, h, "POST", polBase+"/draft/chunks/c1/approve", `{"reviewToken":"tok"}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "tok", token)
	require.Equal(t, float64(5), decode(t, rec.Body.Bytes())["policyVersion"])

	rec = do(t, h, "POST", polBase+"/draft/chunks/c1/approve", `{}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())

	rec = do(t, h, "POST", polBase+"/draft/chunks/c1/reject", `{"reason":"nope"}`, authHeader())
	require.Equal(t, 204, rec.Code, rec.Body.String())
	require.Equal(t, "nope", reason)

	rec = do(t, h, "POST", polBase+"/draft/approve-all", `{"includeSecurityFlagged":true,"approvals":[{"chunkId":"c1","reviewToken":"t1"}]}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.True(t, flagged)
	require.Equal(t, []types.DraftChunkApproval{{ChunkID: "c1", ReviewToken: "t1"}}, approveAl)
	require.Equal(t, float64(2), decode(t, rec.Body.Bytes())["chunksApproved"])

	rec = do(t, h, "POST", polBase+"/draft/clear", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, float64(4), decode(t, rec.Body.Bytes())["chunksCleared"])

	rec = do(t, h, "GET", polBase+"/draft/history", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 2)

	rec = do(t, h, "PUT", polBase+"/draft/chunks/c1", `{"name":"edited","binaries":[{"path":"/bin/x"}]}`, authHeader())
	require.Equal(t, 204, rec.Code, rec.Body.String())
	require.Equal(t, "c1", editChunk)
	require.Equal(t, "edited", editRule.Name)

	rec = do(t, h, "POST", polBase+"/draft/chunks/c1/undo", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "c1", undone)
	require.Equal(t, "uh", decode(t, rec.Body.Bytes())["policyHash"])
}
