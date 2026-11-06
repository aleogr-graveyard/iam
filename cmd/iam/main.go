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
	cfg, err := loadEnv(os.Getenv)
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
