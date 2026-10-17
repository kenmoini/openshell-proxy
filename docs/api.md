# API reference

All routes live under `/v1` unless noted. The machine-readable contract is [`openapi.json`](../openapi.json) (also served at `/openapi.json`, with interactive docs at `/docs`). This page is derived from the route registrations in `internal/api/*.go`; the code is the source of truth. The proxy registers 61 `/v1` operations plus three raw routes (`/healthz`, `/readyz`, `/openapi.json`) and `/docs`.

Authorization is decided by the OpenShell gateway, not the proxy. The "Gateway role / scope" column shows what the gateway typically requires (`scope / role`); "gateway-enforced" means the spec does not pin a specific scope. In every path below `{ws}` is `/v1/workspaces/{workspace}` and `{sb}` is `{ws}/sandboxes/{sandbox}`.

## Conventions

- **Auth**: `Authorization: Bearer <token>` on every route except `/healthz`, `/readyz`, `/openapi.json`, `/docs` and `/v1/health`. See the README for the pass-through model.
- **Pagination**: list routes accept `pageSize` (0 to 1000), `pageToken`, `labelSelector` and `allWorkspaces`, and return `{"items": [...], "nextPageToken": "..."}`. Policy revision lists accept only `pageSize` and `pageToken`.
- **`allowMissing`**: delete routes accept `?allowMissing=true` so deleting an absent resource succeeds instead of returning 404.
- **Deletion results**: deletes return a `DeletionResult` body with `outcome` of `Completed`, `Accepted` or `AlreadyAbsent`. The HTTP status is 200 for `Completed` and `AlreadyAbsent`, and 202 when the gateway only `Accepted` the deletion (asynchronous teardown).
- **`If-Match`**: optimistic concurrency. The value is the numeric resource version, bare (`7`) or quoted (`"7"`). Weak validators, `*` and non-numeric values give 400. It maps to the gateway's expected resource version, and a mismatch is reported as 412 (see [errors](errors.md)).
- **`X-Request-ID`**: a valid inbound value is propagated, otherwise one is generated. It is returned on every response, forwarded to the gateway, and echoed as `requestId` in problem bodies.
- **`Retry-After`**: sent with 429 and 503 (and mirrored as `retryAfterSeconds`). It is advisory.
- **`Idempotency-Key`**: not supported in v1. The gateway has a per-mutation `request_id`, but the SDK does not expose it. Retry mutations only when you know the outcome (see `outcomeUnknown` in [errors](errors.md)).
- **Errors**: `application/problem+json`, documented in [errors.md](errors.md).

## Platform

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `/healthz` | Proxy liveness | none | Raw route, tokenless. |
| GET | `/readyz` | Proxy readiness (checks the gateway) | none | Raw route, tokenless; 503 when the gateway is unhealthy. |
| GET | `/openapi.json` | OpenAPI document | none | Served by huma, tokenless. `/docs` serves interactive docs. |
| GET | `/v1/health` | Gateway health check | none | Tokenless. |
| GET | `/v1/whoami` | Identity of the calling token | gateway-enforced | |
| GET | `/v1/gateway/info` | Gateway version, drivers and extensions | `config:read / platform_admin` | |
| GET | `/v1/gateway/config` | Get gateway-global settings | gateway-enforced | |
| PATCH | `/v1/gateway/config` | Update gateway-global settings or policy | gateway-enforced | Optional `If-Match`; optional `workspace` query. |

## Workspaces

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `/v1/workspaces` | List workspaces | gateway-enforced | Paginated. |
| POST | `/v1/workspaces` | Create a workspace | `platform_admin` | 201. |
| GET | `{ws}` | Get a workspace | gateway-enforced | |
| DELETE | `{ws}` | Delete a workspace | `platform_admin` | 200 or 202; `allowMissing`. |
| GET | `{ws}/members` | List workspace members | gateway-enforced | Paginated. |
| POST | `{ws}/members` | Add a workspace member | `admin` | 201. |
| DELETE | `{ws}/members/{subject}` | Remove a workspace member | `admin` | 200 or 202; `allowMissing`. |

## Sandboxes

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{ws}/sandboxes` | List sandboxes | `sandbox:read / user` | Paginated. |
| POST | `{ws}/sandboxes` | Create a sandbox | `sandbox:write / user` | 201. Body may reference a template. |
| GET | `{sb}` | Get a sandbox | `sandbox:read / user` | |
| DELETE | `{sb}` | Delete a sandbox | `sandbox:write / user` | 200 or 202; `allowMissing`. |
| POST | `{sb}/start` | Start a stopped sandbox | `sandbox:write / user` | `?wait=true&timeout=` blocks until Ready; timeout capped by `timeouts.waitMax`. |
| POST | `{sb}/stop` | Stop a running sandbox | `sandbox:write / user` | `?wait=true&timeout=` blocks until Stopped. |
| POST | `{sb}/wait` | Wait for a sandbox to become ready or stopped | `sandbox:read / user` | Body `{for: ready or stopped, timeout, pollInterval}`. A terminal phase gives 409 `SANDBOX_TERMINAL_PHASE`. |
| GET | `{sb}/logs` | Fetch sandbox logs | `sandbox:read / user` | `lines`, `since`, `sources`, `minLevel`. Collected, not streamed. |
| GET | `{sb}/providers` | List providers attached to a sandbox | `sandbox:read / user` | Paginated. |
| PUT | `{sb}/providers/{provider}` | Attach a provider to a sandbox | `sandbox:write / user` | Optional `If-Match`. |
| DELETE | `{sb}/providers/{provider}` | Detach a provider from a sandbox | `sandbox:write / user` | Optional `If-Match`. |
| GET | `{sb}/config` | Get effective sandbox configuration | gateway-enforced | |
| PATCH | `{sb}/config` | Update sandbox configuration | gateway-enforced | Optional `If-Match`. |

## Sandbox templates

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{ws}/sandbox-templates` | List sandbox workload templates | gateway-enforced | Paginated. |
| POST | `{ws}/sandbox-templates` | Create a sandbox workload template | `admin` | 201. |
| GET | `{ws}/sandbox-templates/{name}` | Get a sandbox workload template | gateway-enforced | |
| DELETE | `{ws}/sandbox-templates/{name}` | Delete a sandbox workload template | `admin` | 200 or 202; `allowMissing`. |

