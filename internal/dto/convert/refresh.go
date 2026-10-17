package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// RefreshStatusFromSDK maps an SDK refresh status to its DTO.
func RefreshStatusFromSDK(s *v1.RefreshStatus) dto.RefreshStatus {
	return dto.RefreshStatus{
		Provider:             s.Provider,
		ProviderID:           s.ProviderID,
		CredentialKey:        s.CredentialKey,
		Strategy:             string(s.Strategy),
		Status:               s.Status,
		ExpiresAt:            TimePtr(s.ExpiresAt),
		NextRefreshAt:        TimePtr(s.NextRefreshAt),
		LastRefreshAt:        TimePtr(s.LastRefreshAt),
		LastError:            s.LastError,
		RecoveryAction:       s.RecoveryAction.String(),
		FailureCode:          s.FailureCode,
		ProviderErrorSubtype: s.ProviderErrorSubtype,
		LastErrorAt:          TimePtr(s.LastErrorAt),
	}
}

// RefreshConfigToSDK builds the SDK refresh config; provider and credential
// key come from the URL path.
func RefreshConfigToSDK(provider, key string, req dto.RefreshConfigRequest) *v1.RefreshConfig {
	return &v1.RefreshConfig{
		Provider:           provider,
		CredentialKey:      key,
		Strategy:           v1.RefreshStrategy(req.Strategy),
		Material:           req.Material,
		SecretMaterialKeys: req.SecretMaterialKeys,
		ExpiresAt:          req.ExpiresAt,
	}
}
