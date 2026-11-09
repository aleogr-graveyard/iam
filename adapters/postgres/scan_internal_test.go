package postgres

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/tenancy"
)

// fakeRow hands fixed values to the scan destinations of the adapter.
type fakeRow []any

func (r fakeRow) Scan(dest ...any) error {
	if len(dest) != len(r) {
		return errors.New("fakeRow: column count")
	}
	for i, d := range dest {
		switch d := d.(type) {
		case idDest:
			if err := d.ScanUUID(pgtype.UUID{Bytes: r[i].(ids.ID), Valid: true}); err != nil {
				return err
			}
		case utcDest:
			if err := d.ScanTimestamptz(pgtype.Timestamptz{Time: r[i].(time.Time), Valid: true}); err != nil {
				return err
			}
		default:
			reflect.ValueOf(d).Elem().Set(reflect.ValueOf(r[i]))
		}
	}
	return nil
}

// TestScan_ForeignTenantRowIsCorrupt: a row of another tenant that reached a
// scan (both the query filter and RLS would have to fail first, so no
// database can produce it) is ErrCorrupt, never returned.
func TestScan_ForeignTenantRowIsCorrupt(t *testing.T) {
	own, foreign := ids.TenantID{ID: ids.ID{1}}, ids.TenantID{ID: ids.ID{2}}
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	tx := &tenantTx{tenant: own}
	scans := map[string]func(pgx.Row) error{
		"organization": func(r pgx.Row) error { _, err := orgRepo{tx}.scan(r); return err },
		"user":         func(r pgx.Row) error { _, err := userRepo{tx}.scan(r); return err },
		"membership":   func(r pgx.Row) error { _, err := memberRepo{tx}.scan(r); return err },
		"application":  func(r pgx.Row) error { _, err := appRepo{tx}.scan(r); return err },
	}
	rows := map[string]func(tenant ids.ID) fakeRow{
		"organization": func(tn ids.ID) fakeRow { return fakeRow{ids.ID{3}, tn, "sales", "Sales", now} },
		"user": func(tn ids.ID) fakeRow {
			return fakeRow{ids.ID{4}, tn, "ana@example.com", false, tenancy.StatusActive, now}
		},
		"membership": func(tn ids.ID) fakeRow { return fakeRow{tn, ids.ID{3}, ids.ID{4}, now} },
		"application": func(tn ids.ID) fakeRow {
			return fakeRow{ids.ID{5}, tn, "App", tenancy.ClientPublic, []string{"https://app.example.com/cb"}, []byte(nil), now}
		},
	}
	for entity, scan := range scans {
		if err := scan(rows[entity](own.ID)); err != nil {
			t.Errorf("%s of the Tx's tenant: %v", entity, err)
		}
		if err := scan(rows[entity](foreign.ID)); !errors.Is(err, ErrCorrupt) {
			t.Errorf("%s of another tenant: %v, want ErrCorrupt", entity, err)
		}
	}
}
