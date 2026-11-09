package memory

import (
	"bytes"
	"context"
	"maps"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
)

// auditLog is the audit chain of one tenant within a transaction: data is the
// transaction's working copy of that tenant. Transactions are serialized, so
// the newest event of data is the head of the chain.
type auditLog struct {
	state  *txState
	tenant ids.TenantID // zero: no chain (a system transaction before the system tenant)
	data   *tenantData
	all    map[ids.ID]*tenantData // committed rows of every tenant; read only
}

// copyEvent copies the slices and the map of an event, so that stored values
// never alias the caller's.
func copyEvent(e audit.Event) audit.Event {
	e.Data = maps.Clone(e.Data)
	e.PrevHash, e.Hash = bytes.Clone(e.PrevHash), bytes.Clone(e.Hash)
	return e
}

func (l auditLog) Append(ctx context.Context, e audit.Event) (audit.Event, error) {
	if err := l.state.check(ctx); err != nil {
		return audit.Event{}, err
	}
	if l.tenant.IsZero() {
		return audit.Event{}, storage.ErrInvalidTenant
	}
	var seq int64
	var prev []byte
	if n := len(l.data.events); n > 0 {
		seq, prev = l.data.events[n-1].Sequence, l.data.events[n-1].Hash
	}
	sealed, err := storage.SealAudit(l.tenant, seq+1, prev, e)
	if err != nil {
		return audit.Event{}, err
	}
	if l.idTaken(sealed.ID.ID) {
		return audit.Event{}, conflict("audit event", "id")
	}
	l.data.events = append(l.data.events, sealed)
	l.data.eventIDs[sealed.ID.ID] = struct{}{}
	return copyEvent(sealed), nil
}

// idTaken reports whether any tenant's chain holds id: event IDs are unique
// across tenants.
func (l auditLog) idTaken(id ids.ID) bool {
	if _, ok := l.data.eventIDs[id]; ok {
		return true
	}
	for t, d := range l.all {
		if _, ok := d.eventIDs[id]; ok && t != l.tenant.ID {
			return true
		}
	}
	return false
}

// Events relies on the chain numbering: event n is at index n-1.
func (l auditLog) Events(ctx context.Context, after int64, limit int) ([]audit.Event, error) {
	if err := l.state.check(ctx); err != nil {
		return nil, err
	}
	n, err := storage.AuditPage(after, limit)
	if err != nil {
		return nil, err
	}
	events := l.data.events
	start := int(min(after, int64(len(events))))
	page := events[start:min(start+n, len(events))]
	out := make([]audit.Event, len(page))
	for i, e := range page {
		out[i] = copyEvent(e)
	}
	return out, nil
}