## Exec

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| POST | `{sb}/exec` | Run a command and collect its output | `sandbox:write / user` | Body `{command, env, workDir, noLoginShell, encoding}`; returns `{exitCode, stdout, stderr}`. Body limit `limits.maxExecBody`, deadline `timeouts.exec`. Streaming (SSE) and interactive (WebSocket) exec are not in this release. |

## Services

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{sb}/services` | List sandbox services | gateway-enforced | Paginated. |
| POST | `{sb}/services` | Expose a sandbox service | gateway-enforced | 201. |
| GET | `{sb}/services/{name}` | Get a sandbox service | gateway-enforced | |
| DELETE | `{sb}/services/{name}` | Delete a sandbox service | gateway-enforced | 200 or 202; `allowMissing`. |

## Providers

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{ws}/providers` | List providers | `provider:read / user` | Paginated. |
| POST | `{ws}/providers` | Create a provider | `provider:write / admin` | 201. |
| GET | `{ws}/providers/{name}` | Get a provider | `provider:read / user` | |
| PUT | `{ws}/providers/{name}` | Update a provider | `provider:write / admin` | `?ensure=true` creates it if absent. |
| DELETE | `{ws}/providers/{name}` | Delete a provider | `provider:write / admin` | 200 or 202; `allowMissing`. |
| GET | `{ws}/providers/{provider}/refresh` | Get credential refresh status | `provider:read / user` | Optional `credentialKey` query. |
| PUT | `{ws}/providers/{provider}/refresh/{key}` | Configure credential refresh | `provider:write / admin` | |
| DELETE | `{ws}/providers/{provider}/refresh/{key}` | Delete credential refresh configuration | `provider:write / admin` | 200 or 202; `allowMissing`. |
| POST | `{ws}/providers/{provider}/refresh/{key}/rotate` | Rotate a credential now | `provider:write / admin` | |

Provider request bodies: `id`, `createdAt`, `deletionTimestamp` and `spec.credentialHandles` are server-owned and ignored on create and update. The workspace comes from the path; a non-empty `workspace` in the body that differs from the path is rejected with 400. `resourceVersion` in the body is kept as optimistic concurrency for updates, so a `GET` response can be edited and sent back to `PUT` unchanged.

## Provider profiles

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{ws}/provider-profiles` | List provider profiles | `provider:read / user` | Paginated. |
| POST | `{ws}/provider-profiles/import` | Import provider profiles | `provider:write / admin` | |
| POST | `{ws}/provider-profiles/lint` | Lint provider profiles | gateway-enforced | |
| GET | `{ws}/provider-profiles/{id}` | Get a provider profile | `provider:read / user` | |
| PUT | `{ws}/provider-profiles/{id}` | Update a provider profile | `provider:write / admin` | `If-Match` required; missing gives 428. |
| DELETE | `{ws}/provider-profiles/{id}` | Delete a provider profile | `provider:write / admin` | 200 or 202; `allowMissing`. |

## Policy

| Method | Path | Summary | Gateway role / scope | Notes |
|---|---|---|---|---|
| GET | `{sb}/policy/draft` | Get the draft policy | gateway-enforced | Optional `status` filter. |
| GET | `{sb}/policy/draft/history` | Get draft policy history | gateway-enforced | |
| POST | `{sb}/policy/draft/approve-all` | Approve all draft chunks | `config:write / admin` | |
| POST | `{sb}/policy/draft/clear` | Clear draft chunks | `config:write / admin` | |
| PUT | `{sb}/policy/draft/chunks/{chunk}` | Edit a draft chunk's proposed rule | `config:write / admin` | 204. |
| POST | `{sb}/policy/draft/chunks/{chunk}/approve` | Approve a draft chunk | `config:write / admin` | |
| POST | `{sb}/policy/draft/chunks/{chunk}/reject` | Reject a draft chunk | `config:write / admin` | 204. |
| POST | `{sb}/policy/draft/chunks/{chunk}/undo` | Undo a draft chunk approval | `config:write / admin` | |
| GET | `{sb}/policy/status` | Get a sandbox policy revision status | gateway-enforced | `version` query; 0 means latest. |
| GET | `{sb}/policy/revisions` | List sandbox policy revisions | gateway-enforced | `pageSize`, `pageToken`. |
| GET | `{ws}/policy/global/status` | Get a gateway-global policy revision status | gateway-enforced | `version` query. |
| GET | `{ws}/policy/global/revisions` | List gateway-global policy revisions | gateway-enforced | `pageSize`, `pageToken`. |

## Not in this release

These are planned as a follow-up: exec streaming (SSE), sandbox watch (SSE), interactive exec, SSH tunnel and TCP forward (WebSocket), and SSH session create/revoke. `Files()` is unsupported by the SDK. Action verbs are sub-path segments (`/start`, `/approve-all`) rather than `:verb` suffixes because Go's `ServeMux` rejects `{name}:verb` patterns.
