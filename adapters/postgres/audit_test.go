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
	if evs, err := f.events(store, tenant); err != nil || len(evs) != 1 || audit.VerifyChain(tenant, evs) != nil {
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
		if err := audit.VerifyChain(tenant, evs); err != nil {
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
		// The owner can switch the chain trigger off (a superuser here, with
		// session_replication_role); VerifyChain still sees the change.
		"insert event": `SET session_replication_role = replica;
			UPDATE iam.audit_events SET sequence = 4 WHERE sequence = 3;
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
			if err := audit.VerifyChain(tenant, evs); !errors.Is(err, audit.ErrBrokenChain) {
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

// TestAudit_DatabaseRefusesInsertsThatBreakTheChain (I1): the app role,
// reaching around the adapter with a raw INSERT, cannot leave a gap, a repeat
// or a wrong link, which every later Append would chain onto. A raw row that
// links correctly but carries a forged hash is accepted: proving a row came
// from the service is authenticity, for the later signed checkpoints; the
// chain still shows it (VerifyChain fails at that row).
func TestAudit_DatabaseRefusesInsertsThatBreakTheChain(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, appPool(t, migratedDB(t)))
	tenant := f.createTenant(store, "acme")
	first := f.appendEvent(store, tenant)
	const insert = `INSERT INTO iam.audit_events VALUES ($1, $2, $3, now(), 1, 'a.b', 'system', 'x', '', 'success', '{}', $4, $5)`
	forged := make([]byte, 32)
	rawInTenant(t, store, tenant, func(ctx context.Context, raw pgx.Tx) {
		for name, args := range map[string][]any{
			"gap":        {f.id().String(), tenant.String(), 3, first.Hash, forged},
			"repeat":     {f.id().String(), tenant.String(), 1, []byte{}, forged},
			"wrong link": {f.id().String(), tenant.String(), 2, forged, forged},
			"empty link": {f.id().String(), tenant.String(), 2, []byte{}, forged},
			"far future": {f.id().String(), tenant.String(), 1 << 40, first.Hash, forged},
		} {
			sp, err := raw.Begin(ctx)
			if err != nil {
				t.Fatalf("savepoint: %v", err)
			}
			_, err = sp.Exec(ctx, insert, args...)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.ConstraintName != "audit_events_chain_check" {
				t.Errorf("%s: error %v, want the chain check (23514)", name, err)
			}
			_ = sp.Rollback(ctx)
		}
		if _, err := raw.Exec(ctx, insert, f.id().String(), tenant.String(), 2, first.Hash, forged); err != nil {
			t.Errorf("well-linked raw insert: %v, want accepted (authenticity gap)", err)
		}
	})
	evs, err := f.events(store, tenant)
	if err != nil || len(evs) != 2 {
		t.Fatalf("chain: %d events, %v; want 2", len(evs), err)
	}
	if err := audit.VerifyChain(tenant, evs); !errors.Is(err, audit.ErrBrokenChain) {
		t.Errorf("VerifyChain over the forged row = %v, want ErrBrokenChain", err)
	}
}

// TestAudit_AppendWaitingPastLockTimeoutIsBusy (M2): an append that waits for
// another transaction's append to the same chain longer than lock_timeout is
// storage.ErrBusy, never a conflict or a generic error.
func TestAudit_AppendWaitingPastLockTimeoutIsBusy(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, appPool(t, migratedDB(t)))
	tenant := f.createTenant(store, "acme")
	first, second := f.auditEvent(), f.auditEvent()
	held, release := make(chan struct{}), make(chan struct{})
	var holder error
	var wg sync.WaitGroup
	wg.Go(func() {
		holder = store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
			_, err := tx.Audit().Append(ctx, first)
			close(held)
			<-release
			return err
		})
	})
	<-held
	err := store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Audit().Append(ctx, second)
		return err
	})
	close(release)
	wg.Wait()
	if !errors.Is(err, storage.ErrBusy) || errors.Is(err, storage.ErrConflict) {
		t.Errorf("Append behind a held lock = %v, want ErrBusy", err)
	}
	if holder != nil {
		t.Fatalf("holder: %v", holder)
	}
	if evs, err := f.events(store, tenant); err != nil || len(evs) != 1 {
		t.Errorf("chain: %d events, %v; want the holder's 1", len(evs), err)
	}
}

// TestStore_ReadCommittedDespiteServerDefault (M1): transactions are READ
// COMMITTED whatever default_transaction_isolation says, so an append that
// waited for the chain's lock sees the event committed meanwhile.
func TestStore_ReadCommittedDespiteServerDefault(t *testing.T) {
	db := migratedDB(t)
	adminExec(t, db, "ALTER ROLE "+quote(db.AppRole)+" SET default_transaction_isolation = 'repeatable read'")
	store := newStore(t, appPool(t, db))
	rawInSystem(t, store, func(ctx context.Context, raw pgx.Tx) {
		var level string
		if err := raw.QueryRow(ctx, "SHOW transaction_isolation").Scan(&level); err != nil || level != "read committed" {
			t.Errorf("isolation = %q (%v), want read committed", level, err)
		}
	})
	f := newFixture(t)
	tenant := f.createTenant(store, "acme")
	var ready sync.WaitGroup
	ready.Add(2)
	events := []audit.Event{f.auditEvent(), f.auditEvent()}
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
		t.Fatalf("concurrent first appends: %v", err)
	}
}
