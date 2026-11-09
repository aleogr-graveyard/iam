package memory

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// Errors name the entity and the field, never the value.
func notFound(entity string) error { return fmt.Errorf("%w: %s", storage.ErrNotFound, entity) }

func conflict(entity, field string) error {
	return fmt.Errorf("%w: %s %s", storage.ErrConflict, entity, field)
}

// same is the copy function of values that hold no slices.
func same[V any](v V) V { return v }

// copyApp copies the slices of an application, so that stored values never
// alias the caller's.
func copyApp(a tenancy.Application) tenancy.Application {
	a.RedirectURIs = slices.Clone(a.RedirectURIs)
	a.SecretHash = bytes.Clone(a.SecretHash)
	return a
}

// get returns a copy of m[id], or ErrNotFound. The zero ID is never stored
// (Validate refuses it), so it finds nothing.
func get[V any](m map[ids.ID]V, id ids.ID, entity string, copyFn func(V) V) (V, error) {
	v, ok := m[id]
	if !ok {
		var zero V
		return zero, notFound(entity)
	}
	return copyFn(v), nil
}

// page returns copies of the values of m with an ID strictly after `after`, in
// ascending ID order, at most storage.ClampLimit(limit) of them.
func page[V any](m map[ids.ID]V, after ids.ID, limit int, copyFn func(V) V) ([]V, error) {
	n, err := storage.ClampLimit(limit)
	if err != nil {
		return nil, err
	}
	keys := slices.SortedFunc(maps.Keys(m), compareIDs)
	start, found := slices.BinarySearchFunc(keys, after, compareIDs)
	if found {
		start++
	}
	keys = keys[start:min(start+n, len(keys))]
	out := make([]V, len(keys))
	for i, k := range keys {
		out[i] = copyFn(m[k])
	}
	return out, nil
}

func compareIDs(a, b ids.ID) int { return bytes.Compare(a[:], b[:]) }

// keyed holds rows by ID with a second unique key (a slug or an e-mail).
// Stored values are never mutated in place, so a shallow copy of the maps is
// a full copy.
type keyed[V any] struct {
	entity, keyName string
	byID            map[ids.ID]V
	byKey           map[string]ids.ID
}

func newKeyed[V any](entity, keyName string) keyed[V] {
	return keyed[V]{entity: entity, keyName: keyName, byID: map[ids.ID]V{}, byKey: map[string]ids.ID{}}
}

func (k keyed[V]) clone() keyed[V] {
	k.byID, k.byKey = maps.Clone(k.byID), maps.Clone(k.byKey)
	return k
}

// insert stores v, refusing a duplicate ID or key.
func (k keyed[V]) insert(id ids.ID, key string, v V) error {
	if _, dup := k.byID[id]; dup {
		return conflict(k.entity, "id")
	}
	if _, dup := k.byKey[key]; dup {
		return conflict(k.entity, k.keyName)
	}
	k.byID[id] = v
	k.byKey[key] = id
	return nil
}

// lookup finds a row by its key.
func (k keyed[V]) lookup(key string) (V, error) {
	id, ok := k.byKey[key]
	if !ok {
		var zero V
		return zero, notFound(k.entity)
	}
	return k.byID[id], nil
}

// memberKey identifies a membership within a tenant.
type memberKey struct{ org, user ids.ID }

// tenantData holds the rows of one tenant.
type tenantData struct {
	orgs    keyed[tenancy.Organization] // by slug
	users   keyed[tenancy.User]         // by normalized e-mail
	members map[memberKey]tenancy.Membership
	apps    map[ids.ID]tenancy.Application
	// events is the audit chain in sequence order, append-only; eventIDs
	// indexes its IDs.
	events   []audit.Event
	eventIDs map[ids.ID]struct{}
}

// clone copies d; a nil d (a tenant with nothing committed yet) gives an
// empty tenantData.
func (d *tenantData) clone() *tenantData {
	if d == nil {
		return &tenantData{
			orgs:     newKeyed[tenancy.Organization]("organization", "slug"),
			users:    newKeyed[tenancy.User]("user", "email"),
			members:  map[memberKey]tenancy.Membership{},
			apps:     map[ids.ID]tenancy.Application{},
			eventIDs: map[ids.ID]struct{}{},
		}
	}
	// Clip: an append to the copy's events must never write into the
	// committed backing array.
	return &tenantData{
		orgs: d.orgs.clone(), users: d.users.clone(), members: maps.Clone(d.members), apps: maps.Clone(d.apps),
		events: slices.Clip(d.events), eventIDs: maps.Clone(d.eventIDs),
	}
}
