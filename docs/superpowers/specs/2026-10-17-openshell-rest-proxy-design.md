# OpenShell REST Proxy — Design Spec (approved 2026-10-17)

## Context

OpenShell's gateway exposes a single gRPC service (`openshell.v1.OpenShell`, 84 RPCs). There is no HTTP/JSON surface: no `google.api.http` annotations, no grpc-gateway, no Connect/gRPC-Web listener (verified in `proto/openshell.proto` and the gateway configuration docs). Automation tooling (CI jobs, webhooks, low-code platforms, shell scripts) wants plain REST. This project builds `openshell-proxy`: a Go service exposing a RESTful API that proxies to a gateway through the official Go SDK, forwarding each caller's OIDC bearer token so the gateway keeps enforcing per-user RBAC.

State: `/Users/kenmoini/Development/openshell-proxy` is an empty git repo. Go 1.27.1 and `openshell` CLI 0.1.4 are installed locally. SDK source for reference is checked out read-only at `/tmp/claude-501/osrepo/repo/sdk/go` (HEAD `b58e24c`, plus tag `v0.1.4` fetched).

## Decisions (made by user)

| Topic | Decision |
|---|---|
| Caller auth | **Pass-through bearer.** Callers send `Authorization: Bearer <OIDC JWT>` for the gateway's issuer/audience; proxy forwards it verbatim per request. Gateway is the authorization authority. |
| Scope | **Everything the SDK offers** (incl. interactive exec, SSH tunnel, TCP forward over WebSocket; policy drafts; config). `Files()` excluded (SDK stub). |
| Config | **YAML file** with env-var overrides. |
| HTTP stack | **huma v2 (v2.39.1) on stdlib `net/http` ServeMux**; `huma/v2/sse` for SSE; `github.com/coder/websocket` (already an SDK dep) for WebSocket routes mounted raw on the same mux. |

Calls I'm making (not user-decided, flag if wrong): JSON is camelCase with proxy-owned DTOs; `uint64` resource versions serialized as strings; RFC 9457 `application/problem+json` errors; module path `github.com/kenmoini/openshell-proxy`; env prefix `OSPROXY_`.

## SDK research summary

### What the SDK gives us
- Import `v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"`; Go ≥ 1.25.13; deps grpc 1.83, x/oauth2, x/sync, coder/websocket.
- `v1.NewClient(v1.Config{Address, TLS, Auth, Timeout, RetryPolicy, Logger})`. **Only `Address`, `TLS`, `Auth` are used** (`client.go:66-100`); `Timeout`/`RetryPolicy` are ignored → the proxy owns all deadlines; no SDK retries. `http://` address ⇒ plaintext, else TLS (`internal/grpc/conn.go`).
- `AuthProvider` == gRPC `PerRPCCredentials`: `GetRequestMetadata(ctx, ...)` receives the **call context**, so a single shared client can inject a per-request token read from ctx. `RequireTransportSecurity()` true is rejected with `http://` addresses.
- Subclient interfaces (files `sandbox.go`, `exec.go`, `workspace.go`, `provider.go`, `profile.go`, `refresh.go`, `service.go`, `policy.go`, `config.go`, `health.go`, `ssh.go`, `tcp.go`, `sandbox_template.go`): Sandboxes (Create/Get/List/Stop/Start/Delete/Attach|Detach|ListProviders/WaitReady/WaitStopped/Watch/GetLogs), SandboxTemplates, Exec (Run/Stream/Interactive), Services, Providers(+Profiles, +Refresh, Ensure), Workspaces(+members), Policy (draft review + status + revisions), Config (GetSandbox/GetGateway/Update), Health (Check/GetGatewayInfo/GetCurrentUser), SSH (CreateSession/RevokeSession/Tunnel), TCP (Forward/Listen).
- Pagination: `List()` returns lazy `*Pager[T]`; `NextPage(ctx)`; `ListOptions{PageSize, PageToken, LabelSelector, AllWorkspaces}`. Policy list uses its own `ListPolicyOption`s.
- Errors: `*types.StatusError{Code, Message, Cause, GRPCCode, FieldViolations, ErrorInfo{Reason,Domain,Metadata}, RetryDelay}`. The converter **folds** `ResourceExhausted→ErrorUnavailable`, `FailedPrecondition→ErrorConflict`, `Aborted→ErrorConflict`, `OutOfRange→ErrorInvalidArgument` (`internal/converter/errors.go`), so HTTP mapping must branch on `GRPCCode` first.
- `fake` package implements `ClientInterface` in memory, but **only** Sandboxes/Templates/Providers/Workspaces/Health and Policy `List`/`GetStatus` are real; Exec, Services, Config, Profiles, Refresh, SSH, TCP, and the other Policy methods return `ErrorUnimplemented`.
- `oidc.NewClientCredentialsAuth` and `gateway.NewClient/LoadConfig` exist but are not needed for pass-through mode (documented as a possible later "service account" mode).

