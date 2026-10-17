package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/danielgtaylor/huma/v2"
	"github.com/kenmoini/openshell-proxy/internal/config"
)

const testToken = "test-token"

func testConfig(mutate func(*config.Config)) *config.Config {
	cfg := config.Default()
	cfg.Gateway.Address = "127.0.0.1:1"
	cfg.Gateway.AllowPlaintext = true
	cfg.Auth.Mode = config.ModeBearerPassthrough
	cfg.Auth.LocalSanityCheck = false
	if mutate != nil {
		mutate(&cfg)
	}
	return &cfg
}

func testDeps(client v1.ClientInterface, mutate func(*config.Config)) Deps {
	return Deps{
		Client: client,
		Cfg:    testConfig(mutate),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func newTestAPI(t *testing.T, client v1.ClientInterface, mutate func(*config.Config)) (http.Handler, huma.API) {
	t.Helper()
	return New(testDeps(client, mutate))
}

// do issues a request straight through the handler; no sockets are opened.
func do(t *testing.T, h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rd)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func authHeader() map[string]string {
	return map[string]string{"Authorization": "Bearer " + testToken}
}
