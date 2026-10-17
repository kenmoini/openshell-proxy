package api

import (
	"context"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const svcBase = "/v1/workspaces/ws1/sandboxes/sb1/services"

func svcAPI(t *testing.T, stub *gatewaytest.ServiceStub) func(method, path, body string) (int, map[string]any) {
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithServices(stub)), nil)
	return func(method, path, body string) (int, map[string]any) {
		rec := do(t, h, method, path, body, authHeader())
		return rec.Code, decode(t, rec.Body.Bytes())
	}
}

func TestServices_Expose201(t *testing.T) {
	var name string
	var port uint32
	var domain bool
	var mode v1.ServiceAuthorizationMode
	stub := &gatewaytest.ServiceStub{ExposeFn: func(_ context.Context, ws, sb, n string, p uint32, dom bool, opts ...v1.ExposeServiceOptions) (*v1.ServiceEndpoint, error) {
		require.Equal(t, "ws1", ws)
		require.Equal(t, "sb1", sb)
		name, port, domain = n, p, dom
		if len(opts) > 0 {
			mode = opts[0].AuthorizationMode
		}
		return &v1.ServiceEndpoint{Name: n, TargetPort: p, Domain: dom, URL: "https://x.example", AuthorizationMode: mode}, nil
	}}
	do := svcAPI(t, stub)
	code, m := do("POST", svcBase, `{"name":"web","targetPort":8080,"domain":true,"authorizationMode":"BearerPassthrough"}`)
	require.Equal(t, 201, code, m)
	require.Equal(t, "web", name)
	require.EqualValues(t, 8080, port)
	require.True(t, domain)
	require.Equal(t, v1.ServiceAuthorizationModeBearerPassthrough, mode)
	require.Equal(t, "https://x.example", m["url"])
	require.Equal(t, "BearerPassthrough", m["authorizationMode"])

	code, _ = do("POST", svcBase, `{"name":"web","targetPort":0}`)
	require.Equal(t, 422, code)
}

func TestServices_ListAndGet(t *testing.T) {
	stub := &gatewaytest.ServiceStub{
		ListFn: func(string, string, ...v1.ListOptions) (*v1.Pager[*v1.ServiceEndpoint], error) {
			return gatewaytest.StaticPager([]*v1.ServiceEndpoint{{Name: "web", TargetPort: 80, URL: "u"}}, ""), nil
		},
		GetFn: func(_ context.Context, _, _, n string) (*v1.ServiceEndpoint, error) {
			return &v1.ServiceEndpoint{Name: n, TargetPort: 80, URL: "u", AuthorizationMode: v1.ServiceAuthorizationModeStrip}, nil
		},
	}
	do := svcAPI(t, stub)
	code, m := do("GET", svcBase, "")
	require.Equal(t, 200, code, m)
	require.Len(t, m["items"], 1)

	code, m = do("GET", svcBase+"/web", "")
	require.Equal(t, 200, code, m)
	require.Equal(t, "web", m["name"])
	require.Equal(t, "Strip", m["authorizationMode"])
}

func TestServices_ListEmptyIsArray(t *testing.T) {
	stub := &gatewaytest.ServiceStub{ListFn: func(string, string, ...v1.ListOptions) (*v1.Pager[*v1.ServiceEndpoint], error) {
		return gatewaytest.StaticPager[*v1.ServiceEndpoint](nil, ""), nil
	}}
	do := svcAPI(t, stub)
	code, m := do("GET", svcBase, "")
	require.Equal(t, 200, code, m)
	require.NotNil(t, m["items"])
	require.Len(t, m["items"], 0)
}

func TestServices_Delete(t *testing.T) {
	stub := &gatewaytest.ServiceStub{DeleteFn: func(_ context.Context, _, _, n string, _ ...v1.DeleteOptions) (*v1.DeletionResult, error) {
		require.Equal(t, "web", n)
		return &v1.DeletionResult{Outcome: v1.DeletionCompleted}, nil
	}}
	do := svcAPI(t, stub)
	code, m := do("DELETE", svcBase+"/web", "")
	require.Equal(t, 200, code, m)
	require.Equal(t, "Completed", m["outcome"])
}
