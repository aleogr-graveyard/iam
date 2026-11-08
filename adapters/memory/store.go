// Package memory is the in-memory storage adapter (contract C3): the reference
// implementation the contract suite (storagetest) is written against, and the
// store of unit tests. Nothing survives the process.
//
// Transactions are serialized by one mutex held while fn runs. Each one works
// on a copy of the data it may change (copy-on-write at the scope of one
// tenant, or of the tenant registry) and swaps it in only when fn returns nil,
// so an error or a panic leaves the committed data untouched. Stored values
// never alias the caller's: slices are copied on the way in and on the way
// out. IDs are unique across tenants: a create checks the committed rows of
// every other tenant, which cannot change while the mutex is held. A nested
// InTenant or InSystem made with the ctx given to fn is refused
// (storage.ErrNestedTx); one made with another ctx deadlocks.
package memory

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
)

// Store is safe for concurrent use.
type Store struct {
	allowInsecureHTTP bool

	mu      sync.Mutex
	reg     registry
	tenants map[ids.ID]*tenantData // committed rows of each tenant; nil until its first commit
}

// Option configures a Store.
type Option func(*Store)

// WithAllowInsecureHTTP sets whether applications may have http redirect URIs
// other than loopback (tenancy.Application.Validate). It comes from
// configuration and is for development only; the default is false.
func WithAllowInsecureHTTP(allow bool) Option {
	return func(s *Store) { s.allowInsecureHTTP = allow }
}

// New returns an empty Store.
func New(opts ...Option) *Store {
	s := &Store{reg: newRegistry(), tenants: map[ids.ID]*tenantData{}}
	for _, o := range opts {
		o(s)
	}
	return s
}

// InTenant implements storage.Store.
func (s *Store) InTenant(ctx context.Context, tenant ids.TenantID, fn func(context.Context, storage.Tx) error) error {
	txCtx, err := begin(ctx)
	if err != nil {
		return err
	}
	if tenant.IsZero() {
		return storage.ErrInvalidTenant
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.reg.tenants.byID[tenant.ID]; !ok {
		return storage.ErrInvalidTenant
	}
	tx := &tenantTx{tenant: tenant, data: s.tenants[tenant.ID].clone(), all: s.tenants, allowInsecureHTTP: s.allowInsecureHTTP}
	defer tx.finish()
	if err := run(txCtx, func(ctx context.Context) error { return fn(ctx, tx) }); err != nil {
		return err
	}
	s.tenants[tenant.ID] = tx.data
	return nil
}

// InSystem implements storage.Store.
func (s *Store) InSystem(ctx context.Context, fn func(context.Context, storage.SystemTx) error) error {
	txCtx, err := begin(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := &systemTx{reg: s.reg.clone()}
	defer tx.finish()
	if err := run(txCtx, func(ctx context.Context) error { return fn(ctx, tx) }); err != nil {
		return err
	}
	s.reg = tx.reg
	return nil
}

// begin refuses a canceled ctx and a nested transaction.
func begin(ctx context.Context) (context.Context, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return storage.EnterTx(ctx)
}

// run calls fn and refuses to commit when ctx ended meanwhile, as a database
// transaction bound to ctx would.
func run(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	return ctx.Err()
}

// txState marks the end of a transaction; methods called afterwards fail.
type txState struct{ done atomic.Bool }

func (s *txState) finish() { s.done.Store(true) }

// check is called first by every method of a Tx or SystemTx.
func (s *txState) check(ctx context.Context) error {
	if s.done.Load() {
		return storage.ErrTxDone
	}
	return ctx.Err()
}
