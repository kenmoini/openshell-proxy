# OpenShell REST Proxy (Core) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Go HTTP/JSON service that proxies REST calls to an OpenShell gRPC gateway through the official Go SDK, forwarding each caller's OIDC bearer token, covering health, workspaces, sandboxes, templates, exec (collected), services, providers/profiles/refresh, policy and config.

**Architecture:** One shared `*v1.Client` whose `AuthProvider` reads the caller's bearer token from the request context, so the gateway enforces RBAC per caller. huma v2 on the stdlib `net/http` mux provides typed handlers, validation and OpenAPI. Proxy-owned DTOs (the SDK types have no JSON tags) are converted to/from SDK types; SDK `StatusError`s are mapped to RFC 9457 problem responses.

**Tech Stack:** Go (module `go 1.25`), `github.com/NVIDIA/OpenShell/sdk/go` pinned to commit `5a7728c304b5528e954e9ae900cf21f84e1a12b1` (release v0.1.4), `github.com/danielgtaylor/huma/v2@v2.39.1` (+ `adapters/humago`), `gopkg.in/yaml.v3`, `github.com/rs/cors`, `github.com/stretchr/testify`.

**Spec:** `docs/superpowers/specs/2026-10-17-openshell-rest-proxy-design.md` (route table, error table, gaps G1–G13). Streaming (SSE/WebSocket) and hardening/release are a follow-up plan.

## Global Constraints

- Module path `github.com/kenmoini/openshell-proxy`; Go toolchain floor `go 1.25` in go.mod (SDK requires ≥ 1.25.13; local is 1.27.1).
- SDK pinned to `5a7728c304b5528e954e9ae900cf21f84e1a12b1`; code uses the **v0.1.4 signatures**: `Sandboxes().Create(ctx, ws, name, spec, labels, opts ...v1.CreateOptions)`, `v1.CreateOptions{Annotations, ServiceExposures}`; no `WithLabels` helpers exist at this pin.
- API layer depends only on `v1.ClientInterface` (never `*v1.Client`) so `fake`/stubs can be injected.
- JSON: lowerCamel field names; enums as strings; `time.Time` as RFC 3339 (`omitempty` pointers when optional); `uint64` resource/config revisions as **strings** on output, number-or-string on input.
- Errors: `application/problem+json` everywhere, including huma's own validation errors (installed via `huma.NewError` override).
- Env override rule: `OSPROXY_<SECTION>_<FIELD>` in SCREAMING_SNAKE (e.g. `OSPROXY_GATEWAY_ADDRESS`, `OSPROXY_AUTH_MODE`, `OSPROXY_TIMEOUTS_REQUEST`).
- Never log `Authorization` headers, tokens, request bodies, `Provider.Spec.Credentials`, `RefreshConfig.Material`, `SSHSession.Token`.
- `http.Server.WriteTimeout` = 0 (streaming later); per-request deadlines come from `timeouts.request` via middleware.
- Every handler passes `r.Context()`-derived ctx to the SDK (the auth provider reads the token from it).

## Review Focus

1. `Authorization` header variants: lowercase `bearer x`, multiple spaces, `Basic ...`, `Bearer` with empty token → all 401 problem with `WWW-Authenticate: Bearer` (test in Task 5).
2. Oversized JSON body (> `limits.maxBody`) → 413 problem, connection still usable (test in Task 4).
3. `If-Match` forms: `"7"`, `7`, `W/"7"` (weak → 400), `*` (→ 400) (test in Task 11).
4. URL-encoded path params (member subject `oidc|abc%2Fdef`, profile IDs with `:`) reach the SDK decoded (test in Task 13).
5. Gateway unreachable: `/readyz` → 503 problem; `/v1/whoami` → 503 problem with `grpcCode: UNAVAILABLE`; process keeps serving `/healthz` (test in Task 10).

---

### Task 1: Module bootstrap and Makefile

**Files:**
- Create: `go.mod`, `cmd/openshell-proxy/main.go` (placeholder `func main(){}`), `Makefile`, `.golangci.yml`, `.gitignore`, `docs/sdk-pinning.md`

**Interfaces:**
- Produces: a compiling module with the SDK, huma, yaml, cors, testify resolved; `make build|test|lint|update-sdk SDK_REF=<ref>`.

- [ ] **Step 1: Init module and pin dependencies**

```bash
cd /Users/kenmoini/Development/openshell-proxy
go mod init github.com/kenmoini/openshell-proxy
go get github.com/NVIDIA/OpenShell/sdk/go@5a7728c304b5528e954e9ae900cf21f84e1a12b1
go get github.com/danielgtaylor/huma/v2@v2.39.1 gopkg.in/yaml.v3 github.com/rs/cors github.com/stretchr/testify
```

- [ ] **Step 2: Write placeholder main and a smoke test that imports the SDK**

`cmd/openshell-proxy/main_test.go`: `TestSDKImports` constructs `v1.Config{Address: "127.0.0.1:1", Auth: v1.NoAuth()}` and asserts `v1.NewClient` returns no error (gRPC dials lazily) and `Close()` succeeds.

- [ ] **Step 3: Run `go build ./... && go test ./...`** — Expected: PASS; `go list -m github.com/NVIDIA/OpenShell/sdk/go` prints a `v0.0.0-...-5a7728c304b5` pseudo-version.

- [ ] **Step 4: Makefile + docs/sdk-pinning.md** — targets `build`, `test` (`go test -race ./...`), `lint` (`golangci-lint run`), `run` (`go run ./cmd/openshell-proxy --config configs/example.yaml`), `openapi` (`go run ./cmd/openshell-proxy --print-openapi > openapi.json`), `update-sdk` (`go get github.com/NVIDIA/OpenShell/sdk/go@$(SDK_REF) && go mod tidy && go test ./...`). `docs/sdk-pinning.md` explains: no module tags, pin to the gateway release commit, re-pin procedure, known signature change after v0.1.4.

- [ ] **Step 5: Commit** — `git add -A && git commit -m "chore: bootstrap module, pin OpenShell SDK to v0.1.4 commit"`

---

### Task 2: Configuration loading

**Files:**
- Create: `internal/config/config.go`, `internal/config/load.go`, `internal/config/validate.go`, `configs/example.yaml`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces:
  ```go
  type Config struct { Server ServerConfig; Gateway GatewayConfig; Auth AuthConfig; Timeouts TimeoutsConfig; Limits LimitsConfig; CORS CORSConfig; Log LogConfig }
  type ServerConfig struct { Addr string; TLS TLSFiles; ReadHeaderTimeout, IdleTimeout time.Duration }
  type TLSFiles struct { CertFile, KeyFile string }
  type GatewayConfig struct { Address string; CAFile string; InsecureSkipVerify bool; AllowPlaintext bool }
  type AuthConfig struct { Mode Mode; LocalSanityCheck bool; AllowQueryToken bool }   // Mode = "bearer-passthrough" | "none"
  type TimeoutsConfig struct { Request, Exec, WaitMax, Stream, SSEHeartbeat time.Duration }
  type LimitsConfig struct { MaxBody, MaxExecBody, WSMaxFrame int64 }   // yaml accepts "1MiB"/"4MiB"/bytes
  type CORSConfig struct { AllowedOrigins []string; AllowCredentials bool }
  type LogConfig struct { Level, Format string }
  func Default() Config
  func Load(path string, getenv func(string) string) (*Config, error)   // defaults → yaml (path may be "") → env → Validate
  func (c *Config) Validate() error
  func (g GatewayConfig) DialAddress() string   // "http://"+Address when AllowPlaintext && no scheme; else Address stripped of "https://"
  ```
