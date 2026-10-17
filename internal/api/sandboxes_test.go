package api

import (
	"context"
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const sbBase = "/v1/workspaces/ws1/sandboxes"

func sbClient(opts ...gatewaytest.Option) *gatewaytest.Client {
	c := gatewaytest.New(opts...)
	c.AddWorkspace(&v1.Workspace{Name: "ws1"})
	return c
}

func TestSandboxes_CreateGetList(t *testing.T) {
	h, _ := newTestAPI(t, sbClient(), nil)
	rec := do(t, h, "POST", sbBase, `{"name":"sb1","labels":{"a":"b"},"spec":{"template":{"image":"python:3.12"}}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	require.Equal(t, "sb1", decode(t, rec.Body.Bytes())["name"])

	rec = do(t, h, "GET", sbBase+"/sb1", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	spec := decode(t, rec.Body.Bytes())["spec"].(map[string]any)
	require.Equal(t, "python:3.12", spec["template"].(map[string]any)["image"])

	rec = do(t, h, "GET", sbBase, "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Len(t, decode(t, rec.Body.Bytes())["items"], 1)
}

func TestSandboxes_CreateFromTemplate(t *testing.T) {
	c := sbClient()
	c.AddSandboxTemplate("ws1", &v1.SandboxWorkloadTemplate{Name: "tpl"})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "POST", sbBase, `{"name":"sb2","template":"tpl","spec":{}}`, authHeader())
	require.Equal(t, 201, rec.Code, rec.Body.String())
	from := decode(t, rec.Body.Bytes())["createdFromWorkloadTemplate"].(map[string]any)
	require.Equal(t, "tpl", from["name"])
}

func TestSandboxes_StopStartWait(t *testing.T) {
	c := sbClient()
	c.AddSandbox("ws1", &v1.Sandbox{Name: "sb1", Status: v1.SandboxStatus{Phase: v1.SandboxReady}})
	h, _ := newTestAPI(t, c, nil)

	rec := do(t, h, "POST", sbBase+"/sb1/stop?wait=true", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "Stopped", decode(t, rec.Body.Bytes())["status"].(map[string]any)["phase"])

	rec = do(t, h, "POST", sbBase+"/sb1/start?wait=true&timeout=5s", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, "Ready", decode(t, rec.Body.Bytes())["status"].(map[string]any)["phase"])

	rec = do(t, h, "POST", sbBase+"/sb1/wait", `{"for":"ready"}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())

	rec = do(t, h, "POST", sbBase+"/sb1/stop", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())

	rec = do(t, h, "POST", sbBase+"/sb1/wait", `{"for":"bogus"}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
	rec = do(t, h, "POST", sbBase+"/sb1/wait", `{"for":"ready","timeout":"abc"}`, authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())
}

func TestSandboxes_WaitTerminalIs409(t *testing.T) {
	stub := &gatewaytest.SandboxStub{WaitReadyFn: func(context.Context, string, string, ...v1.WaitOptions) (*v1.Sandbox, error) {
		return nil, &v1.StatusError{Code: v1.ErrorInternal, Message: "sandbox \"x\" is in error state"}
	}}
	h, _ := newTestAPI(t, sbClient(gatewaytest.WithSandboxOverrides(stub)), nil)
	rec := do(t, h, "POST", sbBase+"/x/wait", `{"for":"ready"}`, authHeader())
	require.Equal(t, 409, rec.Code, rec.Body.String())
	require.Equal(t, "SANDBOX_TERMINAL_PHASE", decode(t, rec.Body.Bytes())["reason"])
}

func TestWaitDeadline_Capped(t *testing.T) {
	max := time.Minute
	d, err := waitDeadline("", max)
	require.NoError(t, err)
	require.Equal(t, max, d)
	d, err = waitDeadline("5s", max)
	require.NoError(t, err)
	require.Equal(t, 5*time.Second, d)
	d, err = waitDeadline("1h", max)
	require.NoError(t, err)
	require.Equal(t, max, d)
	for _, bad := range []string{"0s", "-1s", "soon"} {
		_, err = waitDeadline(bad, max)
		require.Error(t, err, bad)
	}
}

func TestSandboxes_WaitTimeoutExpired(t *testing.T) {
	c := sbClient()
	c.AddSandbox("ws1", &v1.Sandbox{Name: "sb1", Status: v1.SandboxStatus{Phase: v1.SandboxProvisioning}})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "POST", sbBase+"/sb1/wait", `{"for":"ready","timeout":"1ns"}`, authHeader())
	require.Equal(t, 504, rec.Code, rec.Body.String())
	require.Equal(t, true, decode(t, rec.Body.Bytes())["outcomeUnknown"])
}

func TestSandboxes_ProvidersAttachDetach(t *testing.T) {
	c := sbClient()
	c.AddSandbox("ws1", &v1.Sandbox{Name: "sb1"})
	c.AddProvider("ws1", &v1.Provider{Name: "p", Type: "github"})
	h, _ := newTestAPI(t, c, nil)

	rec := do(t, h, "PUT", sbBase+"/sb1/providers/p", "", map[string]string{"Authorization": authHeader()["Authorization"], "If-Match": `"0"`})
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, true, decode(t, rec.Body.Bytes())["attached"])

	rec = do(t, h, "GET", sbBase+"/sb1/providers", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	items := decode(t, rec.Body.Bytes())["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, "p", items[0].(map[string]any)["name"])

	rec = do(t, h, "DELETE", sbBase+"/sb1/providers/p", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, true, decode(t, rec.Body.Bytes())["detached"])

	rec = do(t, h, "PUT", sbBase+"/sb1/providers/p", "", map[string]string{"Authorization": authHeader()["Authorization"], "If-Match": "W/x"})
	require.Equal(t, 400, rec.Code, rec.Body.String())
}

func TestSandboxes_Logs(t *testing.T) {
	// The SDK fake does not implement GetLogs, so stub it.
	calls := 0
	var gotOpts []v1.LogOption
	stub := &gatewaytest.SandboxStub{GetLogsFn: func(_ context.Context, ws, sb string, opts ...v1.LogOption) (*v1.LogResult, error) {
		calls++
		gotOpts = opts
		require.Equal(t, "ws1", ws)
		require.Equal(t, "sb1", sb)
		return &v1.LogResult{}, nil
	}}
	h, _ := newTestAPI(t, sbClient(gatewaytest.WithSandboxOverrides(stub)), nil)
	rec := do(t, h, "GET", sbBase+"/sb1/logs", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	body := decode(t, rec.Body.Bytes())
	require.Equal(t, []any{}, body["lines"])
	require.EqualValues(t, 0, body["bufferTotal"])

	rec = do(t, h, "GET", sbBase+"/sb1/logs?lines=50&since=2026-01-01T00:00:00Z&sources=gateway,%20sandbox,&minLevel=WARN", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Equal(t, 2, calls)
	cfg := types.ApplyLogOptions(gotOpts)
	require.EqualValues(t, 50, cfg.Lines())
	require.True(t, cfg.Since().Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	require.Equal(t, []string{"gateway", "sandbox"}, cfg.Sources())
	require.Equal(t, "WARN", cfg.MinLevel())

	rec = do(t, h, "GET", sbBase+"/sb1/logs?since=yesterday", "", authHeader())
	require.Equal(t, 400, rec.Code, rec.Body.String())
	require.Equal(t, 2, calls)
}

func TestSandboxes_DeleteAccepted202(t *testing.T) {
	stub := &gatewaytest.SandboxStub{DeleteFn: func(context.Context, string, string, ...v1.DeleteOptions) (*v1.DeletionResult, error) {
		return &v1.DeletionResult{Outcome: v1.DeletionAccepted, SandboxID: "sb-1"}, nil
	}}
	h, _ := newTestAPI(t, sbClient(gatewaytest.WithSandboxOverrides(stub)), nil)
	rec := do(t, h, "DELETE", sbBase+"/sb1", "", authHeader())
	require.Equal(t, 202, rec.Code, rec.Body.String())
	require.Equal(t, "Accepted", decode(t, rec.Body.Bytes())["outcome"])
}

func TestSandboxes_NamedStartIsReachable(t *testing.T) {
	c := sbClient()
	c.AddSandbox("ws1", &v1.Sandbox{Name: "start"})
	h, _ := newTestAPI(t, c, nil)
	rec := do(t, h, "GET", sbBase+"/start", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
}
