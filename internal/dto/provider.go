package dto

import (
	"log/slog"
	"sort"
	"time"
)

// CredentialHandle references a credential held by an external driver.
type CredentialHandle struct {
	Driver   string            `json:"driver,omitempty"`
	Handle   string            `json:"handle,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ProviderSpec is the desired configuration of a provider.
type ProviderSpec struct {
	Credentials         map[string]string           `json:"credentials,omitempty"`
	Config              map[string]string           `json:"config,omitempty"`
	CredentialExpiresAt map[string]time.Time        `json:"credentialExpiresAt,omitempty"`
	ProfileWorkspace    string                      `json:"profileWorkspace,omitempty"`
	CredentialHandles   map[string]CredentialHandle `json:"credentialHandles,omitempty"`
}

// Provider is a credential provider. It is used both as the request body and
// the response shape.
type Provider struct {
	ObjectMeta
	Type string       `json:"type,omitempty"`
	Spec ProviderSpec `json:"spec,omitempty"`
}

// LogValue redacts the provider for structured logging: only the name, type
// and the sorted credential keys are emitted, never credential values.
func (p Provider) LogValue() slog.Value {
	keys := make([]string, 0, len(p.Spec.Credentials))
	for k := range p.Spec.Credentials {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return slog.GroupValue(
		slog.String("name", p.Name),
		slog.String("type", p.Type),
		slog.Any("credentialKeys", keys),
	)
}

var _ slog.LogValuer = Provider{}
