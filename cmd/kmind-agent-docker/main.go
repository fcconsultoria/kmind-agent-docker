package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fcconsultoria/kmind-agent-docker/internal/agent"
	"github.com/fcconsultoria/kmind-agent-docker/internal/config"
	"github.com/fcconsultoria/kmind-agent-docker/internal/identity"
)

var version = "dev"

func main() {
	configPath := flag.String("config", config.ConfigPathFromEnv(), "path to YAML configuration")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	installation, err := identity.LoadOrCreate(cfg.StateDir)
	if err != nil {
		slog.Error("initialize identity", "error", err)
		os.Exit(1)
	}
	runner, err := agent.New(cfg, installation, version, slog.Default())
	if err != nil {
		slog.Error("initialize agent", "error", err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("agent stopped", "error", err)
		os.Exit(1)
	}
}
