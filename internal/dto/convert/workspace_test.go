package convert

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceFromSDK_Golden(t *testing.T) {
	del := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	w := &v1.Workspace{
		ID: "ws-id-1", Name: "team-a", Workspace: "root",
		CreatedAt:         time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Labels:            map[string]string{"env": "dev"},
		Annotations:       map[string]string{"note": "hi"},
		ResourceVersion:   18446744073709551615,
		DeletionTimestamp: &del,
		Phase:             v1.WorkspacePhase("Terminating"),
	}
	got, err := json.Marshal(WorkspaceFromSDK(w))
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/workspace.json")
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(got))
}

func TestMemberFromSDK_NameFallbackAndRole(t *testing.T) {
	m := MemberFromSDK(&v1.WorkspaceMember{PrincipalSubject: "alice", Role: v1.WorkspaceRoleAdmin})
	require.Equal(t, "alice", m.Name)
	require.Equal(t, "Admin", m.Role)
	require.Equal(t, v1.WorkspaceRoleUser, RoleToSDK("User"))
}
