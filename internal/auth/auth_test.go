package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kenmoini/openshell-proxy/internal/config"
)

var fixedNow = time.Unix(1_800_000_000, 0)

func jwt(payload string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc([]byte(payload)) + "." + "sig"
}

func TestParseBearer(t *testing.T) {
	tests := []struct {
		in    string
		token string
		ok    bool
	}{
		{"Bearer abc", "abc", true},
		{"bearer abc", "abc", true},
		{"BEARER abc", "abc", true},
		{"Bearer  abc", "abc", true},
		{"  Bearer abc  ", "abc", true},
		{"Basic abc", "", false},
		{"Bearer", "", false},
		{"Bearer a b", "", false},
		{"", "", false},
		{"abc", "", false},
	}
	for _, tc := range tests {
		got, ok := ParseBearer(tc.in)
		assert.Equal(t, tc.ok, ok, tc.in)
		assert.Equal(t, tc.token, got, tc.in)
	}
}

func TestSanityCheck(t *testing.T) {
	skew := 60 * time.Second
	exp := func(d time.Duration) string {
		return jwt(`{"exp":` + strconv.FormatInt(fixedNow.Add(d).Unix(), 10) + `}`)
	}
	assert.NoError(t, SanityCheck(jwt(`{"sub":"u"}`), fixedNow, skew))
	assert.NoError(t, SanityCheck(exp(time.Hour), fixedNow, skew))
	assert.ErrorIs(t, SanityCheck(exp(-time.Hour), fixedNow, skew), ErrExpired)
	assert.NoError(t, SanityCheck(exp(-30*time.Second), fixedNow, skew))
	assert.NoError(t, SanityCheck(jwt(`{"exp":1.8e9}`), fixedNow, skew))
	assert.ErrorIs(t, SanityCheck("abc", fixedNow, skew), ErrMalformed)
	assert.ErrorIs(t, SanityCheck("a..c", fixedNow, skew), ErrMalformed)
	assert.ErrorIs(t, SanityCheck("a.b.c.d", fixedNow, skew), ErrMalformed)
	assert.ErrorIs(t, SanityCheck("a.!!!.c", fixedNow, skew), ErrMalformed)
	assert.ErrorIs(t, SanityCheck(jwt(`[1]`), fixedNow, skew), ErrMalformed)
	assert.ErrorIs(t, SanityCheck(jwt(`not json`), fixedNow, skew), ErrMalformed)
}

func run(t *testing.T, o Options, path, authz string) (*httptest.ResponseRecorder, *string) {
	t.Helper()
	var seen *string
	h := Middleware(o)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tok, ok := TokenFrom(r.Context()); ok {
			seen = &tok
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authz != "" {
		req.Header.Set("Authorization", authz)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, seen
}

func opts() Options {
	return Options{Mode: config.ModeBearerPassthrough, LocalSanityCheck: true, Now: func() time.Time { return fixedNow }}
}

func reason(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &m))
	return m["reason"].(string)
}

func TestMiddleware_PassthroughRequiresToken(t *testing.T) {
	rec, seen := run(t, opts(), "/v1/sandboxes", "")
	assert.Equal(t, 401, rec.Code)
	assert.Equal(t, "TOKEN_MISSING", reason(t, rec))
	assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
	assert.Nil(t, seen)

	tok := jwt(`{"sub":"u"}`)
	rec, seen = run(t, opts(), "/v1/sandboxes", "Bearer "+tok)
	assert.Equal(t, 200, rec.Code)
	require.NotNil(t, seen)
	assert.Equal(t, tok, *seen)

	rec, _ = run(t, opts(), "/v1/sandboxes", "Basic abc")
	assert.Equal(t, 401, rec.Code)
	assert.Equal(t, "TOKEN_MALFORMED", reason(t, rec))

	rec, _ = run(t, opts(), "/v1/sandboxes", "Bearer abc")
	assert.Equal(t, 401, rec.Code)
	assert.Equal(t, "TOKEN_MALFORMED", reason(t, rec))

	rec, _ = run(t, opts(), "/v1/sandboxes", "Bearer "+jwt(`{"exp":1}`))
	assert.Equal(t, 401, rec.Code)
	assert.Equal(t, "TOKEN_EXPIRED", reason(t, rec))
	assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"))

	o := opts()
	o.LocalSanityCheck = false
	rec, _ = run(t, o, "/v1/sandboxes", "Bearer abc")
	assert.Equal(t, 200, rec.Code)
}

func TestMiddleware_ExemptPaths(t *testing.T) {
	for _, p := range []string{"/healthz", "/readyz", "/openapi.json", "/openapi.yaml", "/docs", "/v1/health"} {
		rec, _ := run(t, opts(), p, "")
		assert.Equal(t, 200, rec.Code, p)
	}
	o := opts()
	o.Exempt = []string{"/custom"}
	rec, _ := run(t, o, "/custom", "")
	assert.Equal(t, 200, rec.Code)
	rec, _ = run(t, o, "/healthz/x", "")
	assert.Equal(t, 401, rec.Code)
	// token still placed in ctx on exempt path
	_, seen := run(t, opts(), "/healthz", "Bearer abc")
	require.NotNil(t, seen)
	assert.Equal(t, "abc", *seen)
}

func TestMiddleware_ModeNone(t *testing.T) {
	o := Options{Mode: config.ModeNone}
	rec, seen := run(t, o, "/v1/sandboxes", "")
	assert.Equal(t, 200, rec.Code)
	assert.Nil(t, seen)
	rec, seen = run(t, o, "/v1/sandboxes", "Bearer abc")
	assert.Equal(t, 200, rec.Code)
	require.NotNil(t, seen)
	assert.Equal(t, "abc", *seen)
	rec, seen = run(t, o, "/v1/sandboxes", "Basic zzz")
	assert.Equal(t, 200, rec.Code)
	assert.Nil(t, seen)
}

func TestMiddleware_QueryToken(t *testing.T) {
	tok := jwt(`{"sub":"u"}`)
	rec, _ := run(t, opts(), "/v1/x?access_token="+tok, "")
	assert.Equal(t, 401, rec.Code)

	o := opts()
	o.AllowQueryToken = true
	rec, seen := run(t, o, "/v1/x?access_token="+tok, "")
	assert.Equal(t, 200, rec.Code)
	require.NotNil(t, seen)
	assert.Equal(t, tok, *seen)
}
