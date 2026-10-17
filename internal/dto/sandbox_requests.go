package dto

// SandboxCreate is the request body for creating a sandbox.
type SandboxCreate struct {
	Name             string            `json:"name" minLength:"1"`
	Spec             SandboxSpec       `json:"spec"`
	Labels           map[string]string `json:"labels,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	ServiceExposures []ServiceExposure `json:"serviceExposures,omitempty"`
	// Template names a workspace sandbox template to create from.
	Template string `json:"template,omitempty"`
}

// ServiceExposure exposes a loopback HTTP service while the sandbox is created.
type ServiceExposure struct {
	Service           string `json:"service" minLength:"1"`
	TargetPort        uint32 `json:"targetPort"`
	AuthorizationMode string `json:"authorizationMode,omitempty" enum:",Strip,BearerPassthrough"`
}

// WaitRequest asks the proxy to wait for a sandbox phase. Durations are Go
// duration strings such as "30s".
type WaitRequest struct {
	For          string `json:"for" enum:"ready,stopped"`
	Timeout      string `json:"timeout,omitempty"`
	PollInterval string `json:"pollInterval,omitempty"`
}

// ProviderSummary is a provider attached to a sandbox.
type ProviderSummary struct {
	ObjectMeta
	Type string `json:"type,omitempty"`
}

// ProviderAttachment is the result of attaching a provider to a sandbox.
type ProviderAttachment struct {
	Sandbox  Sandbox `json:"sandbox"`
	Attached bool    `json:"attached"`
}

// ProviderDetachment is the result of detaching a provider from a sandbox.
type ProviderDetachment struct {
	Sandbox  Sandbox `json:"sandbox"`
	Detached bool    `json:"detached"`
}
