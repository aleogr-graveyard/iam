package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func do(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func assertJSON(t *testing.T, rec *httptest.ResponseRecorder, code int, body string) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != body {
		t.Fatalf("body = %q, want %q", got, body)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestHealthz_OK(t *testing.T) {
	assertJSON(t, do(t, New(":0").Handler(), http.MethodGet, "/healthz"), 200, `{"status":"ok"}`)
}

func TestReadyz_AllChecksPass(t *testing.T) {
	ok := func(context.Context) error { return nil }
	s := New(":0", WithReadinessChecks(
		ReadinessCheck{Name: "a", Check: ok},
		ReadinessCheck{Name: "b", Check: ok},
	))
	assertJSON(t, do(t, s.Handler(), http.MethodGet, "/readyz"), 200, `{"status":"ok"}`)
}

func TestReadyz_FailingCheckReturns503WithNameOnly(t *testing.T) {
	s := New(":0", WithReadinessChecks(
		ReadinessCheck{Name: "database", Check: func(context.Context) error {
			return errors.New("dial postgres://u:secret@h")
		}},
		ReadinessCheck{Name: "cache", Check: func(context.Context) error { return nil }},
	))
	rec := do(t, s.Handler(), http.MethodGet, "/readyz")
	assertJSON(t, rec, 503, `{"status":"unavailable","failed":["database"]}`)
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatal("body leaks error detail")
	}
}

func TestReadyz_FailedNamesSorted(t *testing.T) {
	bad := func(context.Context) error { return errors.New("x") }
	s := New(":0", WithReadinessChecks(
		ReadinessCheck{Name: "b", Check: bad},
		ReadinessCheck{Name: "a", Check: bad},
	))
	assertJSON(t, do(t, s.Handler(), http.MethodGet, "/readyz"), 503, `{"status":"unavailable","failed":["a","b"]}`)
}

func TestReadyz_CheckTimeout(t *testing.T) {
	s := New(":0",
		WithReadinessTimeout(50*time.Millisecond),
		WithReadinessChecks(ReadinessCheck{Name: "slow", Check: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}}),
	)
	start := time.Now()
	rec := do(t, s.Handler(), http.MethodGet, "/readyz")
	if d := time.Since(start); d > time.Second {
		t.Fatalf("took %v", d)
	}
	assertJSON(t, rec, 503, `{"status":"unavailable","failed":["slow"]}`)
}

func TestReadyz_UnresponsiveCheckDoesNotBlock(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	s := New(":0",
		WithReadinessTimeout(50*time.Millisecond),
		WithReadinessChecks(ReadinessCheck{Name: "stuck", Check: func(context.Context) error {
			<-release // ignores ctx
			return nil
		}}),
	)
	start := time.Now()
	rec := do(t, s.Handler(), http.MethodGet, "/readyz")
	if d := time.Since(start); d > time.Second {
		t.Fatalf("took %v", d)
	}
	assertJSON(t, rec, 503, `{"status":"unavailable","failed":["stuck"]}`)
}

func TestHealthEndpoints_MethodNotAllowed(t *testing.T) {
	for _, p := range []string{"/healthz", "/readyz"} {
		rec := do(t, New(":0").Handler(), http.MethodPost, p)
		if rec.Code != 405 {
			t.Fatalf("%s: status = %d", p, rec.Code)
		}
		if got := rec.Header().Get("Allow"); got != "GET" {
			t.Fatalf("%s: Allow = %q", p, got)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("%s: Cache-Control = %q", p, got)
		}
	}
}

func BenchmarkHealthz(b *testing.B) {
	h := New(":0").Handler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	b.ReportAllocs()
	for b.Loop() {
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
}
