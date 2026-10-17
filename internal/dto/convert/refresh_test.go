package convert

import (
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
	"github.com/stretchr/testify/require"
)

func TestRefreshStatus_FromSDK(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d := RefreshStatusFromSDK(&v1.RefreshStatus{
		Provider: "p", ProviderID: "pid", CredentialKey: "K", Strategy: v1.RefreshStrategyExternal, Status: "ok",
		ExpiresAt: now, NextRefreshAt: now, LastRefreshAt: now, LastError: "e", RecoveryAction: 2,
		FailureCode: "fc", ProviderErrorSubtype: "st", LastErrorAt: now,
	})
	require.Equal(t, "reauthorize", d.RecoveryAction)
	require.Equal(t, "External", d.Strategy)
	require.Equal(t, now, *d.ExpiresAt)
	require.Equal(t, now, *d.LastErrorAt)
	require.Nil(t, RefreshStatusFromSDK(&v1.RefreshStatus{}).ExpiresAt)
}

func TestRefreshConfigToSDK(t *testing.T) {
	exp := time.Now()
	c := RefreshConfigToSDK("p", "K", dto.RefreshConfigRequest{Strategy: "Static", Material: map[string]string{"a": "b"}, SecretMaterialKeys: []string{"a"}, ExpiresAt: &exp})
	require.Equal(t, &v1.RefreshConfig{Provider: "p", CredentialKey: "K", Strategy: v1.RefreshStrategyStatic, Material: map[string]string{"a": "b"}, SecretMaterialKeys: []string{"a"}, ExpiresAt: &exp}, c)
}
