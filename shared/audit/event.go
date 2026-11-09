// Package audit defines the audit event (contract C4) and its hash chain
// (ADR 0005): every tenant has one append-only chain of events, each sealed
// with the hash of the one before it, so that changing, removing, inserting
// or reordering a stored event is detected by VerifyChain (TM-TEN-004).
//
// # Schema (version 1)
//
// An Event records who (Actor) did what (Action) to which resource (Target),
// with what result (Outcome), when (Time) and in which tenant (TenantID),
// plus a few short key-value facts (Data). Validate holds every rule; an
// event that fails it is never sealed or stored. Any change to the fields, to
// their rules or to the canonical form below is a new SchemaVersion, made
// through a reviewed change of contract C4.
//
// # Personal data and secrets
//
// An event is kept for as long as the chain and is copied to whoever reviews
// it, so it must never hold raw personal data or secrets: no e-mail address,
// name, phone, IP address, password, token, code or key, in Data, Target or
// Actor.ID. Refer to people and resources by their IDs. Validate refuses, as a
// cheap guard against the most common mistake, any of those values that
// contains something shaped like an e-mail address (an '@' followed by a
// domain with a dot); the rule itself is the emitter's responsibility.
//
// Target and Data carry only IDs, resource names built from IDs and slugs,
// and enum-like values the service chose (a status, a client type, a
// count), never text a user typed (a display name, a free-text reason). Not
// only for privacy: Validate's refusal fails the append and with it the
// business transaction it belongs to, so user-supplied text that merely
// looks like an address (a name with an '@') would make a valid change fail.
//
// # Times
//
// Time comes from the emitter's injected clock, never from a client. It is
// stored in UTC at microsecond precision (the precision of a Postgres
// timestamptz, contract C3), so that the hashed value is the stored one; the
// storage adapters normalize it before sealing.
//
// # Hash chain
//
// The events of a tenant are numbered by Sequence from 1, with no gap. The
// canonical form of an event is the RFC 8785 JSON (shared/canonjson) of
// every field but Hash, under the names Canonical shows, with IDs in their
// canonical text form, Time as "2006-01-02T15:04:05.000000Z" and PrevHash in
// lower-case hex. Then
//
//	Hash = SHA-256(PrevHash || canonical form)
//
// The genesis is the empty string: the first event of a chain has Sequence 1
// and an empty PrevHash; every other event's PrevHash is the Hash of the event
// before it. The order of a chain is its Sequence, never the order of event
// IDs (UUIDv7 IDs only sort to the millisecond).
//
// The chain proves integrity, not authenticity: whoever can write the store
// directly (the schema owner) can rewrite a whole chain from the first changed
// event on, and can drop its newest events, without VerifyChain noticing.
// Anchoring the head of each chain outside the database (a signed checkpoint)
// closes both; it is later work.
package audit

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/aleogr/iam/shared/ids"
)

// SchemaVersion is the version of the event schema this package writes.
const SchemaVersion = 1

// Limits of an event, in bytes.
const (
	MaxActionBytes    = 64
	MaxActorIDBytes   = 128
	MaxTargetBytes    = 512
	MaxDataKeys       = 32
	MaxDataValueBytes = 256
)

// Actor types.
const (
	ActorUser        = "user"
	ActorOperator    = "operator"
	ActorSystem      = "system"
	ActorApplication = "application"
)

// Outcomes.
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
	OutcomeDenied  = "denied"
)

var (
	actorTypes = map[string]bool{ActorUser: true, ActorOperator: true, ActorSystem: true, ActorApplication: true}
	outcomes   = map[string]bool{OutcomeSuccess: true, OutcomeFailure: true, OutcomeDenied: true}
)

// ErrInvalid reports an event that fails Validate. Its messages name fields,
// never values.
var ErrInvalid = errors.New("audit: invalid event")

var (
	// actionPattern is two or more dotted lower-case segments, such as
	// "tenant.created" or "authn.password.verified".
	actionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
	dataKey       = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
)

// Actor is who caused an event: a user, an operator of the platform, the
// system itself or an application (an OAuth client), by its ID.
type Actor struct {
	Type string
	ID   string
}

// Event is one audit event. See the package doc for the rules.
type Event struct {
	ID            ids.EventID
	TenantID      ids.TenantID
	Sequence      int64
	Time          time.Time
	SchemaVersion int
	// Action is dotted lower case, such as "tenant.created".
	Action string
	Actor  Actor
	// Target is the resource name, such as "tenants/acme/organizations/<id>";
	// it may be empty.
	Target string
	// Outcome is OutcomeSuccess, OutcomeFailure or OutcomeDenied.
	Outcome string
	// Data holds short facts about the event; never raw personal data or
	// secrets (package doc). Nil and empty are the same.
	Data map[string]string
	// PrevHash is empty for the first event of a chain and the previous
	// event's Hash otherwise; Hash is set by Seal.
	PrevHash, Hash []byte
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalid}, args...)...)
}