- Defaults: `:8080`, `10s`, `120s`; `bearer-passthrough`, sanity `true`, query token `false`; `30s, 10m, 10m, 1h, 15s`; `1MiB, 4MiB, 1MiB`; log `info`/`json`.

- [ ] **Step 1: Write failing tests**

```go
func TestLoad_DefaultsWhenNoFile(t *testing.T)      // Load("", noEnv) == Default(); Validate fails only on empty gateway address
func TestLoad_YAMLOverridesDefaults(t *testing.T)   // temp yaml sets gateway.address, timeouts.request: 45s, limits.maxBody: 2MiB → 2097152
func TestLoad_EnvOverridesYAML(t *testing.T)        // OSPROXY_GATEWAY_ADDRESS, OSPROXY_AUTH_MODE=none, OSPROXY_TIMEOUTS_REQUEST=5s win over yaml
func TestValidate_PlaintextRules(t *testing.T)      // allowPlaintext:false + address "http://x" → error; allowPlaintext:true + caFile set → error; mode bearer-passthrough + allowPlaintext → error "auth.mode must be none"
func TestValidate_ModeAndTLSFiles(t *testing.T)     // unknown mode → error; certFile without keyFile → error
func TestDialAddress(t *testing.T)                  // ("gw:443",false)→"gw:443"; ("https://gw:443",false)→"gw:443"; ("127.0.0.1:17670",true)→"http://127.0.0.1:17670"
```

- [ ] **Step 2: Run `go test ./internal/config/ -v`** — Expected: FAIL (package missing).
- [ ] **Step 3: Implement** `config.go` (types + `Default`), `load.go` (`Load`: yaml.v3 decode with `KnownFields(true)`; env walk over an explicit table of `(envName, setter)` pairs; byte-size parser accepting `KiB|MiB|GiB` suffixes and plain integers), `validate.go`, `configs/example.yaml` (all keys with comments).
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(config): YAML config with env overrides and validation"`

---

### Task 3: Problem responses and SDK error mapping

**Files:**
- Create: `internal/httperr/problem.go`, `internal/httperr/map.go`, `internal/httperr/install.go`
- Test: `internal/httperr/map_test.go`, `internal/httperr/problem_test.go`

**Interfaces:**
- Produces:
  ```go
  type Problem struct {
      Type string `json:"type,omitempty"`; Title string `json:"title"`; Status int `json:"status"`
      Detail string `json:"detail,omitempty"`; Instance string `json:"instance,omitempty"`
      Errors []*huma.ErrorDetail `json:"errors,omitempty"`
      Reason string `json:"reason,omitempty"`; Domain string `json:"domain,omitempty"`; Metadata map[string]string `json:"metadata,omitempty"`
      GRPCCode string `json:"grpcCode,omitempty"`; RetryAfterSeconds *int `json:"retryAfterSeconds,omitempty"`
      OutcomeUnknown bool `json:"outcomeUnknown,omitempty"`; RequestID string `json:"requestId,omitempty"`
  }
  func (p *Problem) Error() string; func (p *Problem) GetStatus() int; func (p *Problem) GetHeaders() http.Header  // huma.StatusError + huma.HeadersError
  func New(status int, detail string) *Problem
  type Options struct { IfMatch bool }
  func FromSDK(err error, opts Options) *Problem        // table in spec; nil → nil
  func Write(w http.ResponseWriter, r *http.Request, p *Problem)   // sets Content-Type application/problem+json + GetHeaders()
  func Install()   // sets huma.NewError to build *Problem (status, message, errs → Errors via huma.ErrorDetail when err implements huma.ErrorDetailer)
  ```
- Mapping facts: switch on `StatusError.GRPCCode` (`codes.Code`) first; `GRPCCode==0` → by `StatusError.Code`; terminal-phase rule: `Code==ErrorInternal && GRPCCode==0 && Cause==nil` → 409 with `Reason="SANDBOX_TERMINAL_PHASE"`; `RetryDelay` → `Retry-After` header (ceil seconds, min 1) and `RetryAfterSeconds`; 401 adds `WWW-Authenticate: Bearer`; `FieldViolations` → `Errors[i] = {Location: "body."+Field, Message: Description}`; `Title` = `http.StatusText(status)`; `GRPCCode` string via `codes.Code.String()`.

- [ ] **Step 1: Write failing table test**

```go
func TestFromSDK_Table(t *testing.T) // cases: {name, in error, ifMatch, wantStatus, wantReason, wantRetryAfter bool, wantOutcomeUnknown bool}
// INVALID_ARGUMENT+FieldViolations→400 with Errors[0].Location=="body.spec.image"
// UNAUTHENTICATED→401 header WWW-Authenticate; PERMISSION_DENIED→403; NOT_FOUND→404; ALREADY_EXISTS→409
// ABORTED→409; ABORTED+ifMatch→412; FAILED_PRECONDITION reason REQUEST_OUTCOME_UNCERTAIN→409; FAILED_PRECONDITION other→422; FAILED_PRECONDITION+ifMatch→412
// RESOURCE_EXHAUSTED RetryDelay 1500ms→429 Retry-After "2"; UNAVAILABLE→503 Retry-After "1" (default); DEADLINE_EXCEEDED→504 outcomeUnknown; CANCELLED→504 outcomeUnknown
// UNIMPLEMENTED→501; INTERNAL→502; errors.New("boom")→502; terminal-phase synthesized→409 SANDBOX_TERMINAL_PHASE; GRPCCode 0 + Code ErrorNotFound→404
func TestProblem_JSONShape(t *testing.T)   // omits empty extension fields; always has title/status
func TestInstall_HumaValidationUsesProblem(t *testing.T) // huma.Error422UnprocessableEntity("x", &huma.ErrorDetail{...}) yields *Problem with Errors populated
```
Build SDK errors with `&v1.StatusError{Code: ..., GRPCCode: int32(codes.X), ErrorInfo: &v1.ErrorInfo{Reason: ...}, RetryDelay: &d, FieldViolations: ...}` (`v1.StatusError` is an alias of `types.StatusError`).

- [ ] **Step 2: Run `go test ./internal/httperr/ -v`** — Expected: FAIL.
- [ ] **Step 3: Implement** the three files per the Interfaces block.
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(httperr): RFC 9457 problems and gRPC→HTTP mapping"`

---

### Task 4: stdlib middlewares

**Files:**
- Create: `internal/middleware/requestid.go`, `recover.go`, `logging.go`, `bodylimit.go`, `cors.go`, `chain.go`
- Test: `internal/middleware/middleware_test.go`

**Interfaces:**
- Produces:
  ```go
  func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler   // first mw outermost
  func RequestID(next http.Handler) http.Handler       // reads X-Request-ID (≤128 chars, [A-Za-z0-9._-]) else generates 16-byte hex; sets response header; stores in ctx
  func RequestIDFrom(ctx context.Context) string
  func Recover(log *slog.Logger) func(http.Handler) http.Handler   // panic → httperr.Write 500 problem, logs stack
  func Logging(log *slog.Logger) func(http.Handler) http.Handler   // one slog.Info per request: method, path (query stripped), status, bytes, durationMs, requestId; never headers
  func BodyLimit(n int64) func(http.Handler) http.Handler          // http.MaxBytesReader; also rejects Content-Length > n up front with 413 problem
  func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler // rs/cors; no-op when no origins configured
  ```

