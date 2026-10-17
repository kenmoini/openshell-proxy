package api

import (
	"encoding/json"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestWorkspaces_CRUD(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	rec := do(t, h, "POST", "/v1/workspaces", `{"name":"team-a","labels":{"env":"dev"}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	require.Equal(t, "team-a", decode(t, rec.Body.Bytes())["name"])

	rec = do(t, h, "GET", "/v1/workspaces/team-a", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "team-a", decode(t, rec.Body.Bytes())["name"])

	rec = do(t, h, "GET", "/v1/workspaces", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 1)

	rec = do(t, h, "DELETE", "/v1/workspaces/team-a", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "Completed", decode(t, rec.Body.Bytes())["outcome"])

	rec = do(t, h, "GET", "/v1/workspaces/team-a", "", authHeader())
	require.Equal(t, 404, rec.Code, rec.Body.String())
	require.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")
}

func TestMembers_AddListRemove(t *testing.T) {
	fc := gatewaytest.New()
	fc.AddWorkspace(&v1.Workspace{Name: "ws1"})
	h, _ := newTestAPI(t, fc, nil)

	rec := do(t, h, "POST", "/v1/workspaces/ws1/members", `{"principalSubject":"alice","role":"User"}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	require.Equal(t, "User", decode(t, rec.Body.Bytes())["role"])

	rec = do(t, h, "GET", "/v1/workspaces/ws1/members", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	items := decode(t, rec.Body.Bytes())["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, "alice", items[0].(map[string]any)["principalSubject"])

	rec = do(t, h, "POST", "/v1/workspaces/ws1/members", `{"principalSubject":"bob","role":"Root"}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())

	rec = do(t, h, "DELETE", "/v1/workspaces/ws1/members/alice", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
}

func TestMembers_SubjectPathIsDecoded(t *testing.T) {
	fc := gatewaytest.New()
	fc.AddWorkspace(&v1.Workspace{Name: "ws1"})
	fc.AddMember("ws1", &v1.WorkspaceMember{PrincipalSubject: "oidc|abc/def", Role: v1.WorkspaceRoleUser})
	h, _ := newTestAPI(t, fc, nil)

	rec := do(t, h, "DELETE", "/v1/workspaces/ws1/members/oidc%7Cabc%2Fdef", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "Completed", decode(t, rec.Body.Bytes())["outcome"])

	rec = do(t, h, "GET", "/v1/workspaces/ws1/members", "", authHeader())
	require.Empty(t, decode(t, rec.Body.Bytes())["items"])
}

func TestWorkspaceDelete_AllowMissing(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	rec := do(t, h, "DELETE", "/v1/workspaces/nope", "", authHeader())
	require.Equal(t, 404, rec.Code, rec.Body.String())
	rec = do(t, h, "DELETE", "/v1/workspaces/nope?allowMissing=true", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "AlreadyAbsent", decode(t, rec.Body.Bytes())["outcome"])
}
