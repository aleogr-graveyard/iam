package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aleogr/iam/adapters/memory"
	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/storage/storagetest"
	"github.com/aleogr/iam/shared/tenancy"
)

var _ storage.Store = (*memory.Store)(nil)

func TestStore_Contract(t *testing.T) {
	storagetest.Run(t, func(*testing.T) storage.Store { return memory.New() })
}

// The contract suite checks the default (insecure http refused); the option
// set from configuration in development accepts it.
func TestStore_AllowInsecureHTTPOption(t *testing.T) {
	gen := ids.Generator{Clock: &clock.Fixed{T: time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)}}
	newID := func() ids.ID {
		id, err := gen.New()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		store *memory.Store
		ok    bool
	}{
		{memory.New(), false},
		{memory.New(memory.WithAllowInsecureHTTP(false)), false},
		{memory.New(memory.WithAllowInsecureHTTP(true)), true},
	} {
		tenant := tenancy.Tenant{ID: ids.TenantID{ID: newID()}, Slug: "acme", Name: "Acme", Status: tenancy.StatusActive, CreatedAt: now}
		if err := tc.store.InSystem(t.Context(), func(ctx context.Context, tx storage.SystemTx) error {
			return tx.CreateTenant(ctx, tenant)
		}); err != nil {
			t.Fatal(err)
		}
		app := tenancy.Application{
			ID: ids.ApplicationID{ID: newID()}, Name: "Dev", Type: tenancy.ClientPublic,
			RedirectURIs: []string{"http://app.example.com/cb"}, CreatedAt: now,
		}
		err := tc.store.InTenant(t.Context(), tenant.ID, func(ctx context.Context, tx storage.Tx) error {
			return tx.Applications().Create(ctx, app)
		})
		if tc.ok && err != nil {
			t.Errorf("allowed: Create = %v, want nil", err)
		}
		if !tc.ok && !errors.Is(err, storage.ErrInvalid) {
			t.Errorf("refused: Create = %v, want ErrInvalid", err)
		}
	}
}
