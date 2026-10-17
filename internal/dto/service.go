package dto

// ServiceEndpoint is an exposed sandbox service.
type ServiceEndpoint struct {
	ID                string `json:"id,omitempty"`
	SandboxID         string `json:"sandboxId,omitempty"`
	Sandbox           string `json:"sandbox,omitempty"`
	Name              string `json:"name"`
	TargetPort        uint32 `json:"targetPort"`
	Domain            bool   `json:"domain"`
	URL               string `json:"url"`
	Workspace         string `json:"workspace,omitempty"`
	AuthorizationMode string `json:"authorizationMode,omitempty"`
}

// ServiceExpose is the request body for exposing a sandbox service.
type ServiceExpose struct {
	Name              string `json:"name" minLength:"1"`
	TargetPort        uint32 `json:"targetPort" minimum:"1" maximum:"65535"`
	Domain            bool   `json:"domain,omitempty"`
	AuthorizationMode string `json:"authorizationMode,omitempty" enum:"Strip,BearerPassthrough"`
}
