package main

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestEnvOrDefault_SetValue(t *testing.T) {
	const key = "IMG_PROXY_TEST_ENV_SET"
	t.Setenv(key, "custom-value")
	if got := envOrDefault(key, "fallback"); got != "custom-value" {
		t.Errorf("envOrDefault(%q, fallback) = %q, want custom-value", key, got)
	}
}

func TestEnvOrDefault_EmptyValue(t *testing.T) {
	const key = "IMG_PROXY_TEST_ENV_EMPTY"
	t.Setenv(key, "")
	if got := envOrDefault(key, "fallback"); got != "fallback" {
		t.Errorf("envOrDefault(%q, fallback) = %q, want fallback", key, got)
	}
}

func TestNewHTTPServer(t *testing.T) {
	srv := newHTTPServer("127.0.0.1:9999", http.NewServeMux())
	if srv.Addr != "127.0.0.1:9999" {
		t.Errorf("Addr = %q, want 127.0.0.1:9999", srv.Addr)
	}
	if srv.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, readHeaderTimeout)
	}
}

func TestNewGenerator_Defaults(t *testing.T) {
	t.Setenv("CODEX_BIN", "")
	t.Setenv("CODEX_TIMEOUT", "")
	gen := newGenerator(slogDiscard())
	if gen.Bin != "codex" {
		t.Errorf("Bin = %q, want codex", gen.Bin)
	}
}

func TestNewGenerator_Override(t *testing.T) {
	t.Setenv("CODEX_BIN", "/custom/codex")
	t.Setenv("CODEX_TIMEOUT", "30s")
	gen := newGenerator(slogDiscard())
	if gen.Bin != "/custom/codex" {
		t.Errorf("Bin = %q, want /custom/codex", gen.Bin)
	}
	if gen.Timeout.String() != "30s" {
		t.Errorf("Timeout = %v, want 30s", gen.Timeout)
	}
}

func slogDiscard() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
