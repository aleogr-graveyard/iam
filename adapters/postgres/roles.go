package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
)

// ErrInvalidAppRole reports a database role the adapter refuses to run as or
// to grant to: a name that is not a plain identifier starting with iam_ (the
// service never touches roles of other projects, ADR 0001); a role that does
// not exist; a role that is, or can become (pg_has_role MEMBER: inherited or
// through SET ROLE), a superuser, a BYPASSRLS role, the migrating role or an
// owner of anything in the iam schema; or a role whose reachable privileges
// differ from appPrivileges.
var ErrInvalidAppRole = errors.New("postgres: invalid app role")

// querier is what the role checks need: a pool or a transaction.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// reachableRoles is every role $1 can act as: itself and every role it is a
// member of, directly or not, with or without INHERIT.
const reachableRoles = `pg_roles r WHERE pg_has_role($1::oid, r.oid, 'MEMBER')`

// roleRisksSQL reports what the roles reachable from $1 can do around RLS.
// Ownership covers the schema and the relations, functions and types in it.
const roleRisksSQL = `
WITH iam_owners AS (
    SELECT nspowner AS owner FROM pg_namespace WHERE nspname = 'iam'
    UNION SELECT c.relowner FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'iam'
    UNION SELECT p.proowner FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = 'iam'
    UNION SELECT t.typowner FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE n.nspname = 'iam'
)
SELECT coalesce(bool_or(r.rolsuper), false),
       coalesce(bool_or(r.rolbypassrls), false),
       coalesce(bool_or(r.rolname = current_user), false),
       coalesce(bool_or(r.oid IN (SELECT owner FROM iam_owners)), false)
FROM ` + reachableRoles

type roleRisks struct{ super, bypassRLS, currentUser, owner bool }

func risksOf(ctx context.Context, q querier, role uint32) (roleRisks, error) {
	var r roleRisks
	err := q.QueryRow(ctx, roleRisksSQL, role).Scan(&r.super, &r.bypassRLS, &r.currentUser, &r.owner)
	return r, err
}

// unsafe names the first risk, ignoring currentUser unless asked.
func (r roleRisks) unsafe(withCurrentUser bool) error {
	switch {
	case r.super:
		return fmt.Errorf("%w: role is or can become a superuser", ErrInvalidAppRole)
	case r.bypassRLS:
		return fmt.Errorf("%w: role is or can become a role that bypasses row level security", ErrInvalidAppRole)
	case withCurrentUser && r.currentUser:
		return fmt.Errorf("%w: role is or can become the migration role", ErrInvalidAppRole)
	case r.owner:
		return fmt.Errorf("%w: role is or can become an owner of objects in the iam schema", ErrInvalidAppRole)
	}
	return nil
}

// checkAppRole is Migrate's check of the role it grants to; it returns the
// role's OID.
func checkAppRole(ctx context.Context, q querier, appRole string) (uint32, error) {
	var oid uint32
	err := q.QueryRow(ctx, "SELECT oid FROM pg_roles WHERE rolname = $1", appRole).Scan(&oid)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("%w: role does not exist", ErrInvalidAppRole)
	}
	if err != nil {
		return 0, err
	}
	r, err := risksOf(ctx, q, oid)
	if err != nil {
		return 0, err
	}
	return oid, r.unsafe(true)
}

// checkConnectionRole is the store's startup check of the roles it connects
// as (the lab login role iam-run, a member of the app role, included): the
// current user, and the session user, the login, which a connection can always
// return to with RESET ROLE (a superuser login that starts with
// options=-c role=<app> would otherwise pass). It fails closed: any error
// refuses.
func checkConnectionRole(ctx context.Context, q querier) error {
	var current, session uint32
	err := q.QueryRow(ctx, `SELECT (SELECT oid FROM pg_roles WHERE rolname = current_user),
		(SELECT oid FROM pg_roles WHERE rolname = session_user)`).Scan(&current, &session)
	if err != nil {
		return fmt.Errorf("postgres: check connection role: %w", sanitize(err))
	}
	for _, oid := range []uint32{current, session} {
		r, err := risksOf(ctx, q, oid)
		if err != nil {
			return fmt.Errorf("postgres: check connection role: %w", sanitize(err))
		}
		if err := r.unsafe(false); err != nil {
			return err
		}
	}
	return nil
}

