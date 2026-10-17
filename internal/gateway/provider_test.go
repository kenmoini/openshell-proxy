package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kenmoini/openshell-proxy/internal/auth"
	"github.com/kenmoini/openshell-proxy/internal/middleware"
)

func TestProvider_MetadataWithToken(t *testing.T) {
	p := NewProvider(true)

	md, err := p.GetRequestMetadata(context.Background())
	require.NoError(t, err)
	require.Empty(t, md)

	ctx := auth.WithToken(context.Background(), "tok123")
	ctx = withRequestID(t, ctx, "rid-1")
	md, err = p.GetRequestMetadata(ctx)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"authorization": "Bearer tok123", "x-request-id": "rid-1"}, md)

	md, err = p.GetRequestMetadata(auth.WithToken(context.Background(), "only"))
	require.NoError(t, err)
	require.Equal(t, map[string]string{"authorization": "Bearer only"}, md)
}

func TestProvider_RequireTransportSecurity(t *testing.T) {
	require.True(t, NewProvider(true).RequireTransportSecurity())
	require.False(t, NewProvider(false).RequireTransportSecurity())
}

// withRequestID derives a context carrying id by running the real middleware.
func withRequestID(t *testing.T, ctx context.Context, id string) context.Context {
	t.Helper()
	var out context.Context
	h := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { out = r.Context() }))
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set("X-Request-ID", id)
	h.ServeHTTP(httptest.NewRecorder(), req)
	require.Equal(t, id, middleware.RequestIDFrom(out))
	return out
}
