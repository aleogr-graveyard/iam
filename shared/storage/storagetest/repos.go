package storagetest

import (
	"context"
	"time"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

func repoTests() []test {
	return []test{
		{"Organization/CreateAndGet", testOrganizationCreateAndGet},
		{"Organization/Conflicts", testOrganizationConflicts},
		{"Organization/CreateValidates", testOrganizationCreateValidates},
		{"Organization/NotFound", testOrganizationNotFound},
		{"User/CreateAndGet", testUserCreateAndGet},
		{"User/ByEmailNormalizesInput", testUserByEmailNormalizesInput},
		{"User/Conflicts", testUserConflicts},
		{"User/CreateValidates", testUserCreateValidates},
		{"User/NotFound", testUserNotFound},
		{"Application/CreateAndGet", testApplicationCreateAndGet},
		{"Application/Conflicts", testApplicationConflicts},
		{"Application/CreateValidates", testApplicationCreateValidates},
		{"Application/NotFound", testApplicationNotFound},
		{"Application/ReadsReturnCopies", testApplicationReadsReturnCopies},
	}
}

// --- organizations

func testOrganizationCreateAndGet(f *fixture) {
	a := f.tenant("acme")
	want := f.org("sales")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, want) })
	want.TenantID = a
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Organizations().Get(ctx, want.ID)
		wantErr(f.t, "Get", err, nil)
		wantEqual(f.t, "Get", got, want)
		got, err = tx.Organizations().BySlug(ctx, "sales")
		wantErr(f.t, "BySlug", err, nil)
		wantEqual(f.t, "BySlug", got, want)
		return nil
	})
}

func testOrganizationConflicts(f *fixture) {
	a := f.tenant("acme")
	first := f.org("sales")
	sameSlug := f.org("sales")
	sameID := f.org("support")
	sameID.ID = first.ID
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, first) })
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "duplicate slug", tx.Organizations().Create(ctx, sameSlug), storage.ErrConflict)
		wantErr(f.t, "duplicate ID", tx.Organizations().Create(ctx, sameID), storage.ErrConflict)
		return nil
	})
}

func testOrganizationCreateValidates(f *fixture) {
	a := f.tenant("acme")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, f.org("sales")) })
	cases := map[string]func(*tenancy.Organization){
		"zero ID":          func(o *tenancy.Organization) { o.ID = ids.OrganizationID{} },
		"bad slug":         func(o *tenancy.Organization) { o.Slug = "Bad Slug" },
		"reserved slug":    func(o *tenancy.Organization) { o.Slug = tenancy.SystemTenantSlug },
		"empty name":       func(o *tenancy.Organization) { o.Name = "" },
		"zero created_at":  func(o *tenancy.Organization) { o.CreatedAt = time.Time{} },
		"invalid with dup": func(o *tenancy.Organization) { o.Slug, o.Name = "sales", "" },
	}
	for what, mutate := range cases {
		bad := f.org("fresh")
		mutate(&bad)
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			wantInvalid(f.t, what, tx.Organizations().Create(ctx, bad))
			return nil
		})
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		list, err := tx.Organizations().List(ctx, ids.OrganizationID{}, storage.MaxListLimit)
		if err == nil && len(list) != 1 {
			f.t.Errorf("List has %d organizations, want 1: an invalid one was stored", len(list))
		}
		return err
	})
}

func testOrganizationNotFound(f *fixture) {
	a := f.tenant("acme")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, f.org("sales")) })
	unknown := ids.OrganizationID{ID: f.newID()}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, ids.OrganizationID{})
		wantErr(f.t, "Get(zero)", err, storage.ErrNotFound)
		_, err = tx.Organizations().Get(ctx, unknown)
		wantErr(f.t, "Get(unknown)", err, storage.ErrNotFound)
		for _, slug := range []string{"nope", "", "Sales", " sales", tenancy.SystemTenantSlug} {
			_, err = tx.Organizations().BySlug(ctx, slug)
			wantErr(f.t, "BySlug("+slug+")", err, storage.ErrNotFound)
		}
		return nil
	})
}

// --- users

