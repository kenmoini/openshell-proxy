package dto

// Workspace is a tenancy scope that owns sandboxes and providers.
type Workspace struct {
	ObjectMeta
	Phase string `json:"phase"`
}

// WorkspaceCreate is the request body for creating a workspace.
type WorkspaceCreate struct {
	Name   string            `json:"name" minLength:"1"`
	Labels map[string]string `json:"labels,omitempty"`
}

// WorkspaceMember is a principal's membership in a workspace.
type WorkspaceMember struct {
	ObjectMeta
	PrincipalSubject string `json:"principalSubject"`
	Role             string `json:"role"` // "Admin" | "User"
}

// WorkspaceMemberAdd is the request body for adding a member.
type WorkspaceMemberAdd struct {
	PrincipalSubject string `json:"principalSubject" minLength:"1"`
	Role             string `json:"role" enum:"Admin,User"`
}
