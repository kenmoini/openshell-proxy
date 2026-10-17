package gateway

import (
	"context"
	"errors"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

	"github.com/kenmoini/openshell-proxy/internal/auth"
	"github.com/kenmoini/openshell-proxy/internal/config"
)

// Gateway owns the shared SDK client.
type Gateway struct {
	client *v1.Client
	cfg    config.GatewayConfig
}

// New builds the shared client. The SDK dials lazily.
func New(cfg config.GatewayConfig) (*Gateway, error) {
	c, err := v1.NewClient(v1.Config{
		Address: cfg.DialAddress(),
		TLS:     tlsFor(cfg),
		Auth:    NewProvider(!cfg.AllowPlaintext),
	})
	if err != nil {
		return nil, err
	}
	return &Gateway{client: c, cfg: cfg}, nil
}

func tlsFor(cfg config.GatewayConfig) *v1.TLSConfig {
	if cfg.AllowPlaintext {
		return nil
	}
	return &v1.TLSConfig{CAFile: cfg.CAFile, Insecure: cfg.InsecureSkipVerify}
}

// Client returns the shared client; credentials come from each call's context.
func (g *Gateway) Client() v1.ClientInterface { return g.client }

// Close closes the shared client.
func (g *Gateway) Close() error { return g.client.Close() }

// Dedicated returns a short-lived client bound to the caller's token, plus a
// closer. It exists because the SDK's SSH tunnel cleanup revokes sessions with
// a background context, so the per-call provider on the shared client would
// have no token there; a per-caller client carries the token statically.
func (g *Gateway) Dedicated(ctx context.Context) (v1.ClientInterface, func(), error) {
	var provider v1.AuthProvider
	if token, ok := auth.TokenFrom(ctx); ok && token != "" {
		if g.cfg.AllowPlaintext {
			// v1.StaticToken requires TLS; plaintext dev mode needs a variant that does not.
			provider = plaintextToken(token)
		} else {
			provider = v1.StaticToken(token)
		}
	} else if g.cfg.AllowPlaintext {
		provider = v1.NoAuth()
	} else {
		return nil, nil, errors.New("gateway: no caller token available for dedicated client")
	}
	c, err := v1.NewClient(v1.Config{
		Address: g.cfg.DialAddress(),
		TLS:     tlsFor(g.cfg),
		Auth:    provider,
	})
	if err != nil {
		return nil, nil, err
	}
	return c, func() { _ = c.Close() }, nil
}

// plaintextToken is a static bearer token provider that permits plaintext transport.
type plaintextToken string

func (t plaintextToken) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(t)}, nil
}

func (plaintextToken) RequireTransportSecurity() bool { return false }
