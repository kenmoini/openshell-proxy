# Core proxy: deferred follow-ups (from the 2026-10-17 final review triage)

Items the final whole-branch review classified as "can wait". Carry the streaming-related ones into the SSE/WebSocket plan.

- Task 1: minor (deferred): docs/sdk-pinning.md should list the concrete post-v0.1.4 signature change; sandbox note is environment-specific.
- Task 1: minor (deferred): golangci-lint not installed locally; `.golangci.yml` unvalidated. Final review should note; lint runs in CI or when user installs it.
- Task 2: minor (deferred): limits key names duplicated between UnmarshalYAML switch and yaml tags; ParseByteSize exported + int64 overflow on absurd sizes; empty env treated as unset; no zero/negative duration/limit validation; allowPlaintext+https:// not rejected; two weak test assertions; not all 24 env names exercised.
- Task 3: minor (deferred): no compile-time interface assertions; TestWrite lacks header/extension coverage; TestInstall doesn't restore huma.NewError; Write mutates caller's Problem and uses Header.Add; parallel status tables in statusFor/statusFromGRPC lack a comment; RED evidence was compile failure only.
- Task 4: minor (deferred, carry to streaming plan): statusRecorder.Flush doesn't set wrote → Recover could write a 500 body into a started stream; 1xx WriteHeader recorded as final; double-wrapped recorder (Logging+Recover); panic value logged verbatim; no tests for chunked body limit, recover-after-headers, Flush/Hijack passthrough, CORS no-op; no io.ReaderFrom on recorder.
- Task 5: minor (deferred): int64(exp) overflow for absurd float exp (fails closed); query token stays in r.URL (Logging logs path only — verify at final; streaming plan must scrub ws URLs); no segment length cap; `switch err` instead of errors.Is; missing tests for padded base64 and exp-as-string.
- Task 6: minor (deferred): no direct unit test of plaintextToken; closer discards Close error; dedicated clients carry no x-request-id.
- Task 7: minor (deferred): shared ErrUnimplemented pointer; StatusErr lacks Cause/FieldViolations; thin coverage of fold table and unset-stub paths; missing doc comments on With* options; StaticPager test also covers StatusErr.
- Task 8: minor (deferred → Task 11 adds TestIfMatchHeader_Version): /readyz bypasses huma timeout middleware (has its own); OpenAPI() re-calls Install (idempotent).
- Task 9: minor (deferred): ServeTLS early failure doesn't close listener; server failure logged twice; "starting" logged before bind; docs show Go-formatted durations; newLogger silently falls back (Validate already rejects bad level/format per Task 2); second SIGINT during grace doesn't force exit.
- Tasks 10-12: minor (deferred): unavailable test relies on port 1 refusing (could blackhole in some CI → accept either code or use an ephemeral closed port); thin /readyz body assertion; Uint64String edge cases untested; Page.Items nil-safety by convention; NilPage test relies on SDK exhaustion semantics; one discarded Marshal error in a test.
- Task 13: minor (deferred): label/annotation maps aliased not copied; weak require.Empty on possibly-null items; converters not nil-safe for nil pointer + nil error.
- Task 14: minor (deferred): SandboxFromSDK(nil) panics (value return); Phase lacks enum tag; maps/slices aliased by reference; test local named `min` shadows builtin.
- Task 15: minor (deferred): timeout only validated when wait=true; `deadlineProblem` misnamed + error constructors in wrong file; no HTTP-level cap/bad-timeout tests for start/stop; AttachFromSDK/DetachFromSDK deref without nil check; minLevel unvalidated.
- Tasks 16-18: minor (deferred): AuthModeToSDK lacks doc comment; readyWithin normalisation undocumented on DTO; exec 422 test doesn't assert stub not called; ServiceExpose.Domain omitempty inconsistency; no test for exec MaxBodyBytes override.
- Task 19: minor (deferred; raise at final review): ProviderToSDK forwards client-supplied server-owned fields (ID, CreatedAt, DeletionTimestamp, ResourceVersion, CredentialHandles) — consider dropping on create/update; profile PUT doesn't compare body id with path; import/lint accept empty items; name-mismatch test shallow; no plain-PUT-on-missing test; credentialKey passthrough unasserted; omitempty on non-pointer structs misleading; huma 422 may echo offending values in bodies (not logs); no RED-first run.
- Tasks 20-21: minor (deferred): empty bytes setting omitted by omitempty; Bytes not copied; untested body-field expectedResourceVersion fallback, header-over-body precedence, malformed If-Match on these routes, unknown type 422; approvals:[] treated as absent (undocumented).
- Tasks 20-21: minor (deferred): resolver error order nondeterministic (map iteration); ErrorDetail.Value set to type not member value.
- Task 22: review approved. minor (deferred): api.md says "four raw routes" but lists three + /docs; README 202 note cosmetic.

## Final reviewer notes (can wait)
- Slowloris: `ReadTimeout=0` and huma body reads ignore the context; a client can trickle up to 4 MiB indefinitely. Planned for the hardening phase.
- Shutdown: a long in-flight exec/wait outlasting the 20 s grace makes `Run` return an error and main exit 1 on a normal SIGTERM; treat as warning + `srv.Close()`.
- huma's `$schema` link transformer adds `$schema` to bodies and `/schemas/*.json`, `/openapi-3.0.*` are behind auth (401); disable the link transformer or exempt the paths.
- Converters panic (→ 500 via Recover) on a nil SDK pointer with nil error; a guard returning 502 would be cleaner.
- `auth.allowQueryToken` applies to every route; narrow it to WebSocket routes when streaming lands; scrub `access_token` from any logged URL.
- Status recorder `Flush` doesn't mark the response started (matters for SSE); carry into the streaming plan.