// appPrivileges is exactly what the store uses, by table of the iam schema.
// A migration that grants more updates it in the same change (expand).
var appPrivileges = map[string][]string{
	"applications":  {"INSERT", "SELECT"},
	"audit_events":  {"INSERT", "SELECT"},
	"memberships":   {"DELETE", "INSERT", "SELECT"},
	"organizations": {"INSERT", "SELECT"},
	"tenants":       {"INSERT", "SELECT"},
	"users":         {"INSERT", "SELECT"},
}

// reachedPrivilegesSQL lists the table privileges the roles reachable from
// $1 hold on relations of the iam schema, PUBLIC's included.
const reachedPrivilegesSQL = `
SELECT c.relname, p.priv
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
CROSS JOIN unnest(ARRAY['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'REFERENCES', 'TRIGGER']) AS p(priv)
WHERE n.nspname = 'iam' AND c.relkind IN ('r', 'p', 'v', 'm', 'f')
  AND EXISTS (SELECT 1 FROM ` + reachableRoles + ` AND has_table_privilege(r.oid, c.oid, p.priv))
`

// reachedOtherSQL reports the other privileges the roles reachable from $1
// hold in the iam schema: on the schema itself, on any sequence, on any
// function (EXECUTE, PUBLIC's included), and any column-level grant (to one
// of them or to PUBLIC, grantee 0), which table privileges do not show.
const reachedOtherSQL = `
SELECT coalesce(bool_or(has_schema_privilege(r.oid, 'iam', 'USAGE')), false),
       coalesce(bool_or(has_schema_privilege(r.oid, 'iam', 'CREATE')), false),
       coalesce(bool_or(EXISTS (
           SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
           -- CASE: has_sequence_privilege fails on a relation that is not a
           -- sequence, and the planner may test it before relkind.
           WHERE n.nspname = 'iam'
             AND CASE WHEN c.relkind = 'S' THEN has_sequence_privilege(r.oid, c.oid, 'USAGE, SELECT, UPDATE') ELSE false END)), false),
       coalesce(bool_or(EXISTS (
           SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
           WHERE n.nspname = 'iam' AND has_function_privilege(r.oid, p.oid, 'EXECUTE'))), false),
       coalesce(bool_or(EXISTS (
           SELECT 1 FROM pg_attribute a
           JOIN pg_class c ON c.oid = a.attrelid JOIN pg_namespace n ON n.oid = c.relnamespace
           CROSS JOIN LATERAL aclexplode(a.attacl) x
           WHERE n.nspname = 'iam' AND a.attacl IS NOT NULL AND x.grantee IN (r.oid, 0))), false)
FROM ` + reachableRoles

// verifyPrivileges refuses an app role whose reachable privileges differ
// from appPrivileges, plus USAGE without CREATE on the schema, and nothing on
// sequences, functions or single columns: none of them is used today, so
// there is no allow-list for them yet.
func verifyPrivileges(ctx context.Context, q querier, role uint32) error {
	rows, err := q.Query(ctx, reachedPrivilegesSQL, role)
	if err != nil {
		return err
	}
	got, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var table, priv string
		err := row.Scan(&table, &priv)
		return table + " " + priv, err
	})
	if err != nil {
		return err
	}
	var want []string
	for table, privs := range appPrivileges {
		for _, p := range privs {
			want = append(want, table+" "+p)
		}
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		return fmt.Errorf("%w: table privileges differ from the expected set", ErrInvalidAppRole)
	}
	return verifyOtherPrivileges(ctx, q, role)
}

// verifyOtherPrivileges checks what table privileges do not show.
func verifyOtherPrivileges(ctx context.Context, q querier, role uint32) error {
	var usage, create, sequence, function, column bool
	if err := q.QueryRow(ctx, reachedOtherSQL, role).Scan(&usage, &create, &sequence, &function, &column); err != nil {
		return err
	}
	switch {
	case !usage || create:
		return fmt.Errorf("%w: schema privileges differ from USAGE only", ErrInvalidAppRole)
	case sequence:
		return fmt.Errorf("%w: role holds a privilege on a sequence", ErrInvalidAppRole)
	case function:
		return fmt.Errorf("%w: role can execute a function", ErrInvalidAppRole)
	case column:
		return fmt.Errorf("%w: role holds a column-level privilege", ErrInvalidAppRole)
	}
	return nil
}
