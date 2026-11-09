package storagetest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sync"
	"time"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// auditTests are the rules of the audit chain (storage package doc, Audit;
// contract C4).
func auditTests() []test {
	return []test{
		{"Audit/AppendAssignsChain", testAuditAppendAssignsChain},
		{"Audit/AppendChainsPerTenant", testAuditChainsPerTenant},
		{"Audit/AppendIsAtomicWithTx", testAuditAppendIsAtomicWithTx},
		{"Audit/ConcurrentAppendsAreGapless", testAuditConcurrentAppends},
		{"Audit/CrossTenantIsolation", testAuditCrossTenantIsolation},
		{"Audit/InvalidEventRefused", testAuditInvalidEventRefused},
		{"Audit/EventsPaging", testAuditEventsPaging},
		{"Audit/SystemChain", testAuditSystemChain},
		{"Audit/UseAfterReturnFails", testAuditUseAfterReturnFails},
	}
}

// auditEvent is a valid event as a caller builds it: ID and Time set, chain
// fields left for Append.
func (f *fixture) auditEvent(action string) audit.Event {
	return audit.Event{
		ID: ids.EventID{ID: f.newID()}, Time: f.clk.Now(), Action: action,
		Actor:   audit.Actor{Type: audit.ActorOperator, ID: "op-1"},
		Target:  "tenants/acme",
		Outcome: audit.OutcomeSuccess,
		Data:    map[string]string{"n": action},
	}
}

// appendEvents appends n events to tenant in one transaction each.
func (f *fixture) appendEvents(tenant ids.TenantID, n int) []audit.Event {
	f.t.Helper()
	out := make([]audit.Event, 0, n)
	for i := range n {
		e := f.auditEvent(fmt.Sprintf("test.event_%d", i))
		f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
			got, err := tx.Audit().Append(ctx, e)
			out = append(out, got)
			return err
		})
	}
	return out
}

// chainOf reads tenant's whole chain, page by page.
func (f *fixture) chainOf(tenant ids.TenantID) []audit.Event {
	f.t.Helper()
	var all []audit.Event
	f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
		for after := int64(0); ; {
			page, err := tx.Audit().Events(ctx, after, storage.MaxListLimit)
			if err != nil || len(page) == 0 {
				return err
			}
			all = append(all, page...)
			after = page[len(page)-1].Sequence
		}
	})
	return all
}

// wantChain checks that tenant's stored chain is intact and equals want.
func (f *fixture) wantChain(tenant ids.TenantID, want []audit.Event) {
	f.t.Helper()
	got := f.chainOf(tenant)
	if err := audit.VerifyChain(tenant, got); err != nil {
		f.t.Errorf("stored chain: %v", err)
	}
	if want == nil {
		want = []audit.Event{}
	}
	if got == nil {
		got = []audit.Event{}
	}
	wantEvents(f.t, "stored chain", got, want)
}

func wantEvents(t interface {
	Helper()
	Errorf(string, ...any)
}, what string, got, want []audit.Event) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: %d events, want %d", what, len(got), len(want))
		return
	}
	for i := range got {
		g, w := got[i], want[i]
		g.Time, w.Time = g.Time.UTC(), w.Time.UTC()
		if !reflect.DeepEqual(g, w) {
			t.Errorf("%s[%d]:\n got  %+v\n want %+v", what, i, g, w)
		}
	}
}