func testUserCreateAndGet(f *fixture) {
	a := f.tenant("acme")
	want := f.user("ana@example.com")
	want.EmailVerified = true
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Users().Create(ctx, want) })
	want.TenantID = a
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Users().Get(ctx, want.ID)
		wantErr(f.t, "Get", err, nil)
		wantEqual(f.t, "Get", got, want)
		got, err = tx.Users().ByEmail(ctx, "ana@example.com")
		wantErr(f.t, "ByEmail", err, nil)
		wantEqual(f.t, "ByEmail", got, want)
		return nil
	})
}

// testUserByEmailNormalizesInput: the domain is mapped to its A-label and
// lowercased; the local part keeps its case (tenancy.NormalizeEmail), so
// "A@..." and "a@..." are two mailboxes.
func testUserByEmailNormalizesInput(f *fixture) {
	a := f.tenant("acme")
	lower := f.user("a@xn--bcher-kva.de")
	upper := f.user("A@xn--bcher-kva.de")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Users().Create(ctx, lower); err != nil {
			return err
		}
		return tx.Users().Create(ctx, upper)
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		for input, want := range map[string]ids.UserID{
			"a@Bücher.DE":        lower.ID,
			" a@BÜCHER.de\t":     lower.ID,
			"a@xn--BCHER-KVA.de": lower.ID,
			"A@Bücher.DE":        upper.ID,
			"A@bücher.de":       upper.ID, // NFD domain
		} {
			got, err := tx.Users().ByEmail(ctx, input)
			if err != nil || got.ID != want {
				f.t.Errorf("ByEmail(%q) = %v, %v; want %v", input, got.ID, err, want)
			}
		}
		return nil
	})
}

// testUserConflicts also checks uniqueness among rows created in the same
// transaction.
func testUserConflicts(f *fixture) {
	a := f.tenant("acme")
	first := f.user("ana@example.com")
	sameID := f.user("bia@example.com")
	sameID.ID = first.ID
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		users := tx.Users()
		if err := users.Create(ctx, first); err != nil {
			return err
		}
		wantErr(f.t, "duplicate e-mail", users.Create(ctx, f.user(first.Email)), storage.ErrConflict)
		wantErr(f.t, "duplicate ID", users.Create(ctx, sameID), storage.ErrConflict)
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		list, err := tx.Users().List(ctx, ids.UserID{}, storage.MaxListLimit)
		if err == nil && len(list) != 1 {
			f.t.Errorf("List has %d users, want 1", len(list))
		}
		return err
	})
}

func testUserCreateValidates(f *fixture) {
	a := f.tenant("acme")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return tx.Users().Create(ctx, f.user("ana@example.com"))
	})
	cases := map[string]func(*tenancy.User){
		"zero ID":            func(u *tenancy.User) { u.ID = ids.UserID{} },
		"invalid e-mail":     func(u *tenancy.User) { u.Email = "not-an-email" },
		"not normalized":     func(u *tenancy.User) { u.Email = "bia@Example.COM" },
		"unknown status":     func(u *tenancy.User) { u.Status = "deleted" },
		"zero created_at":    func(u *tenancy.User) { u.CreatedAt = time.Time{} },
		"not normalized dup": func(u *tenancy.User) { u.Email = "ana@EXAMPLE.com" },
	}
	for what, mutate := range cases {
		bad := f.user("bia@example.com")
		mutate(&bad)
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			wantInvalid(f.t, what, tx.Users().Create(ctx, bad))
			return nil
		})
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		list, err := tx.Users().List(ctx, ids.UserID{}, storage.MaxListLimit)
		if err == nil && len(list) != 1 {
			f.t.Errorf("List has %d users, want 1: an invalid one was stored", len(list))
		}
		return err
	})
}

func testUserNotFound(f *fixture) {
	a := f.tenant("acme")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return tx.Users().Create(ctx, f.user("ana@example.com"))
	})
	unknown := ids.UserID{ID: f.newID()}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Users().Get(ctx, ids.UserID{})
		wantErr(f.t, "Get(zero)", err, storage.ErrNotFound)
		_, err = tx.Users().Get(ctx, unknown)
		wantErr(f.t, "Get(unknown)", err, storage.ErrNotFound)
		for _, email := range []string{"bia@example.com", "Ana@example.com", "", "ana", "ana@@example.com", "ana@example.com\x00"} {
			_, err = tx.Users().ByEmail(ctx, email)
			wantErr(f.t, "ByEmail", err, storage.ErrNotFound)
		}
		return nil
	})
}

