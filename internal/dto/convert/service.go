package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// AuthModeFromSDK maps an SDK authorization mode to its API string; the
// unspecified mode maps to the empty string.
func AuthModeFromSDK(m v1.ServiceAuthorizationMode) string {
	switch m {
	case v1.ServiceAuthorizationModeStrip:
		return "Strip"
	case v1.ServiceAuthorizationModeBearerPassthrough:
		return "BearerPassthrough"
	}
	return ""
}

// ServiceFromSDK maps an SDK service endpoint to its DTO.
func ServiceFromSDK(s *v1.ServiceEndpoint) dto.ServiceEndpoint {
	return dto.ServiceEndpoint{
		ID:                s.ID,
		SandboxID:         s.SandboxID,
		Sandbox:           s.Sandbox,
		Name:              s.Name,
		TargetPort:        s.TargetPort,
		Domain:            s.Domain,
		URL:               s.URL,
		Workspace:         s.Workspace,
		AuthorizationMode: AuthModeFromSDK(s.AuthorizationMode),
	}
}
