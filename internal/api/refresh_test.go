package api

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

func TestRefresh_ConfigureRotateStatus(t *testing.T) {
	var cfg *v1.RefreshConfig
	var rotP, rotK, stP, stK string
	stub := &gatewaytest.RefreshStub{
		ConfigureFn: func(_ context.Context, ws string, c *v1.RefreshConfig) (*v1.RefreshStatus, error) {
			require.Equal(t, "ws1", ws)
			cfg = c
			return &v1.RefreshStatus{Provider: c.Provider, CredentialKey: c.CredentialKey, Strategy: c.Strategy, Status: "ok"}, nil
		},
		RotateFn: func(_ context.Context, _, p, k string) (*v1.RefreshStatus, error) {
			rotP, rotK = p, k
			return &v1.RefreshStatus{Provider: p, CredentialKey: k, RecoveryAction: 1}, nil
		},
		GetStatusFn: func(_ context.Context, _, p, k string) ([]*v1.RefreshStatus, error) {
			stP, stK = p, k
			return []*v1.RefreshStatus{{Provider: p, CredentialKey: "A"}, {Provider: p, CredentialKey: "B"}}, nil
		},
		DeleteFn: func(context.Context, string, string, string, ...v1.DeleteOptions) (*v1.DeletionResult, error) {
			return &v1.DeletionResult{Outcome: v1.DeletionCompleted}, nil
		},
	}
	// Capture logs to prove material never reaches them.
	var logs bytes.Buffer
	deps := testDeps(gatewaytest.New(gatewaytest.WithRefresh(stub)), nil)
	deps.Log = slog.New(slog.NewJSONHandler(&logs, nil))
	h, _ := New(deps)

	const base = "/v1/workspaces/ws1/providers/gh/refresh"
	rec := do(t, h, "PUT", base+"/TOKEN", `{"strategy":"OAuth2RefreshToken","material":{"refresh_token":"mat-secret-123"},"secretMaterialKeys":["refresh_token"]}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "gh", cfg.Provider)
	require.Equal(t, "TOKEN", cfg.CredentialKey)
	require.Equal(t, v1.RefreshStrategyOAuth2RefreshToken, cfg.Strategy)
	require.Equal(t, "mat-secret-123", cfg.Material["refresh_token"])
	require.Equal(t, []string{"refresh_token"}, cfg.SecretMaterialKeys)
	require.Equal(t, "ok", decode(t, rec.Body.Bytes())["status"])

	rec = do(t, h, "POST", base+"/TOKEN/rotate", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "gh", rotP)
	require.Equal(t, "TOKEN", rotK)
	require.Equal(t, "retry", decode(t, rec.Body.Bytes())["recoveryAction"])

	rec = do(t, h, "GET", base+"?credentialKey=TOKEN", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "gh", stP)
	require.Equal(t, "TOKEN", stK)
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 2)

	rec = do(t, h, "DELETE", base+"/TOKEN", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())

	require.NotContains(t, logs.String(), "mat-secret-123")
}

func TestRefresh_EmptyStatusItemsNotNull(t *testing.T) {
	stub := &gatewaytest.RefreshStub{GetStatusFn: func(context.Context, string, string, string) ([]*v1.RefreshStatus, error) {
		return nil, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithRefresh(stub)), nil)
	rec := do(t, h, "GET", "/v1/workspaces/ws1/providers/gh/refresh", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, []any{}, decode(t, rec.Body.Bytes())["items"])
}
