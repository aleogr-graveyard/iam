// Command iam is the single binary of the identity platform.
//
//	iam          serves the platform
//	iam migrate  applies the database migrations and exits
//
// Both read the same configuration (shared/config). The migration job
// connects with IAM_DATABASE_URL as the migration role and grants the
// service's role, IAM_DATABASE_APP_ROLE, what it needs (adapters/postgres).
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/httpserver"
	"github.com/aleogr/iam/shared/config"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log, os.Args[1:]); err != nil {
		log.Error("iam stopped", "error", err)
		os.Exit(1)
	}
}

var errUsage = errors.New("usage: iam [migrate]")

func run(log *slog.Logger, args []string) error {
	if len(args) > 1 || (len(args) == 1 && args[0] != "migrate") {
		return errUsage
	}
	cfg, findings, err := config.Load(os.Getenv, os.ReadFile)
	logFindings(log, findings)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(args) == 1 {
		return migrate(ctx, log, cfg)
	}
	return serve(ctx, log, cfg)
}

// migrate applies the migrations with the migration job's pool of one
// connection (database.MigrationJob).
func migrate(ctx context.Context, log *slog.Logger, cfg config.Config) error {
	if cfg.DatabaseURL == "" {
		return errors.New("migrate: IAM_DATABASE_URL is not set")
	}
	log.Info("iam migrating", "version", version)
	d := cfg.Database
	pool, err := database.OpenMigrator(ctx, cfg.DatabaseURL, database.Limits{
		StatementTimeout: d.MigrationStatementTimeout, IdleInTransactionTimeout: d.IdleInTransactionTimeout,
		LockTimeout: d.MigrationLockTimeout, MaxConnIdleTime: d.MaxConnIdleTime,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	err = postgres.Migrate(ctx, pool, postgres.MigrateConfig{
		AppRole: cfg.DatabaseAppRole, StatementTimeout: d.MigrationStatementTimeout, LockTimeout: d.MigrationLockTimeout,
	})
	if err != nil {
		return err
	}
	log.Info("iam migrations applied")
	return nil
}

func serve(ctx context.Context, log *slog.Logger, cfg config.Config) error {
	log.Info("iam starting", "version", version, "addr", cfg.HTTPAddr)

	var opts []httpserver.Option
	if cfg.DatabaseURL != "" {
		d := cfg.Database
		pool, err := database.Open(ctx, cfg.DatabaseURL, database.Limits{
			StatementTimeout: d.StatementTimeout, IdleInTransactionTimeout: d.IdleInTransactionTimeout,
			LockTimeout: d.LockTimeout, MaxConnIdleTime: d.MaxConnIdleTime,
		})
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
