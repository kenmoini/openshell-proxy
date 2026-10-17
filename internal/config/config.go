// Package config loads, overrides and validates the proxy configuration.
package config

import "time"

// Mode selects how the proxy authenticates callers.
type Mode string

const (
	// ModeBearerPassthrough forwards the caller's bearer token to the gateway.
	ModeBearerPassthrough Mode = "bearer-passthrough"
	// ModeNone performs no authentication.
	ModeNone Mode = "none"
)

// Config is the complete proxy configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Gateway  GatewayConfig  `yaml:"gateway"`
	Auth     AuthConfig     `yaml:"auth"`
	Timeouts TimeoutsConfig `yaml:"timeouts"`
	Limits   LimitsConfig   `yaml:"limits"`
	CORS     CORSConfig     `yaml:"cors"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig configures the HTTP listener.
type ServerConfig struct {
	Addr              string        `yaml:"addr"`
	TLS               TLSFiles      `yaml:"tls"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout"`
	IdleTimeout       time.Duration `yaml:"idleTimeout"`
}

// TLSFiles are optional server certificate and key paths.
type TLSFiles struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

// GatewayConfig describes the upstream OpenShell gRPC gateway.
type GatewayConfig struct {
	Address            string `yaml:"address"`
	CAFile             string `yaml:"caFile"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	AllowPlaintext     bool   `yaml:"allowPlaintext"`
}

// AuthConfig configures caller authentication.
type AuthConfig struct {
	Mode             Mode `yaml:"mode"`
	LocalSanityCheck bool `yaml:"localSanityCheck"`
	AllowQueryToken  bool `yaml:"allowQueryToken"`
}

// TimeoutsConfig holds per-operation deadlines.
type TimeoutsConfig struct {
	Request      time.Duration `yaml:"request"`
	Exec         time.Duration `yaml:"exec"`
	WaitMax      time.Duration `yaml:"waitMax"`
	Stream       time.Duration `yaml:"stream"`
	SSEHeartbeat time.Duration `yaml:"sseHeartbeat"`
}

// LimitsConfig holds size limits in bytes. YAML accepts "1MiB" style values
// or plain integers.
type LimitsConfig struct {
	MaxBody     int64 `yaml:"maxBody"`
	MaxExecBody int64 `yaml:"maxExecBody"`
	WSMaxFrame  int64 `yaml:"wsMaxFrame"`
}

// CORSConfig configures cross-origin access.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowedOrigins"`
	AllowCredentials bool     `yaml:"allowCredentials"`
}

// LogConfig configures logging.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Default returns the configuration used when nothing is overridden.
func Default() Config {
	return Config{
		Server: ServerConfig{
			Addr:              ":8080",
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		Auth: AuthConfig{Mode: ModeBearerPassthrough, LocalSanityCheck: true},
		Timeouts: TimeoutsConfig{
			Request:      30 * time.Second,
			Exec:         10 * time.Minute,
			WaitMax:      10 * time.Minute,
			Stream:       time.Hour,
			SSEHeartbeat: 15 * time.Second,
		},
		Limits: LimitsConfig{MaxBody: 1 << 20, MaxExecBody: 4 << 20, WSMaxFrame: 1 << 20},
		Log:    LogConfig{Level: "info", Format: "json"},
	}
}
