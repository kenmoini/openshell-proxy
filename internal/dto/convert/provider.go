package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// ProviderFromSDK maps an SDK provider to its DTO.
func ProviderFromSDK(p *v1.Provider) dto.Provider {
	out := dto.Provider{
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
		Spec: dto.ProviderSpec{
			Credentials:         p.Spec.Credentials,
			Config:              p.Spec.Config,
			CredentialExpiresAt: p.Spec.CredentialExpiresAt,
			ProfileWorkspace:    p.Spec.ProfileWorkspace,
		},
	}
	if len(p.Spec.CredentialHandles) > 0 {
		out.Spec.CredentialHandles = make(map[string]dto.CredentialHandle, len(p.Spec.CredentialHandles))
		for k, h := range p.Spec.CredentialHandles {
			out.Spec.CredentialHandles[k] = dto.CredentialHandle{Driver: h.Driver, Handle: h.Handle, Metadata: h.Metadata}
		}
	}
	return out
}

// ProviderToSDK maps a provider request body to the SDK type. Only
// client-owned fields are forwarded: name, type, labels, annotations, spec
// (minus server-resolved credential handles) and the body-level optimistic
// concurrency token ResourceVersion. ID, creation and deletion timestamps are
// server-owned and dropped; Workspace comes from the request path.
func ProviderToSDK(workspace string, p dto.Provider) *v1.Provider {
	return &v1.Provider{
		Name:            p.Name,
		Type:            p.Type,
		Labels:          p.Labels,
		Annotations:     p.Annotations,
		ResourceVersion: uint64(p.ResourceVersion),
		Workspace:       workspace,
		Spec: v1.ProviderSpec{
			Credentials:         p.Spec.Credentials,
			Config:              p.Spec.Config,
			CredentialExpiresAt: p.Spec.CredentialExpiresAt,
			ProfileWorkspace:    p.Spec.ProfileWorkspace,
		},
	}
}