### Gateway auth/RBAC facts that constrain the design
- Token is gRPC metadata `authorization: Bearer <jwt>`; gateway validates iss/aud/exp/sub against JWKS, reads roles from `roles_claim` (default `realm_access.roles`; admin `openshell-admin`, user `openshell-user`) and optional scopes (`sandbox:read|write`, `provider:*`, `config:*`, `workspace:*`, `openshell:all`).
- Each RPC has an `(openshell.options.v1.authorization)` annotation: `auth_mode`, `scope`, `workspace_role` (user|admin) or `global_role` (platform_admin). `Health` is unauthenticated; `GetGatewayInfo`, `CreateWorkspace`, `DeleteWorkspace` are platform-admin. Route table below records these.
- Workspace membership is per subject; `allWorkspaces=true` list flag is platform-admin only. `GetProviderCredentials` is mTLS-operator-only and not in the SDK → out of scope.

### Gaps / issues / risks (and how the plan handles them)
| # | Issue | Handling |
|---|---|---|
| G1 | SDK auth is per-client, not per-request | Custom ctx-aware `AuthProvider`; one shared `*v1.Client`. |
| G2 | `SSH().Tunnel` revokes the session in a cleanup using `context.Background()` (`ssh_client.go:142`) → with ctx auth, revoke would run unauthenticated and leak sessions | SSH tunnel WebSocket uses a short-lived **dedicated** client built with `v1.StaticToken(callerToken)`; everything else uses the shared client. |
| G3 | `Timeout`/`RetryPolicy` ignored by `NewClient` | Per-request `context.WithTimeout` in handlers; streaming routes opt out; document that `Retry-After` is advisory. |
| G4 | `WaitReady/WaitStopped` synthesize `ErrorInternal` with `GRPCCode==0` for terminal phases (`sandbox_client.go:314-327`) | Mapper rule: `ErrorInternal && GRPCCode==0 && Cause==nil` → 409 `SANDBOX_TERMINAL_PHASE`, not 502. |
| G5 | Idempotency: proto has `request_id` on mutations; SDK never sets it | Document `Idempotency-Key` as unsupported in v1; consider upstream issue. |
| G6 | No module tags on `sdk/go`; `@latest` = pseudo-version off `main`; API changed between v0.1.4 and HEAD (sandbox `Create(..., labels, opts ...CreateOptions)` → `Create(..., opts ...CreateOption)`, new `WithLabels/WithAnnotations/WithServiceExposures`) | **Pin to the v0.1.4 tag commit `5a7728c`** to match the installed 0.1.4 gateway; `make update-sdk SDK_REF=` + `docs/sdk-pinning.md`. Code written against the v0.1.4 signatures. |
| G7 | `Files()` is a stub (`ErrTransportNotAvailable`) | Excluded; noted in docs. Possible later: exec-based transfer. |
| G8 | Byte-stream RPCs (Interactive, SSH tunnel, TCP forward) aren't REST | WebSocket endpoints with documented protocols. `TCP().Listen` excluded (binds local ports). |
| G9 | Domain types have no JSON tags | Proxy-owned DTOs + converters with round-trip tests. |
| G10 | `fake` coverage gaps | `internal/gateway/gatewaytest` embeds `*fake.Client` and allows swapping any sub-interface with hand-written stubs; integration test is the real safety net. |
| G11 | Secrets in payloads (`SSHSession.Token`, `Provider.Spec.Credentials`, `RefreshConfig.Material`) | DTOs implement `slog.LogValuer` redaction; never log bodies; SSH revoke token goes in body, never URL. |
| G12 | Dev against plaintext local gateway | Config `gateway.allowPlaintext: true` + `auth.mode: none`; provider's `RequireTransportSecurity` follows the flag; validation forbids CA/cert with `http://`. |
| G13 | `http.Server.WriteTimeout` must be 0 for SSE/WS | Rely on `ReadHeaderTimeout`, `IdleTimeout`, per-route ctx deadlines. |

