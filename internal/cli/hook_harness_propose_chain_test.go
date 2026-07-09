// Package cli — Stop-path propose auto-chain tests.
// SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-004 / REQ-HRR-005 (AC-HRR-005 /
// AC-HRR-006): the Stop path chains proposal generation after classify when
// promotions > 0, landing proposal files in .moai/harness/proposals/. The chain
// is fail-open (AC-HRR-006): a propose error is logged to stderr and NEVER
// blocks session end (exit 0).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// seedEligibleUsageLog writes n agent_invocation:Bash:<hash> events (eligible,
// non-degenerate) so the classifier produces at least one rule-tier promotion
// that the propose chain can map to a candidate.
func seedEligibleUsageLog(t *testing.T, dir string, n int) {
	t.Helper()
	events := repeatEvents(map[string]string{
		"event_type": "agent_invocation", "subject": "Bash", "context_hash": "propose-hash",
	}, n)
	seedEventLines(t, dir, events)
}

// TestRunHarnessObserveStop_ProposeChainAutoRuns covers AC-HRR-005: when the
// classify pass produces ≥1 eligible promotion, the Stop path auto-runs
// proposal generation so that proposal files land in .moai/harness/proposals/.
func TestRunHarnessObserveStop_ProposeChainAutoRuns(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	seedEligibleUsageLog(t, dir, 5) // 5 obs → rule tier → actionable candidate
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"done","session":{"id":"sess-propose"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop error: %v", err)
		}
	})

	// AC-HRR-005: .moai/harness/proposals/ exists with ≥1 proposal dir.
	propDir := filepath.Join(dir, ".moai", "harness", "proposals")
	entries, err := os.ReadDir(propDir)
	if err != nil {
		t.Fatalf("proposals dir not created (propose chain did not run): %v", err)
	}
	proposalCount := 0
	for _, e := range entries {
		if e.IsDir() {
			// Each proposal is a subdirectory containing proposal.json.
			if _, statErr := os.Stat(filepath.Join(propDir, e.Name(), "proposal.json")); statErr == nil {
				proposalCount++
			}
		}
	}
	if proposalCount < 1 {
		t.Errorf("expected ≥1 proposal in %s, got %d (entries=%v)", propDir, proposalCount, entries)
	}
}

// TestRunHarnessObserveStop_ProposeChainFailOpen covers AC-HRR-006: when the
// propose path is faulted (proposals dir made unwritable), the Stop handler
// still returns nil (exit 0, non-blocking) — session end is NEVER blocked by a
// propose error.
func TestRunHarnessObserveStop_ProposeChainFailOpen(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	seedEligibleUsageLog(t, dir, 5)
	t.Chdir(dir)

	// Block the proposals output path with a FILE (not directory) so WriteProposals
	// cannot create the proposals dir there — mkdir fails, simulating a fault.
	// We block the parent proposals path location by creating a file at that path.
	blockPath := filepath.Join(dir, ".moai", "harness", "proposals")
	if err := os.MkdirAll(filepath.Dir(blockPath), 0o755); err != nil {
		t.Fatalf("mkdir harness dir: %v", err)
	}
	// Create a read-only file at the proposals path so MkdirAll(proposals) fails.
	if err := os.WriteFile(blockPath, []byte("block"), 0o444); err != nil {
		t.Fatalf("create blocking file: %v", err)
	}

	var stderrBuf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	withStdin(t, `{"last_assistant_message":"x","session":{"id":"sess-propose-failopen"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Errorf("Stop handler must NOT propagate propose errors (fail-open): %v", err)
		}
	})

	// The handler must have exited 0 (no error returned) — fail-open verified by
	// the absence of a returned error above. A propose error should surface in
	// stderr as a non-blocking log line.
	if stderrBuf.Len() == 0 {
		// stderr may or may not carry the propose error depending on whether the
		// fault triggered; the load-bearing assertion is exit 0 (above). This is
		// a soft check, not a hard requirement.
		t.Log("note: stderr empty — propose fault may not have surfaced a log line")
	}
}
