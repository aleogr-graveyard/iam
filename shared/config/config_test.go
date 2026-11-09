package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

var goodKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func noFile(string) ([]byte, error) { return nil, errors.New("no files in this test") }

func prod(extra map[string]string) map[string]string {
	m := map[string]string{
		"IAM_PUBLIC_BASE_URL": "https://iam.example.com",
		"IAM_CURSOR_KEY":      goodKey,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestLoad_Defaults(t *testing.T) {
	cfg, findings, err := Load(env(prod(nil)), noFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != Production || cfg.HTTPAddr != ":8080" || cfg.DatabaseAppRole != "iam_app" ||
		cfg.ShutdownTimeout != 10*time.Second || cfg.DatabaseURL != "" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.Argon2 != (Argon2Settings{MemoryKiB: 19456, Iterations: 2, Parallelism: 1, MaxConcurrent: 2}) {
		t.Fatalf("argon2 defaults: %+v", cfg.Argon2)
	}
	if cfg.Database != (DatabaseLimits{
		StatementTimeout: 5 * time.Second, IdleInTransactionTimeout: 10 * time.Second, LockTimeout: 2 * time.Second,
		MaxConnIdleTime: time.Minute, MigrationStatementTimeout: 5 * time.Minute, MigrationLockTimeout: 30 * time.Second,
	}) {
		t.Fatalf("database defaults: %+v", cfg.Database)
	}
	if len(findings) != 0 {
		t.Fatalf("findings: %+v", findings)
	}
}

func TestLoad_PrecedenceFileThenEnv(t *testing.T) {
	file := `{"http_addr": ":9000", "shutdown_timeout": "3s", "argon2_iterations": 3, "database_app_role": "from_file"}`
	read := func(p string) ([]byte, error) {
		if p != "/etc/iam.json" {
			t.Fatalf("path %q", p)
		}
		return []byte(file), nil
	}
	cfg, _, err := Load(env(prod(map[string]string{
		"IAM_CONFIG_FILE":      "/etc/iam.json",
		"IAM_SHUTDOWN_TIMEOUT": "5s",
	})), read)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":9000" || cfg.Argon2.Iterations != 3 || cfg.DatabaseAppRole != "from_file" {
		t.Fatalf("file values lost: %+v", cfg)
	}
	if cfg.ShutdownTimeout != 5*time.Second {
		t.Fatalf("environment must win over file: %v", cfg.ShutdownTimeout)
	}
}

func TestLoad_FileRejectsUnknownAndDuplicateKeys(t *testing.T) {
	for _, file := range []string{`{"nope": 1}`, `{"http_addr": "a", "http_addr": "b"}`, `{"config_file": "x"}`, `{"http_addr": true}`, `[`} {
		_, _, err := Load(env(prod(map[string]string{"IAM_CONFIG_FILE": "f"})), func(string) ([]byte, error) { return []byte(file), nil })
		if err == nil {
			t.Fatalf("%s: accepted", file)
		}
	}
}

func TestLoad_EmptyEnvCountsAsUnset(t *testing.T) {
	file := `{"cursor_key": "` + goodKey + `"}`
	cfg, _, err := Load(env(map[string]string{
		"IAM_PUBLIC_BASE_URL": "https://iam.example.com",
		"IAM_CONFIG_FILE":     "f",
		"IAM_CURSOR_KEY":      "",
	}), func(string) ([]byte, error) { return []byte(file), nil })
	if err != nil || cfg.CursorKey != goodKey {
		t.Fatalf("file value must survive an empty variable: %v", err)
	}
}

func TestLoad_FileValueTypes(t *testing.T) {
	for _, file := range []string{`{"http_addr": 8080}`, `{"database_app_role": 1}`} {
		_, _, err := Load(env(prod(map[string]string{"IAM_CONFIG_FILE": "f"})), func(string) ([]byte, error) { return []byte(file), nil })
		if err == nil {
			t.Fatalf("%s: number accepted for a string field", file)
		}
	}
}

func TestLoad_UnknownKeyNotEchoed(t *testing.T) {
	_, _, err := Load(env(prod(map[string]string{"IAM_CONFIG_FILE": "f"})), func(string) ([]byte, error) { return []byte(`{"hunter2": 1}`), nil })
	if err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoad_ParseErrorsNeverEchoValues(t *testing.T) {
	for _, name := range []string{"IAM_SHUTDOWN_TIMEOUT", "IAM_ARGON2_MEMORY_KIB", "IAM_ARGON2_PARALLELISM"} {
		_, _, err := Load(env(prod(map[string]string{name: "s3cr3t-token"})), noFile)
		if err == nil || !strings.Contains(err.Error(), name) || strings.Contains(err.Error(), "s3cr3t") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestLoad_MalformedPublicBaseURL(t *testing.T) {
	for _, u := range []string{"https://x?", "https://x#", "https://:443", "https://x/?a=b", "https://x/", "HTTPS://x", "https://u:p@x", "ftp://x", "https://x/a#f", "https://x:", "https://x:/a"} {
		_, _, err := Load(env(prod(map[string]string{"IAM_PUBLIC_BASE_URL": u})), noFile)
		if err == nil {
			t.Errorf("%q accepted", u)
		}
	}
	for _, u := range []string{"https://iam.example.com", "https://iam.example.com:8443/base"} {
		if _, _, err := Load(env(prod(map[string]string{"IAM_PUBLIC_BASE_URL": u})), noFile); err != nil {
			t.Errorf("%q rejected: %v", u, err)
		}
	}
}

func TestLoad_ReportsAllInvalidFields(t *testing.T) {
	_, _, err := Load(env(prod(map[string]string{
		"IAM_SHUTDOWN_TIMEOUT":   "abc",
		"IAM_ARGON2_PARALLELISM": "many",
	})), noFile)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, name := range []string{"IAM_SHUTDOWN_TIMEOUT", "IAM_ARGON2_PARALLELISM"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not name %s: %v", name, err)
		}
	}
}

func TestLoad_ProhibitionsFailStartup(t *testing.T) {
	cases := map[string]map[string]string{
		"http issuer in lab":     prod(map[string]string{"IAM_ENVIRONMENT": "lab", "IAM_PUBLIC_BASE_URL": "http://iam.example.com"}),
		"http issuer in default": prod(map[string]string{"IAM_PUBLIC_BASE_URL": "http://iam.example.com"}),
		"empty cursor key":       prod(map[string]string{"IAM_CURSOR_KEY": "", "IAM_ENVIRONMENT": "lab"}),
		"unknown environment":    prod(map[string]string{"IAM_ENVIRONMENT": "staging"}),
	}
	for name, m := range cases {
		_, findings, err := Load(env(m), noFile)
		if err == nil || !strings.Contains(err.Error(), "ADR 0012") {
			t.Errorf("%s: err = %v", name, err)
		}
		found := false
		for _, f := range findings {
			found = found || f.Severity == Prohibited
		}
		if !found {
			t.Errorf("%s: no Prohibited finding", name)
		}
	}
}

func TestLoad_DevelopmentRelaxesOnlyWhatIsAllowed(t *testing.T) {
	cfg, findings, err := Load(env(map[string]string{
		"IAM_ENVIRONMENT":     "development",
		"IAM_PUBLIC_BASE_URL": "http://127.0.0.1:8080",
	}), noFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CursorKey != "" {
		t.Fatal("config must not invent a key")
	}
	if len(findings) != 2 || findings[0].Severity != Warning {
		t.Fatalf("findings: %+v", findings)
	}
}

func TestLoad_WeakSettingsProduceAlerts(t *testing.T) {
	_, findings, err := Load(env(prod(map[string]string{"IAM_ARGON2_MEMORY_KIB": "8192"})), noFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Severity != Critical || findings[0].Code != "ARGON2_MEMORY_LOW" ||
		!strings.Contains(findings[0].Reference, "OWASP") {
		t.Fatalf("findings: %+v", findings)
	}
	_, findings, err = Load(env(prod(map[string]string{"IAM_ARGON2_ITERATIONS": "1"})), noFile)
	if err != nil || len(findings) != 1 || findings[0].Code != "ARGON2_ITERATIONS_LOW" {
		t.Fatalf("iterations: %+v %v", findings, err)
	}
}

func TestLoad_SecretsNeverInErrors(t *testing.T) {
	_, _, err := Load(env(prod(map[string]string{"IAM_DATABASE_URL": "postgres://u:hunter2@h"})), noFile) //nolint:gosec // fake credential: the test asserts it never leaks
	if err == nil || !strings.Contains(err.Error(), "IAM_DATABASE_URL") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("secret leaked: %v", err)
	}
	_, _, err = Load(env(prod(map[string]string{"IAM_CURSOR_KEY": "sekrit-not-base64!"})), noFile)
	if err == nil || strings.Contains(err.Error(), "sekrit") {
		t.Fatalf("cursor key: %v", err)
	}
}

func TestLoad_FileReadErrorDoesNotLeak(t *testing.T) {
	_, _, err := Load(env(prod(map[string]string{"IAM_CONFIG_FILE": "/x"})), func(string) ([]byte, error) {
		return nil, errors.New("boom hunter2")
	})
	if err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("err = %v", err)
	}
}

func TestDocument_ListsEveryField(t *testing.T) {
	var b bytes.Buffer
	if err := Document(&b); err != nil {
		t.Fatal(err)
	}
	var cfg Config
	fs := fieldsOf(reflectValue(&cfg))
	if len(fs) < 12 {
		t.Fatalf("only %d fields discovered", len(fs))
	}
	for _, f := range fs {
		if !strings.Contains(b.String(), "`"+f.env+"`") {
			t.Errorf("%s missing from the reference", f.env)
		}
		if f.doc == "" {
			t.Errorf("%s has no doc tag", f.env)
		}
	}
	if strings.Contains(b.String(), goodKey) {
		t.Fatal("secret in documentation")
	}
}

// messageVocabulary collects the fixed text Load can produce, by loading a
// configuration that trips every rule, plus the generated reference.
func messageVocabulary() string {
	_, _, err := Load(env(map[string]string{
		"IAM_ENVIRONMENT": "bogus", "IAM_PUBLIC_BASE_URL": "http://x/", "IAM_DATABASE_URL": "zz", "IAM_CURSOR_KEY": "!",
		"IAM_SHUTDOWN_TIMEOUT": "zz", "IAM_ARGON2_MEMORY_KIB": "zz", "IAM_ARGON2_ITERATIONS": "0", "IAM_DATABASE_APP_ROLE": " ",
		"IAM_CONFIG_FILE": "f",
	}), func(string) ([]byte, error) { return []byte(`{"x": 1, "http_addr": 1, "http_addr": 2}`), nil })
	var doc bytes.Buffer
	_ = Document(&doc)
	return err.Error() + doc.String()
}

func FuzzLoad(f *testing.F) {
	vocab := messageVocabulary()
	f.Add("https://iam.example.com", goodKey, "postgres://u:hunter2secret@h")
	f.Add("https://user:pw1234@x", goodKey, "postgres://u:hunter2secret@h/")
	for _, s := range []string{"", "10s", "abc", "0s", "-1h", "9999999999999h"} {
		f.Add(s, s, s)
	}
	f.Add("https://iam.example.com", goodKey, "postgres://h/db")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		m := map[string]string{
			"IAM_SHUTDOWN_TIMEOUT": a, "IAM_PUBLIC_BASE_URL": a, "IAM_CURSOR_KEY": b,
			"IAM_DATABASE_URL": c, "IAM_ENVIRONMENT": b, "IAM_ARGON2_MEMORY_KIB": c,
		}
		cfg, _, err := Load(env(m), noFile)
		if err != nil {
			msg := err.Error()
			for _, v := range m {
				if len(v) >= 4 && !strings.Contains(vocab, v) && strings.Contains(msg, v) {
					t.Fatalf("a field value was echoed in the error")
				}
			}
			return
		}
		assertSafe(t, cfg)
	})
}

func assertSafe(t *testing.T, cfg Config) {
	t.Helper()
	if cfg.ShutdownTimeout <= 0 || !cfg.Environment.known() || cfg.Database.StatementTimeout < time.Millisecond {
		t.Fatalf("accepted invalid config: %+v", cfg)
	}
	if cfg.Environment != Development && (!strings.HasPrefix(strings.ToLower(cfg.PublicBaseURL), "https://") || cfg.CursorKey == "") {
		t.Fatalf("prohibition bypassed: %+v", cfg)
	}
}

func reflectValue(c *Config) reflect.Value { return reflect.ValueOf(c).Elem() }

func TestLoad_DatabaseLimitsAreBounded(t *testing.T) {
	for _, name := range []string{
		"IAM_DATABASE_STATEMENT_TIMEOUT", "IAM_DATABASE_IDLE_IN_TRANSACTION_TIMEOUT", "IAM_DATABASE_LOCK_TIMEOUT",
		"IAM_DATABASE_MAX_CONN_IDLE_TIME", "IAM_MIGRATION_STATEMENT_TIMEOUT", "IAM_MIGRATION_LOCK_TIMEOUT",
	} {
		for _, v := range []string{"0s", "-1s", "500us", "2h"} {
			if _, _, err := Load(env(prod(map[string]string{name: v})), noFile); err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("%s=%s: error %v, want it refused by name", name, v, err)
			}
		}
		if _, _, err := Load(env(prod(map[string]string{name: "1ms"})), noFile); err != nil {
			t.Errorf("%s=1ms: %v", name, err)
		}
	}
}
