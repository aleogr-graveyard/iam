package audit

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/aleogr/iam/shared/canonjson"
	"github.com/aleogr/iam/shared/crypto"
	"github.com/aleogr/iam/shared/ids"
)

// HashBytes is the length of Hash and of a non-empty PrevHash.
const HashBytes = crypto.ChecksumBytes

// timeLayout is the canonical text of Time: UTC, always six fractional digits.
const timeLayout = "2006-01-02T15:04:05.000000Z"

// ErrBrokenChain reports a slice of events that is not an intact chain.
var ErrBrokenChain = errors.New("audit: broken chain")

// Canonical returns the canonical form of e that Seal hashes: the RFC 8785
// JSON of every field but Hash (package doc). It does not validate e.
func Canonical(e Event) ([]byte, error) {
	data := e.Data
	if data == nil {
		data = map[string]string{}
	}
	return canonjson.Marshal(map[string]any{
		"schema_version": e.SchemaVersion,
		"id":             e.ID.String(),
		"tenant_id":      e.TenantID.String(),
		"sequence":       e.Sequence,
		"time":           e.Time.UTC().Format(timeLayout),
		"action":         e.Action,
		"actor":          map[string]string{"type": e.Actor.Type, "id": e.Actor.ID},
		"target":         e.Target,
		"outcome":        e.Outcome,
		"data":           data,
		"prev_hash":      hex.EncodeToString(e.PrevHash),
	})
}

// hash computes SHA-256(prev || canonical form of e).
func hash(e Event) ([]byte, error) {
	canon, err := Canonical(e)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	return crypto.Checksum(append(bytes.Clone(e.PrevHash), canon...)), nil
}

// Seal returns e linked to prev, the Hash of the event before it (empty for
// the first event of a chain), with its Hash computed. It sets PrevHash to a
// copy of prev, ignores any Hash e already has, and refuses an event that
// fails Validate (ErrInvalid).
func Seal(prev []byte, e Event) (Event, error) {
	e.PrevHash, e.Hash = nil, nil
	if len(prev) > 0 {
		e.PrevHash = bytes.Clone(prev)
	}
	if err := e.Validate(); err != nil {
		return Event{}, err
	}
	h, err := hash(e)
	if err != nil {
		return Event{}, err
	}
	e.Hash = h
	return e, nil
}

// VerifyChain checks that events are tenant's whole chain, from its first
// event, in order: every event is valid and of tenant, Sequence runs
// 1, 2, 3... without gap or repeat, every PrevHash is the Hash before it (the
// first is empty) and every Hash is the hash of its event. It returns nil for
// no events and ErrBrokenChain naming the first sequence position that fails.
//
// It cannot see newest events dropped from the end (package doc). Hashes are
// compared in constant time although they are not secret: it costs nothing
// and leaves no timing question to review.
func VerifyChain(tenant ids.TenantID, events []Event) error {
	if tenant.IsZero() {
		return fmt.Errorf("%w: zero tenant", ErrBrokenChain)
	}
	var prev []byte
	for i, e := range events {
		pos := int64(i) + 1
		if err := e.Validate(); err != nil {
			return fmt.Errorf("%w: position %d: %w", ErrBrokenChain, pos, err)
		}
		switch {
		case e.TenantID != tenant:
			return fmt.Errorf("%w: position %d belongs to another tenant", ErrBrokenChain, pos)
		case e.Sequence != pos:
			return fmt.Errorf("%w: position %d has sequence %d", ErrBrokenChain, pos, e.Sequence)
		case !crypto.ConstantTimeEqual(e.PrevHash, prev):
			return fmt.Errorf("%w: position %d does not link to the event before it", ErrBrokenChain, pos)
		}
		h, err := hash(e)
		if err != nil {
			return fmt.Errorf("%w: position %d: %w", ErrBrokenChain, pos, err)
		}
		if !crypto.ConstantTimeEqual(e.Hash, h) {
			return fmt.Errorf("%w: position %d does not match its hash", ErrBrokenChain, pos)
		}
		prev = e.Hash
	}
	return nil
}
