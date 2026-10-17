// Command openshell-proxy is a REST/JSON proxy in front of the OpenShell gRPC gateway.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kenmoini/openshell-proxy/internal/api"
	"github.com/kenmoini/openshell-proxy/internal/config"
	"github.com/kenmoini/openshell-proxy/internal/gateway"
	"github.com/kenmoini/openshell-proxy/internal/server"
	"github.com/kenmoini/openshell-proxy/internal/version"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	path := flag.String("config", "", "path to YAML config file (default: environment only)")
	printOpenAPI := flag.Bool("print-openapi", false, "print the OpenAPI document as JSON and exit")
	flag.Parse()

	if *printOpenAPI {
		cfg := config.Default()
		logger := newLogger(cfg.Log)
		spec, err := api.OpenAPI(api.Deps{Cfg: &cfg, Log: logger})
		if err != nil {
			return fmt.Errorf("generate openapi: %w", err)
		}
		_, err = os.Stdout.Write(append(spec, '\n'))
		return err
	}

	cfg, err := config.Load(*path, os.Getenv)
	if err != nil {
		return err
	}
	logger := newLogger(cfg.Log)

	g, err := gateway.New(cfg.Gateway)
	if err != nil {
		return fmt.Errorf("gateway: %w", err)
	}
	defer g.Close()

	h, _ := api.New(api.Deps{Client: g.Client(), Dedicated: g.Dedicated, Cfg: cfg, Log: logger})
	srv := server.New(cfg.Server, h, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting openshell-proxy",
		"addr", cfg.Server.Addr,
		"gatewayAddress", cfg.Gateway.Address,
		"authMode", string(cfg.Auth.Mode),
		"version", version.Version)
	if err := srv.Run(ctx); err != nil {
		logger.Error("server failed", "error", err)
		return err
	}
	logger.Info("shutdown complete")
	return nil
}

func newLogger(c config.LogConfig) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	if c.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, opts))
}
