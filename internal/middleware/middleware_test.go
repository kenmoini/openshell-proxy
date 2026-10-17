package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestRequestID_GeneratedAndEchoed(t *testing.T) {
	var ctxID string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestIDFrom(r.Context())
	}))
	hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	got := rec.Header().Get("X-Request-ID")
	if !hex32.MatchString(got) || ctxID != got {
		t.Fatalf("generated: header=%q ctx=%q", got, ctxID)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "abc-123_x.y")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-ID") != "abc-123_x.y" || ctxID != "abc-123_x.y" {
		t.Fatalf("valid id not echoed: %q ctx=%q", rec.Header().Get("X-Request-ID"), ctxID)
	}

	for _, bad := range []string{"bad id!", strings.Repeat("a", 129)} {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Request-ID", bad)
		h.ServeHTTP(rec, req)
		if !hex32.MatchString(rec.Header().Get("X-Request-ID")) {
			t.Fatalf("invalid id %q not replaced: %q", bad, rec.Header().Get("X-Request-ID"))
		}
	}
	if RequestIDFrom(httptest.NewRequest("GET", "/", nil).Context()) != "" {
		t.Fatal("expected empty id without middleware")
	}
}

func TestRecover_WritesProblem(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	h := Recover(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("secret boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("content-type %q", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != float64(500) {
		t.Fatalf("body %v", body)
	}
	if strings.Contains(rec.Body.String(), "secret boom") {
		t.Fatal("panic value leaked")
	}
	if !strings.Contains(buf.String(), "secret boom") {
		t.Fatal("panic not logged")
	}
}

func TestLogging_RedactsQueryAndHeaders(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte("hello"))
	}), RequestID, Logging(log))
	req := httptest.NewRequest("GET", "/v1/x?access_token=s3cret", nil)
	req.Header.Set("Authorization", "Bearer abc")
	h.ServeHTTP(httptest.NewRecorder(), req)

	out := buf.String()
	for _, bad := range []string{"access_token=", "s3cret", "Authorization", "Bearer"} {
		if strings.Contains(out, bad) {
			t.Fatalf("log leaked %q: %s", bad, out)
		}
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &line); err != nil {
		t.Fatalf("expected one json line: %v: %s", err, out)
	}
	if line["path"] != "/v1/x" || line["method"] != "GET" || line["status"] != float64(201) || line["bytes"] != float64(5) {
		t.Fatalf("attrs: %v", line)
	}
	if line["requestId"] == "" || line["requestId"] == nil {
		t.Fatalf("missing requestId: %v", line)
	}
	if _, ok := line["durationMs"]; !ok {
		t.Fatal("missing durationMs")
	}
}

func TestBodyLimit_Returns413(t *testing.T) {
	srv := httptest.NewServer(BodyLimit(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})))
	defer srv.Close()
	c := srv.Client()

	resp, err := c.Post(srv.URL, "application/octet-stream", bytes.NewReader(make([]byte, 2048)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 413 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("content-type %q", ct)
	}

	resp, err = c.Post(srv.URL, "application/octet-stream", bytes.NewReader(make([]byte, 10)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("second request status %d", resp.StatusCode)
	}
}

func TestChain_Order(t *testing.T) {
	var order []string
	mk := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), mk("a"), mk("b"), mk("c"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if got := strings.Join(order, ","); got != "a,b,c,handler" {
		t.Fatalf("order %s", got)
	}
}
