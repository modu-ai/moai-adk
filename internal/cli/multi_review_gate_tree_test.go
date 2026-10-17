package cli

// Card t1632, sync-audit r10 delta F11: a config-orphaned worktree persists its
// convergence result as <session>--tree-<key>.json. The Claude hook reads every
// tree's result (loadConvergenceResults); the Codex chain's member 7 must read
// them all too, or the same project blocks on one runtime and passes on the
// other.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestCodexStopChain_MultiReviewMemberReadsEveryTreeResult(t *testing.T) {
	withMultiChangeDetector(t, true)
	const session = "sess-tree"
	const treeFile = session + "--tree-0123456789ab.json"

	setup := func(t *testing.T, base, tree ConvergenceResult) string {
		t.Helper()
		root := writeMultiState(t, base, session)
		dir := filepath.Join(root, ".moai", "state", "audit-multi")
		b, err := json.Marshal(tree)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, treeFile), b, 0o644); err != nil {
			t.Fatal(err)
		}
		cfgDir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(cfgDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte("workflow:\n  multi:\n    review_gate:\n      enabled: true\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return root
	}

	for name, tc := range map[string]struct {
		tree      ConvergenceResult
		wantBlock bool
	}{
		"tree result with a required gate unmet": {requiredGateUnmetResult(), true},
		"tree result with a required fail":       {requiredFailResult(), true},
		"every tree passes":                      {allPassResult(), false},
	} {
		t.Run(name, func(t *testing.T) {
			root := setup(t, allPassResult(), tc.tree)

			// The Claude hook is the reference: it reads every tree's result.
			claude, err := HandleMultiReviewGate(gateInput(false), true, root, session)
			if err != nil {
				t.Fatal(err)
			}
			if blocked := claude != nil && claude.Decision == hook.DecisionBlock; blocked != tc.wantBlock {
				t.Fatalf("claude hook blocked=%v, want %v (%+v)", blocked, tc.wantBlock, claude)
			}

			got := newCodexStopChain(root, stopInput(session, false)).multiReviewMember(context.Background())
			if denied := got.Class == reasonGateFailed; denied != tc.wantBlock {
				t.Fatalf("codex member denied=%v, want %v (decision=%v class=%v status=%v)", denied, tc.wantBlock, got.Decision, got.Class, got.Status)
			}
		})
	}
}
