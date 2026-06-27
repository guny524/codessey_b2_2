package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeGen records what it received and returns a canned result.
type fakeGen struct {
	png       []byte
	err       error
	gotPrompt string
	gotSize   string
}

func (f *fakeGen) Generate(_ context.Context, prompt, size string) ([]byte, error) {
	f.gotPrompt = prompt
	f.gotSize = size
	return f.png, f.err
}

// panicGen fails the test if Generate is ever called (used to prove the
// handler rejects bad input before reaching generation).
type panicGen struct{}

func (panicGen) Generate(_ context.Context, _, _ string) ([]byte, error) {
	panic("Generate must not be called")
}

func do(t *testing.T, s *Server, method, path, body, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func decodeB64(t *testing.T, rec *httptest.ResponseRecorder) []byte {
	t.Helper()
	var resp imageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("data len = %d, want 1", len(resp.Data))
	}
	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		t.Fatalf("b64 decode: %v", err)
	}
	return b
}

func TestGenerate_Success(t *testing.T) {
	fg := &fakeGen{png: []byte("PNGDATA")}
	s := NewServer(fg, "", nil)
	body := `{"prompt":"a cat","size":"512x512","n":2,"model":"x","response_format":"url"}`
	rec := do(t, s, http.MethodPost, "/v1/images/generations", body, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if got := decodeB64(t, rec); string(got) != "PNGDATA" {
		t.Fatalf("png = %q, want PNGDATA", got)
	}
	if fg.gotPrompt != "a cat" || fg.gotSize != "512x512" {
		t.Fatalf("generator got prompt=%q size=%q", fg.gotPrompt, fg.gotSize)
	}
}

func TestGenerate_NoSizeMeansNoHint(t *testing.T) {
	fg := &fakeGen{png: []byte("X")}
	s := NewServer(fg, "", nil)
	rec := do(t, s, http.MethodPost, "/v1/images/generations", `{"prompt":"a cat"}`, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if fg.gotSize != "" {
		t.Fatalf("size = %q, want empty", fg.gotSize)
	}
}

func TestGenerate_InvalidSize(t *testing.T) {
	for _, body := range []string{
		`{"prompt":"a cat","size":"1024"}`,
		`{"prompt":"a cat","size":"1024x"}`,
		`{"prompt":"a cat","size":"x1024"}`,
		`{"prompt":"a cat","size":"10x20x30"}`,
		`{"prompt":"a cat","size":"abcxdef"}`,
	} {
		t.Run(body, func(t *testing.T) {
			s := NewServer(panicGen{}, "", nil)
			rec := do(t, s, http.MethodPost, "/v1/images/generations", body, "")
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGenerate_MissingPrompt(t *testing.T) {
	s := NewServer(panicGen{}, "", nil)
	rec := do(t, s, http.MethodPost, "/v1/images/generations", `{"prompt":"   "}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGenerate_BadJSON(t *testing.T) {
	s := NewServer(panicGen{}, "", nil)
	rec := do(t, s, http.MethodPost, "/v1/images/generations", `{not json`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGenerate_GeneratorError(t *testing.T) {
	s := NewServer(&fakeGen{err: errors.New("boom")}, "", nil)
	rec := do(t, s, http.MethodPost, "/v1/images/generations", `{"prompt":"a cat"}`, "")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestGenerate_AuthRequired(t *testing.T) {
	s := NewServer(&fakeGen{png: []byte("X")}, "secret", nil)

	rec := do(t, s, http.MethodPost, "/v1/images/generations", `{"prompt":"a cat"}`, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no key: status = %d, want 401", rec.Code)
	}

	rec = do(t, s, http.MethodPost, "/v1/images/generations", `{"prompt":"a cat"}`, "Bearer secret")
	if rec.Code != http.StatusOK {
		t.Fatalf("good key: status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
}

func TestGenerate_MethodNotAllowed(t *testing.T) {
	s := NewServer(panicGen{}, "", nil)
	rec := do(t, s, http.MethodGet, "/v1/images/generations", "", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	s := NewServer(panicGen{}, "", nil)
	rec := do(t, s, http.MethodGet, "/healthz", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
