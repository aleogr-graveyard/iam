// Command iam is the single binary of the identity platform.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/httpserver"
	"github.com/aleogr/iam/shared/config"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("iam stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, findings, err := config.Load(os.Getenv, os.ReadFile)
	logFindings(log, findings)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("iam starting", "version", version, "addr", cfg.HTTPAddr)

	var opts []httpserver.Option
	if cfg.DatabaseURL != "" {
		pool, err := database.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer pool.Close()
		opts = append(opts, httpserver.WithReadinessChecks(database.ReadinessCheck(pool)))
	}
	return httpserver.New(cfg.HTTPAddr, opts...).Run(ctx, cfg.ShutdownTimeout)
}

// logFindings logs each finding at the level of its severity. Prohibited
// findings are also fatal, but Load reports those through its error.
func logFindings(log *slog.Logger, findings []config.Finding) {
	levels := map[config.Severity]slog.Level{
		config.Info:       slog.LevelInfo,
		config.Warning:    slog.LevelWarn,
		config.Critical:   slog.LevelError,
		config.Prohibited: slog.LevelError,
	}
	for _, f := range findings {
		log.Log(context.Background(), levels[f.Severity], f.Message,
			"severity", f.Severity.String(), "code", f.Code, "reference", f.Reference)
	}
}
