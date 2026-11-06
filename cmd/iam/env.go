package main

import (
	"fmt"
	"time"
)

// envConfig is a stopgap until typed configuration lands (contract C1).
type envConfig struct {
	HTTPAddr        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

func loadEnv(getenv func(string) string) (envConfig, error) {
	cfg := envConfig{
		HTTPAddr:        ":8080",
		DatabaseURL:     getenv("IAM_DATABASE_URL"),
		ShutdownTimeout: 10 * time.Second,
	}
	if v := getenv("IAM_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := getenv("IAM_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return envConfig{}, fmt.Errorf("IAM_SHUTDOWN_TIMEOUT: %w", err)
		}
		if d <= 0 {
			return envConfig{}, fmt.Errorf("IAM_SHUTDOWN_TIMEOUT: must be greater than zero, got %q", v)
		}
		cfg.ShutdownTimeout = d
	}
	return cfg, nil
}