// --- applications

func testApplicationCreateAndGet(f *fixture) {
	a := f.tenant("acme")
	for _, want := range []tenancy.Application{f.app("Public"), f.confidentialApp("Confidential")} {
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Applications().Create(ctx, want) })
		want.TenantID = a
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			got, err := tx.Applications().Get(ctx, want.ID)
			wantErr(f.t, "Get "+want.Name, err, nil)
			wantEqual(f.t, "Get "+want.Name, got, want)
			return nil
		})
	}
}

func testApplicationConflicts(f *fixture) {
	a := f.tenant("acme")
	first := f.app("First")
	sameID := f.app("Second")
	sameID.ID = first.ID
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Applications().Create(ctx, first) })
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "duplicate ID", tx.Applications().Create(ctx, sameID), storage.ErrConflict)
		return nil
	})
}

func testApplicationCreateValidates(f *fixture) {
	a := f.tenant("acme")
	cases := map[string]func(*tenancy.Application){
		"zero ID":           func(x *tenancy.Application) { x.ID = ids.ApplicationID{} },
		"empty name":        func(x *tenancy.Application) { x.Name = "" },
		"unknown type":      func(x *tenancy.Application) { x.Type = "other" },
		"secret on public":  func(x *tenancy.Application) { x.SecretHash = make([]byte, 32) },
		"no redirect":       func(x *tenancy.Application) { x.RedirectURIs = nil },
		"wildcard redirect": func(x *tenancy.Application) { x.RedirectURIs = []string{"https://*.example.com/cb"} },
		"insecure http by default": func(x *tenancy.Application) {
			x.RedirectURIs = []string{"http://app.example.com/cb"}
		},
		"zero created_at": func(x *tenancy.Application) { x.CreatedAt = time.Time{} },
	}
	for what, mutate := range cases {
		bad := f.app("Bad")
		mutate(&bad)
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			wantInvalid(f.t, what, tx.Applications().Create(ctx, bad))
			return nil
		})
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		list, err := tx.Applications().List(ctx, ids.ApplicationID{}, storage.MaxListLimit)
		if err == nil && len(list) != 0 {
			f.t.Errorf("List has %d applications, want 0: an invalid one was stored", len(list))
		}
		return err
	})
}

func testApplicationNotFound(f *fixture) {
	a := f.tenant("acme")
	unknown := ids.ApplicationID{ID: f.newID()}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Applications().Get(ctx, ids.ApplicationID{})
		wantErr(f.t, "Get(zero)", err, storage.ErrNotFound)
		_, err = tx.Applications().Get(ctx, unknown)
		wantErr(f.t, "Get(unknown)", err, storage.ErrNotFound)
		return nil
	})
}

// testApplicationReadsReturnCopies: neither the caller's input after Create
// nor a returned value aliases stored data.
func testApplicationReadsReturnCopies(f *fixture) {
	a := f.tenant("acme")
	in := f.confidentialApp("Copy")
	want := in
	want.TenantID = a
	want.RedirectURIs = append([]string(nil), in.RedirectURIs...)
	want.SecretHash = append([]byte(nil), in.SecretHash...)
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Applications().Create(ctx, in); err != nil {
			return err
		}
		in.RedirectURIs[0], in.SecretHash[0] = "https://evil.example/cb", 0xff
		got, err := tx.Applications().Get(ctx, in.ID)
		if err != nil {
			return err
		}
		got.RedirectURIs[1], got.SecretHash[1] = "https://evil.example/cb2", 0xfe
		list, err := tx.Applications().List(ctx, ids.ApplicationID{}, 10)
		if err != nil || len(list) != 1 {
			f.t.Fatalf("List = %d items, %v", len(list), err)
		}
		list[0].RedirectURIs[0], list[0].SecretHash[2] = "https://evil.example/cb3", 0xfd
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Applications().Get(ctx, want.ID)
		wantErr(f.t, "Get", err, nil)
		wantEqual(f.t, "stored application after mutating copies", got, want)
		return nil
	})
}
