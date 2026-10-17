package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// WorkspaceFromSDK maps an SDK workspace to its DTO.
func WorkspaceFromSDK(w *v1.Workspace) dto.Workspace {
	return dto.Workspace{
		ObjectMeta: dto.ObjectMeta{
			ID:                w.ID,
			Name:              w.Name,
			Workspace:         w.Workspace,
			CreatedAt:         TimePtr(w.CreatedAt),
			DeletionTimestamp: w.DeletionTimestamp,
			Labels:            w.Labels,
			Annotations:       w.Annotations,
			ResourceVersion:   dto.Uint64String(w.ResourceVersion),
		},
		Phase: string(w.Phase),
	}
}

// MemberFromSDK maps an SDK workspace member to its DTO. The name falls back
// to the principal subject when the gateway leaves it empty.
func MemberFromSDK(m *v1.WorkspaceMember) dto.WorkspaceMember {
	name := m.Name
	if name == "" {
		name = m.PrincipalSubject
	}
	return dto.WorkspaceMember{
		ObjectMeta: dto.ObjectMeta{
			ID:              m.ID,
			Name:            name,
			CreatedAt:       TimePtr(m.CreatedAt),
			Labels:          m.Labels,
			Annotations:     m.Annotations,
			ResourceVersion: dto.Uint64String(m.ResourceVersion),
		},
		PrincipalSubject: m.PrincipalSubject,
		Role:             string(m.Role),
	}
}

// RoleToSDK maps the API role string to the SDK role. Input is validated by
// the request schema, so anything else is Unknown.
func RoleToSDK(s string) v1.WorkspaceRole {
	switch s {
	case "Admin":
		return v1.WorkspaceRoleAdmin
	case "User":
		return v1.WorkspaceRoleUser
	}
	return v1.WorkspaceRoleUnknown
}
