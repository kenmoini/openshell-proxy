package convert

import (
	"encoding/json"
	"os"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func maximalEndpoint() v1.PolicyNetworkEndpoint {
	matchers := map[string]v1.L7QueryMatcher{
		"q": {Glob: "a*", Any: []string{"x", "y"}},
		"r": {Glob: "b?"},
	}
	return v1.PolicyNetworkEndpoint{
		Host: "api.example.com", Port: 443, Ports: []uint32{443, 8443}, Protocol: "rest",
		TLS: v1.NetworkTLSModeSkip, Enforcement: v1.NetworkEnforcementModeEnforce, Access: v1.NetworkAccessPresetReadWrite,
		Rules: []v1.L7Rule{
			{Allow: &v1.L7Allow{
				Method: "GET", Path: "/v1/**", Command: "cmd", Query: matchers,
				OperationType: "query", OperationName: "Op", Fields: []string{"f1", "f2"}, Params: matchers,
			}},
			{},
		},
		AllowedIPs: []string{"10.0.0.0/8", "192.168.1.1"},
		DenyRules: []v1.L7DenyRule{{
			Method: "DELETE", Path: "/admin", Command: "rm", Query: matchers,
			OperationType: "mutation", OperationName: "Nuke", Fields: []string{"z"}, Params: matchers,
		}},
		AllowEncodedSlash: true, PersistedQueries: "allowlist",
		GraphqlPersistedQueries: map[string]v1.GraphqlOperation{
			"abc": {OperationType: "query", OperationName: "Get", Fields: []string{"id", "name"}},
		},
		GraphqlMaxBodyBytes: 4096, Path: "/graphql",
		WebsocketCredentialRewrite: true, RequestBodyCredentialRewrite: true,
		AllowUninspectedCredentials: true, ProviderCredentialed: true, AdvisorProposed: true,
		CredentialSigning: "sigv4", SigningService: "s3", SigningRegion: "us-east-1",
		JSONRPCMaxBodyBytes: 2048,
		Mcp: &types.McpOptions{
			StrictToolNames: ptr(true), AllowAllKnownMcpMethods: ptr(false), Versions: []string{"2025-11-25"},
		},
		CredentialBinding: &types.NetworkCredentialBinding{Provider: "github"},
	}
}

func maximalPolicy() *v1.SandboxPolicy {
	return &v1.SandboxPolicy{
		Version:    7,
		Filesystem: &v1.FilesystemPolicy{IncludeWorkdir: true, ReadOnly: []string{"/usr"}, ReadWrite: []string{"/tmp", "/work"}},
		Landlock:   &v1.LandlockPolicy{Compatibility: "best_effort"},
		Process:    &v1.ProcessPolicy{RunAsUser: "sandbox", RunAsGroup: "sandbox"},
		NetworkPolicies: map[string]v1.NetworkPolicyRule{
			"github": {
				Name:      "github",
				Endpoints: []v1.PolicyNetworkEndpoint{maximalEndpoint(), {Host: "plain.example.com", Port: 80}},
				Binaries:  []v1.PolicyNetworkBinary{{Path: "/usr/bin/git"}, {Path: "/usr/bin/curl"}},
			},
		},
		NetworkMiddlewares: map[string]types.NetworkMiddlewareConfig{
			"audit": {
				Name: "audit", Middleware: "logger",
				Config:  map[string]any{"level": "debug", "n": float64(3)},
				OnError: "fail_closed",
				Endpoints: &types.MiddlewareEndpointSelector{
					Include: []string{"*.example.com"}, Exclude: []string{"internal.example.com"},
				},
				Order: 5,
			},
		},
	}
}

func TestPolicy_Golden(t *testing.T) {
	got, err := json.Marshal(PolicyFromSDK(maximalPolicy()))
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/policy.json")
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(got))
}

func TestPolicy_RoundTrip(t *testing.T) {
	p := maximalPolicy()
	require.Equal(t, p, PolicyToSDK(PolicyFromSDK(p)))
	require.Nil(t, PolicyFromSDK(nil))
	require.Nil(t, PolicyToSDK(nil))

	// nil vs empty is preserved.
	empty := &v1.SandboxPolicy{
		NetworkPolicies:    map[string]v1.NetworkPolicyRule{},
		NetworkMiddlewares: nil,
		Filesystem:         &v1.FilesystemPolicy{ReadOnly: []string{}},
	}
	require.Equal(t, empty, PolicyToSDK(PolicyFromSDK(empty)))
}