- [ ] **Step 1: Write failing tests**

```go
func TestRequestID_GeneratedAndEchoed(t *testing.T)   // no header → 32-hex response header and ctx value; provided valid header → echoed; invalid chars → replaced
func TestRecover_WritesProblem(t *testing.T)          // handler panics → 500, Content-Type application/problem+json, body.status==500
func TestLogging_RedactsQueryAndHeaders(t *testing.T) // buffer handler: log line contains path "/v1/x", not "access_token=", not "Authorization"
func TestBodyLimit_Returns413(t *testing.T)           // 2 KiB body with limit 1 KiB → 413 problem; next request on same client works (Review Focus 2)
func TestChain_Order(t *testing.T)                    // records order of 3 middlewares: outer→inner
```

- [ ] **Step 2: Run tests** — Expected: FAIL.
- [ ] **Step 3: Implement.** Body-limit 413 is produced by a small wrapper that checks `Content-Length` and otherwise relies on `MaxBytesReader`; handlers that hit the reader error get a `*http.MaxBytesError` which Task 8's huma error hook maps to 413.
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(middleware): request id, recover, logging, body limit, cors"`

---

### Task 5: Bearer pass-through auth middleware

**Files:**
- Create: `internal/auth/context.go`, `internal/auth/bearer.go`, `internal/auth/sanity.go`, `internal/auth/middleware.go`
- Test: `internal/auth/auth_test.go`

**Interfaces:**
- Consumes: `httperr.Write`, `config.AuthConfig`.
- Produces:
  ```go
  func WithToken(ctx context.Context, token string) context.Context
  func TokenFrom(ctx context.Context) (string, bool)
  func ParseBearer(header string) (token string, ok bool)   // scheme case-insensitive, exactly one token, no empty
  var ErrMalformed, ErrExpired error
  func SanityCheck(token string, now time.Time, skew time.Duration) error  // 3 dot-separated segments, base64url payload decodes to JSON; if "exp" present and exp+skew < now → ErrExpired
  type Options struct { Mode config.Mode; LocalSanityCheck bool; AllowQueryToken bool; Exempt []string; Now func() time.Time }
  func Middleware(o Options) func(http.Handler) http.Handler
  ```
- Behavior: exempt exact paths `/healthz`, `/readyz`, `/openapi.json`, `/openapi.yaml`, `/docs`, `/v1/health` (plus any `Exempt`). Mode `none`: forward token if present, never reject. Mode `bearer-passthrough`: missing/malformed → 401 `TOKEN_MISSING`/`TOKEN_MALFORMED`; sanity failure → 401 `TOKEN_EXPIRED`/`TOKEN_MALFORMED`; `?access_token=` accepted only when `AllowQueryToken`. Problems carry `WWW-Authenticate: Bearer`.

- [ ] **Step 1: Write failing tests**

```go
func TestParseBearer(t *testing.T)            // "Bearer abc"→abc; "bearer abc"→abc; "Bearer  abc"→abc; "Basic abc"→!ok; "Bearer"→!ok; "Bearer a b"→!ok  (Review Focus 1)
func TestSanityCheck(t *testing.T)            // valid unsigned-shape JWT ok; exp in past → ErrExpired; exp within 60s skew ok; "abc" → ErrMalformed
func TestMiddleware_PassthroughRequiresToken(t *testing.T)   // missing → 401 problem reason TOKEN_MISSING + WWW-Authenticate; present → next sees TokenFrom(ctx)==token
func TestMiddleware_ExemptPaths(t *testing.T) // /healthz without token → 200
func TestMiddleware_ModeNone(t *testing.T)    // no token → 200; token → still forwarded
func TestMiddleware_QueryToken(t *testing.T)  // ?access_token=x → 401 by default; 200 when AllowQueryToken
```

- [ ] **Step 2: Run tests** — Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(auth): bearer pass-through middleware with local sanity check"`

---

### Task 6: Gateway client factory and ctx-aware AuthProvider

**Files:**
- Create: `internal/gateway/provider.go`, `internal/gateway/gateway.go`
- Test: `internal/gateway/provider_test.go`, `internal/gateway/gateway_test.go`

**Interfaces:**
- Consumes: `auth.TokenFrom`, `middleware.RequestIDFrom`, `config.GatewayConfig`.
- Produces:
  ```go
  type Provider struct{ requireTLS bool }
  func NewProvider(requireTLS bool) *Provider
  func (p *Provider) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error)  // {"authorization":"Bearer <tok>"} if token; {"x-request-id": id} if id; never errors
  func (p *Provider) RequireTransportSecurity() bool
  type Gateway struct { /* shared *v1.Client, cfg */ }
  func New(cfg config.GatewayConfig) (*Gateway, error)                  // v1.NewClient(v1.Config{Address: cfg.DialAddress(), TLS: tlsFor(cfg), Auth: NewProvider(!cfg.AllowPlaintext)})
  func (g *Gateway) Client() v1.ClientInterface
  func (g *Gateway) Dedicated(ctx context.Context) (v1.ClientInterface, func(), error)  // new v1.Client with v1.StaticToken(token) (or NoAuth when no token and plaintext); closer func
  func (g *Gateway) Close() error
  ```
- `tlsFor`: nil when `AllowPlaintext`; else `&v1.TLSConfig{CAFile: cfg.CAFile, Insecure: cfg.InsecureSkipVerify}`.

- [ ] **Step 1: Write failing tests**

```go
func TestProvider_MetadataWithToken(t *testing.T)      // ctx with token+requestId → both keys; without → empty map, nil error
func TestProvider_RequireTransportSecurity(t *testing.T)
func TestNew_PlaintextRequiresNoAuthTLS(t *testing.T)   // cfg{Address:"127.0.0.1:1", AllowPlaintext:true} → ok; cfg{Address:"http://127.0.0.1:1", AllowPlaintext:false} → error (SDK refuses: provider requires TLS)
func TestDedicated_UsesCallerToken(t *testing.T)        // returns client and closer; closer closes without error
```

- [ ] **Step 2: Run tests** — Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(gateway): shared SDK client with per-request bearer injection"`

---

### Task 7: Test double: `gatewaytest`

**Files:**
- Create: `internal/gateway/gatewaytest/client.go`, `internal/gateway/gatewaytest/stubs.go`, `internal/gateway/gatewaytest/errors.go`
- Test: `internal/gateway/gatewaytest/client_test.go`