## Repository layout

```
cmd/openshell-proxy/main.go          --config, --print-openapi
internal/config/                     config.go load.go env.go validate.go
internal/auth/                       context.go middleware.go sanity.go
internal/gateway/                    provider.go (ctx AuthProvider) client.go (Shared/Dedicated)
internal/gateway/gatewaytest/        client.go stubs.go exec_stub.go errors.go
internal/httperr/                    problem.go map.go install.go
internal/dto/                        common.go <resource>.go ... ; convert/ (SDK<->DTO + tests)
internal/api/                        api.go params.go list.go <resource>.go openapi_ws.go
internal/stream/sse/                 exec.go watch.go heartbeat.go
internal/stream/ws/                  accept.go protocol.go interactive.go tunnel.go
internal/middleware/                 requestid logging recover bodylimit timeout cors hopheaders
internal/server/server.go            http.Server, TLS, graceful shutdown, /healthz /readyz
configs/example.yaml  docs/  test/integration/  Makefile  Dockerfile  .golangci.yml
```

Deps: SDK @ `5a7728c`, `github.com/danielgtaylor/huma/v2@v2.39.1`, `github.com/coder/websocket`, `gopkg.in/yaml.v3`, `github.com/rs/cors`, `github.com/stretchr/testify` (test).

## Config shape (YAML + `OSPROXY_*` env overrides)

```yaml
server:   { addr: ":8080", tls: {certFile, keyFile}, readHeaderTimeout: 10s, idleTimeout: 120s }
gateway:  { address: "gateway.example.com:443", caFile: "", insecureSkipVerify: false, allowPlaintext: false }
auth:     { mode: bearer-passthrough | none, localSanityCheck: true, allowQueryToken: false }
timeouts: { request: 30s, exec: 10m, waitMax: 10m, stream: 1h, sseHeartbeat: 15s }
limits:   { maxBody: 1MiB, maxExecBody: 4MiB, wsMaxFrame: 1MiB }
cors:     { allowedOrigins: [], allowCredentials: false }
log:      { level: info, format: json }
```

## Route table (all under `/v1`, bearer required unless noted)

Path style: action verbs are sub-path segments (`/start`, `/approve-all`, ...) because Go's 1.22 mux rejects `{name}:verb` patterns (verified). Conventions: list routes accept `pageSize, pageToken, labelSelector, allWorkspaces` → `{items, nextPageToken}`; deletes accept `?allowMissing=true` and return `DeletionResult` (200 Completed/AlreadyAbsent, 202 Accepted); `If-Match: "<resourceVersion>"` feeds `expectedResourceVersion`; `ws`=`{workspace}`, `sb`=`{sandbox}`. Auth column = proto annotation.

