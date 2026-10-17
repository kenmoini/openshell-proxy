package dto

import (
	"log/slog"
	"sort"
	"time"
)

// RefreshStatus is the state of one credential's refresh lifecycle.
type RefreshStatus struct {
	Provider             string     `json:"provider"`
	ProviderID           string     `json:"providerId,omitempty"`
	CredentialKey        string     `json:"credentialKey"`
	Strategy             string     `json:"strategy,omitempty"`
	Status               string     `json:"status,omitempty"`
	ExpiresAt            *time.Time `json:"expiresAt,omitempty"`
	NextRefreshAt        *time.Time `json:"nextRefreshAt,omitempty"`
	LastRefreshAt        *time.Time `json:"lastRefreshAt,omitempty"`
	LastError            string     `json:"lastError,omitempty"`
	RecoveryAction       string     `json:"recoveryAction,omitempty"`
	FailureCode          string     `json:"failureCode,omitempty"`
	ProviderErrorSubtype string     `json:"providerErrorSubtype,omitempty"`
	LastErrorAt          *time.Time `json:"lastErrorAt,omitempty"`
}

// RefreshStatusList wraps refresh statuses; it is a plain array wrapper, not a
// pager.
type RefreshStatusList struct {
	Items []RefreshStatus `json:"items"`
}

// RefreshConfigRequest configures refresh for a credential. The provider and
// credential key come from the URL path.
type RefreshConfigRequest struct {
	Strategy           string            `json:"strategy" minLength:"1"`
	Material           map[string]string `json:"material,omitempty"`
	SecretMaterialKeys []string          `json:"secretMaterialKeys,omitempty"`
	ExpiresAt          *time.Time        `json:"expiresAt,omitempty"`
}

// LogValue redacts the request for structured logging: only the strategy and
// the sorted material keys are emitted, never material values.
func (r RefreshConfigRequest) LogValue() slog.Value {
	keys := make([]string, 0, len(r.Material))
	for k := range r.Material {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return slog.GroupValue(
		slog.String("strategy", r.Strategy),
		slog.Any("materialKeys", keys),
	)
}

var _ slog.LogValuer = RefreshConfigRequest{}
