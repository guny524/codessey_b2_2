package codex

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	defaultTimeout      = 3 * time.Minute
	truncateOutputBytes = 500
)

// Generator turns a text prompt (and optional WxH size hint) into PNG bytes.
type Generator interface {
	Generate(ctx context.Context, prompt, size string) ([]byte, error)
}

// Runner executes a command and returns its combined output. It is injectable
// so tests can avoid invoking the real codex binary.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// CLI implements Generator by shelling out to `codex exec` with the built-in
// image generation tool.
type CLI struct {
	// Bin is the codex executable name or path. Defaults to "codex".
	Bin string
	// GeneratedDir is the directory the built-in tool drops images into. It is
	// used as a fallback when the explicitly requested output file is absent.
	// Defaults to ~/.codex/generated_images.
	GeneratedDir string
	// Timeout bounds a single generation. Defaults to 3 minutes.
	Timeout time.Duration
	// WorkdirBase is the parent dir for per-request working dirs; empty means
	// the OS temp dir. Point it at a mounted volume to isolate image output.
	WorkdirBase string
	// Run executes the command. Defaults to execRun.
	Run Runner
}

// NewCLI returns a CLI with production defaults.
func NewCLI() *CLI {
	home, _ := os.UserHomeDir()
	return &CLI{
		Bin:          "codex",
		GeneratedDir: filepath.Join(home, ".codex", "generated_images"),
		Timeout:      defaultTimeout,
		Run:          execRun,
	}
}

func execRun(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// promptTemplate constrains codex to a single image-generation action. The
// user-supplied description is wrapped with %q so it is delimited as data
// rather than additional instructions, reducing prompt-injection surface. The
// leading %s is an optional size clause (best-effort hint for codex).
const promptTemplate = `Use your built-in image generation tool to generate one image%s ` +
	`based on this description: %q. ` +
	`Save the resulting PNG file to exactly %s. ` +
	`Do only this and nothing else: do not create any other files, ` +
	`do not run any other commands, do not modify anything else.`

// Generate runs codex to produce an image for prompt and returns the PNG bytes.
// size, when non-empty (e.g. "1024x1024"), is passed to codex as a best-effort
// resolution hint; the actual output size is decided by the codex image model.
func (c *CLI) Generate(ctx context.Context, prompt, size string) (png []byte, err error) {
	if prompt == "" {
		return nil, errors.New("empty prompt")
	}

	workdir, err := os.MkdirTemp(c.WorkdirBase, "image-proxy-*")
	if err != nil {
		return nil, fmt.Errorf("create workdir: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(workdir); removeErr != nil {
			if err == nil {
				png = nil
			}
			err = errors.Join(err, fmt.Errorf("remove workdir: %w", removeErr))
		}
	}()

	outPath := filepath.Join(workdir, "out.png")
	since := time.Now()

	sizeClause := ""
	if size != "" {
		sizeClause = fmt.Sprintf(" at approximately %s resolution", size)
	}

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	args := []string{
		"exec",
		"--dangerously-bypass-approvals-and-sandbox",
		"--skip-git-repo-check",
		"-C", workdir,
		fmt.Sprintf(promptTemplate, sizeClause, prompt, outPath),
	}
	out, runErr := c.Run(ctx, c.Bin, args...)

	// Primary: the file codex was told to write.
	if b, readErr := os.ReadFile(outPath); readErr == nil && len(b) > 0 {
		return b, nil
	}
	// Fallback: the newest PNG the built-in tool dropped while we ran.
	if b, pngErr := newestPNGSince(c.GeneratedDir, since); pngErr == nil {
		return b, nil
	}
	if runErr != nil {
		return nil, fmt.Errorf("codex exec failed: %w: %s", runErr, truncate(out, truncateOutputBytes))
	}
	return nil, fmt.Errorf("codex produced no image; output: %s", truncate(out, truncateOutputBytes))
}

// newestPNGSince returns the bytes of the most recently modified .png under dir
// whose modification time is at or after since.
func newestPNGSince(dir string, since time.Time) ([]byte, error) {
	var newestPath string
	var newestMod time.Time

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".png" {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.ModTime().Before(since) {
			return nil
		}
		if info.ModTime().After(newestMod) {
			newestMod = info.ModTime()
			newestPath = path
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk generated images: %w", err)
	}

	if newestPath == "" {
		return nil, errors.New("no recent png found")
	}
	return os.ReadFile(newestPath)
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
