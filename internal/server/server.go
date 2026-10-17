// Package server owns the HTTP server lifecycle: listen, serve, graceful shutdown.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kenmoini/openshell-proxy/internal/config"
)

const shutdownGrace = 20 * time.Second

// Server wraps an http.Server with context-driven graceful shutdown.
type Server struct {
	srv *http.Server
	cfg config.ServerConfig
	log *slog.Logger

	mu   sync.Mutex
	addr string
}

// New builds a Server. WriteTimeout and ReadTimeout are zero so long-lived
// streams are not cut off.
func New(cfg config.ServerConfig, h http.Handler, log *slog.Logger) *Server {
	return &Server{
		cfg: cfg,
		log: log,
		srv: &http.Server{
			Addr:              cfg.Addr,
			Handler:           h,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			ReadTimeout:       0,
			WriteTimeout:      0,
			ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
		},
	}
}

// Addr returns the bound address once Run has started listening, else "".
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Run serves until ctx is cancelled, then shuts down gracefully. A clean
// shutdown returns nil.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.addr = ln.Addr().String()
	s.mu.Unlock()

	tls := s.cfg.TLS.CertFile != "" && s.cfg.TLS.KeyFile != ""
	errc := make(chan error, 1)
	go func() {
		if tls {
			errc <- s.srv.ServeTLS(ln, s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
		} else {
			errc <- s.srv.Serve(ln)
		}
	}()

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	sctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := s.srv.Shutdown(sctx); err != nil {
		return err
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
