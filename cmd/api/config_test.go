package main

import (
	"net/url"
	"testing"
)

func TestOracleDSNFromEnvBuildsSIDConnection(t *testing.T) {
	t.Setenv("ORACLE_SID", "DEMO")
	t.Setenv("ORACLE_HOST", "127.0.0.1")
	t.Setenv("ORACLE_PORT", "1522")
	t.Setenv("ORACLE_USERNAME", "reader")
	t.Setenv("ORACLE_PASSWORD", "secret")

	dsn, err := oracleDSNFromEnv()
	if err != nil {
		t.Fatalf("oracleDSNFromEnv() error = %v", err)
	}

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	if parsed.Scheme != "oracle" || parsed.Host != "127.0.0.1:1522" {
		t.Fatalf("DSN = %q", dsn)
	}
	if parsed.Query().Get("SID") != "DEMO" {
		t.Fatalf("SID = %q, want DEMO", parsed.Query().Get("SID"))
	}
}

func TestOracleDSNFromEnvRejectsInvalidPort(t *testing.T) {
	t.Setenv("ORACLE_SID", "DEMO")
	t.Setenv("ORACLE_HOST", "127.0.0.1")
	t.Setenv("ORACLE_PORT", "not-a-port")
	t.Setenv("ORACLE_USERNAME", "reader")
	t.Setenv("ORACLE_PASSWORD", "secret")

	if _, err := oracleDSNFromEnv(); err == nil {
		t.Fatal("oracleDSNFromEnv() error = nil, want invalid port error")
	}
}
