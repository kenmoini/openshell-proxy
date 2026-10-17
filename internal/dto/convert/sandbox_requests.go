package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// ServiceExposuresToSDK maps request exposures to SDK exposures.
func ServiceExposuresToSDK(in []dto.ServiceExposure) []v1.ServiceExposure {
	if len(in) == 0 {
		return nil
	}
	out := make([]v1.ServiceExposure, 0, len(in))
	for _, e := range in {
		out = append(out, v1.ServiceExposure{
			Service:           e.Service,
			TargetPort:        e.TargetPort,
			AuthorizationMode: authModeToSDK(e.AuthorizationMode),
		})
	}
	return out
}

func authModeToSDK(s string) v1.ServiceAuthorizationMode {
	switch s {
	case "Strip":
		return v1.ServiceAuthorizationModeStrip
	case "BearerPassthrough":
		return v1.ServiceAuthorizationModeBearerPassthrough
	}
	return 0
}

// ProviderSummaryFromSDK maps an SDK provider to its summary DTO.
func ProviderSummaryFromSDK(p *v1.Provider) dto.ProviderSummary {
	return dto.ProviderSummary{
		ObjectMeta: dto.ObjectMeta{
			ID:                p.ID,
			Name:              p.Name,
			Workspace:         p.Workspace,
			CreatedAt:         TimePtr(p.CreatedAt),
			DeletionTimestamp: p.DeletionTimestamp,
			Labels:            p.Labels,
			Annotations:       p.Annotations,
			ResourceVersion:   dto.Uint64String(p.ResourceVersion),
		},
		Type: p.Type,
	}
}

// AttachFromSDK maps an attach result.
func AttachFromSDK(r *v1.AttachProviderResult) dto.ProviderAttachment {
	out := dto.ProviderAttachment{Attached: r.Attached}
	if r.Sandbox != nil {
		out.Sandbox = SandboxFromSDK(r.Sandbox)
	}
	return out
}

// DetachFromSDK maps a detach result.
func DetachFromSDK(r *v1.DetachProviderResult) dto.ProviderDetachment {
	out := dto.ProviderDetachment{Detached: r.Detached}
	if r.Sandbox != nil {
		out.Sandbox = SandboxFromSDK(r.Sandbox)
	}
	return out
}