**Platform/health**: `GET /healthz` (proxy), `GET /readyz` (`Health().Check`, tokenless), `GET /v1/health`, `GET /v1/whoami` (`GetCurrentUser`), `GET /v1/gateway/info` (`GetGatewayInfo`, config:read/platform_admin), `GET|PATCH /v1/gateway/config` (`Config().GetGateway/Update` global), `GET /openapi.json`, `GET /docs`.

**Workspaces** (`Workspaces()`): `POST /workspaces` (platform_admin) · `GET /workspaces` · `GET|DELETE /workspaces/{ws}` (delete platform_admin) · `GET|POST /workspaces/{ws}/members` (add: admin) · `DELETE /workspaces/{ws}/members/{subject}` (admin).

**Sandboxes** (`Sandboxes()`, sandbox:read|write / user): `POST /workspaces/{ws}/sandboxes` (body `{name, spec, labels, annotations, serviceExposures, template?}` → `Create` or `CreateSandboxFromTemplate`, 201) · `GET` list · `GET|DELETE .../{sb}` · `POST .../{sb}/start` and `POST .../{sb}/stop` (`?wait=&timeout=`) · `POST .../{sb}/wait` (`WaitReady|WaitStopped`, timeout capped by `timeouts.waitMax`) · `GET .../{sb}/watch` (SSE) · `GET .../{sb}/logs` (`?lines&since&sources&minLevel`) · `GET .../{sb}/providers` · `PUT|DELETE .../{sb}/providers/{provider}` (Attach/Detach, `If-Match`).

**Sandbox templates** (`SandboxTemplates()`): `POST|GET /workspaces/{ws}/sandbox-templates` (create: admin) · `GET|DELETE .../{name}` (delete: admin).

**Exec** (`Exec()`, sandbox:write/user): `POST .../{sb}/exec` (`Run`; body `{command, env, workDir, noLoginShell, encoding: utf8|base64}` → `{exitCode, stdout, stderr}`) · `POST .../{sb}/exec/stream` (SSE `stdout|stderr|exit|error|ping`) · `WS .../{sb}/exec/interactive`.

**Services** (`Services()`): `POST|GET .../{sb}/services` · `GET|DELETE .../{sb}/services/{name}`.

**Providers** (`Providers()`, provider:read user / provider:write admin): `POST|GET /workspaces/{ws}/providers` · `GET|PUT|DELETE .../providers/{name}` (`PUT ?ensure=true` → `Ensure`) · profiles: `GET /provider-profiles`, `GET|PUT|DELETE /provider-profiles/{id}` (`If-Match` on PUT), `POST /provider-profiles/import`, `POST /provider-profiles/lint` · refresh: `GET .../providers/{p}/refresh?credentialKey=`, `PUT|DELETE .../providers/{p}/refresh/{key}`, `POST .../refresh/{key}/rotate`.

**Policy** (`Policy()`): `GET .../{sb}/policy/draft` · `POST .../policy/draft/chunks/{chunk}/approve|reject|undo` · `PUT .../policy/draft/chunks/{chunk}` (edit) · `POST .../policy/draft/approve-all` · `POST .../policy/draft/clear` · `GET .../policy/draft/history` · `GET .../{sb}/policy/status?version=` · `GET .../{sb}/policy/revisions` · `GET /workspaces/{ws}/policy/global/status|revisions`. Draft mutations are config:write/admin.

**Config** (`Config()`): `GET|PATCH .../{sb}/config` (`GetSandbox` / `Update`, `If-Match`).

**SSH/TCP** (`SSH()`, `TCP()`, sandbox:write/user): `POST .../{sb}/ssh-sessions` (201, returns token) · `POST /workspaces/{ws}/ssh-sessions/revoke` (token in body) · `WS .../{sb}/ssh/tunnel?port=&serviceId=` (dedicated client, G2) · `WS .../{sb}/tcp/{port}?serviceId=`.