// Validate checks every field but the value of Hash (VerifyChain checks it):
// non-zero IDs, a chain position (Sequence and the PrevHash it implies), the
// schema version, the time, and the bounds and shapes of the text fields.
func (e Event) Validate() error {
	switch {
	case e.ID.IsZero():
		return invalid("id is zero")
	case e.TenantID.IsZero():
		return invalid("tenant_id is zero")
	case e.SchemaVersion != SchemaVersion:
		return invalid("schema_version %d is not %d", e.SchemaVersion, SchemaVersion)
	}
	if err := e.validatePosition(); err != nil {
		return err
	}
	if err := validateTime(e.Time); err != nil {
		return err
	}
	return e.validateText()
}

func (e Event) validatePosition() error {
	switch {
	case e.Sequence < 1:
		return invalid("sequence must be at least 1")
	case e.Sequence == 1 && len(e.PrevHash) != 0:
		return invalid("prev_hash must be empty for sequence 1")
	case e.Sequence > 1 && len(e.PrevHash) != HashBytes:
		return invalid("prev_hash must be %d bytes", HashBytes)
	case len(e.Hash) != 0 && len(e.Hash) != HashBytes:
		return invalid("hash must be %d bytes", HashBytes)
	}
	return nil
}

func validateTime(t time.Time) error {
	switch {
	case t.Location() != time.UTC:
		return invalid("time must be in UTC")
	case !t.Equal(t.Truncate(time.Microsecond)):
		return invalid("time must have microsecond precision")
	case t.Year() < 1970 || t.Year() > 9999:
		return invalid("time out of range")
	}
	return nil
}

func (e Event) validateText() error {
	switch {
	case len(e.Action) > MaxActionBytes || !actionPattern.MatchString(e.Action):
		return invalid("action must be dotted lower case, at most %d bytes", MaxActionBytes)
	case !actorTypes[e.Actor.Type]:
		return invalid("actor type is unknown")
	case !outcomes[e.Outcome]:
		return invalid("outcome is unknown")
	case e.Actor.ID == "":
		return invalid("actor id is empty")
	}
	if err := validateValue("actor id", e.Actor.ID, MaxActorIDBytes); err != nil {
		return err
	}
	if err := validateValue("target", e.Target, MaxTargetBytes); err != nil {
		return err
	}
	return validateData(e.Data)
}

func validateData(data map[string]string) error {
	if len(data) > MaxDataKeys {
		return invalid("data has more than %d keys", MaxDataKeys)
	}
	for k, v := range data {
		if !dataKey.MatchString(k) {
			return invalid("data key must match %s", dataKey)
		}
		if err := validateValue("data value", v, MaxDataValueBytes); err != nil {
			return err
		}
	}
	return nil
}

// validateValue bounds a free-text value and refuses invalid UTF-8, control
// characters and anything shaped like an e-mail address.
func validateValue(field, v string, maxBytes int) error {
	switch {
	case len(v) > maxBytes:
		return invalid("%s is longer than %d bytes", field, maxBytes)
	case !utf8.ValidString(v):
		return invalid("%s is not valid UTF-8", field)
	case strings.ContainsFunc(v, isControlOrFormat):
		return invalid("%s has a control, format or separator character", field)
	case looksLikeEmail(v):
		return invalid("%s looks like an e-mail address (personal data)", field)
	}
	return nil
}

// isControlOrFormat reports the characters that would let a value hide or
// reorder text when an event is displayed: controls (Cc), format characters
// (Cf: bidi overrides such as U+202E, zero-width U+200B) and the line and
// paragraph separators (Zl, Zp).
func isControlOrFormat(r rune) bool {
	return unicode.In(r, unicode.Cc, unicode.Cf, unicode.Zl, unicode.Zp)
}

// looksLikeEmail reports an '@' with something before it and a dotted domain
// after it ("x@y.z"), anywhere in v.
func looksLikeEmail(v string) bool {
	for i := strings.IndexByte(v, '@'); i >= 0; {
		domain := v[i+1:]
		if end := strings.IndexFunc(domain, unicode.IsSpace); end >= 0 {
			domain = domain[:end]
		}
		if dot := strings.IndexByte(domain, '.'); i > 0 && dot > 0 && dot < len(domain)-1 {
			return true
		}
		next := strings.IndexByte(v[i+1:], '@')
		if next < 0 {
			break
		}
		i += 1 + next
	}
	return false
}
