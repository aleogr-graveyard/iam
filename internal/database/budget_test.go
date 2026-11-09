package database_test

import (
	"strings"
	"testing"
	"time"

	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/testdb"
)

// TestConnectionBudget holds the lab ceiling of ADR 0001, item 6: the service
// pool times instances times revisions live during a rollout, plus the
// migration job, plus one person in psql, never exceeds Ceiling.
func TestConnectionBudget(t *testing.T) {
	total := database.ServicePool*database.MaxInstances*database.RolloutRevisions + database.MigrationJob + database.Person
	if total > database.Ceiling {
		t.Fatalf("connection budget %d exceeds the ceiling %d", total, database.Ceiling)
	}
	if database.Ceiling != 6 {
		t.Fatalf("Ceiling = %d; ADR 0001 fixes it at 6", database.Ceiling)
	}
}

// The budget must describe the pools the binary really opens: cmd/iam serves
// with database.Open and migrates with database.OpenMigrator.
func TestConnectionBudget_MatchesPoolConfigs(t *testing.T) {
	const url = "postgres://iam@db.invalid:5432/iam?sslmode=disable"
	for _, tc := range []struct {
		name string
		kind database.PoolKind
		want int32
	}{
		{"service", database.ServicePoolKind, database.ServicePool},
		{"migration", database.MigrationPoolKind, database.MigrationJob},
	} {
		cfg, err := database.PoolConfig(url, tc.kind, testdb.Limits())
		if err != nil {
			t.Fatalf("%s: PoolConfig: %v", tc.name, err)
		}
		if cfg.MaxConns != tc.want {
			t.Errorf("%s pool MaxConns = %d, want %d", tc.name, cfg.MaxConns, tc.want)
		}
		if cfg.MinConns != 0 {
			t.Errorf("%s pool MinConns = %d, want 0 (no idle connections held)", tc.name, cfg.MinConns)
		}
	}
	if database.MaxConns != database.ServicePool {
		t.Fatalf("MaxConns = %d, want ServicePool %d", database.MaxConns, database.ServicePool)
	}
}

// An invalid URL never echoes its credentials (roadmap, inherited from
// milestone 0): the error is a fixed text, whatever part of the URL is wrong.
func TestPoolConfig_InvalidURLDoesNotLeakCredentials(t *testing.T) {
	const password = "s3cr3t-pw"
	for _, raw := range []string{
		"postgres://iam:" + password + "@[::1:5432/iam",
		"postgres://iam:" + password + "@db.example.com:notaport/iam",
		"postgres://iam:" + password + "@db.example.com/iam?sslmode=bogus",
		"postgres://iam:" + password + "@db.example.com/iam?connect_timeout=x",
		"host=db.example.com password=" + password + " sslmode=bogus",
		"host=db.example.com password='" + password,
		"postgres://iam:" + password + "%zz@db.example.com/iam",
	} {
		for _, kind := range []database.PoolKind{database.ServicePoolKind, database.MigrationPoolKind} {
			_, err := database.PoolConfig(raw, kind, testdb.Limits())
			if err == nil {
				t.Errorf("PoolConfig accepted an invalid URL")
				continue
			}
			if strings.Contains(err.Error(), password) || strings.Contains(err.Error(), "db.example.com") {
				t.Errorf("error leaks the URL: %v", err)
			}
		}
	}
}

// Both pool kinds send the limits as session settings, so no statement, lock
// wait or idle transaction can hold a connection indefinitely.
func TestPoolConfig_CarriesLimits(t *testing.T) {
	limits := database.Limits{
		StatementTimeout: 1500 * time.Millisecond, IdleInTransactionTimeout: 7 * time.Second,
		LockTimeout: 250 * time.Millisecond, MaxConnIdleTime: 42 * time.Second,
	}
	want := map[string]string{
		"statement_timeout": "1500", "idle_in_transaction_session_timeout": "7000", "lock_timeout": "250",
	}
	for _, kind := range []database.PoolKind{database.ServicePoolKind, database.MigrationPoolKind} {
		cfg, err := database.PoolConfig("postgres://iam@db.invalid/iam", kind, limits)
		if err != nil {
			t.Fatalf("PoolConfig: %v", err)
		}
		for k, v := range want {
			if got := cfg.ConnConfig.RuntimeParams[k]; got != v {
				t.Errorf("kind %d: %s = %q, want %q", kind, k, got, v)
			}
		}
		if cfg.MaxConnIdleTime != limits.MaxConnIdleTime {
			t.Errorf("kind %d: MaxConnIdleTime = %v, want %v", kind, cfg.MaxConnIdleTime, limits.MaxConnIdleTime)
		}
	}
}

// Limits are required: a missing one would mean "no limit" to Postgres.
func TestPoolConfig_RefusesMissingLimits(t *testing.T) {
	for name, mutate := range map[string]func(*database.Limits){
		"statement": func(l *database.Limits) { l.StatementTimeout = 0 },
		"idle":      func(l *database.Limits) { l.IdleInTransactionTimeout = 0 },
		"lock":      func(l *database.Limits) { l.LockTimeout = 500 * time.Microsecond },
		"conn idle": func(l *database.Limits) { l.MaxConnIdleTime = -time.Second },
	} {
		l := testdb.Limits()
		mutate(&l)
		if _, err := database.PoolConfig("postgres://iam@db.invalid/iam", database.ServicePoolKind, l); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
