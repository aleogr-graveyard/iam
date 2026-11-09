package audit_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
)

var t0 = time.Date(2026, 11, 9, 12, 0, 0, 123456000, time.UTC)

type fixture struct {
	t   *testing.T
	clk *clock.Fixed
	gen ids.Generator
}

func newFixture(t *testing.T) *fixture {
	clk := &clock.Fixed{T: t0}
	return &fixture{t: t, clk: clk, gen: ids.Generator{Clock: clk}}
}

func (f *fixture) id() ids.ID {
	f.t.Helper()
	f.clk.Advance(time.Millisecond)
	id, err := f.gen.New()
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

// event is a valid event at chain position seq, before sealing.
func (f *fixture) event(tenant ids.TenantID, seq int64) audit.Event {
	return audit.Event{
		ID: ids.EventID{ID: f.id()}, TenantID: tenant, Sequence: seq, Time: f.clk.Now(),
		SchemaVersion: audit.SchemaVersion, Action: "organization.created",
		Actor:   audit.Actor{Type: audit.ActorUser, ID: "0190f1a2-0000-7000-8000-000000000001"},
		Target:  "tenants/acme/organizations/sales",
		Outcome: audit.OutcomeSuccess,
		Data:    map[string]string{"slug": "sales", "org_count": "3"},
	}
}

// chain seals n events of one tenant.
func (f *fixture) chain(tenant ids.TenantID, n int) []audit.Event {
	f.t.Helper()
	var out []audit.Event
	var prev []byte
	for i := range n {
		e, err := audit.Seal(prev, f.event(tenant, int64(i+1)))
		if err != nil {
			f.t.Fatalf("Seal %d: %v", i+1, err)
		}
		out = append(out, e)
		prev = e.Hash
	}
	return out
}

func TestSeal_Deterministic(t *testing.T) {
	f := newFixture(t)
	e := f.event(ids.TenantID{ID: f.id()}, 1)
	a, err := audit.Seal(nil, e)
	if err != nil {
		t.Fatal(err)
	}
	// Map iteration order differs between runs; the hash must not.
	for range 20 {
		e2 := e
		e2.Data = maps.Clone(e.Data)
		b, err := audit.Seal(nil, e2)
		if err != nil || !bytes.Equal(a.Hash, b.Hash) {
			t.Fatalf("Seal is not deterministic: %x vs %x (%v)", a.Hash, b.Hash, err)
		}
	}
	if len(a.Hash) != audit.HashBytes {
		t.Errorf("hash length %d, want %d", len(a.Hash), audit.HashBytes)
	}
	if a.PrevHash != nil {
		t.Errorf("first event PrevHash = %x, want empty (genesis)", a.PrevHash)
	}
}

func TestSeal_EveryFieldIsHashed(t *testing.T) {
	f := newFixture(t)
	e := f.event(ids.TenantID{ID: f.id()}, 1)
	a, err := audit.Seal(nil, e)
	if err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*audit.Event){
		"action":   func(e *audit.Event) { e.Action = "organization.deleted" },
		"actor":    func(e *audit.Event) { e.Actor.Type = audit.ActorOperator },
		"target":   func(e *audit.Event) { e.Target = "tenants/acme" },
		"outcome":  func(e *audit.Event) { e.Outcome = audit.OutcomeDenied },
		"data":     func(e *audit.Event) { e.Data = map[string]string{"slug": "sales"} },
		"time":     func(e *audit.Event) { e.Time = e.Time.Add(time.Microsecond) },
		"id":       func(e *audit.Event) { e.ID = ids.EventID{ID: f.id()} },
		"tenant":   func(e *audit.Event) { e.TenantID = ids.TenantID{ID: f.id()} },
		"sequence": func(e *audit.Event) { e.Sequence = 2 },
	} {
		e2 := e
		e2.Data = maps.Clone(e.Data)
		change(&e2)
		prev := []byte(nil)
		if e2.Sequence > 1 {
			prev = a.Hash
		}
		b, err := audit.Seal(prev, e2)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if bytes.Equal(a.Hash, b.Hash) {
			t.Errorf("changing %s does not change the hash", name)
		}
	}
}

