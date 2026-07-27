package quality

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestAstGrepGateReasons pins the ast-grep gate step's pass-path reason: the
// step still passes when sg is absent, but it now says so instead of returning
// the empty string that a clean scan also returns.
//
// The two halves sit at the frames where each is observable. The gate
// constructs its scanner with a fixed binary name, so only the
// scanner-unavailable branch is reachable end-to-end through the gate; the
// distinctness of the two reason classes is asserted on the constants
// directly. Scanner-level classification of a non-sentinel error is covered by
// TestScannerScan_UnavailableSentinel/non-sentinel.
//
// Non-parallel: the reachable half strips PATH with t.Setenv.
func TestAstGrepGateReasons(t *testing.T) {
	t.Run("unavailable reason reaches the gate step", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		projectDir := t.TempDir()
		// The pure-Go suppression sweep runs before the scan and denies on any
		// unpaired ast-grep-ignore, so the fixture carries none.
		if err := os.WriteFile(filepath.Join(projectDir, "main.go"),
			[]byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}

		ok, reason := RunAstGrepGateV2(context.Background(), projectDir, &AstGrepGateConfig{
			Enabled:  true,
			RulesDir: ".moai/config/astgrep-rules",
		})

		if !ok {
			t.Fatalf("RunAstGrepGateV2: want pass when sg is absent, got deny with %q", reason)
		}
		if reason == "" {
			t.Fatal(`RunAstGrepGateV2: want a non-empty pass reason naming the skip, got "" — indistinguishable from a clean scan`)
		}
		if reason != astGrepReasonScannerUnavailable {
			t.Errorf("reason: want the scanner-unavailable constant %q, got %q", astGrepReasonScannerUnavailable, reason)
		}
	})

	t.Run("the two pass reasons are distinct", func(t *testing.T) {
		if astGrepReasonScannerUnavailable == "" {
			t.Error("astGrepReasonScannerUnavailable must be non-empty")
		}
		if astGrepReasonScanDegraded == "" {
			t.Error("astGrepReasonScanDegraded must be non-empty")
		}
		if astGrepReasonScannerUnavailable == astGrepReasonScanDegraded {
			t.Errorf("the scanner-unavailable and degraded-scan reasons collapsed to one string: %q", astGrepReasonScannerUnavailable)
		}
	})
}
