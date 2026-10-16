//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/verify"
)

// TestReviewGate_BlockLockRefusesSymlink — SPEC-REVIEW-GATE-REPAIR-001 F2 (delta
// audit iter2). The per-key block lock is taken by opening its lock path. Before
// the repair the open followed a symbolic link at that path, so the owner record
// was written into the link's target, a file outside the gate's state directory.
// The repaired open refuses a symbolic-link lock path: the target keeps its
// content, the link stays in place, and the turn end allows fail-open under
// codexStateBlockRecordFailed (REQ-RG-004, REQ-RG-007).
func TestReviewGate_BlockLockRefusesSymlink(t *testing.T) {
	root := m3GateFixture(t)
	state := m3ReceiptState(t, root)
	now := time.Now()
	fail := verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     1,
		Verdict:      codexReviewVerdictFail,
		RecordedAt:   now,
	}

	// The sentinel is the link target: a file outside the state directory whose
	// content must survive the turn end unchanged.
	const sentinelBody = "sentinel: the block lock must not write through a symbolic link\n"
	target := filepath.Join(t.TempDir(), "sentinel-target")
	if err := os.WriteFile(target, []byte(sentinelBody), 0o644); err != nil {
		t.Fatalf("write sentinel target: %v", err)
	}
	lockPath := codexReviewBlockLockPath(root, state)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("mkdir lock dir: %v", err)
	}
	if err := os.Symlink(target, lockPath); err != nil {
		t.Fatalf("symlink lock path to sentinel: %v", err)
	}

	out := codexReviewGateFail(root, state, fail, now)

	if got, err := os.ReadFile(target); err != nil {
		t.Errorf("read sentinel target: %v", err)
	} else if string(got) != sentinelBody {
		t.Errorf("the block lock wrote through the symbolic link: the target now holds %q, want the sentinel %q", got, sentinelBody)
	}
	if info, err := os.Lstat(lockPath); err != nil {
		t.Errorf("stat lock path: %v", err)
	} else if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("the lock path must remain a symbolic link, got mode %v", info.Mode())
	}
	m3WantAllowWithState(t, out, codexStateBlockRecordFailed)
}
