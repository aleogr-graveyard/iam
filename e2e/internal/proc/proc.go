// Package proc runs server binaries for end-to-end tests.
package proc

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

const stopTimeout = 15 * time.Second

// Process is a started binary whose output is captured.
type Process struct {
	t    *testing.T
	cmd  *exec.Cmd
	out  *syncBuffer
	done chan struct{}
	once sync.Once
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// Start runs bin with env added to the parent environment. Its stdout and
// stderr are logged if the test fails; at cleanup it receives SIGTERM.
func Start(t *testing.T, bin string, env []string) *Process {
	t.Helper()
	out := &syncBuffer{}
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		t.Fatalf("proc: start %s: %v", bin, err)
	}
	p := &Process{t: t, cmd: cmd, out: out, done: make(chan struct{})}
	go func() {
		_ = cmd.Wait()
		close(p.done)
	}()
	t.Cleanup(func() {
		p.Stop()
		if t.Failed() {
			t.Logf("output of %s:\n%s", bin, out.String())
		}
	})
	return p
}

// Stop sends SIGTERM and waits for exit, killing the process after a timeout.
// It is safe to call more than once.
func (p *Process) Stop() {
	p.once.Do(func() {
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-p.done:
		case <-time.After(stopTimeout):
			_ = p.cmd.Process.Kill()
			<-p.done
			p.t.Errorf("proc: %s did not exit after SIGTERM", p.cmd.Path)
		}
	})
}

// WaitHTTP polls url until it answers with status want or the timeout expires.
func WaitHTTP(t *testing.T, url string, want int, timeout time.Duration) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	last := "no response"
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == want {
				return
			}
			last = resp.Status
		} else {
			last = err.Error()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("proc: %s did not answer %d within %s (last: %s)", url, want, timeout, last)
}

// FreeAddr returns a loopback address with a currently free port.
func FreeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("proc: free address: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().String()
}
