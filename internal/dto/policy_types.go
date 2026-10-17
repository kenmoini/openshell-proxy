package dto

// SandboxPolicy is the top-level security policy of a sandbox.
type SandboxPolicy struct {
	Version            uint32                             `json:"version,omitempty"`
	Filesystem         *FilesystemPolicy                  `json:"filesystem,omitempty"`
	Landlock           *LandlockPolicy                    `json:"landlock,omitempty"`
	Process            *ProcessPolicy                     `json:"process,omitempty"`
	NetworkPolicies    map[string]NetworkPolicyRule       `json:"networkPolicies,omitempty"`
	NetworkMiddlewares map[string]NetworkMiddlewareConfig `json:"networkMiddlewares,omitempty"`
}

// FilesystemPolicy controls directory access.
type FilesystemPolicy struct {
	IncludeWorkdir bool     `json:"includeWorkdir,omitempty"`
	ReadOnly       []string `json:"readOnly,omitempty"`
	ReadWrite      []string `json:"readWrite,omitempty"`
}

// LandlockPolicy configures the Linux Landlock LSM.
type LandlockPolicy struct {
	Compatibility string `json:"compatibility,omitempty"`
}

// ProcessPolicy controls the identity of sandboxed processes.
type ProcessPolicy struct {
	RunAsUser  string `json:"runAsUser,omitempty"`
	RunAsGroup string `json:"runAsGroup,omitempty"`
}

// NetworkMiddlewareConfig configures a supervisor middleware pipeline.
type NetworkMiddlewareConfig struct {
	Name       string                      `json:"name,omitempty"`
	Middleware string                      `json:"middleware,omitempty"`
	Config     map[string]any              `json:"config,omitempty"`
	OnError    string                      `json:"onError,omitempty"`
	Endpoints  *MiddlewareEndpointSelector `json:"endpoints,omitempty"`
	Order      int32                       `json:"order,omitempty"`
}

// MiddlewareEndpointSelector selects destinations by host glob.
type MiddlewareEndpointSelector struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// NetworkPolicyRule is a named network policy rule.
type NetworkPolicyRule struct {
	Name      string                  `json:"name,omitempty"`
	Endpoints []PolicyNetworkEndpoint `json:"endpoints,omitempty"`
	Binaries  []PolicyNetworkBinary   `json:"binaries,omitempty"`
}

// PolicyNetworkBinary identifies a binary subject to network policy.
type PolicyNetworkBinary struct {
	Path string `json:"path,omitempty"`
}

// PolicyNetworkEndpoint is a full network endpoint with access controls.
type PolicyNetworkEndpoint struct {
	Host                         string                      `json:"host,omitempty"`
	Port                         uint32                      `json:"port,omitempty"`
	Ports                        []uint32                    `json:"ports,omitempty"`
	Protocol                     string                      `json:"protocol,omitempty"`
	TLS                          string                      `json:"tls,omitempty" enum:",Skip,Terminate,Passthrough"`
	Enforcement                  string                      `json:"enforcement,omitempty" enum:",Enforce,Audit"`
	Access                       string                      `json:"access,omitempty" enum:",ReadOnly,ReadWrite,Full"`
	Rules                        []L7Rule                    `json:"rules,omitempty"`
	AllowedIPs                   []string                    `json:"allowedIps,omitempty"`
	DenyRules                    []L7DenyRule                `json:"denyRules,omitempty"`
	AllowEncodedSlash            bool                        `json:"allowEncodedSlash,omitempty"`
	PersistedQueries             string                      `json:"persistedQueries,omitempty"`
	GraphqlPersistedQueries      map[string]GraphqlOperation `json:"graphqlPersistedQueries,omitempty"`
	GraphqlMaxBodyBytes          uint32                      `json:"graphqlMaxBodyBytes,omitempty"`
	Path                         string                      `json:"path,omitempty"`
	WebsocketCredentialRewrite   bool                        `json:"websocketCredentialRewrite,omitempty"`
	RequestBodyCredentialRewrite bool                        `json:"requestBodyCredentialRewrite,omitempty"`
	AllowUninspectedCredentials  bool                        `json:"allowUninspectedCredentials,omitempty"`
	ProviderCredentialed         bool                        `json:"providerCredentialed,omitempty"`
	AdvisorProposed              bool                        `json:"advisorProposed,omitempty"`
	CredentialSigning            string                      `json:"credentialSigning,omitempty"`
	SigningService               string                      `json:"signingService,omitempty"`
	SigningRegion                string                      `json:"signingRegion,omitempty"`
	JSONRPCMaxBodyBytes          uint32                      `json:"jsonrpcMaxBodyBytes,omitempty"`
	Mcp                          *McpOptions                 `json:"mcp,omitempty"`
	CredentialBinding            *NetworkCredentialBinding   `json:"credentialBinding,omitempty"`
}

