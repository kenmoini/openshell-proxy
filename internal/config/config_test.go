package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noEnv(string) string { return "" }

func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func writeYAML(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cfg.yaml")
	require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	return p
}

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	cfg, err := Load("", noEnv)
	require.Error(t, err) // empty gateway address
	assert.Nil(t, cfg)

	d := Default()
	assert.Equal(t, ":8080", d.Server.Addr)
	assert.Equal(t, 10*time.Second, d.Server.ReadHeaderTimeout)
	assert.Equal(t, 120*time.Second, d.Server.IdleTimeout)
	assert.Equal(t, ModeBearerPassthrough, d.Auth.Mode)
	assert.True(t, d.Auth.LocalSanityCheck)
	assert.False(t, d.Auth.AllowQueryToken)
	assert.Equal(t, 30*time.Second, d.Timeouts.Request)
	assert.Equal(t, 10*time.Minute, d.Timeouts.Exec)
	assert.Equal(t, 10*time.Minute, d.Timeouts.WaitMax)
	assert.Equal(t, time.Hour, d.Timeouts.Stream)
	assert.Equal(t, 15*time.Second, d.Timeouts.SSEHeartbeat)
	assert.Equal(t, int64(1<<20), d.Limits.MaxBody)
	assert.Equal(t, int64(4<<20), d.Limits.MaxExecBody)
	assert.Equal(t, int64(1<<20), d.Limits.WSMaxFrame)
	assert.Equal(t, "info", d.Log.Level)
	assert.Equal(t, "json", d.Log.Format)

	err = d.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gateway.address")
	d.Gateway.Address = "gw:443"
	assert.NoError(t, d.Validate())
}

func TestLoad_YAMLOverridesDefaults(t *testing.T) {
	p := writeYAML(t, `
gateway:
  address: gw:443
timeouts:
  request: 45s
limits:
  maxBody: 2MiB
  maxExecBody: 8388608
`)
	cfg, err := Load(p, noEnv)
	require.NoError(t, err)
	assert.Equal(t, "gw:443", cfg.Gateway.Address)
	assert.Equal(t, 45*time.Second, cfg.Timeouts.Request)
	assert.Equal(t, int64(2097152), cfg.Limits.MaxBody)
	assert.Equal(t, int64(8388608), cfg.Limits.MaxExecBody)
	assert.Equal(t, int64(1<<20), cfg.Limits.WSMaxFrame)
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	p := writeYAML(t, `
gateway:
  address: yaml:443
auth:
  mode: bearer-passthrough
timeouts:
  request: 45s
`)
	cfg, err := Load(p, envMap(map[string]string{
		"OSPROXY_GATEWAY_ADDRESS":        "env:443",
		"OSPROXY_AUTH_MODE":              "none",
		"OSPROXY_TIMEOUTS_REQUEST":       "5s",
		"OSPROXY_LIMITS_WS_MAX_FRAME":    "512KiB",
		"OSPROXY_CORS_ALLOWED_ORIGINS":   "https://a.example, https://b.example",
		"OSPROXY_AUTH_ALLOW_QUERY_TOKEN": "true",
	}))
	require.NoError(t, err)
	assert.Equal(t, "env:443", cfg.Gateway.Address)
	assert.Equal(t, ModeNone, cfg.Auth.Mode)
	assert.Equal(t, 5*time.Second, cfg.Timeouts.Request)
	assert.Equal(t, int64(512<<10), cfg.Limits.WSMaxFrame)
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, cfg.CORS.AllowedOrigins)
	assert.True(t, cfg.Auth.AllowQueryToken)
}

func TestLoad_Errors(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"), noEnv)
	require.Error(t, err)

	_, err = Load(writeYAML(t, "gateway:\n  address: x:1\n  bogus: 1\n"), noEnv)
	require.Error(t, err)

	base := writeYAML(t, "gateway:\n  address: x:1\n")
	for env, val := range map[string]string{
		"OSPROXY_TIMEOUTS_EXEC":              "nope",
		"OSPROXY_GATEWAY_ALLOW_PLAINTEXT":    "maybe",
		"OSPROXY_LIMITS_MAX_BODY":            "12XB",
		"OSPROXY_SERVER_READ_HEADER_TIMEOUT": "1",
	} {
		_, err = Load(base, envMap(map[string]string{env: val}))
		require.Error(t, err, env)
		assert.Contains(t, err.Error(), env)
	}
}

func TestValidate_PlaintextRules(t *testing.T) {
	c := Default()
	c.Gateway.Address = "http://x"
	require.Error(t, c.Validate())

	c = Default()
	c.Gateway.Address = "127.0.0.1:1"
	c.Gateway.AllowPlaintext = true
	c.Gateway.CAFile = "/ca.pem"
	c.Auth.Mode = ModeNone
	require.Error(t, c.Validate())

	c = Default()
	c.Gateway.Address = "127.0.0.1:1"
	c.Gateway.AllowPlaintext = true
	err := c.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth.mode must be none")

	c.Auth.Mode = ModeNone
	assert.NoError(t, c.Validate())
}

func TestValidate_ModeAndTLSFiles(t *testing.T) {
	c := Default()
	c.Gateway.Address = "gw:443"
	c.Auth.Mode = "weird"
	require.Error(t, c.Validate())

	c = Default()
	c.Gateway.Address = "gw:443"
	c.Server.TLS.CertFile = "c.pem"
	require.Error(t, c.Validate())
	c.Server.TLS.KeyFile = "k.pem"
	assert.NoError(t, c.Validate())
	c.Server.TLS.CertFile = ""
	require.Error(t, c.Validate())

	c = Default()
	c.Gateway.Address = "gw:443"
	c.Log.Level = "loud"
	require.Error(t, c.Validate())
	c.Log.Level = "debug"
	c.Log.Format = "xml"
	require.Error(t, c.Validate())
}

func TestDialAddress(t *testing.T) {
	assert.Equal(t, "gw:443", GatewayConfig{Address: "gw:443"}.DialAddress())
	assert.Equal(t, "gw:443", GatewayConfig{Address: "https://gw:443"}.DialAddress())
	assert.Equal(t, "http://127.0.0.1:17670", GatewayConfig{Address: "127.0.0.1:17670", AllowPlaintext: true}.DialAddress())
	assert.Equal(t, "http://x:1", GatewayConfig{Address: "http://x:1", AllowPlaintext: true}.DialAddress())
}

func TestExampleConfigParses(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "configs", "example.yaml"))
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(b), "gateway"))
	cfg, err := Load(filepath.Join("..", "..", "configs", "example.yaml"), noEnv)
	require.NoError(t, err)
	want := Default()
	want.Gateway.Address = cfg.Gateway.Address
	want.CORS.AllowedOrigins = []string{} // example documents the empty list explicitly
	assert.Equal(t, &want, cfg)
}
