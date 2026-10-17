package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
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

// ProviderToSDK maps a provider DTO to the SDK type. Server-assigned fields
// (ID, created time) are not set except those the client may carry.
func ProviderToSDK(p dto.Provider) *v1.Provider {
	out := &v1.Provider{
		ID:              p.ID,
		Name:            p.Name,
		Type:            p.Type,
		Labels:          p.Labels,
		Annotations:     p.Annotations,
		ResourceVersion: uint64(p.ResourceVersion),
		Workspace:       p.Workspace,
		Spec: v1.ProviderSpec{
			Credentials:         p.Spec.Credentials,
			Config:              p.Spec.Config,
			CredentialExpiresAt: p.Spec.CredentialExpiresAt,
			ProfileWorkspace:    p.Spec.ProfileWorkspace,
		},
	}
	if p.CreatedAt != nil {
		out.CreatedAt = *p.CreatedAt
	}
	out.DeletionTimestamp = p.DeletionTimestamp
	if len(p.Spec.CredentialHandles) > 0 {
		out.Spec.CredentialHandles = make(map[string]types.CredentialHandle, len(p.Spec.CredentialHandles))
		for k, h := range p.Spec.CredentialHandles {
			out.Spec.CredentialHandles[k] = types.CredentialHandle{Driver: h.Driver, Handle: h.Handle, Metadata: h.Metadata}
		}
	}
	return out
}
