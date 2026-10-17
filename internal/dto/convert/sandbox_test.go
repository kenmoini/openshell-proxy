package convert

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/stretchr/testify/require"
)

func maximalSpec() v1.SandboxSpec {
	return v1.SandboxSpec{
		LogLevel:    "debug",
		Environment: map[string]string{"A": "1"},
		Template: &v1.SandboxTemplate{
			Image: "img:1", RuntimeClassName: "gvisor", AgentSocket: "/run/agent.sock",
			Labels: map[string]string{"l": "v"}, Annotations: map[string]string{"a": "v"},
			Environment:    map[string]string{"E": "2"},
			UserNamespaces: ptr(true),
			Resources:      map[string]any{"cpu": "2", "nested": map[string]any{"k": float64(1)}},
			DriverConfig:   map[string]any{"driver": "k8s"},
		},
		Providers: []string{"github", "aws"},
		GPU:       true, GPUCount: ptr(uint32(2)),
		Policy:  maximalPolicy(),
		Command: []string{"/bin/sh", "-c", "sleep 1"},
		TTY:     true, RestartPolicy: v1.SandboxRestartOnFailure,
	}
}

func TestSandboxFromSDK_Golden(t *testing.T) {
	del := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	sb := &v1.Sandbox{
		HostKeyFingerprint: "SHA256:abc",
		ID:                 "sb-1", Name: "dev", Workspace: "team-a",
		CreatedAt:                   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Labels:                      map[string]string{"env": "dev"},
		Annotations:                 map[string]string{"note": "hi"},
		ResourceVersion:             18446744073709551615,
		DeletionTimestamp:           &del,
		CreatedFromWorkloadTemplate: &v1.SandboxWorkloadTemplateProvenance{Name: "tmpl", ResourceVersion: "12"},
		Spec:                        maximalSpec(),
		Status: v1.SandboxStatus{
			AgentPod: "pod-1", AgentFd: "fd-a", SandboxFd: "fd-s",
			Phase: v1.SandboxReady,
			Conditions: []v1.SandboxCondition{
				{Type: "Ready", Status: "True", Reason: "Up", Message: "ok", LastTransitionTime: "2026-01-02T03:04:05Z"},
			},
			CurrentPolicyVersion: 7,
			ExitCode:             ptr(int32(3)),
			EndpointStatuses: []v1.EndpointStatus{{
				EndpointID: "ep-1", Host: "tool.example.com", Ports: []uint32{443}, Path: "/mcp",
				LastResult: v1.EndpointHTTPResponseReceived, LastReportedAt: "2026-01-02T03:04:06Z",
			}},
			ConfigurationAdmission: &types.SandboxConfigurationAdmission{
				State: types.ConfigurationAdmissionAccepted, PolicyVersion: 7, PolicyHash: "deadbeef",
				ConfigRevision: 18446744073709551615, ProviderEnvRevision: 9, Error: "none",
			},
			RestartCount: 2, NextRestartAtMs: 1700000000000, MainProcessStartedAtMs: 1700000001000,
		},
		ServiceURLs: map[string]string{"": "https://a.example.com", "web": "https://web.example.com"},
	}
	got, err := json.Marshal(SandboxFromSDK(sb))
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/sandbox.json")
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(got))
}

func TestSandboxSpec_RoundTrip(t *testing.T) {
	x := maximalSpec()
	require.Equal(t, &x, SandboxSpecToSDK(SpecFromSDK(x)))

	min := v1.SandboxSpec{Environment: map[string]string{}, Providers: []string{}}
	require.Equal(t, &min, SandboxSpecToSDK(SpecFromSDK(min)))
}