**Interfaces:**
- Produces:
  ```go
  type Client struct { *fake.Client /* overrides below */ }
  type Option func(*Client)
  func New(opts ...Option) *Client                          // wraps fake.NewClient(fakeOpts...)
  func WithFakeOptions(opts ...fake.ClientOption) Option    // e.g. fake.WithCurrentUser
  func WithExec(v1.ExecInterface) Option; WithServices(v1.ServiceInterface); WithConfig(v1.ConfigInterface); WithPolicy(v1.PolicyInterface); WithProfiles(v1.ProfileInterface); WithRefresh(v1.RefreshInterface); WithSSH(v1.SSHInterface); WithTCP(v1.TCPInterface)
  // Client overrides Exec(), Services(), Config(), Policy(), SSH(), TCP(); Providers() returns a wrapper that delegates CRUD to fake but Profiles()/Refresh() to the injected stubs.
  // stubs.go — function-field stubs; unset fields return gatewaytest.ErrUnimplemented (a *v1.StatusError{Code: ErrorUnimplemented})
  type ServiceStub struct { ExposeFn func(ctx, ws, sb, name string, port uint32, domain bool, opts ...v1.ExposeServiceOptions) (*v1.ServiceEndpoint, error); GetFn; ListFn func(ws, sb string, opts ...v1.ListOptions) (*v1.Pager[*v1.ServiceEndpoint], error); ListAllFn; DeleteFn }
  type ConfigStub struct { GetSandboxFn; GetGatewayFn; UpdateFn }
  type PolicyStub struct { GetDraftFn; ApproveDraftChunkFn; RejectDraftChunkFn; ApproveAllDraftChunksFn; ClearDraftChunksFn; GetDraftHistoryFn; GetStatusFn; ListFn; ListAllFn; EditDraftChunkFn; UndoDraftChunkFn }
  type ProfileStub struct { ListFn; ListAllFn; GetFn; ImportFn; UpdateFn; LintFn; DeleteFn }
  type RefreshStub struct { GetStatusFn; ConfigureFn; RotateFn; DeleteFn }
  type ExecStub struct { RunFn; StreamFn; InteractiveFn }
  type SSHStub struct { CreateSessionFn; RevokeSessionFn; TunnelFn }; type TCPStub struct { ForwardFn; ListenFn }
  // errors.go
  func StatusErr(code codes.Code, msg string) *v1.StatusError     // fills Code via the SDK's fold rules and GRPCCode
  func StaticPager[T any](items []T, next string) *v1.Pager[T]     // v1.NewPager with a one-shot fetch
  ```

- [ ] **Step 1: Write failing tests**

```go
func TestNew_DelegatesToFake(t *testing.T)        // AddWorkspace then Workspaces().Get works through gatewaytest.Client
func TestWithServices_Overrides(t *testing.T)     // ServiceStub.GetFn called; unset ListFn returns ErrUnimplemented
func TestProviders_WrapperInjectsProfiles(t *testing.T)  // Providers().Profiles() is the stub; Providers().Get hits fake
func TestStaticPager(t *testing.T)                // NextPage returns items + token, second call nil
```

- [ ] **Step 2: Run tests** — Expected: FAIL.
- [ ] **Step 3: Implement.** Stub signatures must match the v0.1.4 interfaces exactly (copy from `openshell/v1/{service,config,policy,profile,refresh,exec,ssh,tcp}.go`).
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "test(gatewaytest): fake-backed client with overridable sub-interfaces"`

---

### Task 8: huma API skeleton, health, whoami, gateway info

**Files:**
- Create: `internal/api/api.go`, `internal/api/params.go`, `internal/api/health.go`, `internal/dto/health.go`, `internal/dto/convert/health.go`
- Test: `internal/api/health_test.go`, `internal/api/apitest_test.go` (helper)

**Interfaces:**
- Consumes: `httperr.Install/FromSDK`, `auth.Middleware`, `middleware.*`, `config.Config`.
- Produces:
  ```go
  type Deps struct { Client v1.ClientInterface; Dedicated func(ctx context.Context) (v1.ClientInterface, func(), error); Cfg *config.Config; Log *slog.Logger }
  func New(d Deps) http.Handler      // mux with: huma API at "/" (humago.New), security scheme "bearer" (http/bearer/JWT), /healthz, /readyz; wrapped by Chain(RequestID, Recover, Logging, CORS, BodyLimit, auth.Middleware)
  func OpenAPI(d Deps) ([]byte, error)   // JSON of the registered API (for --print-openapi)
  // params.go
  type WorkspacePath struct { Workspace string `path:"workspace" minLength:"1"` }
  type SandboxPath struct { WorkspacePath; Sandbox string `path:"sandbox" minLength:"1"` }
  type ListQuery struct { PageSize int `query:"pageSize" minimum:"0" maximum:"1000"`; PageToken string `query:"pageToken"`; LabelSelector string `query:"labelSelector"`; AllWorkspaces bool `query:"allWorkspaces"` }
  func (q ListQuery) SDK() v1.ListOptions
  type AllowMissingQuery struct { AllowMissing bool `query:"allowMissing"` }
  func (q AllowMissingQuery) SDK() v1.DeleteOptions
  type IfMatchHeader struct { IfMatch string `header:"If-Match"` }
  func (h IfMatchHeader) Version() (v uint64, set bool, err error)   // `"7"`→7; `7`→7; `W/"7"` or `*` or non-numeric → error (400)
  func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc)
  func sdkErr(err error, ifMatch bool) error   // httperr.FromSDK → returned to huma
  // dto/health.go
  type Health struct { Healthy bool `json:"healthy"`; Version string `json:"version,omitempty"` }
  type GatewayInfo struct { Status string; Version string; ComputeDrivers []ComputeDriverInfo; Extensions []ExtensionInfo }  (json lowerCamel)
  type CurrentUser struct { Subject, DisplayName, IdentityProvider string; Roles, Scopes []string }
  ```
- Routes: `GET /v1/health` → `Health().Check` (no security), `GET /v1/whoami` → `GetCurrentUser`, `GET /v1/gateway/info` → `GetGatewayInfo`; `GET /readyz` → `Check` and 503 problem when unhealthy/error; `GET /healthz` → `{"status":"ok"}`.
- Every operation sets `Security: []map[string][]string{{"bearer": {}}}` except health/readyz/healthz, and `Errors: []int{400,401,403,404,409,412,422,429,500,502,503,504}` via a shared helper `op(id, method, path, summary string) huma.Operation`.

- [ ] **Step 1: Write failing tests** (helper `newTestServer(t, gatewaytest opts..., cfg mutators...) *httptest.Server` with `auth.Mode=bearer-passthrough` and a fixed test token)

```go
func TestHealth_NoTokenNeeded(t *testing.T)        // GET /v1/health → 200 {"healthy":true,"version":"fake"}
func TestReadyz_UnhealthyIs503(t *testing.T)       // fake.WithHealthResult(&v1.HealthResult{Healthy:false}) → 503 problem
func TestWhoami_RequiresToken(t *testing.T)        // no header → 401 problem; with header → 200 subject from fake.WithCurrentUser
func TestGatewayInfo_MapsFields(t *testing.T)      // fake.WithGatewayInfo → status/version/computeDrivers[0].name
func TestOpenAPI_HasBearerScheme(t *testing.T)     // OpenAPI JSON contains components.securitySchemes.bearer and /v1/whoami security
func TestValidationErrorIsProblem(t *testing.T)    // GET /v1/workspaces?pageSize=5000 (after Task 9; here use a registered probe) → 422 application/problem+json
```

- [ ] **Step 2: Run tests** — Expected: FAIL.
- [ ] **Step 3: Implement.** `humago.New(mux, cfg)` with `huma.DefaultConfig("OpenShell REST Proxy", version.Version)` plus `cfg.Components.SecuritySchemes["bearer"] = &huma.SecurityScheme{Type:"http", Scheme:"bearer", BearerFormat:"JWT"}`; call `httperr.Install()` once in `New`.
- [ ] **Step 4: Run tests** — Expected: PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(api): huma skeleton with health, readiness, whoami and gateway info"`

