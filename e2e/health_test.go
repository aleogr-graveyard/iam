package e2e_test

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/aleogr/iam/e2e/internal/proc"
)

const (
	dbEnv      = "IAM_E2E_DATABASE_URL"
	requireEnv = "IAM_REQUIRE_DB"
	waitFor    = 30 * time.Second
)

// repoRoot returns the directory holding go.work, found by walking up.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.work not found above the working directory")
		}
		dir = parent
	}
}

func build(t *testing.T, root, out, pkg string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", out, pkg) //nolint:gosec // arguments are fixed by the test
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, b)
	}
}

func TestE2E_ClientAReachesIAM(t *testing.T) {
	dbURL := os.Getenv(dbEnv)
	if dbURL == "" {
		if os.Getenv(requireEnv) == "1" {
			t.Fatalf("%s is not set but %s=1", dbEnv, requireEnv)
		}
		t.Skipf("%s not set; skipping end-to-end test", dbEnv)
	}

	root := repoRoot(t)
	bin := t.TempDir()
	iamBin := filepath.Join(bin, "iam")
	aBin := filepath.Join(bin, "a")
	build(t, root, iamBin, "./cmd/iam")
	build(t, root, aBin, "./examples/a")

	iamAddr := proc.FreeAddr(t)
	iam := proc.Start(t, iamBin, []string{
		"IAM_HTTP_ADDR=" + iamAddr,
		"IAM_DATABASE_URL=" + dbURL,
	})
	iam.WaitHTTP(t, "http://"+iamAddr+"/readyz", http.StatusOK, waitFor)

	aAddr := proc.FreeAddr(t)
	a := proc.Start(t, aBin, []string{
		"A_HTTP_ADDR=" + aAddr,
		"A_IAM_URL=http://" + iamAddr,
	})
	aHealth := "http://" + aAddr + "/healthz"
	a.WaitHTTP(t, aHealth, http.StatusOK, waitFor)

	iam.Stop()
	a.WaitHTTP(t, aHealth, http.StatusServiceUnavailable, waitFor)
}