func TestNetworkRule_RoundTrip(t *testing.T) {
	r := maximalPolicy().NetworkPolicies["github"]
	require.Equal(t, &r, NetworkRuleToSDK(*NetworkRuleFromSDK(&r)))
	require.Nil(t, NetworkRuleFromSDK(nil))
}

func TestMergeOps_RoundTrip(t *testing.T) {
	rule := maximalPolicy().NetworkPolicies["github"]
	target := &v1.L7RuleTarget{
		RuleName: "github", Host: "api.example.com", Ports: []uint32{443},
		Path: ptr("/v1"), Binaries: []v1.PolicyNetworkBinary{{Path: "/usr/bin/git"}}, AnyBinary: false,
	}
	ep := maximalEndpoint()
	ops := []v1.PolicyMergeOperation{
		{AddRule: &v1.AddNetworkRule{RuleName: "github", Rule: rule}},
		{RemoveEndpoint: &v1.RemoveNetworkEndpoint{RuleName: "github", Host: "api.example.com", Port: 443}},
		{RemoveRule: &v1.RemoveNetworkRule{RuleName: "github"}},
		{AddDenyRules: &v1.AddDenyRules{Target: target, DenyRules: ep.DenyRules}},
		{AddAllowRules: &v1.AddAllowRules{Target: &v1.L7RuleTarget{RuleName: "r", Host: "h", AnyBinary: true}, Rules: ep.Rules}},
		{RemoveBinary: &v1.RemoveNetworkBinary{RuleName: "github", BinaryPath: "/usr/bin/git"}},
	}
	dtoOps := make([]dto.PolicyMergeOperation, len(ops))
	for i := range ops {
		dtoOps[i] = MergeOpFromSDK(ops[i])
	}
	require.Equal(t, ops, MergeOpsToSDK(dtoOps))
	require.Nil(t, MergeOpsToSDK(nil))
}

func TestEnumStrings(t *testing.T) {
	for _, c := range []struct {
		i v1.NetworkTLSMode
		s string
	}{{0, ""}, {1, "Skip"}, {2, "Terminate"}, {3, "Passthrough"}} {
		require.Equal(t, c.s, tlsString(c.i))
		require.Equal(t, c.i, tlsFromString(c.s))
	}
	for _, c := range []struct {
		i v1.NetworkEnforcementMode
		s string
	}{{0, ""}, {1, "Enforce"}, {2, "Audit"}} {
		require.Equal(t, c.s, enforcementString(c.i))
		require.Equal(t, c.i, enforcementFromString(c.s))
	}
	for _, c := range []struct {
		i v1.NetworkAccessPreset
		s string
	}{{0, ""}, {1, "ReadOnly"}, {2, "ReadWrite"}, {3, "Full"}} {
		require.Equal(t, c.s, accessString(c.i))
		require.Equal(t, c.i, accessFromString(c.s))
	}
	// unknown string -> zero; unknown int -> "".
	require.Equal(t, v1.NetworkTLSMode(0), tlsFromString("bogus"))
	require.Equal(t, v1.NetworkEnforcementMode(0), enforcementFromString("bogus"))
	require.Equal(t, v1.NetworkAccessPreset(0), accessFromString("bogus"))
	require.Equal(t, "", tlsString(99))
	require.Equal(t, "", enforcementString(-1))
	require.Equal(t, "", accessString(42))

	// String-typed enums pass through unchanged.
	s := SandboxFromSDK(&v1.Sandbox{Status: v1.SandboxStatus{Phase: v1.SandboxStopping}, Spec: v1.SandboxSpec{RestartPolicy: v1.SandboxRestartOnFailure}})
	require.Equal(t, "Stopping", s.Status.Phase)
	require.Equal(t, "OnFailure", s.Spec.RestartPolicy)
	require.Equal(t, v1.SandboxRestartAlways, SandboxSpecToSDK(dto.SandboxSpec{RestartPolicy: "Always"}).RestartPolicy)
}