// testAuditAppendAssignsChain: Append forces the tenant and the schema
// version, truncates the time, numbers the events from 1, links and seals
// them, and returns them as Events reads them back.
func testAuditAppendAssignsChain(f *fixture) {
	a, other := f.tenant("acme"), f.tenant("other")
	e := f.auditEvent("tenant.created")
	e.TenantID = other // forced to the Tx's tenant
	e.SchemaVersion = 0
	e.Sequence, e.PrevHash, e.Hash = 99, []byte("forged"), []byte("forged")
	e.Time = e.Time.Add(999 * time.Nanosecond).In(time.FixedZone("BRT", -3*3600))
	var first audit.Event
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		var err error
		first, err = tx.Audit().Append(ctx, e)
		return err
	})
	switch {
	case first.TenantID != a:
		f.t.Errorf("TenantID = %v, want the Tx's tenant %v", first.TenantID, a)
	case first.SchemaVersion != audit.SchemaVersion || first.Sequence != 1 || first.PrevHash != nil:
		f.t.Errorf("first event: schema %d, sequence %d, prev %x; want %d, 1, empty", first.SchemaVersion, first.Sequence, first.PrevHash, audit.SchemaVersion)
	case !first.Time.Equal(f.clk.Now()) || first.Time.Location() != time.UTC:
		f.t.Errorf("Time = %v, want %v truncated to the microsecond, in UTC", first.Time, f.clk.Now())
	case len(first.Hash) != audit.HashBytes:
		f.t.Errorf("Hash = %x, want %d bytes", first.Hash, audit.HashBytes)
	}
	// Neither the caller's event nor the returned one aliases the stored one.
	stored := first
	stored.Data, stored.Hash = maps.Clone(first.Data), bytes.Clone(first.Hash)
	e.Data["n"], first.Data["n"], first.Hash[0] = "changed", "changed", first.Hash[0]^1
	rest := f.appendEvents(a, 3)
	f.wantChain(a, append([]audit.Event{stored}, rest...))
	f.wantChain(other, nil)
}

// testAuditChainsPerTenant: interleaved appends in two tenants give two
// independent chains, each from sequence 1.
func testAuditChainsPerTenant(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	var inA, inB []audit.Event
	for range 3 {
		inA = append(inA, f.appendEvents(a, 1)...)
		inB = append(inB, f.appendEvents(b, 2)...)
	}
	f.wantChain(a, inA)
	f.wantChain(b, inB)
	if inB[0].Sequence != 1 || inA[2].Sequence != 3 || inB[5].Sequence != 6 {
		f.t.Errorf("sequences A %d.., B %d..%d; want per-tenant numbering from 1", inA[2].Sequence, inB[0].Sequence, inB[5].Sequence)
	}
}

// testAuditAppendIsAtomicWithTx: an event appended in a transaction that
// rolls back is gone, and its sequence is reused.
func testAuditAppendIsAtomicWithTx(f *fixture) {
	a := f.tenant("acme")
	kept := f.appendEvents(a, 1)
	err := f.store.InTenant(f.ctx, a, func(ctx context.Context, tx storage.Tx) error {
		e, err := tx.Audit().Append(ctx, f.auditEvent("test.dropped"))
		if err != nil || e.Sequence != 2 {
			f.t.Errorf("Append in the failing tx: sequence %d, %v; want 2", e.Sequence, err)
		}
		if err := tx.Organizations().Create(ctx, f.org("sales")); err != nil {
			f.t.Errorf("Create: %v", err)
		}
		return errBoom
	})
	wantErr(f.t, "InTenant", err, errBoom)
	f.wantEmptyTenant(a)
	f.wantChain(a, kept)
	next := f.appendEvents(a, 1)
	if next[0].Sequence != 2 || !bytes.Equal(next[0].PrevHash, kept[0].Hash) {
		f.t.Errorf("after rollback: sequence %d, want 2 linked to the kept event", next[0].Sequence)
	}
	f.wantChain(a, append(kept, next...))
}

// testAuditConcurrentAppends: appends to one tenant from several goroutines
// (each with other work in its transaction) give one chain without gap or
// fork; appends to another tenant run alongside.
func testAuditConcurrentAppends(f *fixture) {
	const workers, perWorker = 4, 5
	tenants := []ids.TenantID{f.tenant("tenant-a"), f.tenant("tenant-b")}
	events := make([][]audit.Event, workers)
	for w := range workers {
		for i := range perWorker {
			events[w] = append(events[w], f.auditEvent(fmt.Sprintf("test.w%d_e%d", w, i)))
		}
	}
	for _, err := range f.appendConcurrently(func(w int) ids.TenantID { return tenants[min(w/(workers-1), 1)] }, events) {
		wantErr(f.t, "concurrent Append", err, nil)
	}
	for i, want := range []int{(workers - 1) * perWorker, perWorker} {
		chain := f.chainOf(tenants[i])
		if len(chain) != want {
			f.t.Errorf("tenant %d: %d events, want %d", i, len(chain), want)
		}
		if err := audit.VerifyChain(tenants[i], chain); err != nil {
			f.t.Errorf("tenant %d: %v", i, err)
		}
	}
}