Excluded: `Files()`, `TCP().Listen`, `GetProviderCredentials`.

## Key designs

**Ctx-aware AuthProvider** (`internal/gateway/provider.go`): `GetRequestMetadata` returns `authorization: Bearer <tok>` when `auth.TokenFrom(ctx)` is set and `x-request-id` from the request-id middleware; returns empty metadata (never an error) when no token so tokenless `Health` works. `RequireTransportSecurity()` = `!cfg.gateway.allowPlaintext`. `gateway.New(cfg)` builds one shared client; `Dedicated(ctx)` builds a `StaticToken` client for SSH tunnel lifetimes.

**Auth middleware** (`internal/auth`): mode `bearer-passthrough` requires a well-formed bearer (401 + `WWW-Authenticate: Bearer` otherwise); optional local sanity check (3 segments, decodable payload, `exp` with 60 s skew) is a fast-fail only, never the authority. Exempt: `/healthz`, `/readyz`, `/openapi.json`, `/docs`, `/v1/health`. WS routes take the header; `?access_token=` only when `allowQueryToken` (scrubbed from logs).

**Error mapping** (`internal/httperr`): `Problem` (RFC 9457 + `reason`, `domain`, `metadata`, `grpcCode`, `errors[]` from FieldViolations, `retryAfterSeconds`, `outcomeUnknown`, `requestId`), implements `huma.StatusError` + headers (`Retry-After`, `WWW-Authenticate`). Installed via `huma.NewError` override so huma's own 4xx share the envelope. Switch on `GRPCCode` first:

| gRPC | HTTP |
|---|---|
| INVALID_ARGUMENT, OUT_OF_RANGE | 400 + `errors[]` |
| UNAUTHENTICATED | 401 |
| PERMISSION_DENIED | 403 |
| NOT_FOUND | 404 |
| ALREADY_EXISTS | 409 |
| ABORTED | 409 (412 when `If-Match` was sent) |
| FAILED_PRECONDITION | 412 if `If-Match`; 409 if reason starts `REQUEST_`; else 422 |
| RESOURCE_EXHAUSTED | 429 + Retry-After |
| UNAVAILABLE | 503 + Retry-After |
| DEADLINE_EXCEEDED, CANCELLED | 504, `outcomeUnknown: true` |
| UNIMPLEMENTED | 501 |
| INTERNAL, UNKNOWN, DATA_LOSS, non-StatusError | 502 |
| `GRPCCode==0` (SDK-local) | by `Code`; G4 terminal-phase rule → 409 |

**SSE** (`huma/v2/sse.Register`): typed event map; exec loop on `ExecStream.Next()` until EOF then `ExitCode()`; watch loop on `ResultChan()` → `added|modified|deleted|error`; `ping` heartbeat every `sseHeartbeat` (huma SSE has no comment API); headers `Cache-Control: no-cache`, `X-Accel-Buffering: no`; client disconnect cancels ctx and closes the stream.

**WebSocket protocols** (`internal/stream/ws`): interactive: first text frame `{"type":"start","command":[],"cols","rows","env","workDir","noLoginShell"}` within 5 s; binary frames = stdin; text `resize|closeStdin|cancel` map to `Resize`, `v1.CloseInteractiveInput`, `v1.CancelInteractive`; server sends binary merged output (SDK merges stdout/stderr in `Interactive`), then `{"type":"exit","exitCode"}` and close 1000, or `{"type":"error", problem}` + close 1011/1008. Tunnel/forward: `websocket.NetConn` ↔ `io.ReadWriteCloser` with two `io.Copy`s. SDK errors before `Accept` become HTTP problems; after, close codes with reason. Documented in OpenAPI as manual path items with an `x-websocket` extension plus `docs/websocket-protocols.md`.

