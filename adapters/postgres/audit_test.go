package postgres_test

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
)

func (f *fixture) auditEvent() audit.Event {
	return audit.Event{
		ID: ids.EventID{ID: f.id()}, Time: f.now, Action: "test.appended",
		Actor: audit.Actor{Type: audit.ActorSystem, ID: "test"}, Target: "tenants/acme", Outcome: audit.OutcomeSuccess,
		Data: map[string]string{"k": "v"},
	}
}

func (f *fixture) appendEvent(s storage.Store, tenant ids.TenantID) audit.Event {
	f.t.Helper()
	var sealed audit.Event
	if err := s.InTenant(f.t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		var err error
		sealed, err = tx.Audit().Append(ctx, f.auditEvent())
		return err
	}); err != nil {
		f.t.Fatalf("Append: %v", err)
	}
	return sealed
}

func (f *fixture) events(s storage.Store, tenant ids.TenantID) ([]audit.Event, error) {
	var out []audit.Event
	err := s.InTenant(f.t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		var err error
		out, err = tx.Audit().Events(ctx, 0, storage.MaxListLimit)
		return err
	})
	return out, err
}

// TestAudit_AppRoleCannotUpdateOrDelete: the service's role cannot change,
// remove or empty audit events, its own tenant's included (TM-TEN-004).
func TestAudit_AppRoleCannotUpdateOrDelete(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	store := newStore(t, appPool(t, db))
	tenant := f.createTenant(store, "acme")
	f.appendEvent(store, tenant)
	rawInTenant(t, store, tenant, func(ctx context.Context, raw pgx.Tx) {
		for _, sql := range []string{
			"UPDATE iam.audit_events SET data = '{}'",
			"DELETE FROM iam.audit_events",
			"TRUNCATE iam.audit_events",
		} {
			sp, err := raw.Begin(ctx)
			if err != nil {
				t.Fatalf("savepoint: %v", err)
			}
			_, err = sp.Exec(ctx, sql)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != codeInsufficientPrivilege {
				t.Errorf("%s: error %v, want insufficient privilege (42501)", sql, err)
			}
			_ = sp.Rollback(ctx)
		}
	})
	if evs, err := f.events(store, tenant); err != nil || len(evs) != 1 || audit.VerifyChain(evs) != nil {
		t.Errorf("chain after the attempts: %d events, %v; want the 1 intact event", len(evs), err)
	}
}

// TestAudit_ConcurrentFirstAppends: two transactions that both reach Append
// on a chain with no event yet (no row to lock) give sequences 1 and 2, never
// two events at 1 or a conflict. The barrier makes both transactions open,
// on the service's pool of 2, before either appends.
func TestAudit_ConcurrentFirstAppends(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, appPool(t, migratedDB(t)))
	for round := range 5 {
		tenant := f.createTenant(store, "acme-"+string(rune('a'+round)))
		events := []audit.Event{f.auditEvent(), f.auditEvent()}
		var ready sync.WaitGroup
		ready.Add(len(events))
		errs := make([]error, len(events))
		var wg sync.WaitGroup
		for i, e := range events {
			wg.Go(func() {
				errs[i] = store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
					ready.Done()
					ready.Wait()
					_, err := tx.Audit().Append(ctx, e)
					return err
				})
			})
		}
		wg.Wait()
		if err := errors.Join(errs...); err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		evs, err := f.events(store, tenant)
		if err != nil || len(evs) != 2 {
			t.Fatalf("round %d: %d events, %v; want 2", round, len(evs), err)
		}
		if err := audit.VerifyChain(evs); err != nil {
			t.Errorf("round %d: %v", round, err)
		}
	}
}

// TestAudit_TamperingBehindTheAdapterIsDetected is the abuse test of
// TM-TEN-004 at the database: the schema owner, which the app role is not,
// edits, deletes or inserts rows directly. Reads still succeed where the rows
// are well formed, and VerifyChain reports the change; a malformed row is
// ErrCorrupt.
func TestAudit_TamperingBehindTheAdapterIsDetected(t *testing.T) {
	for name, sql := range map[string]string{
		"edit data":    `UPDATE iam.audit_events SET data = '{"k":"w"}' WHERE sequence = 2`,
		"edit time":    `UPDATE iam.audit_events SET occurred_at = occurred_at + interval '1 second' WHERE sequence = 1`,
		"delete event": `DELETE FROM iam.audit_events WHERE sequence = 2`,
		"swap order": `UPDATE iam.audit_events SET sequence = 100 WHERE sequence = 2;
			UPDATE iam.audit_events SET sequence = 2 WHERE sequence = 3;
			UPDATE iam.audit_events SET sequence = 3 WHERE sequence = 100`,
		"insert event": `UPDATE iam.audit_events SET sequence = 4 WHERE sequence = 3;
			INSERT INTO iam.audit_events SELECT gen_random_uuid(), tenant_id, 3, occurred_at, schema_version, action,
			actor_type, actor_id, target, outcome, data, hash, hash FROM iam.audit_events WHERE sequence = 2`,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			db := migratedDB(t)
			store := newStore(t, appPool(t, db))
			tenant := f.createTenant(store, "acme")
			for range 3 {
				f.appendEvent(store, tenant)
			}
			adminExec(t, db, sql)
			evs, err := f.events(store, tenant)
			if err != nil {
				t.Fatalf("Events: %v", err)
			}
			if err := audit.VerifyChain(evs); !errors.Is(err, audit.ErrBrokenChain) {
				t.Errorf("VerifyChain = %v, want ErrBrokenChain", err)
			}
		})
	}
	t.Run("malformed row", func(t *testing.T) {
		f := newFixture(t)
		db := migratedDB(t)
		store := newStore(t, appPool(t, db))
		tenant := f.createTenant(store, "acme")
		f.appendEvent(store, tenant)
		adminExec(t, db, `UPDATE iam.audit_events SET action = 'Not Dotted'`)
		if _, err := f.events(store, tenant); !errors.Is(err, postgres.ErrCorrupt) {
			t.Errorf("Events = %v, want ErrCorrupt", err)
		}
	})
}

// TestAudit_StoredRowMatchesSealedEvent: the columns hold exactly what was
// hashed, so a verifier reading the table directly gets the same chain.
func TestAudit_StoredRowMatchesSealedEvent(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	store := newStore(t, appPool(t, db))
	tenant := f.createTenant(store, "acme")
	first, second := f.appendEvent(store, tenant), f.appendEvent(store, tenant)
	admin := connect(t, db.AdminURL)
	for _, e := range []audit.Event{first, second} {
		var prev, hash string
		err := admin.QueryRow(t.Context(), "SELECT encode(prev_hash, 'hex'), encode(hash, 'hex') FROM iam.audit_events WHERE id = $1",
			e.ID.String()).Scan(&prev, &hash)
		if err != nil || prev != hex.EncodeToString(e.PrevHash) || hash != hex.EncodeToString(e.Hash) {
			t.Errorf("sequence %d: stored prev %s hash %s (%v); want %x %x", e.Sequence, prev, hash, err, e.PrevHash, e.Hash)
		}
	}
}
