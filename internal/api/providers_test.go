package api

import (
	"testing"

	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

func TestProviders_CRUDAndEnsure(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	const base = "/v1/workspaces/ws1/providers"
	rec := do(t, h, "POST", base, `{"name":"gh","type":"github","spec":{"credentials":{"TOKEN":"x"}}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	require.Equal(t, "gh", decode(t, rec.Body.Bytes())["name"])

	rec = do(t, h, "GET", base+"/gh", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "github", decode(t, rec.Body.Bytes())["type"])

	rec = do(t, h, "GET", base, "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 1)

	rec = do(t, h, "PUT", base+"/gh", `{"name":"gh","type":"github","spec":{"config":{"url":"u"}}}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())

	rec = do(t, h, "PUT", base+"/new?ensure=true", `{"name":"new","type":"github"}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "new", decode(t, rec.Body.Bytes())["name"])

	rec = do(t, h, "DELETE", base+"/gh", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
}

func TestProviders_UpdateNameMismatch400(t *testing.T) {
	c := gatewaytest.New()
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "PUT", "/v1/workspaces/ws1/providers/gh", `{"name":"other","type":"github"}`, authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())
}