---

### Task 9: Server, main, end-to-end smoke

**Files:**
- Create: `internal/server/server.go`, `internal/version/version.go`, `cmd/openshell-proxy/main.go` (replace placeholder), `docs/configuration.md`
- Test: `internal/server/server_test.go`

**Interfaces:**
- Produces:
  ```go
  type Server struct{ /* http.Server */ }
  func New(cfg config.ServerConfig, h http.Handler, log *slog.Logger) *Server   // ReadHeaderTimeout, IdleTimeout from cfg; WriteTimeout 0
  func (s *Server) Run(ctx context.Context) error   // ListenAndServe or ListenAndServeTLS; on ctx.Done → Shutdown(20s)
  func (s *Server) Addr() string                    // actual bound address (for ":0" in tests)
  // main: flags --config <path> (default "" → env only), --print-openapi; builds slog from cfg.Log; gateway.New; api.New; signal.NotifyContext(SIGINT, SIGTERM); exit code 1 on error.
  ```

- [ ] **Step 1: Write failing test**

```go
func TestServer_RunAndGracefulShutdown(t *testing.T)   // Addr ":0", handler returns 200; cancel ctx → Run returns nil within 2s
```

- [ ] **Step 2: Run test** — Expected: FAIL.
- [ ] **Step 3: Implement** server, version (`var Version = "dev"`, ldflags-settable), main, `docs/configuration.md` (every key, default, env name).
- [ ] **Step 4: Run tests and smoke** — `go test ./... -race` PASS; `OSPROXY_GATEWAY_ADDRESS=127.0.0.1:1 OSPROXY_AUTH_MODE=none OSPROXY_GATEWAY_ALLOW_PLAINTEXT=true go run ./cmd/openshell-proxy` then `curl -s localhost:8080/healthz` → `{"status":"ok"}`; `curl -si localhost:8080/v1/whoami` → 503 problem with `"grpcCode":"Unavailable"` (Review Focus 5); `go run ./cmd/openshell-proxy --print-openapi | head`.
- [ ] **Step 5: Commit** — `git commit -m "feat: server lifecycle and CLI entrypoint"`

---

### Task 10: Readiness/unavailable behavior test (Review Focus 5)

**Files:**
- Test: `internal/api/unavailable_test.go`

- [ ] **Step 1: Write test** `TestUnavailableGateway_Problems`: build `api.New` with a real `gateway.New(config.GatewayConfig{Address:"127.0.0.1:1", AllowPlaintext:true})` and `auth.Mode=none`; assert `/healthz` 200, `/readyz` 503 problem, `/v1/whoami` 503 problem `grpcCode:"Unavailable"`, all within `timeouts.request=2s`.
- [ ] **Step 2: Run** — Expected: PASS (if it fails, fix `withTimeout` application in Task 8 handlers).
- [ ] **Step 3: Commit** — `git commit -m "test(api): unavailable gateway surfaces 503 problems"`

---

### Task 11: Shared DTO primitives and converters

**Files:**
- Create: `internal/dto/common.go`, `internal/dto/convert/common.go`
- Test: `internal/dto/common_test.go`, `internal/api/params_test.go`

**Interfaces:**
- Produces:
  ```go
  type Uint64String uint64   // MarshalJSON → "123"; UnmarshalJSON accepts 123 or "123"
  type ObjectMeta struct { ID string `json:"id,omitempty"`; Name string `json:"name"`; Workspace string `json:"workspace,omitempty"`; CreatedAt *time.Time `json:"createdAt,omitempty"`; DeletionTimestamp *time.Time `json:"deletionTimestamp,omitempty"`; Labels map[string]string `json:"labels,omitempty"`; Annotations map[string]string `json:"annotations,omitempty"`; ResourceVersion Uint64String `json:"resourceVersion,omitempty"` }
  type Page[T any] struct { Items []T `json:"items"`; NextPageToken string `json:"nextPageToken,omitempty"` }
  type DeletionResult struct { Outcome string `json:"outcome"`; SandboxID string `json:"sandboxId,omitempty"` }   // "Completed"|"Accepted"|"AlreadyAbsent"|"Unspecified"
  // convert/common.go
  func TimePtr(t time.Time) *time.Time   // nil when zero
  func Deletion(r *v1.DeletionResult) dto.DeletionResult
  func DeletionStatus(r *v1.DeletionResult) int   // 202 for Accepted, else 200
  ```

- [ ] **Step 1: Write failing tests**

```go
func TestUint64String_RoundTrip(t *testing.T)   // 18446744073709551615 marshals as string; unmarshals from 7 and "7"; rejects "x"
func TestIfMatchHeader_Version(t *testing.T)    // `"7"`→7; `7`→7; ``→unset; `W/"7"`→error; `*`→error; `"abc"`→error  (Review Focus 3)
func TestDeletion_Mapping(t *testing.T)         // DeletionAccepted→"Accepted",202; DeletionAlreadyAbsent→200
```

- [ ] **Step 2: Run** — Expected: FAIL. **Step 3: Implement.** **Step 4: Run** — PASS.
- [ ] **Step 5: Commit** — `git commit -m "feat(dto): shared JSON primitives and deletion mapping"`

---

### Task 12: Generic list helper

**Files:**
- Create: `internal/api/list.go`
- Test: `internal/api/list_test.go`

**Interfaces:**
- Produces: `func listPage[T, D any](ctx context.Context, pager *v1.Pager[T], conv func(T) D) (*dto.Page[D], error)` — exactly one `NextPage(ctx)`; nil page → empty `Items` (not null) and no token.

- [ ] **Step 1: Test** `TestListPage_SingleFetch` with `gatewaytest.StaticPager([]*v1.Workspace{...}, "tok2")` → 2 items, `NextPageToken=="tok2"`; `TestListPage_EmptyIsEmptyArray` JSON `"items":[]`.
- [ ] **Step 2–4:** fail → implement → pass.
- [ ] **Step 5: Commit** — `git commit -m "feat(api): generic single-page list helper"`

---

### Task 13: Workspaces and members

**Files:**
- Create: `internal/dto/workspace.go`, `internal/dto/convert/workspace.go`, `internal/api/workspaces.go`
- Test: `internal/dto/convert/workspace_test.go`, `internal/api/workspaces_test.go`

**Interfaces:**
- Produces:
  ```go
  type Workspace struct { ObjectMeta; Phase string `json:"phase"` }
  type WorkspaceCreate struct { Name string `json:"name" minLength:"1"`; Labels map[string]string `json:"labels,omitempty"` }
  type WorkspaceMember struct { ObjectMeta; PrincipalSubject string `json:"principalSubject"`; Role string `json:"role"` }   // role "Admin"|"User"
  type WorkspaceMemberAdd struct { PrincipalSubject string `json:"principalSubject" minLength:"1"`; Role string `json:"role" enum:"Admin,User"` }
  func WorkspaceFromSDK(*v1.Workspace) dto.Workspace; MemberFromSDK(*v1.WorkspaceMember) dto.WorkspaceMember; RoleToSDK(string) v1.WorkspaceRole
  ```
