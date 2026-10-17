# Configuration

Configuration is resolved in this order (later wins): built-in defaults, the YAML file given by `--config`, then `OSPROXY_*` environment variables. Unknown YAML keys are rejected. Durations use Go syntax (`30s`, `10m`); sizes accept an integer or a `KiB`/`MiB`/`GiB` suffix.

| Key | Type | Default | Env var | Notes |
|---|---|---|---|---|
| `server.addr` | string | `:8080` | `OSPROXY_SERVER_ADDR` | Listen address. |
| `server.tls.certFile` | string | `""` | `OSPROXY_SERVER_TLS_CERT_FILE` | Must be set together with `keyFile`; both set enables HTTPS. |
| `server.tls.keyFile` | string | `""` | `OSPROXY_SERVER_TLS_KEY_FILE` | Must be set together with `certFile`. |
| `server.readHeaderTimeout` | duration | `10s` | `OSPROXY_SERVER_READ_HEADER_TIMEOUT` | |
| `server.idleTimeout` | duration | `2m0s` (120s) | `OSPROXY_SERVER_IDLE_TIMEOUT` | |
| `gateway.address` | string | `""` (required) | `OSPROXY_GATEWAY_ADDRESS` | OpenShell gRPC gateway `host:port`. `http://` prefix requires `allowPlaintext`. |
| `gateway.caFile` | string | `""` | `OSPROXY_GATEWAY_CA_FILE` | CA bundle for the gateway; must not be set with `allowPlaintext`. |
| `gateway.insecureSkipVerify` | bool | `false` | `OSPROXY_GATEWAY_INSECURE_SKIP_VERIFY` | Skips gateway certificate verification. Development only. |
| `gateway.allowPlaintext` | bool | `false` | `OSPROXY_GATEWAY_ALLOW_PLAINTEXT` | Plain gRPC to the gateway. Requires `auth.mode: none`. |
| `auth.mode` | string | `bearer-passthrough` | `OSPROXY_AUTH_MODE` | `bearer-passthrough` or `none`. Plaintext gateways force `none`. |
| `auth.localSanityCheck` | bool | `true` | `OSPROXY_AUTH_LOCAL_SANITY_CHECK` | Cheap local token shape check before forwarding. |
| `auth.allowQueryToken` | bool | `false` | `OSPROXY_AUTH_ALLOW_QUERY_TOKEN` | Accept the token as a query parameter (for clients that cannot set headers). |
| `timeouts.request` | duration | `30s` | `OSPROXY_TIMEOUTS_REQUEST` | Per-request deadline for unary calls. |
| `timeouts.exec` | duration | `10m0s` (10m) | `OSPROXY_TIMEOUTS_EXEC` | |
| `timeouts.waitMax` | duration | `10m0s` (10m) | `OSPROXY_TIMEOUTS_WAIT_MAX` | |
| `timeouts.stream` | duration | `1h0m0s` (1h) | `OSPROXY_TIMEOUTS_STREAM` | |
| `timeouts.sseHeartbeat` | duration | `15s` | `OSPROXY_TIMEOUTS_SSE_HEARTBEAT` | |
| `limits.maxBody` | size | `1MiB` | `OSPROXY_LIMITS_MAX_BODY` | |
| `limits.maxExecBody` | size | `4MiB` | `OSPROXY_LIMITS_MAX_EXEC_BODY` | |
| `limits.wsMaxFrame` | size | `1MiB` | `OSPROXY_LIMITS_WS_MAX_FRAME` | |
| `cors.allowedOrigins` | list of string | `[]` | `OSPROXY_CORS_ALLOWED_ORIGINS` | Env value is comma separated. |
| `cors.allowCredentials` | bool | `false` | `OSPROXY_CORS_ALLOW_CREDENTIALS` | |
| `log.level` | string | `info` | `OSPROXY_LOG_LEVEL` | `debug`, `info`, `warn`, `error`. |
| `log.format` | string | `json` | `OSPROXY_LOG_FORMAT` | `json` or `text`. |

## Validation rules

- `gateway.address` is required.
- `gateway.allowPlaintext: true` requires `auth.mode: none` and forbids `gateway.caFile`.
- `server.tls.certFile` and `server.tls.keyFile` are set together or not at all.

## Running

Against a TLS gateway with bearer passthrough:

```sh
OSPROXY_GATEWAY_ADDRESS=gateway.example.com:443 go run ./cmd/openshell-proxy --config proxy.yaml
```

Local development against a plaintext gateway (no auth):

```sh
OSPROXY_GATEWAY_ADDRESS=127.0.0.1:17670 OSPROXY_AUTH_MODE=none OSPROXY_GATEWAY_ALLOW_PLAINTEXT=true \
  go run ./cmd/openshell-proxy
```

Print the OpenAPI document (needs no config or gateway):

```sh
go run ./cmd/openshell-proxy --print-openapi
```