// TestSeal_CanonicalForm pins the hashed bytes of schema version 1: a change
// here is a change of contract C4 and needs a new SchemaVersion.
func TestSeal_CanonicalForm(t *testing.T) {
	tenant, _ := ids.Parse("0190f1a2-0000-7000-8000-00000000000a")
	id, _ := ids.Parse("0190f1a2-0000-7000-8000-00000000000b")
	e := audit.Event{
		ID: ids.EventID{ID: id}, TenantID: ids.TenantID{ID: tenant}, Sequence: 1, SchemaVersion: 1,
		Time: time.Date(2026, 11, 9, 12, 0, 0, 5000, time.UTC), Action: "tenant.created",
		Actor: audit.Actor{Type: audit.ActorSystem, ID: "bootstrap"}, Target: "tenants/acme",
		Outcome: audit.OutcomeSuccess, Data: map[string]string{"slug": "acme"},
	}
	got, err := audit.Canonical(e)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"action":"tenant.created","actor":{"id":"bootstrap","type":"system"},"data":{"slug":"acme"},` +
		`"id":"0190f1a2-0000-7000-8000-00000000000b","outcome":"success","prev_hash":"","schema_version":1,` +
		`"sequence":1,"target":"tenants/acme","tenant_id":"0190f1a2-0000-7000-8000-00000000000a",` +
		`"time":"2026-11-09T12:00:00.000005Z"}`
	if string(got) != want {
		t.Fatalf("canonical form:\n got  %s\n want %s", got, want)
	}
	sealed, err := audit.Seal(nil, e)
	if err != nil {
		t.Fatal(err)
	}
	// SHA-256 of the bytes above (empty PrevHash: the genesis), computed
	// independently with sha256sum.
	const wantHash = "2127342acbe9e2917e2a9cfa32d4d45f383a7d48bcb08329ca5968e2fde5142c"
	if hex.EncodeToString(sealed.Hash) != wantHash {
		t.Errorf("hash = %x, want %s", sealed.Hash, wantHash)
	}
	// Second vector: sequence 2 linked to the first, so prev is both hashed
	// in front of the canonical form and written into it; also computed with
	// an independent SHA-256 over the hand-written bytes.
	id2, _ := ids.Parse("0190f1a2-0000-7000-8000-00000000000d")
	next := audit.Event{
		ID: ids.EventID{ID: id2}, TenantID: e.TenantID, Sequence: 2, SchemaVersion: 1,
		Time: time.Date(2026, 11, 9, 12, 0, 1, 0, time.UTC), Action: "organization.created",
		Actor:  audit.Actor{Type: audit.ActorUser, ID: "0190f1a2-0000-7000-8000-00000000000c"},
		Target: "tenants/acme/organizations/sales", Outcome: audit.OutcomeSuccess,
		Data: map[string]string{"slug": "sales", "org_count": "2"},
	}
	sealed2, err := audit.Seal(sealed.Hash, next)
	if err != nil {
		t.Fatal(err)
	}
	const wantHash2 = "7a64d83c1184f39f1bea64a0e9ffd472c6beb73865d29be1ff3848f3aa1900da"
	if hex.EncodeToString(sealed2.Hash) != wantHash2 {
		t.Errorf("sequence 2 hash = %x, want %s", sealed2.Hash, wantHash2)
	}
	if err := audit.VerifyChain(e.TenantID, []audit.Event{sealed, sealed2}); err != nil {
		t.Errorf("the two vectors: %v", err)
	}
	// Nil and empty Data hash alike.
	e.Data = nil
	a, _ := audit.Seal(nil, e)
	e.Data = map[string]string{}
	b, _ := audit.Seal(nil, e)
	if !bytes.Equal(a.Hash, b.Hash) {
		t.Error("nil and empty Data hash differently")
	}
}

func TestSeal_LinksAndDoesNotAlias(t *testing.T) {
	f := newFixture(t)
	evs := f.chain(ids.TenantID{ID: f.id()}, 2)
	if !bytes.Equal(evs[1].PrevHash, evs[0].Hash) {
		t.Fatal("second event does not link to the first")
	}
	prev := bytes.Clone(evs[0].Hash)
	e := f.event(evs[0].TenantID, 2)
	sealed, err := audit.Seal(prev, e)
	if err != nil {
		t.Fatal(err)
	}
	prev[0] ^= 0xff
	if bytes.Equal(sealed.PrevHash, prev) {
		t.Error("PrevHash aliases the caller's slice")
	}
	// A stale Hash in the input is ignored, not hashed.
	e.Hash = []byte("stale")
	again, err := audit.Seal(evs[0].Hash, e)
	if err != nil || !bytes.Equal(again.Hash, sealed.Hash) {
		t.Errorf("Seal with a stale Hash: %x, %v; want %x", again.Hash, err, sealed.Hash)
	}
}

func TestSeal_RefusesWrongChainPosition(t *testing.T) {
	f := newFixture(t)
	tenant := ids.TenantID{ID: f.id()}
	if _, err := audit.Seal(make([]byte, 32), f.event(tenant, 1)); !errors.Is(err, audit.ErrInvalid) {
		t.Errorf("sequence 1 with a PrevHash: %v, want ErrInvalid", err)
	}
	if _, err := audit.Seal(nil, f.event(tenant, 2)); !errors.Is(err, audit.ErrInvalid) {
		t.Errorf("sequence 2 without a PrevHash: %v, want ErrInvalid", err)
	}
	if _, err := audit.Seal(make([]byte, 31), f.event(tenant, 2)); !errors.Is(err, audit.ErrInvalid) {
		t.Errorf("short PrevHash: %v, want ErrInvalid", err)
	}
}

func TestValidate(t *testing.T) {
	f := newFixture(t)
	valid := f.event(ids.TenantID{ID: f.id()}, 1)
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid event: %v", err)
	}
	long := func(n int) string { return strings.Repeat("a", n) }
	manyKeys := map[string]string{}
	for i := range audit.MaxDataKeys + 1 {
		manyKeys["k"+strings.Repeat("x", i)] = "v"
	}
	for name, change := range map[string]func(*audit.Event){
		"zero ID":              func(e *audit.Event) { e.ID = ids.EventID{} },
		"zero tenant":          func(e *audit.Event) { e.TenantID = ids.TenantID{} },
		"sequence 0":           func(e *audit.Event) { e.Sequence = 0 },
		"negative sequence":    func(e *audit.Event) { e.Sequence = -1 },
		"schema version 0":     func(e *audit.Event) { e.SchemaVersion = 0 },
		"schema version 2":     func(e *audit.Event) { e.SchemaVersion = 2 },
		"zero time":            func(e *audit.Event) { e.Time = time.Time{} },
		"nanoseconds":          func(e *audit.Event) { e.Time = e.Time.Add(time.Nanosecond) },
		"not UTC":              func(e *audit.Event) { e.Time = e.Time.In(time.FixedZone("BRT", -3*3600)) },
		"year 10000":           func(e *audit.Event) { e.Time = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"before 1970":          func(e *audit.Event) { e.Time = time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC) },
		"empty action":         func(e *audit.Event) { e.Action = "" },
		"undotted action":      func(e *audit.Event) { e.Action = "created" },
		"upper-case action":    func(e *audit.Event) { e.Action = "Tenant.created" },
		"action with space":    func(e *audit.Event) { e.Action = "tenant .created" },
		"trailing dot":         func(e *audit.Event) { e.Action = "tenant." },
		"long action":          func(e *audit.Event) { e.Action = "a." + long(audit.MaxActionBytes) },
		"unknown actor type":   func(e *audit.Event) { e.Actor.Type = "admin" },
		"empty actor ID":       func(e *audit.Event) { e.Actor.ID = "" },
		"long actor ID":        func(e *audit.Event) { e.Actor.ID = long(audit.MaxActorIDBytes + 1) },
		"actor ID control":     func(e *audit.Event) { e.Actor.ID = "a\nb" },
		"actor ID e-mail":      func(e *audit.Event) { e.Actor.ID = "ana@example.com" },
		"unknown outcome":      func(e *audit.Event) { e.Outcome = "ok" },
		"long target":          func(e *audit.Event) { e.Target = long(audit.MaxTargetBytes + 1) },
		"target control":       func(e *audit.Event) { e.Target = "tenants/\x00" },
		"target invalid UTF-8": func(e *audit.Event) { e.Target = "tenants/\xff" },
		"target e-mail":        func(e *audit.Event) { e.Target = "tenants/acme/users/ana@example.com" },
		"too many keys":        func(e *audit.Event) { e.Data = manyKeys },
		"empty key":            func(e *audit.Event) { e.Data = map[string]string{"": "v"} },
		"upper-case key":       func(e *audit.Event) { e.Data = map[string]string{"Slug": "v"} },
		"key with digit first": func(e *audit.Event) { e.Data = map[string]string{"1a": "v"} },
		"long key":             func(e *audit.Event) { e.Data = map[string]string{"a" + long(64): "v"} },
		"long value":           func(e *audit.Event) { e.Data = map[string]string{"k": long(audit.MaxDataValueBytes + 1)} },
		"value control":        func(e *audit.Event) { e.Data = map[string]string{"k": "a\u0085b"} },
		"value invalid UTF-8":  func(e *audit.Event) { e.Data = map[string]string{"k": "\xc3"} },
		"target bidi override": func(e *audit.Event) { e.Target = "tenants/\u202eacme" },
		"actor ID zero width":  func(e *audit.Event) { e.Actor.ID = "op\u200b1" },
		"value line separator": func(e *audit.Event) { e.Data = map[string]string{"k": "a\u2028b"} },
		"value para separator": func(e *audit.Event) { e.Data = map[string]string{"k": "a\u2029b"} },
		"value BOM":            func(e *audit.Event) { e.Data = map[string]string{"k": "\ufeffa"} },
		"value e-mail":         func(e *audit.Event) { e.Data = map[string]string{"k": "ana@example.com"} },
		"value e-mail in text": func(e *audit.Event) { e.Data = map[string]string{"k": "sent to ana@mail.example.org"} },
		"sequence 1 prev hash": func(e *audit.Event) { e.PrevHash = make([]byte, 32) },
		"wrong hash length":    func(e *audit.Event) { e.Hash = make([]byte, 31) },
	} {
		e := valid
		e.Data = maps.Clone(valid.Data)
		change(&e)
		err := e.Validate()
		if !errors.Is(err, audit.ErrInvalid) {
			t.Errorf("%s: Validate = %v, want ErrInvalid", name, err)
			continue
		}
		for _, leak := range []string{"ana@", "example.com", "aaaaaaaa", "\xff"} {
			if strings.Contains(err.Error(), leak) {
				t.Errorf("%s: error %q echoes the value", name, err)
			}
		}
	}
	for name, change := range map[string]func(*audit.Event){
		"empty target":         func(e *audit.Event) { e.Target = "" },
		"nil data":             func(e *audit.Event) { e.Data = nil },
		"at sign without dot":  func(e *audit.Event) { e.Data = map[string]string{"k": "v@1"} },
		"max key":              func(e *audit.Event) { e.Data = map[string]string{"a" + long(63): "v"} },
		"max value":            func(e *audit.Event) { e.Data = map[string]string{"k": long(audit.MaxDataValueBytes)} },
		"non-ASCII value":      func(e *audit.Event) { e.Data = map[string]string{"k": "São Paulo"} },
		"every actor type":     func(e *audit.Event) { e.Actor.Type = audit.ActorApplication },
		"three-part action":    func(e *audit.Event) { e.Action = "authn.password.verified" },
		"underscore in action": func(e *audit.Event) { e.Action = "tenant.status_changed" },
	} {
		e := valid
		e.Data = maps.Clone(valid.Data)
		change(&e)
		if err := e.Validate(); err != nil {
			t.Errorf("%s: Validate = %v, want nil", name, err)
		}
	}
}

func TestVerifyChain_AcceptsValidChains(t *testing.T) {
	f := newFixture(t)
	tenant := ids.TenantID{ID: f.id()}
	for _, n := range []int{0, 1, 5} {
		if err := audit.VerifyChain(tenant, f.chain(tenant, n)); err != nil {
			t.Errorf("chain of %d: %v", n, err)
		}
	}
	// An intact chain of another tenant is not this tenant's chain.
	other := f.chain(ids.TenantID{ID: f.id()}, 2)
	if err := audit.VerifyChain(tenant, other); !errors.Is(err, audit.ErrBrokenChain) {
		t.Errorf("another tenant's chain: %v, want ErrBrokenChain", err)
	}
	if err := audit.VerifyChain(ids.TenantID{}, nil); !errors.Is(err, audit.ErrBrokenChain) {
		t.Errorf("zero tenant: %v, want ErrBrokenChain", err)
	}
}

func TestVerifyChain_DetectsTampering(t *testing.T) {
	f := newFixture(t)
	tenant := ids.TenantID{ID: f.id()}
	other := f.chain(ids.TenantID{ID: f.id()}, 3)
	for name, tamper := range map[string]func([]audit.Event) []audit.Event{
		"modify data": func(evs []audit.Event) []audit.Event {
			evs[1].Data = map[string]string{"slug": "other"}
			return evs
		},
		"add data key": func(evs []audit.Event) []audit.Event {
			evs[1].Data = maps.Clone(evs[1].Data)
			evs[1].Data["extra"] = "x"
			return evs
		},
		"modify action": func(evs []audit.Event) []audit.Event { evs[2].Action = "organization.deleted"; return evs },
		"modify time":   func(evs []audit.Event) []audit.Event { evs[0].Time = evs[0].Time.Add(time.Second); return evs },
		"modify hash": func(evs []audit.Event) []audit.Event {
			evs[3].Hash = bytes.Clone(evs[3].Hash)
			evs[3].Hash[0] ^= 1
			return evs
		},
		"reorder":       func(evs []audit.Event) []audit.Event { evs[1], evs[2] = evs[2], evs[1]; return evs },
		"drop middle":   func(evs []audit.Event) []audit.Event { return append(evs[:2:2], evs[3:]...) },
		"drop first":    func(evs []audit.Event) []audit.Event { return evs[1:] },
		"change prev":   func(evs []audit.Event) []audit.Event { evs[2].PrevHash = bytes.Clone(evs[0].Hash); return evs },
		"duplicate seq": func(evs []audit.Event) []audit.Event { return append(evs[:3:3], evs[2:]...) },
		"insert event": func(evs []audit.Event) []audit.Event {
			ins, err := audit.Seal(evs[1].Hash, f.event(evs[0].TenantID, 3))
			if err != nil {
				t.Fatal(err)
			}
			return append(append(evs[:2:2], ins), evs[2:]...)
		},
		"resealed edit": func(evs []audit.Event) []audit.Event {
			// Recomputing the edited event's hash breaks the next link.
			e := evs[1]
			e.Target = "tenants/evil"
			resealed, err := audit.Seal(e.PrevHash, e)
			if err != nil {
				t.Fatal(err)
			}
			evs[1] = resealed
			return evs
		},
		"mixed tenant": func(evs []audit.Event) []audit.Event { evs[2] = other[2]; return evs },
		// Last, linked and hashed correctly: only the tenant check sees it.
		"foreign event resealed last": func(evs []audit.Event) []audit.Event {
			sealed, err := audit.Seal(evs[3].Hash, f.event(other[0].TenantID, 5))
			if err != nil {
				t.Fatal(err)
			}
			evs[4] = sealed
			return evs
		},
		"schema version": func(evs []audit.Event) []audit.Event { evs[1].SchemaVersion = 2; return evs },
		"sequence shift": func(evs []audit.Event) []audit.Event {
			for i := range evs {
				evs[i].Sequence++
			}
			return evs
		},
	} {
		evs := tamper(f.chain(tenant, 5))
		if err := audit.VerifyChain(tenant, evs); !errors.Is(err, audit.ErrBrokenChain) {
			t.Errorf("%s: VerifyChain = %v, want ErrBrokenChain", name, err)
		}
	}
}
