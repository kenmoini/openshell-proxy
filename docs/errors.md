# Errors

Every error response is `application/problem+json` (RFC 9457) with extension members. This includes errors raised by request validation and by the proxy's own middleware, so clients need only one error parser.

## Problem shape

```json
{
  "title": "Not Found",
  "status": 404,
  "detail": "sandbox \"demo\" not found",
  "instance": "/v1/workspaces/default/sandboxes/demo",
  "errors": [{"location": "body.name", "message": "must not be empty"}],
  "reason": "SANDBOX_NOT_FOUND",
  "domain": "openshell.nvidia.com",
  "metadata": {"sandbox": "demo"},
  "grpcCode": "NotFound",
  "retryAfterSeconds": 1,
  "outcomeUnknown": false,
  "requestId": "0b8f..."
}
```

| Field | Type | Meaning |
|---|---|---|
| `type` | string | Problem type URI; may be omitted. |
| `title` | string | Standard HTTP status text. |
| `status` | integer | HTTP status code. |
| `detail` | string | Human-readable message; may be omitted. |
| `instance` | string | Request path; may be omitted. |
| `errors` | array | Per-field validation problems, each `{location, message}`. Gateway field violations appear here as `body.<field>`. |
| `reason` | string | Machine-readable reason from the gateway's `ErrorInfo`, or one set by the proxy (for example `SANDBOX_TERMINAL_PHASE`, `TOKEN_MISSING`). |
| `domain` | string | `ErrorInfo` domain from the gateway. |
| `metadata` | object | `ErrorInfo` metadata (string map) from the gateway. |
| `grpcCode` | string | Original gRPC code name (for example `NotFound`) when the gateway returned one. |
| `retryAfterSeconds` | integer | Suggested wait before retrying; also sent as the `Retry-After` header. |
| `outcomeUnknown` | boolean | `true` when the request may or may not have taken effect (see 504). Omitted when false. |
| `requestId` | string | The `X-Request-ID` of the request, for correlating logs. |

Empty members are omitted from the JSON.

## gRPC to HTTP mapping

The mapper looks at the gRPC code first. Errors from the SDK without a gRPC code (`GRPCCode == 0`, SDK-local) are mapped by the SDK error code instead (second table).

| gRPC code | HTTP | Notes |
|---|---|---|
| `INVALID_ARGUMENT`, `OUT_OF_RANGE` | 400 | Field violations go in `errors[]`. |
| `UNAUTHENTICATED` | 401 | Adds `WWW-Authenticate: Bearer`. |
| `PERMISSION_DENIED` | 403 | The gateway enforces roles and workspace membership. |
| `NOT_FOUND` | 404 | |
| `ALREADY_EXISTS` | 409 | |
| `ABORTED` | 409 | 412 when the request carried `If-Match`. |
| `FAILED_PRECONDITION` | 412, 409 or 422 | 412 when the request carried `If-Match`; otherwise 409 when `reason` is `REQUEST_OUTCOME_UNCERTAIN` (the `REQUEST_*` idempotency family); otherwise 422. |
| `RESOURCE_EXHAUSTED` | 429 | `Retry-After`. |
| `UNAVAILABLE` | 503 | `Retry-After`. |
| `DEADLINE_EXCEEDED`, `CANCELLED` | 504 | `outcomeUnknown: true`. |
| `UNIMPLEMENTED` | 501 | |
| anything else (`INTERNAL`, `UNKNOWN`, `DATA_LOSS`, ...) | 502 | |

SDK-local errors (no gRPC code), mapped by SDK error code:

| SDK error | HTTP | Notes |
|---|---|---|
| invalid argument | 400 | |
| unauthenticated | 401 | `WWW-Authenticate: Bearer`. |
| permission denied | 403 | |
| not found | 404 | |
| already exists | 409 | |
| conflict | 409 | 412 when the request carried `If-Match`. |
| unavailable | 503 | `Retry-After`. |
| deadline exceeded, cancelled | 504 | `outcomeUnknown: true`. |
| unimplemented | 501 | |
| internal with no gRPC code and no wrapped cause (a sandbox reached a terminal phase while waiting) | 409 | `reason: SANDBOX_TERMINAL_PHASE`. |
| anything else | 502 | |

Rules that apply on top of the tables:

- **Terminal phase**: waiting on a sandbox that entered a terminal phase (for example failed) is a client-visible state, not a server fault, so it is 409 `SANDBOX_TERMINAL_PHASE` rather than 502.
- **If-Match**: `ABORTED`, `FAILED_PRECONDITION` and SDK conflicts become 412 instead of 409/422 when `If-Match` was sent. A `PUT` of a provider profile without `If-Match` is rejected with 428.
- **Retry-After**: 429 and 503 always carry `Retry-After` and `retryAfterSeconds`. A delay supplied by the gateway is rounded up to whole seconds (minimum 1); otherwise it is 1. It is advisory.
- **504 and `outcomeUnknown`**: a deadline or cancellation can fire after the gateway applied a mutation. The proxy sets `outcomeUnknown: true`, so read the resource before retrying. Because `Idempotency-Key` is unsupported, blind retries of creates can duplicate or conflict.
- **502 detail**: when the upstream failure is not a recognizable gateway status (a transport error, for example), `detail` is the fixed string `upstream error`; internal error text is not leaked. Context deadline or cancellation errors are 504, not 502.

## Errors raised by the proxy

| HTTP | `reason` | When |
|---|---|---|
| 401 | `TOKEN_MISSING` | No `Authorization` header on a route that needs one (`auth.mode: bearer-passthrough`). |
| 401 | `TOKEN_MALFORMED` | The header is not `Bearer <token>`, or the token is not a decodable JWT (local sanity check). |
| 401 | `TOKEN_EXPIRED` | The token's `exp` has passed (60 s skew allowed). Needs `auth.localSanityCheck: true`. |
| 413 | | Request body larger than `limits.maxBody` (or `limits.maxExecBody` for exec). |
| 428 | | `If-Match` required but absent (provider profile update). |

401 responses carry `WWW-Authenticate: Bearer` (with `error="invalid_token"` for an expired token). The local checks are fast-fail only; the gateway remains the authority and may still return 401 or 403 for a token the proxy accepted. Validation errors from the request schema (wrong types, enum values, missing required fields) are 422 with `errors[]` in huma's standard form, wrapped in the same problem envelope.
