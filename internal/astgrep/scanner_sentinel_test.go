package astgrep_test

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestScannerScan_UnavailableSentinel pins the three-way classification Scan
// must make when asked to run: scanner-absent, scanner-rejected, and clean.
//
// Returning ([]Finding{}, nil) for the absent case is what made a missing
// scanner indistinguishable from a clean scan, so the first subtest asserts an
// error is returned at all before asserting which error it is.
//
// Every subtest is non-parallel: the absent case strips PATH with t.Setenv,
// which is incompatible with t.Parallel().
func TestScannerScan_UnavailableSentinel(t *testing.T) {
	t.Run("unavailable", func(t *testing.T) {
		// isSGAvailable resolves the binary with exec.LookPath, which reads the
		// process PATH; an empty directory removes sg from it.
		t.Setenv("PATH", t.TempDir())

		s := astgrep.NewScanner(&astgrep.ScannerConfig{
			RulesDir: t.TempDir(),
			SGBinary: "sg",
		})
		findings, err := s.Scan(context.Background(), ".")

		if err == nil {
			t.Fatal("Scan: want an error when sg is unresolvable, got nil — an empty-and-successful result is indistinguishable from a clean scan")
		}
		if !errors.Is(err, astgrep.ErrScannerUnavailable) {
			t.Errorf("errors.Is(err, ErrScannerUnavailable): want true, got false for %v", err)
		}
		if !strings.Contains(err.Error(), "sg") {
			t.Errorf("error must name the binary it sought, got: %v", err)
		}
		if !strings.Contains(err.Error(), "ast-grep.github.io") {
			t.Errorf("error must carry install guidance, got: %v", err)
		}
		if len(findings) != 0 {
			t.Errorf("findings: want empty on the unavailable path, got %d", len(findings))
		}
	})

	t.Run("non-sentinel", func(t *testing.T) {
		// A t.TempDir() path matches no trustedBinaryPrefixes entry, so
		// ValidateBinary rejects it before the availability probe runs. That
		// error must stay distinguishable from the unavailable sentinel —
		// otherwise the sentinel would swallow unrelated scan failures.
		s := astgrep.NewScanner(&astgrep.ScannerConfig{
			RulesDir: t.TempDir(),
			SGBinary: filepath.Join(t.TempDir(), "sg"),
		})
		_, err := s.Scan(context.Background(), ".")

		if err == nil {
			t.Fatal("Scan: want an error for an untrusted binary path, got nil")
		}
		if errors.Is(err, astgrep.ErrScannerUnavailable) {
			t.Errorf("sentinel over-matches: an untrusted-binary error must not satisfy errors.Is(..., ErrScannerUnavailable); got %v", err)
		}
	})

	t.Run("clean scan", func(t *testing.T) {
		if _, err := exec.LookPath("sg"); err != nil {
			t.Skip("sg not installed on this host — the clean-scan half of the discrimination is not runnable here")
		}
		// An absent rules directory short-circuits Scan just after the
		// availability probe, so this exercises the resolvable-binary path
		// without depending on any rule content.
		s := astgrep.NewScanner(&astgrep.ScannerConfig{
			RulesDir: filepath.Join(t.TempDir(), "absent-rules"),
			SGBinary: "sg",
		})
		findings, err := s.Scan(context.Background(), ".")

		if err != nil {
			t.Fatalf("Scan: want a nil error when sg resolves, got %v", err)
		}
		if errors.Is(err, astgrep.ErrScannerUnavailable) {
			t.Error("a clean scan must not satisfy errors.Is(..., ErrScannerUnavailable)")
		}
		if len(findings) != 0 {
			t.Errorf("findings: want empty for an absent rules dir, got %d", len(findings))
		}
	})
}
