package tmux

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunFunc is the function signature for executing shell commands.
// It allows injection of custom runners for testing.
type RunFunc func(ctx context.Context, name string, args ...string) (string, error)

// Detector checks whether tmux is available on the system.
type Detector interface {
	// IsAvailable returns true if tmux is installed and executable.
	IsAvailable() bool

	// Version returns the tmux version string (e.g., "3.4"), or an error.
	Version() (string, error)
}

// SystemDetector implements Detector using the system PATH.
type SystemDetector struct {
	run RunFunc
}

// Compile-time interface compliance check.
var _ Detector = (*SystemDetector)(nil)

// DetectorOption configures a SystemDetector.
type DetectorOption func(*SystemDetector)

// WithRunFunc sets a custom command runner (used for testing).
func WithRunFunc(fn RunFunc) DetectorOption {
	return func(d *SystemDetector) {
		d.run = fn
	}
}

// NewDetector creates a new SystemDetector.
func NewDetector(opts ...DetectorOption) *SystemDetector {
	d := &SystemDetector{
		run: defaultRun,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// IsAvailable returns true if tmux can be executed.
func (d *SystemDetector) IsAvailable() bool {
	_, err := d.run(context.Background(), "tmux", "-V")
	return err == nil
}

// Version returns the parsed tmux version string.
// It expects output in the format "tmux X.Y" and returns "X.Y".
func (d *SystemDetector) Version() (string, error) {
	output, err := d.run(context.Background(), "tmux", "-V")
	if err != nil {
		return "", fmt.Errorf("get tmux version: %w", err)
	}

	ver := parseVersion(output)
	if ver == "" {
		return "", fmt.Errorf("parse tmux version from %q: unexpected format", output)
	}

	return ver, nil
}

// parseVersion extracts the version from "tmux X.Y" output.
func parseVersion(output string) string {
	output = strings.TrimSpace(output)
	parts := strings.Fields(output)
	if len(parts) < 2 {
		return ""
	}
	// "tmux 3.4" -> "3.4"
	return parts[1]
}

// defaultRun executes a command using os/exec with context support.
func defaultRun(ctx context.Context, name string, args ...string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, ErrTmuxNotFound)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s: %s", name, errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}
