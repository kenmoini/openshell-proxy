package convert

import (
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/stretchr/testify/require"
)

func TestPolicyRevisionFromSDK(t *testing.T) {
	r := PolicyRevisionFromSDK(v1.SandboxPolicyRevision{Version: 2, Status: types.PolicyLoadStatusFailed, LoadedAt: time.Time{}})
	require.Equal(t, "Failed", r.Status)
	require.Nil(t, r.LoadedAt)
}

func TestDraftPolicyFromSDK_NonNil(t *testing.T) {
	d := DraftPolicyFromSDK(&v1.DraftPolicy{DraftVersion: 3, Chunks: []v1.PolicyChunk{{ID: "c"}}})
	require.Len(t, d.Chunks, 1)
	require.NotNil(t, d.Chunks[0].DenialSummaryIDs)
	require.Equal(t, uint64(3), uint64(d.DraftVersion))
	require.NotNil(t, DraftPolicyFromSDK(&v1.DraftPolicy{}).Chunks)
	require.NotNil(t, DraftHistoryFromSDK(nil).Items)
}
