package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kenmoini/openshell-proxy/internal/config"
	"github.com/kenmoini/openshell-proxy/internal/gateway"
)

func TestUnavailableGateway_Problems(t *testing.T) {
	start := time.Now()

	g, err := gateway.New(config.GatewayConfig{Address: "127.0.0.1:1", AllowPlaintext: true})
	require.NoError(t, err)
	defer g.Close()

	h, _ := newTestAPI(t, g.Client(), func(c *config.Config) {
		c.Gateway = config.GatewayConfig{Address: "127.0.0.1:1", AllowPlaintext: true}
		c.Auth.Mode = config.ModeNone
		c.Timeouts.Request = 2 * time.Second
	})

	rec := do(t, h, "GET", "/healthz", "", nil)
	require.Equal(t, 200, rec.Code, rec.Body.String())

	rec = do(t, h, "GET", "/readyz", "", nil)
	require.Equal(t, 503, rec.Code, rec.Body.String())
	require.True(t, strings.HasPrefix(rec.Header().Get("Content-Type"), "application/problem+json"), rec.Header().Get("Content-Type"))

	rec = do(t, h, "GET", "/v1/whoami", "", nil)
	require.Equal(t, 503, rec.Code, rec.Body.String())
	require.True(t, strings.HasPrefix(rec.Header().Get("Content-Type"), "application/problem+json"), rec.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "Unavailable", body["grpcCode"])

	require.Less(t, time.Since(start), 5*time.Second)
}