- Routes (spec): `POST /v1/workspaces` 201 · `GET /v1/workspaces` · `GET|DELETE /v1/workspaces/{workspace}` · `GET|POST /v1/workspaces/{workspace}/members` (201) · `DELETE /v1/workspaces/{workspace}/members/{subject}`.

- [ ] **Step 1: Failing tests**

```go
func TestWorkspaceFromSDK_Golden(t *testing.T)          // full struct → golden JSON in testdata/workspace.json
func TestWorkspaces_CRUD(t *testing.T)                   // POST 201 then GET 200 then list items==1 then DELETE 200 outcome Completed; GET again → 404 problem
func TestMembers_AddListRemove(t *testing.T)            // POST member role User → 201; list contains; DELETE 200
func TestMembers_SubjectPathIsDecoded(t *testing.T)     // subject "oidc|abc/def" encoded as oidc%7Cabc%2Fdef in path → fake receives decoded subject  (Review Focus 4)
func TestWorkspaceDelete_AllowMissing(t *testing.T)     // DELETE unknown → 404; ?allowMissing=true → 200 AlreadyAbsent (fake honors DeleteOptions)
```

- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): workspaces and membership routes"`

---

### Task 14: Sandbox DTOs and converters

**Files:**
- Create: `internal/dto/sandbox.go`, `internal/dto/policy_types.go` (SandboxPolicy + network-policy graph, shared with Task 18), `internal/dto/convert/sandbox.go`, `internal/dto/convert/policy_types.go`
- Test: `internal/dto/convert/sandbox_test.go`, `internal/dto/convert/policy_types_test.go`, `testdata/sandbox.json`, `testdata/policy.json`

**Interfaces:**
- Produces (field names = lowerCamel of the v0.1.4 SDK fields; enums as strings):
  ```go
  type Sandbox struct { ObjectMeta; HostKeyFingerprint string; CreatedFromWorkloadTemplate *TemplateProvenance; Spec SandboxSpec; Status SandboxStatus; ServiceURLs map[string]string }
  type SandboxSpec struct { LogLevel string; Environment map[string]string; Template *SandboxTemplate; Providers []string; GPU bool; GPUCount *uint32 `json:"gpuCount"`; Policy *SandboxPolicy; Command []string; TTY bool `json:"tty"`; RestartPolicy string }
  type SandboxTemplate struct { Image, RuntimeClassName, AgentSocket string; Labels, Annotations, Environment map[string]string; UserNamespaces *bool; Resources, DriverConfig map[string]any }
  type SandboxStatus struct { AgentPod, AgentFd, SandboxFd string; Phase string; Conditions []SandboxCondition; CurrentPolicyVersion uint32; ExitCode *int32; EndpointStatuses []EndpointStatus; ConfigurationAdmission *ConfigurationAdmission; RestartCount uint32; NextRestartAtMs, MainProcessStartedAtMs int64 }
  type SandboxPolicy struct { Version uint32; Filesystem *FilesystemPolicy; Landlock *LandlockPolicy; Process *ProcessPolicy; NetworkPolicies map[string]NetworkPolicyRule; NetworkMiddlewares map[string]NetworkMiddlewareConfig }
  // NetworkPolicyRule, PolicyNetworkEndpoint (all 26 fields; int32 enums as strings: tls "Skip|Terminate|Passthrough", enforcement "Enforce|Audit", access "ReadOnly|ReadWrite|Full", "" for Unspecified; JSONRPCMaxBodyBytes→jsonrpcMaxBodyBytes; Mcp→mcp), L7Rule, L7Allow, L7DenyRule, L7QueryMatcher, McpOptions, GraphqlOperation, PolicyMergeOperation and its 6 operation structs, L7RuleTarget
  func SandboxFromSDK(*v1.Sandbox) dto.Sandbox; SandboxSpecToSDK(dto.SandboxSpec) *v1.SandboxSpec; PolicyFromSDK(*v1.SandboxPolicy) *dto.SandboxPolicy; PolicyToSDK(*dto.SandboxPolicy) *v1.SandboxPolicy; MergeOpsToSDK([]dto.PolicyMergeOperation) []v1.PolicyMergeOperation; NetworkRuleToSDK(dto.NetworkPolicyRule) *v1.NetworkPolicyRule
  ```

- [ ] **Step 1: Failing tests** — `TestSandboxFromSDK_Golden` (maximal fixture → `testdata/sandbox.json`); `TestSandboxSpec_RoundTrip` (`SandboxSpecToSDK(SpecFromSDK(x)) == x` via `assert.Equal`); `TestPolicy_RoundTrip` on a maximal policy with every endpoint field, L7 rules, middlewares, and merge operations; `TestEnumStrings` for tls/enforcement/access/restartPolicy/phase including unknown → "" or passthrough.
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(dto): sandbox and policy converters"`

---

### Task 15: Sandbox routes

**Files:**
- Create: `internal/api/sandboxes.go`, `internal/dto/sandbox_requests.go`, `internal/dto/log.go`, `internal/dto/convert/log.go`
- Test: `internal/api/sandboxes_test.go`

**Interfaces:**
- Produces:
  ```go
  type SandboxCreate struct { Name string `json:"name" minLength:"1"`; Spec SandboxSpec `json:"spec"`; Labels, Annotations map[string]string; ServiceExposures []ServiceExposure; Template string `json:"template,omitempty"` }
  type ServiceExposure struct { Service string; TargetPort uint32; AuthorizationMode string `enum:"Strip,BearerPassthrough"` }
  type WaitRequest struct { For string `json:"for" enum:"ready,stopped"`; Timeout string `json:"timeout,omitempty"`; PollInterval string `json:"pollInterval,omitempty"` }   // Go durations
  type LogLine struct { Timestamp time.Time; Level, Target, Message, Source string; Fields map[string]string }; type LogResult struct { Lines []LogLine; BufferTotal uint32 }
  ```
- Routes: `POST /v1/workspaces/{workspace}/sandboxes` (201; `Template!=""` → `Client.CreateSandboxFromTemplate(ctx, ws, name, template, spec, labels, v1.CreateOptions{Annotations, ServiceExposures})` else `Sandboxes().Create(...)`) · `GET` list · `GET|DELETE .../{sandbox}` · `POST .../{sandbox}/start` and `POST .../{sandbox}/stop` with `?wait=true&timeout=` · `POST .../{sandbox}/wait` · `GET .../{sandbox}/logs?lines&since&sources&minLevel` · `GET .../{sandbox}/providers` · `PUT|DELETE .../{sandbox}/providers/{provider}` (If-Match → `expectedResourceVersion`; absent → 0). Wait timeout = min(requested, `timeouts.waitMax`); default `timeouts.waitMax`; `/start`, `/stop`, `/wait` are exempt from the request-timeout middleware (operation `Metadata["longRunning"]=true`).
- Path style: Go's 1.22 mux rejects `{sandbox}:start` ("bad wildcard segment", verified), so all action verbs are sub-path segments (`/start`, `/stop`, `/wait`, `/approve`, `/import`, ...). A sandbox literally named `start` is still reachable because `/sandboxes/{sandbox}` and `/sandboxes/{sandbox}/start` are different patterns.

- [ ] **Step 1: Failing tests**

