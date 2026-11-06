// Command a is test client A: a minimal relying party used by the end-to-end
// harness. It speaks to the platform only through public HTTP endpoints and
// uses only the standard library, so it is a separate module by design.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultAddr     = "127.0.0.1:8081"
	healthTimeout   = 2 * time.Second
	shutdownTimeout = 10 * time.Second
)

type health struct {
	Status string `json:"status"`
	IAM    string `json:"iam"`
}

func newHandler(iamURL string, client *http.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		code, body := http.StatusOK, health{Status: "ok", IAM: "ok"}
		if !iamHealthy(r.Context(), client, iamURL) {
			code, body = http.StatusServiceUnavailable, health{Status: "unavailable", IAM: "down"}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	})
	return mux
}

func iamHealthy(ctx context.Context, client *http.Client, iamURL string) bool {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iamURL+"/healthz", nil) //nolint:gosec // URL comes from operator configuration, not from requests
	if err != nil {
		return false
	}
	resp, err := client.Do(req) //nolint:gosec // URL comes from operator configuration, not from requests
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("client a stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	iamURL := os.Getenv("A_IAM_URL")
	if iamURL == "" {
		return errors.New("A_IAM_URL is required")
	}
	addr := os.Getenv("A_HTTP_ADDR")
	if addr == "" {
		addr = defaultAddr
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              addr,
		Handler:           newHandler(iamURL, &http.Client{Timeout: healthTimeout}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("client a starting", "addr", addr)

	select {
	case err := <-errc:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
