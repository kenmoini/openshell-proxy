# openshell-proxy

`openshell-proxy` is a REST/JSON proxy in front of the gRPC gateway of [NVIDIA OpenShell](https://github.com/NVIDIA/OpenShell). It exposes workspaces, sandboxes, sandbox templates, exec, services, providers, provider profiles, credential refresh, sandbox and gateway config, and policy draft review as a conventional HTTP API with an OpenAPI 3.1 description, so tools and scripts that cannot speak gRPC can drive OpenShell.

It is built on the official OpenShell Go SDK and adds no policy of its own. Callers present their own OIDC bearer token; the proxy forwards it to the gateway, which remains the single authority for authentication, RBAC and workspace membership.

## Architecture

- Pass-through bearer auth: a context-aware auth provider on one shared SDK client attaches each caller's token to the gRPC call, so no per-user clients or credential storage.
- [huma](https://github.com/danielgtaylor/huma) v2 on the standard library `net/http` mux generates the OpenAPI document from the handlers.
- Proxy-owned request and response DTOs, so the HTTP contract does not depend on SDK or protobuf types, and secrets are redacted from logs.
- Errors are RFC 9457 `application/problem+json` with a fixed gRPC-to-HTTP mapping ([docs/errors.md](docs/errors.md)).
- Configuration is a YAML file with `OSPROXY_*` environment overrides ([docs/configuration.md](docs/configuration.md)).

## Quick start

```sh
go build -o openshell-proxy ./cmd/openshell-proxy

# Edit configs/example.yaml (at minimum gateway.address), or override with env:
OSPROXY_GATEWAY_ADDRESS=gateway.example.com:443 \
OSPROXY_GATEWAY_CA_FILE=/path/to/ca.pem \
  ./openshell-proxy --config configs/example.yaml

# In another shell, with a token your gateway accepts:
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/whoami
```

Useful endpoints: `/healthz` (liveness), `/readyz` (gateway reachable), `/openapi.json`, `/docs` (interactive docs). Print the OpenAPI document without a gateway using `go run ./cmd/openshell-proxy --print-openapi`; `make openapi` regenerates the committed `openapi.json` and `make check-openapi` fails when it is stale.

## Auth model

Callers bring their own OIDC token, issued for the gateway's issuer and audience, in `Authorization: Bearer <token>`. The proxy forwards it verbatim on each request. It never mints, stores or validates tokens beyond an optional local fast-fail check (shape and `exp`); the gateway enforces RBAC and workspace membership and its 401/403 responses pass through. The tokenless routes are `/healthz`, `/readyz`, `/openapi.json`, `/docs` and `/v1/health`.

For local development against a plaintext gateway, set `auth.mode: none` together with `gateway.allowPlaintext: true` (env: `OSPROXY_AUTH_MODE=none`, `OSPROXY_GATEWAY_ALLOW_PLAINTEXT=true`). Do not use this outside a trusted machine.

## Documentation

- [docs/api.md](docs/api.md): routes, conventions (pagination, `If-Match`, `allowMissing`, `X-Request-ID`, `Retry-After`)
- [docs/errors.md](docs/errors.md): problem shape and the gRPC-to-HTTP mapping
- [docs/configuration.md](docs/configuration.md): config keys, env vars, validation
- [docs/sdk-pinning.md](docs/sdk-pinning.md): how the OpenShell SDK is pinned and updated
- [openapi.json](openapi.json): generated OpenAPI document

## Limitations

- `Idempotency-Key` is not supported in v1: the SDK does not expose the gateway's `request_id`.
- `Files()` is unsupported by the SDK, so there are no file transfer routes.
- Streaming exec and watch (SSE) and the interactive exec, SSH and TCP routes (WebSocket) are a follow-up; exec is currently collected (buffered).
- The SDK is pre-1.0 and is pinned to the v0.1.4 commit.

## Verifying against a real gateway

First learn how your gateway authenticates. With the `openshell` CLI:

```sh
openshell status
openshell whoami --output json
```

Then start the proxy in one terminal, using the variant that matches.

OIDC gateway (callers need an IdP token for the gateway's audience):

```sh
OSPROXY_GATEWAY_ADDRESS=<host:port> OSPROXY_GATEWAY_CA_FILE=<ca.pem> make run
export TOKEN=<idp token>
```

For something like Keycloak, use the following:

```sh
OPENSHELL_PROXY_URL="openshell-proxy-openshell.apps.ocp.example.com"
KEYCLOAK_URL="keycloak.example.com"
KEYCLOAK_REALM="openshell"

ACCESS_TOKEN=$(curl -s -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_secret=${YOUR_CLIENT_SECRET}" \
  -d "client_id=openshell-ci" \
  -d "grant_type=client_credentials" \
  "https://${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token" \
  | jq -r '.access_token')

curl \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://${OPENSHELL_PROXY_URL}/v1/whoami

curl \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://${OPENSHELL_PROXY_URL}/v1/workspaces
```

Plaintext local gateway (no auth):

```sh
OSPROXY_GATEWAY_ADDRESS=127.0.0.1:17670 OSPROXY_GATEWAY_ALLOW_PLAINTEXT=true OSPROXY_AUTH_MODE=none make run
```

In a second terminal. Add `-H "Authorization: Bearer $TOKEN"` to each `curl` for the OIDC variant; omit it for the plaintext variant.

```sh
B=http://localhost:8080
H=(-H "Content-Type: application/json")
[ -n "$TOKEN" ] && H+=(-H "Authorization: Bearer $TOKEN")

curl -i $B/readyz                                   # expect 200
curl -s "${H[@]}" $B/v1/whoami
curl -s "${H[@]}" -X POST $B/v1/workspaces/default/sandboxes \
  -d '{"name":"proxy-smoke","spec":{"template":{"image":"python:3.12"}}}'
curl -s "${H[@]}" -X POST $B/v1/workspaces/default/sandboxes/proxy-smoke/wait \
  -d '{"for":"ready"}'
curl -s "${H[@]}" -X POST $B/v1/workspaces/default/sandboxes/proxy-smoke/exec \
  -d '{"command":["echo","hi"]}'                    # expect stdout "hi\n", exitCode 0
curl -i "${H[@]}" -X DELETE $B/v1/workspaces/default/sandboxes/proxy-smoke   # 200 or 202
```

The array syntax above works in bash and zsh. Record the outcomes (status codes and the exec result) in the PR description.
