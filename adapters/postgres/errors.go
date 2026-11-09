package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aleogr/iam/shared/storage"
)

// ErrCorrupt reports a row read back from the database that fails its
// entity's validation. It is never silently skipped.
var ErrCorrupt = errors.New("postgres: stored row is invalid")

// SQLSTATE codes the adapter maps to storage errors.
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
)

// dbError hides the driver's error behind a fixed message. Driver messages
// and details can carry values (an e-mail in "Key (email)=(...) already
// exists", a value pgx could not encode), so Error never prints the cause;
// Unwrap keeps it for errors.Is and errors.As (context.Canceled among them).
type dbError struct {
	msg   string
	cause error
}

func (e *dbError) Error() string { return e.msg }
func (e *dbError) Unwrap() error { return e.cause }

// sanitize turns a driver error into one that is safe to log and return.
// Context errors keep their identity, a closed transaction becomes
// storage.ErrTxDone and a server error keeps only its SQLSTATE.
func sanitize(err error) error {
	var pgErr *pgconn.PgError
	var connErr *pgconn.ConnectError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, storage.ErrTxDone), errors.Is(err, pgx.ErrTxClosed):
		return storage.ErrTxDone
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return &dbError{msg: "postgres: " + contextErr(err).Error(), cause: err}
	case errors.As(err, &pgErr):
		return &dbError{msg: "postgres: database error (SQLSTATE " + pgErr.Code + ")", cause: err}
	case errors.As(err, &connErr):
		return &dbError{msg: "postgres: cannot connect to the database", cause: err}
	case errors.Is(err, ErrInvalidAppRole), errors.Is(err, ErrCorrupt), errors.Is(err, ErrMigrationChanged):
		return err // fixed texts of this package
	}
	return &dbError{msg: "postgres: database error", cause: err}
}

func contextErr(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return context.Canceled
}

// constraintErrors maps each constraint the adapter relies on to the storage
// error it means. Only the entity and the field are named, never a value.
var constraintErrors = map[string]error{
	"tenants_pkey":                        conflict("tenant", "id"),
	"tenants_slug_key":                    conflict("tenant", "slug"),
	"organizations_pkey":                  conflict("organization", "id"),
	"organizations_tenant_id_slug_key":    conflict("organization", "slug"),
	"users_pkey":                          conflict("user", "id"),
	"users_tenant_id_email_key":           conflict("user", "email"),
	"applications_pkey":                   conflict("application", "id"),
	"memberships_pkey":                    conflict("membership", "organization_id, user_id"),
	"memberships_organization_fkey":       notFound("organization"),
	"memberships_user_fkey":               notFound("user"),
	"audit_events_pkey":                   conflict("audit event", "id"),
	"audit_events_tenant_id_sequence_key": conflict("audit event", "sequence"),
}

// writeError maps a failed write: a known unique or foreign key constraint
// gives its storage error, any other unique violation ErrConflict, and the
// rest is sanitized.
func writeError(entity string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if mapped, ok := constraintErrors[pgErr.ConstraintName]; ok &&
			(pgErr.Code == codeUniqueViolation || pgErr.Code == codeForeignKeyViolation) {
			return mapped
		}
		switch pgErr.Code {
		case codeUniqueViolation:
			return conflict(entity, "key")
		case codeForeignKeyViolation:
			return notFound(entity)
		}
	}
	return sanitize(err)
}

func notFound(entity string) error { return fmt.Errorf("%w: %s", storage.ErrNotFound, entity) }

func conflict(entity, field string) error {
	return fmt.Errorf("%w: %s %s", storage.ErrConflict, entity, field)
}

// corrupt reports a stored row that fails validation. The validation error
// is quoted, not wrapped: it names fields only, and a corrupt row must not
// look like invalid input (tenancy.ErrInvalid) to the caller.
func corrupt(entity string, err error) error {
	return fmt.Errorf("%w: %s: %s", ErrCorrupt, entity, err.Error())
}
