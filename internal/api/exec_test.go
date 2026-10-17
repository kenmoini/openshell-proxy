package api

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
)

const execPath = "/v1/workspaces/ws1/sandboxes/sb1/exec"

func TestExec_Run(t *testing.T) {
	var got v1.ExecOptions
	var gotWS, gotSB string
	stub := &gatewaytest.ExecStub{RunFn: func(_ context.Context, ws, sb string, cmd []string, opts ...v1.ExecOptions) (*v1.ExecResult, error) {
		gotWS, gotSB = ws, sb
		if len(opts) > 0 {
			got = opts[0]
		}
		return &v1.ExecResult{ExitCode: 3, Stdout: []byte(strings.Join(cmd, " "))}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithExec(stub)), nil)
	rec := do(t, h, "POST", execPath, `{"command":["echo","hi"],"env":{"A":"b"},"workDir":"/tmp","noLoginShell":true}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.EqualValues(t, 3, m["exitCode"])
	require.Equal(t, "echo hi", m["stdout"])
	require.Equal(t, "", m["stderr"])
	require.Equal(t, "utf8", m["encoding"])
	require.Equal(t, "ws1", gotWS)
	require.Equal(t, "sb1", gotSB)
	require.Equal(t, map[string]string{"A": "b"}, got.Env)
	require.Equal(t, "/tmp", got.WorkDir)
	require.True(t, got.NoLoginShell)
}

func TestExec_Base64WhenBinary(t *testing.T) {
	stub := &gatewaytest.ExecStub{RunFn: func(context.Context, string, string, []string, ...v1.ExecOptions) (*v1.ExecResult, error) {
		return &v1.ExecResult{Stdout: []byte{0xff, 0xfe}}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithExec(stub)), nil)
	rec := do(t, h, "POST", execPath, `{"command":["cat","x"]}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "base64", m["encoding"])
	require.Equal(t, "//4=", m["stdout"])
}

func TestExec_Base64WhenRequested(t *testing.T) {
	stub := &gatewaytest.ExecStub{RunFn: func(context.Context, string, string, []string, ...v1.ExecOptions) (*v1.ExecResult, error) {
		return &v1.ExecResult{Stdout: []byte("hi")}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithExec(stub)), nil)
	rec := do(t, h, "POST", execPath, `{"command":["x"],"encoding":"base64"}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "base64", m["encoding"])
	require.Equal(t, "aGk=", m["stdout"])
}

func TestExec_EmptyCommand422(t *testing.T) {
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithExec(&gatewaytest.ExecStub{})), nil)
	rec := do(t, h, "POST", execPath, `{"command":[]}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
}
