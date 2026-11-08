package storagetest

import (
	"context"
	"strings"

	"github.com/aleogr/iam/shared/storage"
)

func errorTests() []test {
	return []test{
		{"Errors/NoPersonalData", testErrorsNoPersonalData},
	}
}

// marker appears in every value below; no error message may contain it.
const marker = "zq7mark"

// testErrorsNoPersonalData collects the errors of every refused operation that
// carries an e-mail address, a slug or a redirect URI, and checks that none
// echoes the value (TM-LOG: errors reach logs and clients).
func testErrorsNoPersonalData(f *fixture) {
	a := f.tenant("acme-" + marker)
	email := "ana." + marker + "@" + marker + ".example.com"
	existing := f.user(email)
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Users().Create(ctx, existing); err != nil {
			return err
		}
		return tx.Organizations().Create(ctx, f.org("org-"+marker))
	})
	var errs []error
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		users := tx.Users()
		errs = append(errs, users.Create(ctx, f.user(email)))                                                    // conflict
		errs = append(errs, users.Create(ctx, f.user("Ana."+marker+"@"+strings.ToUpper(marker)+".example.com"))) // not normalized
		errs = append(errs, users.Create(ctx, f.user(marker+"@@"+marker)))                                       // invalid
		_, err := users.ByEmail(ctx, "bia."+marker+"@example.com")                                               // not found
		errs = append(errs, err)
		_, err = users.ByEmail(ctx, marker+" @ "+marker) // invalid input
		errs = append(errs, err)
		errs = append(errs, tx.Organizations().Create(ctx, f.org("org-"+marker))) // conflict
		errs = append(errs, tx.Organizations().Create(ctx, f.org("Org-"+marker))) // invalid slug
		_, err = tx.Organizations().BySlug(ctx, "other-"+marker)
		errs = append(errs, err)
		app := f.app("App")
		app.RedirectURIs = []string{"http://" + marker + ".example.com/cb"}
		errs = append(errs, tx.Applications().Create(ctx, app))
		return nil
	})
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		errs = append(errs, tx.CreateTenant(ctx, f.tenantValue("acme-"+marker)))
		_, err := tx.TenantBySlug(ctx, "t-"+marker)
		errs = append(errs, err)
		return nil
	})
	for i, err := range errs {
		switch {
		case err == nil:
			f.t.Errorf("operation %d: no error", i)
		case strings.Contains(strings.ToLower(err.Error()), marker):
			f.t.Errorf("operation %d: error %q contains the input", i, err)
		}
	}
}
