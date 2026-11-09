package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aleogr/iam/shared/tenancy"
)

// Errors of the storage ports. Adapters wrap them; callers test with errors.Is.
var (
	// ErrNotFound reports a row that does not exist in the scope of the
	// transaction, including one that exists only in another tenant.
	ErrNotFound = errors.New("storage: not found")
	// ErrConflict reports a uniqueness violation (ID, slug, e-mail, membership).
	ErrConflict = errors.New("storage: conflict")
	// ErrInvalidTenant reports a zero or unknown tenant given to InTenant.
	ErrInvalidTenant = errors.New("storage: invalid tenant")
	// ErrInvalid reports an invalid argument: an entity that fails validation
	// or a limit of zero or less. It wraps tenancy.ErrInvalid.
	ErrInvalid = fmt.Errorf("storage: %w", tenancy.ErrInvalid)
	// ErrNestedTx reports InTenant or InSystem called inside another
	// transaction's fn.
	ErrNestedTx = errors.New("storage: nested transaction")
	// ErrTxDone reports a Tx or SystemTx used after its fn returned.
	ErrTxDone = errors.New("storage: transaction already finished")
)

// invalidError carries a validation error: errors.Is matches both ErrInvalid
// and every error in the chain of err (tenancy.ErrInvalid among them).
type invalidError struct{ err error }

func (e invalidError) Error() string        { return "storage: " + e.err.Error() }
func (e invalidError) Unwrap() error        { return e.err }
func (e invalidError) Is(target error) bool { return target == ErrInvalid }

// Invalid wraps an entity's validation error (which already wraps
// tenancy.ErrInvalid and never carries a value) so that it is also ErrInvalid.
// Adapters use it for every Validate failure on create.
func Invalid(err error) error {
	if err == nil {
		return nil
	}
	return invalidError{err: err}
}

var errLimit = fmt.Errorf("%w: limit must be positive", ErrInvalid)

// ClampLimit returns the page size an adapter must use for a List limit:
// ErrInvalid for zero or less, MaxListLimit for anything above it.
func ClampLimit(limit int) (int, error) {
	if limit <= 0 {
		return 0, errLimit
	}
	return min(limit, MaxListLimit), nil
}

// Microseconds truncates t toward the past to the microsecond, the precision
// at which every adapter stores times (package doc, Times).
func Microseconds(t time.Time) time.Time { return t.Truncate(time.Microsecond) }

type txKey struct{}

// EnterTx returns the ctx an adapter passes to fn, marked as being inside a
// transaction, or ErrNestedTx when ctx is already marked. Every adapter calls
// it before opening a transaction, so that nesting is refused in all of them.
func EnterTx(ctx context.Context) (context.Context, error) {
	if ctx.Value(txKey{}) != nil {
		return nil, ErrNestedTx
	}
	return context.WithValue(ctx, txKey{}, struct{}{}), nil
}
