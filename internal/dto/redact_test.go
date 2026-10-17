package dto

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_LogValueRedacts(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	p := Provider{ObjectMeta: ObjectMeta{Name: "gh"}, Type: "github",
		Spec: ProviderSpec{Credentials: map[string]string{"TOKEN": "s3cr3t-value", "A": "other-secret"}}}
	log.Info("provider", "p", p)
	out := buf.String()
	require.NotContains(t, out, "s3cr3t-value")
	require.NotContains(t, out, "other-secret")
	require.Contains(t, out, `"credentialKeys":["A","TOKEN"]`)
	require.Contains(t, out, `"name":"gh"`)
}

func TestRefreshConfig_LogValueRedacts(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	log.Info("cfg", "c", RefreshConfigRequest{Strategy: "Static", Material: map[string]string{"refresh_token": "mat-secret"}})
	out := buf.String()
	require.NotContains(t, out, "mat-secret")
	require.Contains(t, out, `"materialKeys":["refresh_token"]`)
	require.Contains(t, out, `"strategy":"Static"`)
}