```go
func TestSandboxes_CreateGetList(t *testing.T)        // POST spec.template.image "python:3.12" → 201 name; GET → spec.template.image; list 1
func TestSandboxes_CreateFromTemplate(t *testing.T)   // AddSandboxTemplate then POST with template → 201 createdFromWorkloadTemplate.name
func TestSandboxes_StopStartWait(t *testing.T)        // POST /stop?wait=true → phase "Stopped"; POST /start?wait=true → "Ready"; POST /wait {"for":"ready"} → 200
func TestSandboxes_WaitTerminalIs409(t *testing.T)    // AddSandbox with Phase Error; POST /wait → 409 reason SANDBOX_TERMINAL_PHASE
func TestSandboxes_WaitTimeoutCapped(t *testing.T)    // cfg waitMax=50ms, sandbox stays Provisioning → 504 problem outcomeUnknown within <1s
func TestSandboxes_NamedStartIsReachable(t *testing.T) // AddSandbox name "start"; GET .../sandboxes/start → 200 (pattern specificity)
func TestSandboxes_ProvidersAttachDetach(t *testing.T) // AddProvider; PUT .../providers/p If-Match "0" → 200 attached true; DELETE → detached true
func TestSandboxes_Logs(t *testing.T)                 // GET logs → 200 {"lines":[],"bufferTotal":0}
func TestSandboxes_DeleteAccepted202(t *testing.T)    // fake delete outcome → status per DeletionStatus
```

- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): sandbox lifecycle, wait, logs and provider attachment routes"`

---

### Task 16: Sandbox templates

**Files:**
- Create: `internal/dto/template.go`, `internal/dto/convert/template.go`, `internal/api/templates.go`
- Test: `internal/api/templates_test.go`, `internal/dto/convert/template_test.go`

**Interfaces:**
- Produces: `type SandboxWorkloadTemplate struct { ObjectMeta; Spec SandboxWorkloadTemplateSpec }`, `SandboxWorkloadTemplateSpec { Workload *SandboxWorkloadConfig; DriverConfig map[string]any; DesiredServiceLevel *SandboxServiceLevel }`, `SandboxWorkloadConfig { Image string; Environment map[string]string; Resources *SandboxResources }`, `SandboxResources { CPU string `json:"cpu"`; Memory string; GPU *SandboxGPURequirements `json:"gpu"` }`, `SandboxServiceLevel { Startup *SandboxStartup }`, `SandboxStartup { ReadyWithin string /* Go duration */; MaxBurst uint32 }`; converters both directions.
- Routes: `POST|GET /v1/workspaces/{workspace}/sandbox-templates`, `GET|DELETE .../{name}`.

- [ ] **Step 1: Tests** — `TestTemplate_RoundTrip` (incl. `ReadyWithin: "90s"`), `TestTemplates_CRUD` against fake (`AddSandboxTemplate`).
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): sandbox workload template routes"`

---

### Task 17: Exec (collected) route

**Files:**
- Create: `internal/dto/exec.go`, `internal/api/exec.go`
- Test: `internal/api/exec_test.go`

**Interfaces:**
- Produces: `type ExecRequest struct { Command []string `json:"command" minItems:"1"`; Env map[string]string; WorkDir string; NoLoginShell bool; Encoding string `json:"encoding,omitempty" enum:"utf8,base64"` }`, `type ExecResult struct { ExitCode int; Stdout, Stderr string; Encoding string }` (default `utf8`; invalid UTF-8 output auto-switches to `base64` and says so).
- Route: `POST /v1/workspaces/{workspace}/sandboxes/{sandbox}/exec` → `Exec().Run(ctx, ws, sb, cmd, v1.ExecOptions{Env, WorkDir, NoLoginShell})`; deadline `timeouts.exec` (exempt from request timeout); body limit `limits.maxExecBody`.

- [ ] **Step 1: Tests** — `TestExec_Run` (ExecStub.RunFn echoes; exitCode 3 → 200 with exitCode 3, not an error), `TestExec_Base64WhenBinary` (stdout `\xff\xfe` → encoding base64), `TestExec_EmptyCommand422`.
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): collected exec route"`

---

### Task 18: Services

**Files:**
- Create: `internal/dto/service.go`, `internal/dto/convert/service.go`, `internal/api/services.go`
- Test: `internal/api/services_test.go`

**Interfaces:**
- Produces: `type ServiceEndpoint struct { ID, SandboxID, Sandbox, Name string; TargetPort uint32; Domain bool; URL string `json:"url"`; Workspace string; AuthorizationMode string }`, `type ServiceExpose struct { Name string `minLength:"1"`; TargetPort uint32 `minimum:"1" maximum:"65535"`; Domain bool; AuthorizationMode string `enum:"Strip,BearerPassthrough"` }`.
- Routes: `POST|GET .../sandboxes/{sandbox}/services`, `GET|DELETE .../services/{name}` → `Services().Expose(ctx, ws, sb, name, port, domain, v1.ExposeServiceOptions{AuthorizationMode})` etc. (check `ExposeServiceOptions` fields in `openshell/v1/service.go` at the pin).

- [ ] **Step 1: Tests** with `gatewaytest.ServiceStub`: `TestServices_Expose201`, `TestServices_ListAndGet`, `TestServices_Delete`.
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): service exposure routes"`

---

### Task 19: Providers, profiles, refresh

**Files:**
- Create: `internal/dto/provider.go`, `internal/dto/profile.go`, `internal/dto/refresh.go`, `internal/dto/convert/{provider,profile,refresh}.go`, `internal/api/providers.go`, `internal/api/profiles.go`, `internal/api/refresh.go`
- Test: `internal/dto/convert/{provider,profile,refresh}_test.go`, `internal/api/{providers,profiles,refresh}_test.go`

**Interfaces:**
- Produces: DTO mirrors of `Provider/ProviderSpec/CredentialHandle`, `ProviderProfile` graph (`ProfileCredential`, `ProfileCredentialRefresh`, `CredentialTokenGrant`, `ProfileDuration{Seconds int64; Nanos int32}`, `NetworkEndpoint`, `NetworkBinary`, `ProfileDiscovery`, `ProfileFile`), `ProfileImportItem{Profile, Source}`, `ImportResult/UpdateResult/LintResult/ProfileDiagnostic`, `RefreshStatus` (`RecoveryAction` as the SDK `String()` value), `RefreshConfigRequest { Strategy string; Material map[string]string; SecretMaterialKeys []string; ExpiresAt *time.Time }`. `Provider`, `RefreshConfigRequest` implement `slog.LogValuer` returning only name/type/keys.
- Routes: providers `POST|GET /providers`, `GET|DELETE /providers/{name}`, `PUT /providers/{name}` (body name must equal path → else 400; `?ensure=true` switches from `Update` to `Ensure`, returning 200 either way); profiles `GET /provider-profiles`, `GET|PUT|DELETE /provider-profiles/{id}` (PUT needs `If-Match` → 428 problem when absent), `POST /provider-profiles/import`, `POST /provider-profiles/lint`; refresh `GET /providers/{p}/refresh?credentialKey=`, `PUT|DELETE /providers/{p}/refresh/{key}`, `POST /providers/{p}/refresh/{key}/rotate`.

- [ ] **Step 1: Tests** — converter round-trips (maximal profile), `TestProvider_LogValueRedacts` (slog JSON handler output lacks credential values), `TestProviders_CRUDAndEnsure` (fake), `TestProviders_UpdateNameMismatch400`, `TestProfiles_UpdateRequiresIfMatch428`, `TestProfiles_ImportLint` (ProfileStub), `TestRefresh_ConfigureRotateStatus` (RefreshStub; assert the test log buffer never contains the material value).
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): providers, provider profiles and credential refresh routes"`

