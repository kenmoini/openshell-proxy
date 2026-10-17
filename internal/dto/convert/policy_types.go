package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// mapSlice converts a slice element-wise, preserving nil vs empty.
func mapSlice[A, B any](in []A, f func(A) B) []B {
	if in == nil {
		return nil
	}
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

// mapMap converts map values, preserving nil vs empty.
func mapMap[K comparable, A, B any](in map[K]A, f func(A) B) map[K]B {
	if in == nil {
		return nil
	}
	out := make(map[K]B, len(in))
	for k, v := range in {
		out[k] = f(v)
	}
	return out
}

// mapPtr converts through a pointer; nil in gives nil out.
func mapPtr[A, B any](in *A, f func(A) B) *B {
	if in == nil {
		return nil
	}
	out := f(*in)
	return &out
}

func tlsString(m v1.NetworkTLSMode) string {
	switch m {
	case v1.NetworkTLSModeSkip:
		return "Skip"
	case v1.NetworkTLSModeTerminate:
		return "Terminate"
	case v1.NetworkTLSModePassthrough:
		return "Passthrough"
	}
	return ""
}

func tlsFromString(s string) v1.NetworkTLSMode {
	switch s {
	case "Skip":
		return v1.NetworkTLSModeSkip
	case "Terminate":
		return v1.NetworkTLSModeTerminate
	case "Passthrough":
		return v1.NetworkTLSModePassthrough
	}
	return v1.NetworkTLSModeUnspecified
}

func enforcementString(m v1.NetworkEnforcementMode) string {
	switch m {
	case v1.NetworkEnforcementModeEnforce:
		return "Enforce"
	case v1.NetworkEnforcementModeAudit:
		return "Audit"
	}
	return ""
}

func enforcementFromString(s string) v1.NetworkEnforcementMode {
	switch s {
	case "Enforce":
		return v1.NetworkEnforcementModeEnforce
	case "Audit":
		return v1.NetworkEnforcementModeAudit
	}
	return v1.NetworkEnforcementModeUnspecified
}

func accessString(m v1.NetworkAccessPreset) string {
	switch m {
	case v1.NetworkAccessPresetReadOnly:
		return "ReadOnly"
	case v1.NetworkAccessPresetReadWrite:
		return "ReadWrite"
	case v1.NetworkAccessPresetFull:
		return "Full"
	}
	return ""
}

func accessFromString(s string) v1.NetworkAccessPreset {
	switch s {
	case "ReadOnly":
		return v1.NetworkAccessPresetReadOnly
	case "ReadWrite":
		return v1.NetworkAccessPresetReadWrite
	case "Full":
		return v1.NetworkAccessPresetFull
	}
	return v1.NetworkAccessPresetUnspecified
}

// PolicyFromSDK maps an SDK sandbox policy to its DTO; nil in gives nil out.
func PolicyFromSDK(p *v1.SandboxPolicy) *dto.SandboxPolicy {
	if p == nil {
		return nil
	}
	return &dto.SandboxPolicy{
		Version: p.Version,
		Filesystem: mapPtr(p.Filesystem, func(f v1.FilesystemPolicy) dto.FilesystemPolicy {
			return dto.FilesystemPolicy{IncludeWorkdir: f.IncludeWorkdir, ReadOnly: f.ReadOnly, ReadWrite: f.ReadWrite}
		}),
		Landlock: mapPtr(p.Landlock, func(l v1.LandlockPolicy) dto.LandlockPolicy {
			return dto.LandlockPolicy{Compatibility: l.Compatibility}
		}),
		Process: mapPtr(p.Process, func(q v1.ProcessPolicy) dto.ProcessPolicy {
			return dto.ProcessPolicy{RunAsUser: q.RunAsUser, RunAsGroup: q.RunAsGroup}
		}),
		NetworkPolicies:    mapMap(p.NetworkPolicies, ruleFromSDK),
		NetworkMiddlewares: mapMap(p.NetworkMiddlewares, middlewareFromSDK),
	}
}

// PolicyToSDK maps a DTO sandbox policy to the SDK type; nil in gives nil out.
func PolicyToSDK(p *dto.SandboxPolicy) *v1.SandboxPolicy {
	if p == nil {
		return nil
	}
	return &v1.SandboxPolicy{
		Version: p.Version,
		Filesystem: mapPtr(p.Filesystem, func(f dto.FilesystemPolicy) v1.FilesystemPolicy {
			return v1.FilesystemPolicy{IncludeWorkdir: f.IncludeWorkdir, ReadOnly: f.ReadOnly, ReadWrite: f.ReadWrite}
		}),
		Landlock: mapPtr(p.Landlock, func(l dto.LandlockPolicy) v1.LandlockPolicy {
			return v1.LandlockPolicy{Compatibility: l.Compatibility}
		}),
		Process: mapPtr(p.Process, func(q dto.ProcessPolicy) v1.ProcessPolicy {
			return v1.ProcessPolicy{RunAsUser: q.RunAsUser, RunAsGroup: q.RunAsGroup}
		}),
		NetworkPolicies:    mapMap(p.NetworkPolicies, ruleToSDK),
		NetworkMiddlewares: mapMap(p.NetworkMiddlewares, middlewareToSDK),
	}
}

func middlewareFromSDK(m types.NetworkMiddlewareConfig) dto.NetworkMiddlewareConfig {
	return dto.NetworkMiddlewareConfig{
		Name: m.Name, Middleware: m.Middleware, Config: m.Config, OnError: m.OnError, Order: m.Order,
		Endpoints: mapPtr(m.Endpoints, func(e types.MiddlewareEndpointSelector) dto.MiddlewareEndpointSelector {
			return dto.MiddlewareEndpointSelector{Include: e.Include, Exclude: e.Exclude}
		}),
	}
}

func middlewareToSDK(m dto.NetworkMiddlewareConfig) types.NetworkMiddlewareConfig {
	return types.NetworkMiddlewareConfig{
		Name: m.Name, Middleware: m.Middleware, Config: m.Config, OnError: m.OnError, Order: m.Order,
		Endpoints: mapPtr(m.Endpoints, func(e dto.MiddlewareEndpointSelector) types.MiddlewareEndpointSelector {
			return types.MiddlewareEndpointSelector{Include: e.Include, Exclude: e.Exclude}
		}),
	}
}

// NetworkRuleFromSDK maps an SDK network rule to its DTO; nil in gives nil out.
func NetworkRuleFromSDK(r *v1.NetworkPolicyRule) *dto.NetworkPolicyRule {
	return mapPtr(r, ruleFromSDK)
}

// NetworkRuleToSDK maps a DTO network rule to the SDK type.
func NetworkRuleToSDK(r dto.NetworkPolicyRule) *v1.NetworkPolicyRule {
	out := ruleToSDK(r)
	return &out
}

func ruleFromSDK(r v1.NetworkPolicyRule) dto.NetworkPolicyRule {
	return dto.NetworkPolicyRule{
		Name:      r.Name,
		Endpoints: mapSlice(r.Endpoints, endpointFromSDK),
		Binaries:  mapSlice(r.Binaries, binaryFromSDK),
	}
}

func ruleToSDK(r dto.NetworkPolicyRule) v1.NetworkPolicyRule {
	return v1.NetworkPolicyRule{
		Name:      r.Name,
		Endpoints: mapSlice(r.Endpoints, endpointToSDK),
		Binaries:  mapSlice(r.Binaries, binaryToSDK),
	}
}

func binaryFromSDK(b v1.PolicyNetworkBinary) dto.PolicyNetworkBinary {
	return dto.PolicyNetworkBinary{Path: b.Path}
}

func binaryToSDK(b dto.PolicyNetworkBinary) v1.PolicyNetworkBinary {
	return v1.PolicyNetworkBinary{Path: b.Path}
}

func endpointFromSDK(e v1.PolicyNetworkEndpoint) dto.PolicyNetworkEndpoint {
	return dto.PolicyNetworkEndpoint{
		Host: e.Host, Port: e.Port, Ports: e.Ports, Protocol: e.Protocol,
		TLS: tlsString(e.TLS), Enforcement: enforcementString(e.Enforcement), Access: accessString(e.Access),
		Rules:                        mapSlice(e.Rules, l7RuleFromSDK),
		AllowedIPs:                   e.AllowedIPs,
		DenyRules:                    mapSlice(e.DenyRules, denyFromSDK),
		AllowEncodedSlash:            e.AllowEncodedSlash,
		PersistedQueries:             e.PersistedQueries,
		GraphqlPersistedQueries:      mapMap(e.GraphqlPersistedQueries, gqlFromSDK),
		GraphqlMaxBodyBytes:          e.GraphqlMaxBodyBytes,
		Path:                         e.Path,
		WebsocketCredentialRewrite:   e.WebsocketCredentialRewrite,
		RequestBodyCredentialRewrite: e.RequestBodyCredentialRewrite,
		AllowUninspectedCredentials:  e.AllowUninspectedCredentials,
		ProviderCredentialed:         e.ProviderCredentialed,
		AdvisorProposed:              e.AdvisorProposed,
		CredentialSigning:            e.CredentialSigning,
		SigningService:               e.SigningService,
		SigningRegion:                e.SigningRegion,
		JSONRPCMaxBodyBytes:          e.JSONRPCMaxBodyBytes,
		Mcp: mapPtr(e.Mcp, func(m types.McpOptions) dto.McpOptions {
			return dto.McpOptions{StrictToolNames: m.StrictToolNames, AllowAllKnownMcpMethods: m.AllowAllKnownMcpMethods, Versions: m.Versions}
		}),
		CredentialBinding: mapPtr(e.CredentialBinding, func(c types.NetworkCredentialBinding) dto.NetworkCredentialBinding {
			return dto.NetworkCredentialBinding{Provider: c.Provider}
		}),
	}
}

func endpointToSDK(e dto.PolicyNetworkEndpoint) v1.PolicyNetworkEndpoint {
	return v1.PolicyNetworkEndpoint{
		Host: e.Host, Port: e.Port, Ports: e.Ports, Protocol: e.Protocol,
		TLS: tlsFromString(e.TLS), Enforcement: enforcementFromString(e.Enforcement), Access: accessFromString(e.Access),
		Rules:                        mapSlice(e.Rules, l7RuleToSDK),
		AllowedIPs:                   e.AllowedIPs,
		DenyRules:                    mapSlice(e.DenyRules, denyToSDK),
		AllowEncodedSlash:            e.AllowEncodedSlash,
		PersistedQueries:             e.PersistedQueries,
		GraphqlPersistedQueries:      mapMap(e.GraphqlPersistedQueries, gqlToSDK),
		GraphqlMaxBodyBytes:          e.GraphqlMaxBodyBytes,
		Path:                         e.Path,
		WebsocketCredentialRewrite:   e.WebsocketCredentialRewrite,
		RequestBodyCredentialRewrite: e.RequestBodyCredentialRewrite,
		AllowUninspectedCredentials:  e.AllowUninspectedCredentials,
		ProviderCredentialed:         e.ProviderCredentialed,
		AdvisorProposed:              e.AdvisorProposed,
		CredentialSigning:            e.CredentialSigning,
		SigningService:               e.SigningService,
		SigningRegion:                e.SigningRegion,
		JSONRPCMaxBodyBytes:          e.JSONRPCMaxBodyBytes,
		Mcp: mapPtr(e.Mcp, func(m dto.McpOptions) types.McpOptions {
			return types.McpOptions{StrictToolNames: m.StrictToolNames, AllowAllKnownMcpMethods: m.AllowAllKnownMcpMethods, Versions: m.Versions}
		}),
		CredentialBinding: mapPtr(e.CredentialBinding, func(c dto.NetworkCredentialBinding) types.NetworkCredentialBinding {
			return types.NetworkCredentialBinding{Provider: c.Provider}
		}),
	}
}

func matcherFromSDK(m v1.L7QueryMatcher) dto.L7QueryMatcher {
	return dto.L7QueryMatcher{Glob: m.Glob, Any: m.Any}
}

func matcherToSDK(m dto.L7QueryMatcher) v1.L7QueryMatcher {
	return v1.L7QueryMatcher{Glob: m.Glob, Any: m.Any}
}

func gqlFromSDK(g types.GraphqlOperation) dto.GraphqlOperation {
	return dto.GraphqlOperation{OperationType: g.OperationType, OperationName: g.OperationName, Fields: g.Fields}
}

func gqlToSDK(g dto.GraphqlOperation) types.GraphqlOperation {
	return types.GraphqlOperation{OperationType: g.OperationType, OperationName: g.OperationName, Fields: g.Fields}
}

func l7RuleFromSDK(r v1.L7Rule) dto.L7Rule {
	return dto.L7Rule{Allow: mapPtr(r.Allow, func(a v1.L7Allow) dto.L7Allow {
		return dto.L7Allow{
			Method: a.Method, Path: a.Path, Command: a.Command, Query: mapMap(a.Query, matcherFromSDK),
			OperationType: a.OperationType, OperationName: a.OperationName, Fields: a.Fields,
			Params: mapMap(a.Params, matcherFromSDK),
		}
	})}
}

func l7RuleToSDK(r dto.L7Rule) v1.L7Rule {
	return v1.L7Rule{Allow: mapPtr(r.Allow, func(a dto.L7Allow) v1.L7Allow {
		return v1.L7Allow{
			Method: a.Method, Path: a.Path, Command: a.Command, Query: mapMap(a.Query, matcherToSDK),
			OperationType: a.OperationType, OperationName: a.OperationName, Fields: a.Fields,
			Params: mapMap(a.Params, matcherToSDK),
		}
	})}
}

func denyFromSDK(d v1.L7DenyRule) dto.L7DenyRule {
	return dto.L7DenyRule{
		Method: d.Method, Path: d.Path, Command: d.Command, Query: mapMap(d.Query, matcherFromSDK),
		OperationType: d.OperationType, OperationName: d.OperationName, Fields: d.Fields,
		Params: mapMap(d.Params, matcherFromSDK),
	}
}

func denyToSDK(d dto.L7DenyRule) v1.L7DenyRule {
	return v1.L7DenyRule{
		Method: d.Method, Path: d.Path, Command: d.Command, Query: mapMap(d.Query, matcherToSDK),
		OperationType: d.OperationType, OperationName: d.OperationName, Fields: d.Fields,
		Params: mapMap(d.Params, matcherToSDK),
	}
}

func targetFromSDK(t v1.L7RuleTarget) dto.L7RuleTarget {
	return dto.L7RuleTarget{
		RuleName: t.RuleName, Host: t.Host, Ports: t.Ports, Path: t.Path,
		Binaries: mapSlice(t.Binaries, binaryFromSDK), AnyBinary: t.AnyBinary,
	}
}

func targetToSDK(t dto.L7RuleTarget) v1.L7RuleTarget {
	return v1.L7RuleTarget{
		RuleName: t.RuleName, Host: t.Host, Ports: t.Ports, Path: t.Path,
		Binaries: mapSlice(t.Binaries, binaryToSDK), AnyBinary: t.AnyBinary,
	}
}

// MergeOpFromSDK maps one SDK merge operation to its DTO.
func MergeOpFromSDK(o v1.PolicyMergeOperation) dto.PolicyMergeOperation {
	return dto.PolicyMergeOperation{
		AddRule: mapPtr(o.AddRule, func(a v1.AddNetworkRule) dto.AddNetworkRule {
			return dto.AddNetworkRule{RuleName: a.RuleName, Rule: ruleFromSDK(a.Rule)}
		}),
		RemoveEndpoint: mapPtr(o.RemoveEndpoint, func(r v1.RemoveNetworkEndpoint) dto.RemoveNetworkEndpoint {
			return dto.RemoveNetworkEndpoint{RuleName: r.RuleName, Host: r.Host, Port: r.Port}
		}),
		RemoveRule: mapPtr(o.RemoveRule, func(r v1.RemoveNetworkRule) dto.RemoveNetworkRule {
			return dto.RemoveNetworkRule{RuleName: r.RuleName}
		}),
		AddDenyRules: mapPtr(o.AddDenyRules, func(a v1.AddDenyRules) dto.AddDenyRules {
			return dto.AddDenyRules{Target: mapPtr(a.Target, targetFromSDK), DenyRules: mapSlice(a.DenyRules, denyFromSDK)}
		}),
		AddAllowRules: mapPtr(o.AddAllowRules, func(a v1.AddAllowRules) dto.AddAllowRules {
			return dto.AddAllowRules{Target: mapPtr(a.Target, targetFromSDK), Rules: mapSlice(a.Rules, l7RuleFromSDK)}
		}),
		RemoveBinary: mapPtr(o.RemoveBinary, func(r v1.RemoveNetworkBinary) dto.RemoveNetworkBinary {
			return dto.RemoveNetworkBinary{RuleName: r.RuleName, BinaryPath: r.BinaryPath}
		}),
	}
}

func mergeOpToSDK(o dto.PolicyMergeOperation) v1.PolicyMergeOperation {
	return v1.PolicyMergeOperation{
		AddRule: mapPtr(o.AddRule, func(a dto.AddNetworkRule) v1.AddNetworkRule {
			return v1.AddNetworkRule{RuleName: a.RuleName, Rule: ruleToSDK(a.Rule)}
		}),
		RemoveEndpoint: mapPtr(o.RemoveEndpoint, func(r dto.RemoveNetworkEndpoint) v1.RemoveNetworkEndpoint {
			return v1.RemoveNetworkEndpoint{RuleName: r.RuleName, Host: r.Host, Port: r.Port}
		}),
		RemoveRule: mapPtr(o.RemoveRule, func(r dto.RemoveNetworkRule) v1.RemoveNetworkRule {
			return v1.RemoveNetworkRule{RuleName: r.RuleName}
		}),
		AddDenyRules: mapPtr(o.AddDenyRules, func(a dto.AddDenyRules) v1.AddDenyRules {
			return v1.AddDenyRules{Target: mapPtr(a.Target, targetToSDK), DenyRules: mapSlice(a.DenyRules, denyToSDK)}
		}),
		AddAllowRules: mapPtr(o.AddAllowRules, func(a dto.AddAllowRules) v1.AddAllowRules {
			return v1.AddAllowRules{Target: mapPtr(a.Target, targetToSDK), Rules: mapSlice(a.Rules, l7RuleToSDK)}
		}),
		RemoveBinary: mapPtr(o.RemoveBinary, func(r dto.RemoveNetworkBinary) v1.RemoveNetworkBinary {
			return v1.RemoveNetworkBinary{RuleName: r.RuleName, BinaryPath: r.BinaryPath}
		}),
	}
}

// MergeOpsToSDK maps DTO merge operations to SDK operations; nil in gives nil out.
func MergeOpsToSDK(ops []dto.PolicyMergeOperation) []v1.PolicyMergeOperation {
	return mapSlice(ops, mergeOpToSDK)
}