**DTOs**: camelCase; enums as strings; RFC3339 times; `uint64` versions as strings (accept number or string on input); `[]byte` base64; `map[string]any` passthrough for `Resources`/`DriverConfig`; shared `ObjectMeta`; generic `Page[T]`. Converters `convert.XFromSDK` / `convert.XToSDK` with round-trip and golden-JSON tests.

## Phases (TDD; each ends with a verification)

0. **Bootstrap**: `go mod init`, pin SDK `@5a7728c`, huma, websocket, yaml, cors, testify; Makefile (`build test lint openapi run integration update-sdk`), `.golangci.yml`, `docs/sdk-pinning.md`. Verify `go build ./...`.
1. **Vertical slice**: config (parse/env/validate incl. G12 rules) → middlewares (requestid, recover, logging, bodylimit, timeout w/ streaming opt-out, hopheaders, cors) → auth middleware → `httperr` table tests → `gateway` provider + factory + `gatewaytest` skeleton → `api.go` (humago, bearer security scheme) + health/whoami/gateway-info → `server.go` (graceful shutdown, `/healthz`, `/readyz`) → `main.go`. Verify: tests green; `curl -H "Authorization: Bearer $TOK" :8080/v1/whoami`; no token → 401 problem+json; `/docs` renders; `/readyz` works tokenless.
2. **Core resources**: DTOs/converters for workspace, sandbox, template, deletion; generic list helper; workspaces (+members), sandboxes (create incl. template branch, start/stop `?wait`, `:wait` cap, logs, providers attach/detach with `If-Match`), templates. Tested against `fake`.
3. **Providers/profiles/refresh**: DTOs with `slog.LogValuer` redaction; `ProfileStub`/`RefreshStub`; handlers incl. `:ensure`, `:import`, `:lint`. Verify Configure bodies never appear in logs.
4. **Services, config, policy**: DTOs incl. full network-policy graph and `SettingValue` tagged union; `ServiceStub/ConfigStub/PolicyStub`; handlers incl. global policy routes. Verify `PATCH config` conflict with `If-Match` → 412.
5. **SSE**: `ExecStub` with scripted streams; exec run/stream + watch handlers; heartbeat with injected clock; disconnect cancels. Verify with `curl -N`.
6. **WebSocket**: protocol types; `accept.go`; interactive with pipe-backed session stub; SSH create/revoke + tunnel (dedicated client) and TCP forward; `openapi_ws.go`; docs. Verify with `websocat` and `ssh -o ProxyCommand=...`.
7. **Hardening/release**: proxy TLS, optional `/metrics`, size/slowloris tests, build-tagged integration test (`OSPROXY_IT_GATEWAY/TOKEN/WORKSPACE`: whoami → create → wait → exec → stream → delete), Dockerfile (distroless, non-root), docs, committed `openapi.json`. Verify `go test -race ./...`, `golangci-lint`, `make integration`.

## Verification (end-to-end)

- Unit: `go test -race ./...` (converters round-trip + golden JSON, httperr table, config, middlewares, auth, provider).
- Handler: `httptest` server over the huma mux with `gatewaytest` (fake + stubs); SSE via line scanner; WS via `websocket.Dial`.
- Manual against a real gateway (user's local 0.1.4 or a dev OIDC gateway): `/readyz`, `/v1/whoami` with a real IdP token, create sandbox → `:wait` → `exec` → `exec/stream` → delete; confirm 401/403 pass straight through from the gateway; confirm SSH tunnel close revokes the session (`openshell` CLI / gateway logs).
- `make integration` build-tagged test as the regression gate when re-pinning the SDK.

## Open items to confirm during build (not blocking)
- Whether the gateway accepts `Config().Update` with empty workspace for `Global=true` (else expose `?workspace=`).
- Whether the user's local gateway is OIDC or plaintext; local dev likely needs `auth.mode: none` + `allowPlaintext: true`. (Reading `~/.config/openshell/gateways/*/metadata.json` was blocked by the permission classifier; the user can run `openshell status` / `openshell whoami` to tell.)
