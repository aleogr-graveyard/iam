package main

import (
	"strings"
	"testing"
	"time"
)

func fakeEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadEnv_Defaults(t *testing.T) {
	got, err := loadEnv(fakeEnv(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := envConfig{HTTPAddr: ":8080", DatabaseURL: "", ShutdownTimeout: 10 * time.Second}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadEnv_Overrides(t *testing.T) {
	got, err := loadEnv(fakeEnv(map[string]string{
		"IAM_HTTP_ADDR":        "127.0.0.1:9",
		"IAM_DATABASE_URL":     "postgres://h/db",
		"IAM_SHUTDOWN_TIMEOUT": "3s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := envConfig{HTTPAddr: "127.0.0.1:9", DatabaseURL: "postgres://h/db", ShutdownTimeout: 3 * time.Second}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadEnv_InvalidShutdownTimeout(t *testing.T) {
	for _, v := range []string{"abc", "0s", "-1s"} {
		_, err := loadEnv(fakeEnv(map[string]string{"IAM_SHUTDOWN_TIMEOUT": v}))
		if err == nil || !strings.Contains(err.Error(), "IAM_SHUTDOWN_TIMEOUT") {
			t.Fatalf("%q: err = %v", v, err)
		}
	}
}

func FuzzLoadEnv(f *testing.F) {
	for _, s := range []string{"", "10s", "abc", "0s", "-1h", "9999999999999h"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, v string) {
		got, err := loadEnv(fakeEnv(map[string]string{"IAM_SHUTDOWN_TIMEOUT": v}))
		if err == nil && got.ShutdownTimeout <= 0 {
			t.Fatalf("%q: accepted non-positive timeout %v", v, got.ShutdownTimeout)
		}
	})
}
