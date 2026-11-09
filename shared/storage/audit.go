package storage

import (
	"fmt"
	"maps"
	"time"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
)

// SealAudit is the part of AuditAppender.Append that every adapter shares
// (package doc, Audit): it forces tenant and audit.SchemaVersion onto e,
// truncates its Time to the microsecond in UTC, copies its Data (nil when
// empty), places it at seq after prev (the Hash of event seq-1; empty for
// seq 1) and seals it. A validation failure is ErrInvalid wrapping
// audit.ErrInvalid. The adapter supplies seq and prev from the newest stored
// event, read under its per-chain serialization, and stores the result.
func SealAudit(tenant ids.TenantID, seq int64, prev []byte, e audit.Event) (audit.Event, error) {
	e.TenantID, e.SchemaVersion, e.Sequence = tenant, audit.SchemaVersion, seq
	e.Time = Microseconds(e.Time).In(time.UTC)
	if len(e.Data) == 0 {
		e.Data = nil
	} else {
		e.Data = maps.Clone(e.Data)
	}
	sealed, err := audit.Seal(prev, e)
	if err != nil {
		return audit.Event{}, Invalid(err)
	}
	return sealed, nil
}

var errAfter = fmt.Errorf("%w: after must not be negative", ErrInvalid)

// AuditPage checks the arguments of AuditLog.Events and returns the page
// size to use (ClampLimit).
func AuditPage(after int64, limit int) (int, error) {
	if after < 0 {
		return 0, errAfter
	}
	return ClampLimit(limit)
}