// NetworkCredentialBinding binds an endpoint to provider credentials.
type NetworkCredentialBinding struct {
	Provider string `json:"provider,omitempty"`
}

// L7Rule wraps an L7 allow rule.
type L7Rule struct {
	Allow *L7Allow `json:"allow,omitempty"`
}

// L7Allow specifies layer-7 allow criteria.
type L7Allow struct {
	Method        string                    `json:"method,omitempty"`
	Path          string                    `json:"path,omitempty"`
	Command       string                    `json:"command,omitempty"`
	Query         map[string]L7QueryMatcher `json:"query,omitempty"`
	OperationType string                    `json:"operationType,omitempty"`
	OperationName string                    `json:"operationName,omitempty"`
	Fields        []string                  `json:"fields,omitempty"`
	Params        map[string]L7QueryMatcher `json:"params,omitempty"`
}

// L7DenyRule specifies layer-7 deny criteria.
type L7DenyRule struct {
	Method        string                    `json:"method,omitempty"`
	Path          string                    `json:"path,omitempty"`
	Command       string                    `json:"command,omitempty"`
	Query         map[string]L7QueryMatcher `json:"query,omitempty"`
	OperationType string                    `json:"operationType,omitempty"`
	OperationName string                    `json:"operationName,omitempty"`
	Fields        []string                  `json:"fields,omitempty"`
	Params        map[string]L7QueryMatcher `json:"params,omitempty"`
}

// L7QueryMatcher matches a parameter by glob or exact values.
type L7QueryMatcher struct {
	Glob string   `json:"glob,omitempty"`
	Any  []string `json:"any,omitempty"`
}

// McpOptions configures MCP-specific policy controls.
type McpOptions struct {
	StrictToolNames         *bool    `json:"strictToolNames,omitempty"`
	AllowAllKnownMcpMethods *bool    `json:"allowAllKnownMcpMethods,omitempty"`
	Versions                []string `json:"versions,omitempty"`
}

// GraphqlOperation describes a persisted GraphQL operation.
type GraphqlOperation struct {
	OperationType string   `json:"operationType,omitempty"`
	OperationName string   `json:"operationName,omitempty"`
	Fields        []string `json:"fields,omitempty"`
}

// PolicyMergeOperation is a single atomic policy mutation; exactly one field is set.
type PolicyMergeOperation struct {
	AddRule        *AddNetworkRule        `json:"addRule,omitempty"`
	RemoveEndpoint *RemoveNetworkEndpoint `json:"removeEndpoint,omitempty"`
	RemoveRule     *RemoveNetworkRule     `json:"removeRule,omitempty"`
	AddDenyRules   *AddDenyRules          `json:"addDenyRules,omitempty"`
	AddAllowRules  *AddAllowRules         `json:"addAllowRules,omitempty"`
	RemoveBinary   *RemoveNetworkBinary   `json:"removeBinary,omitempty"`
}

// AddNetworkRule adds a named rule.
type AddNetworkRule struct {
	RuleName string            `json:"ruleName,omitempty"`
	Rule     NetworkPolicyRule `json:"rule"`
}

// RemoveNetworkEndpoint removes an endpoint from a rule.
type RemoveNetworkEndpoint struct {
	RuleName string `json:"ruleName,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     uint32 `json:"port,omitempty"`
}

// RemoveNetworkRule removes a named rule.
type RemoveNetworkRule struct {
	RuleName string `json:"ruleName,omitempty"`
}

// L7RuleTarget identifies an endpoint and its complete affected scope.
type L7RuleTarget struct {
	RuleName  string                `json:"ruleName,omitempty"`
	Host      string                `json:"host,omitempty"`
	Ports     []uint32              `json:"ports,omitempty"`
	Path      *string               `json:"path,omitempty"`
	Binaries  []PolicyNetworkBinary `json:"binaries,omitempty"`
	AnyBinary bool                  `json:"anyBinary,omitempty"`
}

// AddDenyRules appends deny rules to an endpoint.
type AddDenyRules struct {
	Target    *L7RuleTarget `json:"target,omitempty"`
	DenyRules []L7DenyRule  `json:"denyRules,omitempty"`
}

// AddAllowRules appends allow rules to an endpoint.
type AddAllowRules struct {
	Target *L7RuleTarget `json:"target,omitempty"`
	Rules  []L7Rule      `json:"rules,omitempty"`
}

// RemoveNetworkBinary removes a binary from a rule.
type RemoveNetworkBinary struct {
	RuleName   string `json:"ruleName,omitempty"`
	BinaryPath string `json:"binaryPath,omitempty"`
}
