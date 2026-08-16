package cli

// session_worktree_prmerge_anchor_test.go — t73 surface 3: the PR-merge
// auto-cleanup must skip (preserve) a worktree that anchors a live session,
// in the same immediately-before-removal position as the REQ-SW-024 dirty
// guard.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// writePRMergeAnchorRegistry writes a live-lane entry (stale heartbeat, live
// PID — no per-turn heartbeat driver exists) into the tree-local registry.
func writePRMergeAnchorRegistry(t *testing.T, tree string) {
	t.Helper()
	host, _ := os.Hostname()
	old := time.Now().UTC().Add(-2 * time.Hour)
	entries := []session.Entry{{
		SessionID:     "99999999-8888-7777-6666-555555555555",
		SpecID:        "(none)",
		Phase:         "(none)",
		StartedAt:     old,
		LastHeartbeat: old,
		PID:           os.Getpid(),
		Host:          host,
		CWD:           tree,
	}}
	dir := filepath.Join(tree, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active-sessions.json"), data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

// TestPRMergeCleanup_AnchoredSessionSkipsRemoval (t73 surface 3, priority):
// gh primary path confirms MERGED and the dirty guard passes, yet the
// removal must be skipped because a live session is anchored in the tree.
func TestPRMergeCleanup_AnchoredSessionSkipsRemoval(t *testing.T) {
	tree := t.TempDir()
	writePRMergeAnchorRegistry(t, tree)
	wtBranch := "WT-t73anchor"

	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(tree, wtBranch), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) string { return "MERGED" },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:     func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if removeCalled {
		t.Fatal("PR-merge cleanup removed a worktree with a live anchored session")
	}
	if !strings.Contains(out.String(), "live session anchored") {
		t.Fatalf("expected a preserved notice naming the anchored session, got: %q", out.String())
	}
	if !strings.Contains(out.String(), tree) {
		t.Fatalf("preserved notice must name the worktree path %q, got: %q", tree, out.String())
	}
}
