package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load builds a Config from defaults, then the YAML file at path (skipped when
// path is empty), then OSPROXY_* environment variables, and validates it.
func Load(path string, getenv func(string) string) (*Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	}
	if err := applyEnv(&cfg, getenv); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UnmarshalYAML decodes limits, accepting "1MiB" style strings or integers.
// Keys that are absent keep their current value. Unknown keys are rejected.
func (l *LimitsConfig) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: limits must be a mapping", n.Line)
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		var dst *int64
		switch k.Value {
		case "maxBody":
			dst = &l.MaxBody
		case "maxExecBody":
			dst = &l.MaxExecBody
		case "wsMaxFrame":
			dst = &l.WSMaxFrame
		default:
			return fmt.Errorf("line %d: field %s not found in type config.LimitsConfig", k.Line, k.Value)
		}
		size, err := ParseByteSize(v.Value)
		if err != nil {
			return fmt.Errorf("line %d: limits.%s: %w", v.Line, k.Value, err)
		}
		*dst = size
	}
	return nil
}

// ParseByteSize parses a plain integer or an integer with a KiB, MiB or GiB suffix.
func ParseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	mult := int64(1)
	for suffix, m := range map[string]int64{"KiB": 1 << 10, "MiB": 1 << 20, "GiB": 1 << 30} {
		if strings.HasSuffix(s, suffix) {
			s, mult = strings.TrimSpace(strings.TrimSuffix(s, suffix)), m
			break
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid byte size %q (use an integer or KiB/MiB/GiB suffix)", s)
	}
	return n * mult, nil
}

type envSetter struct {
	name string
	set  func(c *Config, v string) error
}

func str(dst func(c *Config) *string) func(*Config, string) error {
	return func(c *Config, v string) error { *dst(c) = v; return nil }
}

func dur(dst func(c *Config) *time.Duration) func(*Config, string) error {
	return func(c *Config, v string) error {
		d, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*dst(c) = d
		return nil
	}
}

func boolean(dst func(c *Config) *bool) func(*Config, string) error {
	return func(c *Config, v string) error {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		*dst(c) = b
		return nil
	}
}

func size(dst func(c *Config) *int64) func(*Config, string) error {
	return func(c *Config, v string) error {
		n, err := ParseByteSize(v)
		if err != nil {
			return err
		}
		*dst(c) = n
		return nil
	}
}

var envTable = []envSetter{
	{"OSPROXY_SERVER_ADDR", str(func(c *Config) *string { return &c.Server.Addr })},
	{"OSPROXY_SERVER_TLS_CERT_FILE", str(func(c *Config) *string { return &c.Server.TLS.CertFile })},
	{"OSPROXY_SERVER_TLS_KEY_FILE", str(func(c *Config) *string { return &c.Server.TLS.KeyFile })},
	{"OSPROXY_SERVER_READ_HEADER_TIMEOUT", dur(func(c *Config) *time.Duration { return &c.Server.ReadHeaderTimeout })},
	{"OSPROXY_SERVER_IDLE_TIMEOUT", dur(func(c *Config) *time.Duration { return &c.Server.IdleTimeout })},
	{"OSPROXY_GATEWAY_ADDRESS", str(func(c *Config) *string { return &c.Gateway.Address })},
	{"OSPROXY_GATEWAY_CA_FILE", str(func(c *Config) *string { return &c.Gateway.CAFile })},
	{"OSPROXY_GATEWAY_INSECURE_SKIP_VERIFY", boolean(func(c *Config) *bool { return &c.Gateway.InsecureSkipVerify })},
	{"OSPROXY_GATEWAY_ALLOW_PLAINTEXT", boolean(func(c *Config) *bool { return &c.Gateway.AllowPlaintext })},
	{"OSPROXY_AUTH_MODE", func(c *Config, v string) error { c.Auth.Mode = Mode(v); return nil }},
	{"OSPROXY_AUTH_LOCAL_SANITY_CHECK", boolean(func(c *Config) *bool { return &c.Auth.LocalSanityCheck })},
	{"OSPROXY_AUTH_ALLOW_QUERY_TOKEN", boolean(func(c *Config) *bool { return &c.Auth.AllowQueryToken })},
	{"OSPROXY_TIMEOUTS_REQUEST", dur(func(c *Config) *time.Duration { return &c.Timeouts.Request })},
	{"OSPROXY_TIMEOUTS_EXEC", dur(func(c *Config) *time.Duration { return &c.Timeouts.Exec })},
	{"OSPROXY_TIMEOUTS_WAIT_MAX", dur(func(c *Config) *time.Duration { return &c.Timeouts.WaitMax })},
	{"OSPROXY_TIMEOUTS_STREAM", dur(func(c *Config) *time.Duration { return &c.Timeouts.Stream })},
	{"OSPROXY_TIMEOUTS_SSE_HEARTBEAT", dur(func(c *Config) *time.Duration { return &c.Timeouts.SSEHeartbeat })},
	{"OSPROXY_LIMITS_MAX_BODY", size(func(c *Config) *int64 { return &c.Limits.MaxBody })},
	{"OSPROXY_LIMITS_MAX_EXEC_BODY", size(func(c *Config) *int64 { return &c.Limits.MaxExecBody })},
	{"OSPROXY_LIMITS_WS_MAX_FRAME", size(func(c *Config) *int64 { return &c.Limits.WSMaxFrame })},
	{"OSPROXY_CORS_ALLOWED_ORIGINS", func(c *Config, v string) error {
		var out []string
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				out = append(out, o)
			}
		}
		c.CORS.AllowedOrigins = out
		return nil
	}},
	{"OSPROXY_CORS_ALLOW_CREDENTIALS", boolean(func(c *Config) *bool { return &c.CORS.AllowCredentials })},
	{"OSPROXY_LOG_LEVEL", str(func(c *Config) *string { return &c.Log.Level })},
	{"OSPROXY_LOG_FORMAT", str(func(c *Config) *string { return &c.Log.Format })},
}

func applyEnv(c *Config, getenv func(string) string) error {
	for _, e := range envTable {
		v := getenv(e.name)
		if v == "" {
			continue
		}
		if err := e.set(c, v); err != nil {
			return fmt.Errorf("%s: %w", e.name, err)
		}
	}
	return nil
}
