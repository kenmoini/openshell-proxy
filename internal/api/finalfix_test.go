package api

import (
	"context"
	"encoding/json"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

func TestUint64String_AcceptedOverHTTP(t *testing.T) {
	var got []*v1.ConfigUpdate
	stub := &gatewaytest.ConfigStub{UpdateFn: func(_ context.Context, _ string, u *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
		got = append(got, u)
		return &v1.ConfigUpdateResult{Version: 1}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithConfig(stub)), nil)
	rec := do(t, h, "PATCH", cfgPath, `{"settingKey":"a","settingValue":{"type":"bool","bool":true},"expectedResourceVersion":"7"}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, uint64(7), got[0].ExpectedResourceVersion)
	rec = do(t, h, "PATCH", cfgPath, `{"settingKey":"a","settingValue":{"type":"bool","bool":true},"expectedResourceVersion":8}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, uint64(8), got[1].ExpectedResourceVersion)
	rec = do(t, h, "PATCH", cfgPath, `{"settingKey":"a","settingValue":{"type":"bool","bool":true},"expectedResourceVersion":"x"}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
}

func TestProviders_GetPutRoundTrip(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	const base = "/v1/workspaces/ws1/providers"
	rec := do(t, h, "POST", base, `{"name":"gh","type":"github","spec":{"config":{"url":"u"}}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	rec = do(t, h, "GET", base+"/gh", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.NotEmpty(t, m["resourceVersion"], "fake should report a resource version")
	m["labels"] = map[string]any{"env": "prod"}
	b, err := json.Marshal(m)
	require.NoError(t, err)
	rec = do(t, h, "PUT", base+"/gh", string(b), authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "prod", decode(t, rec.Body.Bytes())["labels"].(map[string]any)["env"])
}

func TestTemplates_CreateWithResourceVersionString(t *testing.T) {
	c := gatewaytest.New()
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "POST", tplBase, `{"name":"tpl","resourceVersion":"3","spec":{"workload":{"image":"x"}}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
}

func TestProviders_ServerOwnedFieldsAndWorkspace(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	const base = "/v1/workspaces/ws1/providers"
	rec := do(t, h, "POST", base, `{"name":"gh","type":"github","workspace":"other"}`, authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "workspace in body must match path")
	rec = do(t, h, "PUT", base+"/gh", `{"name":"gh","type":"github","workspace":"other"}`, authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())

	rec = do(t, h, "POST", base, `{"name":"gh","type":"github","workspace":"ws1","id":"client-id","createdAt":"2001-01-01T00:00:00Z"}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.NotEqual(t, "client-id", m["id"])
	require.NotContains(t, rec.Body.String(), "2001-01-01")
}

func TestRequestID_OnHumaWrittenProblems(t *testing.T) {
	stub := gatewaytest.New()
	h, _ := newTestAPI(t, stub, nil)
	rec := do(t, h, "GET", "/v1/workspaces/ws1/providers/missing", "", authHeader())
	require.Equal(t, 404, rec.Code, rec.Body.String())
	id := rec.Header().Get("X-Request-ID")
	require.NotEmpty(t, id)
	require.Equal(t, id, decode(t, rec.Body.Bytes())["requestId"])

	rec = do(t, h, "POST", "/v1/workspaces/ws1/providers", `{"type":"github"}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
	id = rec.Header().Get("X-Request-ID")
	require.NotEmpty(t, id)
	require.Equal(t, id, decode(t, rec.Body.Bytes())["requestId"])
}

func TestValidation422DoesNotEchoCredentials(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(), nil)
	const base = "/v1/workspaces/ws1/providers"
	rec := do(t, h, "POST", base, `{"name":"gh","type":"github","spec":{"credentials":{"TOKEN":12345678}}}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
	require.NotContains(t, rec.Body.String(), "12345678")

	rec = do(t, h, "POST", base, `{"name":"gh","type":"github","spec":{"credentials":{"TOKEN":"sekret-value"},"bogusField":1}}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
	require.NotContains(t, rec.Body.String(), "sekret-value")
}

func TestPolicy_OptionalBodies(t *testing.T) {
	var rejected, approved bool
	stub := &gatewaytest.PolicyStub{
		RejectDraftChunkFn: func(_ context.Context, _, _, _, reason string) error {
			rejected = true
			require.Empty(t, reason)
			return nil
		},
		ApproveAllDraftChunksFn: func(_ context.Context, _, _ string, opts ...v1.ApproveAllOption) (*v1.ApproveAllResult, error) {
			approved = true
			return &v1.ApproveAllResult{}, nil
		},
	}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithPolicy(stub)), nil)
	rec := do(t, h, "POST", polBase+"/draft/chunks/c1/reject", "", authHeader())
	require.Equal(t, 204, rec.Code, rec.Body.String())
	require.True(t, rejected)
	rec = do(t, h, "POST", polBase+"/draft/approve-all", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.True(t, approved)
}