---

### Task 20: Config routes

**Files:**
- Create: `internal/dto/config.go`, `internal/dto/convert/config.go`, `internal/api/configs.go`
- Test: `internal/api/configs_test.go`, `internal/dto/convert/config_test.go`

**Interfaces:**
- Produces: `SettingValue` tagged union `{ type: "string"|"bool"|"int"|"bytes", string?, bool?, int? (int64 as number), bytes? (base64) }`, `EffectiveSetting { Value SettingValue; Scope string }`, `SandboxConfig` (revisions as `Uint64String`), `GatewayConfig { Settings map[string]SettingValue; SettingsRevision Uint64String }`, `ConfigUpdateRequest { Policy *SandboxPolicy; SettingKey string; SettingValue *SettingValue; DeleteSetting bool; MergeOperations []PolicyMergeOperation; Annotations map[string]string; ExpectedResourceVersion *Uint64String }`, `ConfigUpdateResult`.
- Routes: `GET .../sandboxes/{sandbox}/config` → `Config().GetSandbox`; `PATCH .../sandboxes/{sandbox}/config` → `Config().Update(ctx, ws, &v1.ConfigUpdate{Name: sb, ..., ExpectedResourceVersion: fromIfMatchOrBody})`; `GET /v1/gateway/config` → `GetGateway`; `PATCH /v1/gateway/config?workspace=` → `Update(ctx, workspace /* "" default */, &v1.ConfigUpdate{Global: true, ...})`.

- [ ] **Step 1: Tests** — `TestSettingValue_Union` (all four kinds round-trip; missing `type` → 422), `TestConfig_GetSandbox`, `TestConfig_PatchConflict412` (ConfigStub returns ABORTED; If-Match set → 412), `TestGatewayConfig_PatchGlobal` (asserts `Global==true` reached the stub).
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): sandbox and gateway config routes"`

---

### Task 21: Policy routes

**Files:**
- Create: `internal/dto/policy.go`, `internal/dto/convert/policy.go`, `internal/api/policy.go`
- Test: `internal/api/policy_test.go`, `internal/dto/convert/policy_test.go`

**Interfaces:**
- Produces: `PolicyChunk` (all fields; policies via Task 14 `SandboxPolicy`), `DraftPolicy { Chunks; RollingSummary; DraftVersion Uint64String; LastAnalyzedAt *time.Time }`, `DraftHistoryEntry`, `SandboxPolicyRevision { Version uint32; PolicyHash string; Status string /* Loaded... */; LoadError string; CreatedAt, LoadedAt *time.Time; Policy *SandboxPolicy; Provenance map[string]string }`, `PolicyStatusResult { Revision; ActiveVersion uint32 }`, request bodies `ApproveRequest{ReviewToken}`, `RejectRequest{Reason}`, `ApproveAllRequest{IncludeSecurityFlagged bool; Approvals []DraftChunkApproval}`, results `ApproveResult/ApproveAllResult/UndoResult/ClearResult`.
- Routes: `GET .../sandboxes/{sandbox}/policy/draft?status=` · `POST .../policy/draft/chunks/{chunk}/approve|reject|undo` · `PUT .../policy/draft/chunks/{chunk}` (edit, 204) · `POST .../policy/draft/approve-all` · `POST .../policy/draft/clear` · `GET .../policy/draft/history` · `GET .../policy/status?version=` · `GET .../policy/revisions?pageSize&pageToken` → `Policy().List(ws, sb, v1.WithPageSize(int32), v1.WithPageToken(...))` and the two `/policy/global/...` routes using `v1.WithStatusGlobal(true)` / `v1.WithListGlobal(true)` with empty sandbox name.

- [ ] **Step 1: Tests** — `TestPolicy_Revisions` (fake `AddRevision` ×2 → items 2), `TestPolicy_GlobalRevisions` (`AddGlobalRevision`), `TestPolicy_Status` (fake), `TestPolicy_DraftFlow` (PolicyStub: get draft, approve with reviewToken reaches stub, reject 204, approveAll counts, clear, history, edit 204 with rule converted, undo).
- [ ] **Step 2–4:** fail → implement → pass. **Step 5: Commit** — `git commit -m "feat(api): policy draft review, status and revision routes"`

---

### Task 22: OpenAPI artifact, README, manual verification

**Files:**
- Create: `README.md`, `docs/errors.md`, `openapi.json` (generated), `docs/api.md` (route table copied from spec with final paths)
- Modify: `Makefile` (`openapi` target already exists; add `check-openapi` that fails if regeneration differs)

- [ ] **Step 1:** `make openapi` and commit `openapi.json`; write `README.md` (what/why, quick start with `configs/example.yaml`, auth model, pass-through example `curl -H "Authorization: Bearer $TOKEN" …`, limitations: no Idempotency-Key, Files unsupported, streaming in follow-up plan), `docs/errors.md` (problem shape + mapping table).
- [ ] **Step 2: Manual verification against a real gateway.** Ask the user to run `openshell status` and `openshell whoami --output json` to learn auth mode; then either (a) OIDC gateway: `OSPROXY_GATEWAY_ADDRESS=<host:port> OSPROXY_GATEWAY_CA_FILE=<ca> make run` and `TOKEN=<idp token>`; or (b) plaintext local: `OSPROXY_GATEWAY_ADDRESS=127.0.0.1:17670 OSPROXY_GATEWAY_ALLOW_PLAINTEXT=true OSPROXY_AUTH_MODE=none make run`. Then: `/readyz` 200; `/v1/whoami`; `POST /v1/workspaces/default/sandboxes` with `{"name":"proxy-smoke","spec":{"template":{"image":"python:3.12"}}}`; `POST .../proxy-smoke:wait {"for":"ready"}`; `POST .../exec {"command":["echo","hi"]}` → stdout `hi\n`; `DELETE`. Record outcomes in the PR/commit message.
- [ ] **Step 3:** `go test -race ./... && golangci-lint run` — Expected: clean.
- [ ] **Step 4: Commit** — `git commit -m "docs: README, error reference and generated OpenAPI"`

---

## Self-review notes

- Spec coverage: health/whoami/info (T8), workspaces (T13), sandboxes incl. wait/logs/providers (T15), templates (T16), exec collected (T17), services (T18), providers/profiles/refresh (T19), config (T20), policy (T21), error table (T3), auth (T5), config file (T2), G1/G2 factory (T6), G4 rule (T3/T15), G12 plaintext (T2/T6). Deferred to the follow-up plan: exec SSE stream, watch SSE, interactive/SSH/TCP WebSocket, SSH session create/revoke routes, proxy TLS test, Dockerfile, integration test, `/metrics`.
- Interfaces reused across tasks: `httperr.FromSDK/Write/Install` (T3→T4,T5,T8+), `auth.TokenFrom` (T5→T6), `gatewaytest.New/With*/StaticPager` (T7→T8+), `api.Deps`, `WorkspacePath/SandboxPath/ListQuery/IfMatchHeader/AllowMissingQuery` (T8→T13+), `dto.ObjectMeta/Page/Uint64String` (T11→T13+), `convert.PolicyFromSDK/ToSDK/MergeOpsToSDK/NetworkRuleToSDK` (T14→T20,T21).
