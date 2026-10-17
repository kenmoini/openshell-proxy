package config

import (
	"errors"
	"strings"
)

// Validate checks cross-field rules and returns all violations joined.
func (c *Config) Validate() error {
	var errs []error
	add := func(msg string) { errs = append(errs, errors.New(msg)) }

	g := c.Gateway
	if g.Address == "" {
		add("gateway.address is required")
	}
	if !g.AllowPlaintext && strings.HasPrefix(g.Address, "http://") {
		add("gateway.address uses http:// but gateway.allowPlaintext is false")
	}
	if g.InsecureSkipVerify && g.AllowPlaintext {
		add("gateway.insecureSkipVerify must not be set when gateway.allowPlaintext is true")
	}
	if g.AllowPlaintext && g.CAFile != "" {
		add("gateway.caFile must not be set when gateway.allowPlaintext is true")
	}

	switch c.Auth.Mode {
	case ModeBearerPassthrough:
		if g.AllowPlaintext {
			add("auth.mode must be none when gateway.allowPlaintext is true")
		}
	case ModeNone:
	default:
		add("auth.mode must be one of bearer-passthrough, none")
	}

	if (c.Server.TLS.CertFile == "") != (c.Server.TLS.KeyFile == "") {
		add("server.tls.certFile and server.tls.keyFile must be set together")
	}

	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		add("log.level must be one of debug, info, warn, error")
	}
	switch c.Log.Format {
	case "json", "text":
	default:
		add("log.format must be one of json, text")
	}

	positive := func(name string, ok bool) {
		if !ok {
			add(name + " must be greater than zero")
		}
	}
	t, l := c.Timeouts, c.Limits
	positive("timeouts.request", t.Request > 0)
	positive("timeouts.exec", t.Exec > 0)
	positive("timeouts.waitMax", t.WaitMax > 0)
	positive("timeouts.stream", t.Stream > 0)
	positive("timeouts.sseHeartbeat", t.SSEHeartbeat > 0)
	positive("limits.maxBody", l.MaxBody > 0)
	positive("limits.maxExecBody", l.MaxExecBody > 0)
	positive("limits.wsMaxFrame", l.WSMaxFrame > 0)
	return errors.Join(errs...)
}

// DialAddress returns the target to hand to the gateway client: plaintext
// gateways get an http:// scheme, TLS gateways have any https:// stripped.
func (g GatewayConfig) DialAddress() string {
	if g.AllowPlaintext {
		if strings.Contains(g.Address, "://") {
			return g.Address
		}
		return "http://" + g.Address
	}
	return strings.TrimPrefix(g.Address, "https://")
}
