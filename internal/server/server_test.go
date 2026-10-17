package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/kenmoini/openshell-proxy/internal/config"
)

func TestServer_RunAndGracefulShutdown(t *testing.T) {
	started := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			close(started)
			time.Sleep(500 * time.Millisecond)
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, "ok")
	})
	cfg := config.Default().Server
	cfg.Addr = "127.0.0.1:0"
	s := New(cfg, h, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for s.Addr() == "" || s.Addr() == "127.0.0.1:0" {
		if time.Now().After(deadline) {
			t.Fatal("server did not bind")
		}
		time.Sleep(5 * time.Millisecond)
	}
	resp, err := http.Get("http://" + s.Addr() + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}

	// In-flight request must complete across shutdown.
	slowStatus := make(chan int, 1)
	go func() {
		r, err := http.Get("http://" + s.Addr() + "/slow")
		if err != nil {
			slowStatus <- -1
			return
		}
		r.Body.Close()
		slowStatus <- r.StatusCode
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return")
	}
	if st := <-slowStatus; st != 200 {
		t.Fatalf("in-flight request got %d", st)
	}
}
