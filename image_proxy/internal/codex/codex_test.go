package codex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewestPNGSince(t *testing.T) {
	dir := t.TempDir()
	base := time.Now()

	// Old png, modified before the cutoff: must be ignored.
	old := filepath.Join(dir, "old.png")
	writeFile(t, old, "old")
	chtime(t, old, base.Add(-time.Hour))

	// A non-png newer than cutoff: must be ignored.
	writeFile(t, filepath.Join(dir, "note.txt"), "txt")

	// Newest png in a nested session dir: must be selected.
	sess := filepath.Join(dir, "session")
	if err := os.MkdirAll(sess, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(sess, "ig_new.png")
	writeFile(t, want, "newest")
	chtime(t, want, base.Add(time.Minute))

	got, err := newestPNGSince(dir, base)
	if err != nil {
		t.Fatalf("newestPNGSince: %v", err)
	}
	if string(got) != "newest" {
		t.Fatalf("selected wrong file: got %q want %q", got, "newest")
	}
}

func TestNewestPNGSince_NoneRecent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "old.png")
	writeFile(t, p, "old")
	chtime(t, p, time.Now().Add(-time.Hour))

	if _, err := newestPNGSince(dir, time.Now()); err == nil {
		t.Fatal("expected error when no recent png exists")
	}
}

func TestGenerate_EmptyPrompt(t *testing.T) {
	c := &CLI{Run: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		t.Fatal("Run must not be called for empty prompt")
		return nil, nil
	}}
	if _, err := c.Generate(context.Background(), "", ""); err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

// TestGenerate_ReadsRequestedFile verifies the happy path: the fake runner
// writes out.png into the working dir codex was given, the size hint reaches
// the codex prompt, and Generate returns the file bytes.
func TestGenerate_ReadsRequestedFile(t *testing.T) {
	c := &CLI{
		Bin:          "codex",
		GeneratedDir: t.TempDir(),
		Timeout:      time.Minute,
		Run: func(_ context.Context, _ string, args ...string) ([]byte, error) {
			workdir := workdirFromArgs(args)
			if workdir == "" {
				t.Fatalf("no -C workdir in args: %v", args)
			}
			if prompt := args[len(args)-1]; !strings.Contains(prompt, "512x512") {
				t.Errorf("size hint missing from codex prompt: %q", prompt)
			}
			if err := os.WriteFile(filepath.Join(workdir, "out.png"), []byte("PNGDATA"), 0o644); err != nil {
				return nil, err
			}
			return []byte("ok"), nil
		},
	}
	got, err := c.Generate(context.Background(), "a red circle", "512x512")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if string(got) != "PNGDATA" {
		t.Fatalf("got %q want %q", got, "PNGDATA")
	}
}

func TestGenerate_NoImageProduced(t *testing.T) {
	c := &CLI{
		GeneratedDir: t.TempDir(),
		Timeout:      time.Minute,
		Run: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
			return []byte("did nothing"), nil
		},
	}
	if _, err := c.Generate(context.Background(), "x", ""); err == nil {
		t.Fatal("expected error when codex produces no image")
	}
}

// TestGenerate_WorkdirUnderBase verifies the per-request working dir (where the
// image is dropped) is created under WorkdirBase, so it can be a mounted volume.
func TestGenerate_WorkdirUnderBase(t *testing.T) {
	base := t.TempDir()
	c := &CLI{
		GeneratedDir: t.TempDir(),
		Timeout:      time.Minute,
		WorkdirBase:  base,
		Run: func(_ context.Context, _ string, args ...string) ([]byte, error) {
			wd := workdirFromArgs(args)
			if !strings.HasPrefix(wd, base) {
				t.Errorf("workdir %q not under base %q", wd, base)
			}
			if err := os.WriteFile(filepath.Join(wd, "out.png"), []byte("P"), 0o644); err != nil {
				return nil, err
			}
			return []byte("ok"), nil
		},
	}
	if _, err := c.Generate(context.Background(), "x", ""); err != nil {
		t.Fatalf("Generate: %v", err)
	}
}

func workdirFromArgs(args []string) string {
	for i, a := range args {
		if a == "-C" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chtime(t *testing.T, path string, mod time.Time) {
	t.Helper()
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
}
