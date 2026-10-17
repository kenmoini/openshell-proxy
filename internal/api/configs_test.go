package api

import (
	"context"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/gateway/gatewaytest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

const cfgPath = "/v1/workspaces/ws1/sandboxes/sb1/config"

func TestSettingValue_Union(t *testing.T) {
	var got []*v1.ConfigUpdate
	stub := &gatewaytest.ConfigStub{UpdateFn: func(_ context.Context, _ string, u *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
		got = append(got, u)
		return &v1.ConfigUpdateResult{Version: 1}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithConfig(stub)), nil)
	bodies := []string{
		`{"settingKey":"a","settingValue":{"type":"string","string":"hi"}}`,
		`{"settingKey":"b","settingValue":{"type":"bool","bool":true}}`,
		`{"settingKey":"c","settingValue":{"type":"int","int":42}}`,
		`{"settingKey":"d","settingValue":{"type":"bytes","bytes":"aGVsbG8="}}`,
	}
	for _, b := range bodies {
		rec := do(t, h, "PATCH", cfgPath, b, authHeader())
		require.Equal(t, 200, rec.Code, rec.Body.String())
	}
	require.Len(t, got, 4)
	require.Equal(t, v1.SettingValue{Type: "string", StringVal: "hi"}, *got[0].SettingValue)
	require.Equal(t, v1.SettingValue{Type: "bool", BoolVal: true}, *got[1].SettingValue)
	require.Equal(t, v1.SettingValue{Type: "int", IntVal: 42}, *got[2].SettingValue)
	require.Equal(t, v1.SettingValue{Type: "bytes", BytesVal: []byte("hello")}, *got[3].SettingValue)

	rec := do(t, h, "PATCH", cfgPath, `{"settingKey":"a","settingValue":{"string":"hi"}}`, authHeader())
	require.Equal(t, 422, rec.Code, rec.Body.String())
}

func TestConfig_GetSandbox(t *testing.T) {
	stub := &gatewaytest.ConfigStub{GetSandboxFn: func(_ context.Context, ws, sb string) (*v1.SandboxConfig, error) {
		require.Equal(t, "ws1", ws)
		require.Equal(t, "sb1", sb)
		return &v1.SandboxConfig{
			Policy:         &v1.SandboxPolicy{Version: 3},
			PolicyVersion:  3,
			PolicyHash:     "abc",
			ConfigRevision: 9,
			PolicySource:   "sandbox",
			Settings: map[string]v1.EffectiveSetting{
				"k1": {Value: v1.SettingValue{Type: "string", StringVal: "v"}, Scope: "sandbox"},
				"k2": {Value: v1.SettingValue{Type: "int", IntVal: 7}, Scope: "global"},
			},
		}, nil
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithConfig(stub)), nil)
	rec := do(t, h, "GET", cfgPath, "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "9", m["configRevision"])
	require.Equal(t, "abc", m["policyHash"])
	require.Equal(t, float64(3), m["policy"].(map[string]any)["version"])
	s := m["settings"].(map[string]any)
	require.Len(t, s, 2)
	k2 := s["k2"].(map[string]any)
	require.Equal(t, "global", k2["scope"])
	require.Equal(t, "int", k2["value"].(map[string]any)["type"])
	require.Equal(t, float64(7), k2["value"].(map[string]any)["int"])
}

func TestConfig_PatchConflict412(t *testing.T) {
	var got *v1.ConfigUpdate
	stub := &gatewaytest.ConfigStub{UpdateFn: func(_ context.Context, _ string, u *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
		got = u
		return nil, gatewaytest.StatusErr(codes.Aborted, "version conflict")
	}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithConfig(stub)), nil)
	hdr := authHeader()
	hdr["If-Match"] = `"3"`
	rec := do(t, h, "PATCH", cfgPath, `{"deleteSetting":true,"settingKey":"a"}`, hdr)
	require.Equal(t, 412, rec.Code, rec.Body.String())
	require.NotNil(t, got)
	require.Equal(t, uint64(3), got.ExpectedResourceVersion)
	require.Equal(t, "sb1", got.Name)
	require.True(t, got.DeleteSetting)
}

func TestGatewayConfig_PatchGlobal(t *testing.T) {
	var got *v1.ConfigUpdate
	var gotWS string
	stub := &gatewaytest.ConfigStub{
		GetGatewayFn: func(context.Context) (*v1.GatewayConfig, error) {
			return &v1.GatewayConfig{SettingsRevision: 5, Settings: map[string]v1.SettingValue{"x": {Type: "bool", BoolVal: true}}}, nil
		},
		UpdateFn: func(_ context.Context, ws string, u *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
			got, gotWS = u, ws
			return &v1.ConfigUpdateResult{SettingsRevision: 6}, nil
		}}
	h, _ := newTestAPI(t, gatewaytest.New(gatewaytest.WithConfig(stub)), nil)
	rec := do(t, h, "GET", "/v1/gateway/config", "", authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	m := decode(t, rec.Body.Bytes())
	require.Equal(t, "5", m["settingsRevision"])
	require.Equal(t, true, m["settings"].(map[string]any)["x"].(map[string]any)["bool"])

	rec = do(t, h, "PATCH", "/v1/gateway/config?workspace=wsX", `{"settingKey":"x","settingValue":{"type":"bool","bool":false}}`, authHeader())
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.True(t, got.Global)
	require.Equal(t, "wsX", gotWS)
	require.Equal(t, "6", decode(t, rec.Body.Bytes())["settingsRevision"])
}
