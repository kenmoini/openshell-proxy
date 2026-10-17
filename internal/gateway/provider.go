// Package gateway builds the shared OpenShell SDK client and injects each
// caller's bearer token per RPC.
package gateway

import (
	"context"

	"github.com/kenmoini/openshell-proxy/internal/auth"
	"github.com/kenmoini/openshell-proxy/internal/middleware"
)

// Provider is a v1.AuthProvider that reads the caller's token and request ID
// from the call context, so one shared client serves every caller.
type Provider struct{ requireTLS bool }

// NewProvider returns a Provider. requireTLS controls RequireTransportSecurity.
func NewProvider(requireTLS bool) *Provider { return &Provider{requireTLS: requireTLS} }

// GetRequestMetadata never fails. With no token it returns no authorization
// entry, which the gateway accepts for unauthenticated RPCs such as Health.
func (p *Provider) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	if tok, ok := auth.TokenFrom(ctx); ok && tok != "" {
		md["authorization"] = "Bearer " + tok
	}
	if id := middleware.RequestIDFrom(ctx); id != "" {
		md["x-request-id"] = id
	}
	return md, nil
}

// RequireTransportSecurity reports whether credentials may only travel over TLS.
func (p *Provider) RequireTransportSecurity() bool { return p.requireTLS }
