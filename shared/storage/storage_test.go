package storage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

func TestClampLimit(t *testing.T) {
	cases := []struct {
		in, want int
		err      bool
	}{
		{-1, 0, true},
		{0, 0, true},
		{1, 1, false},
		{storage.MaxListLimit, storage.MaxListLimit, false},
		{storage.MaxListLimit + 1, storage.MaxListLimit, false},
		{int(^uint(0) >> 1), storage.MaxListLimit, false},
	}
	for _, c := range cases {
		got, err := storage.ClampLimit(c.in)
		if c.err {
			if !errors.Is(err, storage.ErrInvalid) || !errors.Is(err, tenancy.ErrInvalid) {
				t.Errorf("ClampLimit(%d) error = %v, want ErrInvalid wrapping tenancy.ErrInvalid", c.in, err)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("ClampLimit(%d) = %d, %v; want %d, nil", c.in, got, err, c.want)
		}
	}
}

func TestInvalid_MatchesBothSentinelsAndKeepsTheChain(t *testing.T) {
	cause := tenancy.Organization{}.Validate()
	err := storage.Invalid(cause)
	for _, target := range []error{storage.ErrInvalid, tenancy.ErrInvalid, cause} {
		if !errors.Is(err, target) {
			t.Errorf("errors.Is(%v, %v) = false", err, target)
		}
	}
	for _, target := range []error{storage.ErrConflict, storage.ErrNotFound} {
		if errors.Is(err, target) {
			t.Errorf("errors.Is(%v, %v) = true", err, target)
		}
	}
	if storage.Invalid(nil) != nil {
		t.Error("Invalid(nil) != nil")
	}
}

func TestEnterTx_RefusesNesting(t *testing.T) {
	ctx, err := storage.EnterTx(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := storage.EnterTx(ctx); !errors.Is(err, storage.ErrNestedTx) {
		t.Fatalf("nested EnterTx error = %v, want ErrNestedTx", err)
	}
}

func TestMicroseconds_TruncatesTowardThePast(t *testing.T) {
	at := time.Date(2026, 11, 9, 12, 0, 0, 123456999, time.UTC)
	if got := storage.Microseconds(at); !got.Equal(time.Date(2026, 11, 9, 12, 0, 0, 123456000, time.UTC)) {
		t.Fatalf("Microseconds = %v", got)
	}
	if !storage.Microseconds(time.Time{}).IsZero() {
		t.Fatal("the zero time is not kept zero")
	}
}
