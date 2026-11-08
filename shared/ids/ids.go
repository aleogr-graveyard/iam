// Package ids defines the identifiers used across the service: UUIDv7 values
// (RFC 9562 section 5.7) wrapped in one distinct type per entity, so that a
// tenant identifier cannot be passed where a user identifier is expected.
//
//	func Load(t ids.TenantID, u ids.UserID) {}
//
//	Load(userID, tenantID) // does not compile: the types are distinct
//
// An explicit conversion such as TenantID(userID) does compile; it must be done
// deliberately, never to silence a type error.
//
// The zero value means "unset" and never comes from parsing: Parse,
// UnmarshalText and JSON refuse the all-zero UUID. Wire structs mark ID fields
// with `json:",omitzero"`, which works on Go 1.24+ through the promoted IsZero.
// Presence of a required ID is checked by the storage ports and the domain
// constructors.
//
// Time comes from an injected clock.Clock and randomness from shared/crypto.
package ids

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/crypto"
)

// ErrInvalid reports text that is not a canonical UUIDv7.
var ErrInvalid = errors.New("ids: invalid identifier")

// ErrNoClock reports a Generator used without a clock.
var ErrNoClock = errors.New("ids: generator has no clock")

// ErrTimeOutOfRange reports a clock reading outside the 48-bit millisecond
// range that a UUIDv7 can carry.
var ErrTimeOutOfRange = errors.New("ids: time outside the UUIDv7 range")

const (
	textLen   = 36
	maxMillis = int64(1)<<48 - 1
	hexDigits = "0123456789abcdef"
)

// ID is a UUIDv7. The zero value is the all-zero UUID and means "unset".
type ID [16]byte

// IsZero reports whether the ID is all zero.
func (id ID) IsZero() bool { return id == ID{} }

// String returns the canonical lowercase 8-4-4-4-12 form.
func (id ID) String() string {
	var b [textLen]byte
	j := 0
	for i, v := range id {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			b[j] = '-'
			j++
		}
		b[j], b[j+1] = hexDigits[v>>4], hexDigits[v&0x0f]
		j += 2
	}
	return string(b[:])
}

// Parse accepts only the canonical lowercase 8-4-4-4-12 form of a UUIDv7 with
// the RFC 4122/9562 variant. Any other version or variant, and the all-zero
// UUID, are refused: every identifier of this service is generated as a
// UUIDv7, and the zero value means "unset".
func Parse(s string) (ID, error) {
	if len(s) != textLen {
		return ID{}, fmt.Errorf("%w: length %d", ErrInvalid, len(s))
	}
	id, ok := decodeHex(s)
	if !ok {
		return ID{}, fmt.Errorf("%w: malformed", ErrInvalid)
	}
	if id.IsZero() {
		return ID{}, fmt.Errorf("%w: zero", ErrInvalid)
	}
	if id[6]>>4 != 7 || id[8]>>6 != 0b10 {
		return ID{}, fmt.Errorf("%w: not a UUIDv7", ErrInvalid)
	}
	return id, nil
}

// decodeHex decodes a 36-byte 8-4-4-4-12 string of lowercase hex digits.
func decodeHex(s string) (ID, bool) {
	var id ID
	j := 0
	for i := 0; i < textLen; i++ {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if s[i] != '-' {
				return ID{}, false
			}
			continue
		}
		hi, lo := strings.IndexByte(hexDigits, s[i]), strings.IndexByte(hexDigits, s[i+1])
		if hi < 0 || lo < 0 {
			return ID{}, false
		}
		id[j] = byte(hi)<<4 | byte(lo) //nolint:gosec // both are 0..15
		j++
		i++
	}
	return id, true
}

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler with the rules of Parse.
// On error the receiver is left unchanged.
func (id *ID) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Generator creates UUIDv7 values from an injected clock.
type Generator struct {
	Clock clock.Clock
}

// New returns a new UUIDv7: a 48-bit big-endian Unix-millisecond timestamp from
// the clock, the version and variant bits, and 74 random bits. It fails when
// the clock is before the Unix epoch or past the 48-bit range.
func (g Generator) New() (ID, error) {
	if g.Clock == nil {
		return ID{}, ErrNoClock
	}
	ms := g.Clock.Now().UnixMilli()
	if ms < 0 || ms > maxMillis {
		return ID{}, ErrTimeOutOfRange
	}
	rnd, err := crypto.RandomBytes(10)
	if err != nil {
		return ID{}, fmt.Errorf("ids: %w", err)
	}
	var id ID
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(ms)) //nolint:gosec // range checked above
	copy(id[:6], ts[2:])
	copy(id[6:], rnd)
	id[6] = 0x70 | id[6]&0x0f
	id[8] = 0x80 | id[8]&0x3f
	return id, nil
}

// TenantID identifies a tenant.
type TenantID struct{ ID }

// OrganizationID identifies an organization.
type OrganizationID struct{ ID }

// UserID identifies a user.
type UserID struct{ ID }

// ApplicationID identifies an application.
type ApplicationID struct{ ID }

// EventID identifies an audit event.
type EventID struct{ ID }