// appendConcurrently runs one goroutine per element of events, which appends
// its events to tenantOf(worker) in one transaction each, after a read; it
// returns every error.
func (f *fixture) appendConcurrently(tenantOf func(int) ids.TenantID, events [][]audit.Event) []error {
	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup
	for w := range events {
		wg.Go(func() {
			for _, e := range events[w] {
				err := f.store.InTenant(f.ctx, tenantOf(w), func(ctx context.Context, tx storage.Tx) error {
					if _, err := tx.Organizations().List(ctx, ids.OrganizationID{}, 1); err != nil {
						return err
					}
					_, err := tx.Audit().Append(ctx, e)
					return err
				})
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	return errs
}

// testAuditCrossTenantIsolation: B never reads A's events, and cannot reuse
// an event ID of A (IDs are unique across tenants).
func testAuditCrossTenantIsolation(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	inA := f.appendEvents(a, 2)
	f.wantChain(b, nil)
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		dup := f.auditEvent("test.duplicate")
		dup.ID = inA[0].ID
		_, err := tx.Audit().Append(ctx, dup)
		wantErr(f.t, "Append with A's event ID in B", err, storage.ErrConflict)
		_, err = tx.Audit().Append(ctx, f.auditEvent("test.own")) // the Tx is still usable
		return err
	})
	f.wantChain(a, inA)
	if chain := f.chainOf(b); len(chain) != 1 || chain[0].Sequence != 1 || chain[0].TenantID != b {
		f.t.Errorf("tenant B chain = %+v, want its one event at sequence 1", chain)
	}
}

// testAuditInvalidEventRefused: an invalid event is ErrInvalid wrapping
// audit.ErrInvalid, carries no value in its error, and does not consume a
// sequence.
func testAuditInvalidEventRefused(f *fixture) {
	a := f.tenant("acme")
	for name, change := range map[string]func(*audit.Event){
		"zero ID":        func(e *audit.Event) { e.ID = ids.EventID{} },
		"zero time":      func(e *audit.Event) { e.Time = time.Time{} },
		"bad action":     func(e *audit.Event) { e.Action = "Created" },
		"bad actor type": func(e *audit.Event) { e.Actor.Type = "root" },
		"bad outcome":    func(e *audit.Event) { e.Outcome = "" },
		"bad data key":   func(e *audit.Event) { e.Data = map[string]string{"Bad-Key": "x"} },
		"e-mail in data": func(e *audit.Event) { e.Data = map[string]string{"user": "zq7ana@example.com"} },
	} {
		e := f.auditEvent("test.invalid")
		change(&e)
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			_, err := tx.Audit().Append(ctx, e)
			if !errors.Is(err, storage.ErrInvalid) || !errors.Is(err, audit.ErrInvalid) {
				f.t.Errorf("%s: Append = %v, want storage.ErrInvalid wrapping audit.ErrInvalid", name, err)
			}
			if errors.Is(err, storage.ErrConflict) || errors.Is(err, tenancy.ErrInvalid) {
				f.t.Errorf("%s: %v is also ErrConflict or tenancy.ErrInvalid", name, err)
			}
			if err != nil && bytes.Contains([]byte(err.Error()), []byte("zq7ana")) {
				f.t.Errorf("%s: error %q echoes the value", name, err)
			}
			return nil
		})
	}
	if got := f.appendEvents(a, 1); got[0].Sequence != 1 {
		f.t.Errorf("first valid event after refusals: sequence %d, want 1", got[0].Sequence)
	}
}

