package convert

import (
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/stretchr/testify/require"
)

func TestProvider_RoundTrip(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	del := created.Add(time.Hour)
	in := &v1.Provider{
		ID: "id1", Name: "gh", Type: "github", CreatedAt: created,
		Labels: map[string]string{"a": "b"}, Annotations: map[string]string{"c": "d"},
		ResourceVersion: 1<<60 + 7, Workspace: "ws", DeletionTimestamp: &del,
		Spec: v1.ProviderSpec{
			Credentials:         map[string]string{"TOKEN": "x"},
			Config:              map[string]string{"url": "u"},
			CredentialExpiresAt: map[string]time.Time{"TOKEN": created},
			ProfileWorkspace:    "pw",
			CredentialHandles:   map[string]types.CredentialHandle{"TOKEN": {Driver: "vault", Handle: "h", Metadata: map[string]string{"m": "n"}}},
		},
	}
	require.Equal(t, in, ProviderToSDK(ProviderFromSDK(in)))
}
