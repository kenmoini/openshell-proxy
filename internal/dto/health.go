// Package dto holds the JSON request/response shapes of the REST API.
package dto

// Health is the gateway liveness result.
type Health struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version,omitempty"`
}

// ComputeDriverInfo describes a compute backend available on the gateway.
type ComputeDriverInfo struct {
	Name          string `json:"name"`
	DriverName    string `json:"driverName,omitempty"`
	DriverVersion string `json:"driverVersion,omitempty"`
}

// ExtensionInfo describes one negotiated gateway extension.
type ExtensionInfo struct {
	Kind                  string   `json:"kind"`
	ConfiguredName        string   `json:"configuredName,omitempty"`
	ImplementationName    string   `json:"implementationName,omitempty"`
	ImplementationVersion string   `json:"implementationVersion,omitempty"`
	ProtocolMajor         uint32   `json:"protocolMajor"`
	ProtocolMinor         uint32   `json:"protocolMinor"`
	SupportedCapabilities []string `json:"supportedCapabilities,omitempty"`
	RequiredCapabilities  []string `json:"requiredCapabilities,omitempty"`
}

// GatewayInfo is operational metadata about the gateway.
type GatewayInfo struct {
	Status         string              `json:"status"`
	Version        string              `json:"version,omitempty"`
	ComputeDrivers []ComputeDriverInfo `json:"computeDrivers"`
	Extensions     []ExtensionInfo     `json:"extensions"`
}

// CurrentUser is the authenticated caller's identity.
type CurrentUser struct {
	Subject          string   `json:"subject"`
	DisplayName      string   `json:"displayName,omitempty"`
	IdentityProvider string   `json:"identityProvider,omitempty"`
	Roles            []string `json:"roles"`
	Scopes           []string `json:"scopes"`
}
