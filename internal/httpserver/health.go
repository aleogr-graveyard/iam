package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
)

type healthBody struct {
	Status string   `json:"status"`
	Failed []string `json:"failed,omitempty"`
}

func (s *Server) registerHealth() {
	for path, h := range map[string]http.HandlerFunc{
		"/healthz": s.healthz,
		"/readyz":  s.readyz,
	} {
		s.mux.HandleFunc("GET "+path, h)
		s.mux.HandleFunc(path, methodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", http.MethodGet)
	writeJSON(w, http.StatusMethodNotAllowed, struct {
		Status string `json:"status"`
	}{"method not allowed"})
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
}

// readyz runs every check concurrently under one deadline. A check that
// ignores its context is reported as failed once the deadline passes; only
// names are exposed, never error text.
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.readinessTimeout)
	defer cancel()

	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(s.checks))
	for _, c := range s.checks {
		go func() { results <- result{c.Name, c.Check(ctx)} }()
	}

	pending := make(map[string]bool, len(s.checks))
	for _, c := range s.checks {
		pending[c.Name] = true
	}
	var failed []string
collect:
	for range s.checks {
		select {
		case res := <-results:
			delete(pending, res.name)
			if res.err != nil {
				failed = append(failed, res.name)
			}
		case <-ctx.Done():
			break collect
		}
	}
	for name := range pending {
		failed = append(failed, name)
	}

	if len(failed) == 0 {
		writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
		return
	}
	sort.Strings(failed)
	writeJSON(w, http.StatusServiceUnavailable, healthBody{Status: "unavailable", Failed: failed})
}