// testAuditEventsPaging: Events starts strictly after `after`, honors and
// clamps the limit, and refuses a limit of zero or a negative `after`.
func testAuditEventsPaging(f *fixture) {
	a := f.tenant("acme")
	all := f.appendEvents(a, 5)
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		log := tx.Audit()
		page, err := log.Events(ctx, 2, 2)
		wantErr(f.t, "Events(2, 2)", err, nil)
		wantEvents(f.t, "Events(2, 2)", page, all[2:4])
		page, err = log.Events(ctx, 0, storage.MaxListLimit+1)
		wantErr(f.t, "Events(0, max+1)", err, nil)
		wantEvents(f.t, "Events(0, max+1)", page, all)
		page, err = log.Events(ctx, 5, 10)
		if err != nil || len(page) != 0 {
			f.t.Errorf("Events past the end = %d events, %v; want none", len(page), err)
		}
		_, err = log.Events(ctx, 0, 0)
		wantErr(f.t, "Events limit 0", err, storage.ErrInvalid)
		_, err = log.Events(ctx, -1, 10)
		wantErr(f.t, "Events after -1", err, storage.ErrInvalid)
		return nil
	})
}

// testAuditSystemChain (ruling R1): InSystem audit fails with
// ErrInvalidTenant until the system tenant exists, also in the transaction
// that creates it; afterwards it lands in the system tenant's chain only.
func testAuditSystemChain(f *fixture) {
	a := f.tenant("acme")
	appendSys := func(ctx context.Context, tx storage.SystemTx) (audit.Event, error) {
		return tx.Audit().Append(ctx, f.auditEvent("tenant.created"))
	}
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := appendSys(ctx, tx)
		wantErr(f.t, "system Append before the system tenant", err, storage.ErrInvalidTenant)
		if _, err := tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, systemName, f.clk); err != nil {
			return err
		}
		_, err = appendSys(ctx, tx)
		wantErr(f.t, "system Append in the tx that creates it", err, storage.ErrInvalidTenant)
		return nil
	})
	var sys tenancy.Tenant
	var sealed []audit.Event
	for range 2 {
		f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
			var err error
			if sys, err = tx.SystemTenant(ctx); err != nil {
				return err
			}
			e, err := appendSys(ctx, tx)
			sealed = append(sealed, e)
			return err
		})
	}
	if sealed[0].TenantID != sys.ID || sealed[0].Sequence != 1 || sealed[1].Sequence != 2 {
		f.t.Errorf("system events: tenant %v, sequences %d, %d; want the system tenant %v, 1, 2", sealed[0].TenantID, sealed[0].Sequence, sealed[1].Sequence, sys.ID)
	}
	f.wantChain(sys.ID, sealed)
	f.wantChain(a, nil)
	// A rolled-back system append is dropped too.
	err := f.store.InSystem(f.ctx, func(ctx context.Context, tx storage.SystemTx) error {
		_, err := appendSys(ctx, tx)
		return errors.Join(err, errBoom)
	})
	wantErr(f.t, "InSystem", err, errBoom)
	f.wantChain(sys.ID, sealed)
}

func testAuditUseAfterReturnFails(f *fixture) {
	a := f.tenant("acme")
	var log storage.AuditLog
	f.inTenant(a, func(_ context.Context, tx storage.Tx) error { log = tx.Audit(); return nil })
	_, err := log.Append(f.ctx, f.auditEvent("test.late"))
	wantErr(f.t, "Append after return", err, storage.ErrTxDone)
	_, err = log.Events(f.ctx, 0, 10)
	wantErr(f.t, "Events after return", err, storage.ErrTxDone)
	var sysLog storage.AuditAppender
	f.inSystem(func(_ context.Context, tx storage.SystemTx) error { sysLog = tx.Audit(); return nil })
	_, err = sysLog.Append(f.ctx, f.auditEvent("test.late"))
	wantErr(f.t, "system Append after return", err, storage.ErrTxDone)
	f.wantChain(a, nil)
}
