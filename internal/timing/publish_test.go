package timing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPublishAppendsToStepSummary proves the calibration line reaches the
// GitHub Actions job summary — the surface that stays readable on a PASSING
// run, where `go test` discards the package's own output.
func TestPublishAppendsToStepSummary(t *testing.T) {
	summary := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summary)

	publish("first: ratio=1.02x")
	publish("second: ratio=1.09x")

	got, err := os.ReadFile(summary)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	for _, want := range []string{"first: ratio=1.02x", "second: ratio=1.09x"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("summary missing %q; got:\n%s", want, got)
		}
	}
}

// TestPublishNoopWithoutStepSummary pins the local-run contract: no env var,
// no file, no error.
func TestPublishNoopWithoutStepSummary(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "")
	publish("ignored: ratio=1.00x")
}

// TestPublishIgnoresUnwritableTarget proves an unwritable summary path cannot
// turn a passing latency assertion red.
func TestPublishIgnoresUnwritableTarget(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", filepath.Join(t.TempDir(), "no-such-dir", "summary.md"))
	publish("ignored: ratio=1.00x")
}
