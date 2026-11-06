// Package httpserver is the HTTP front of the iam binary: health endpoints
// today, the public APIs in later milestones.
package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

const (
	defaultReadinessTimeout = 2 * time.Second
	readHeaderTimeout       = 5 * time.Second
)

// ReadinessCheck is a named dependency probe run by /readyz.
type ReadinessCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

// Option configures a Server.
type Option func(*Server)

// WithReadinessChecks adds checks evaluated by /readyz.
func WithReadinessChecks(checks ...ReadinessCheck) Option {
	return func(s *Server) { s.checks = append(s.checks, checks...) }
}

// WithReadinessTimeout bounds how long /readyz waits for all checks.
func WithReadinessTimeout(d time.Duration) Option {
	return func(s *Server) { s.readinessTimeout = d }
}

// Server is the HTTP server of the iam binary.
type Server struct {
	addr             string
	mux              *http.ServeMux
	checks           []ReadinessCheck
	readinessTimeout time.Duration
}

// New builds a Server listening on addr once Run is called.
func New(addr string, opts ...Option) *Server {
	s := &Server{
		addr:             addr,
		mux:              http.NewServeMux(),
		readinessTimeout: defaultReadinessTimeout,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.registerHealth()
	return s
}

// Handler returns the root handler.
func (s *Server) Handler() http.Handler { return s.mux }

// Run listens on the configured address and serves until ctx is done, then
// shuts down gracefully within shutdownTimeout. It returns nil on a clean
// shutdown.
func (s *Server) Run(ctx context.Context, shutdownTimeout time.Duration) error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.Serve(ctx, l, shutdownTimeout)
}

// Serve is Run on a caller-provided listener.
func (s *Server) Serve(ctx context.Context, l net.Listener, shutdownTimeout time.Duration) error {
	srv := &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(l) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		return err
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
