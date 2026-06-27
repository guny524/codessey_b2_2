// Command image_proxy serves an OpenAI-compatible
// /v1/images/generations endpoint backed by the Codex CLI's built-in image
// generation (ChatGPT OAuth). No OPENAI_API_KEY and no per-call API charge.
//
// Configuration (environment variables):
//
//	IMAGE_PROXY_ADDR     listen address      (default 127.0.0.1:8080)
//	IMAGE_PROXY_API_KEY  optional bearer key (default off; set when exposed)
//	CODEX_BIN            codex executable    (default "codex")
//	CODEX_TIMEOUT        per-request timeout (default 3m, e.g. "90s")
//	IMAGE_PROXY_WORKDIR  base dir for image output (default OS temp; set to a volume)
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"image_proxy/internal/codex"
	"image_proxy/internal/server"
)

const (
	defaultAddr       = "127.0.0.1:8080"
	readHeaderTimeout = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	gen := newGenerator(logger)
	addr := envOrDefault("IMAGE_PROXY_ADDR", defaultAddr)
	apiKey := os.Getenv("IMAGE_PROXY_API_KEY")

	srv := server.NewServer(gen, apiKey, logger)
	httpSrv := newHTTPServer(addr, srv.Handler())

	logger.Info("image_proxy listening",
		"addr", addr, "auth", apiKey != "", "codex_bin", gen.Bin, "timeout", gen.Timeout)
	if err := httpSrv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// newGenerator builds the codex-backed image generator from the environment.
func newGenerator(logger *slog.Logger) *codex.CLI {
	gen := codex.NewCLI()
	if bin := os.Getenv("CODEX_BIN"); bin != "" {
		gen.Bin = bin
	}
	if d := os.Getenv("CODEX_TIMEOUT"); d != "" {
		if dur, err := time.ParseDuration(d); err == nil {
			gen.Timeout = dur
		} else {
			logger.Warn("invalid CODEX_TIMEOUT, using default", "value", d, "default", gen.Timeout)
		}
	}
	if wd := os.Getenv("IMAGE_PROXY_WORKDIR"); wd != "" {
		gen.WorkdirBase = wd
	}
	return gen
}

func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
