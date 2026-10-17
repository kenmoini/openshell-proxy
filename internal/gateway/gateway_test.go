package gateway

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kenmoini/openshell-proxy/internal/auth"
	"github.com/kenmoini/openshell-proxy/internal/config"
)

func TestNew_PlaintextRequiresNoAuthTLS(t *testing.T) {
	g, err := New(config.GatewayConfig{Address: "127.0.0.1:1", AllowPlaintext: true})
	require.NoError(t, err)
	require.NotNil(t, g.Client())
	require.NoError(t, g.Close())

	_, err = New(config.GatewayConfig{Address: "http://127.0.0.1:1", AllowPlaintext: false})
	require.Error(t, err)
}

func TestDedicated_UsesCallerToken(t *testing.T) {
	g, err := New(config.GatewayConfig{Address: "127.0.0.1:1", AllowPlaintext: true})
	require.NoError(t, err)
	defer g.Close()

	c, closer, err := g.Dedicated(auth.WithToken(context.Background(), "tok"))
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NotPanics(t, closer)

	// plaintext mode without a token falls back to no auth
	c, closer, err = g.Dedicated(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)
	closer()

	// TLS mode without a token has no identity to use
	tg, err := New(config.GatewayConfig{Address: "127.0.0.1:1", InsecureSkipVerify: true})
	require.NoError(t, err)
	defer tg.Close()
	_, _, err = tg.Dedicated(context.Background())
	require.Error(t, err)

	c, closer, err = tg.Dedicated(auth.WithToken(context.Background(), "tok"))
	require.NoError(t, err)
	require.NotNil(t, c)
	closer()
}
