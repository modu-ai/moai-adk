package cli

// SPEC-CODEX-GATE-SCOPE-001 — the codex review gate's review-target scoping.
//
// File layout (test-first, per acceptance.md §C):
//   - M1 regression line (observed GREEN on the pre-change tree):
//     TestCodexReviewGate_TreeScopeRequestShapeUnchanged pins the non-card
//     session's review request shape-identical to its form before this SPEC
//     (REQ-CGS-003, the SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 line).
//   - M2 RED set (observed FAILING before the implementation): the scope
//     discriminator (AC-CGS-003/004/005/008), the card-scope request
//     (AC-CGS-001), the frozen-projectDir bypass (AC-CGS-006), the scoped
//     self-gate (AC-CGS-007), the scope-bound receipt key (AC-CGS-009/012),
//     the two-path parity (AC-CGS-010), and the scope log (AC-CGS-013).

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// cardScopeGit runs one git subcommand in dir with a pinned identity, so the
// fixture repositories are reproducible and never consult the operator's git
// configuration (same shape as stopFixture.git, but not bound to a fixture).
func cardScopeGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir,
		"-c", "user.name=fx", "-c", "user.email=fx@example.com",
		"-c", "commit.gpgsign=false", "-c", "maintenance.auto=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// --- M1: the tree-scope regression line (AC-CGS-002, GREEN pre-change) ---

// TestCodexReviewGate_TreeScopeRequestShapeUnchanged pins the non-card
// session's assembled review request: a session on a plain branch keeps the
// exact pre-SPEC shape — target uncommittedChanges (the REQ-CRT-006
// serialization) with cwd = the resolved tree. The card-scope work added by
// this SPEC sits BESIDE this shape, never over it (REQ-CGS-003, spec.md §F).
func TestCodexReviewGate_TreeScopeRequestShapeUnchanged(t *testing.T) {
	f := newStopFixture(t) // branch "main" — not a card branch
	f.dirty(t, "tree scope regression")

	sess := withCodexSession(t, codexSessionScript("- [P1] tree scope findings"))

	input := &hook.HookInput{SessionID: "sess-tree", CWD: f.root}
	out, err := HandleCodexReviewGate(input, true /* enabled */, f.root)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a fail verdict must BLOCK; got %+v", out)
	}

	if len(sess.sent) < 3 {
		t.Fatalf("expected >=3 sent requests; got %d (%v)", len(sess.sent), sess.sent)
	}
	// cwd rides the thread/start request (openCodexSessionResolved); the target
	// rides review/start. Both are the gate's assembled review-request fields.
	threadReq := sentRequest(t, sess.sent[1])
	if threadReq["method"] != codexMethodThreadStart {
		t.Fatalf("2nd request = %v, want %s", threadReq["method"], codexMethodThreadStart)
	}
	threadParams, _ := threadReq["params"].(map[string]any)
	if got, _ := threadParams["cwd"].(string); got != f.root {
		t.Errorf("thread/start cwd = %q, want the resolved tree %q", got, f.root)
	}
	req := sentRequest(t, sess.sent[2])
	if req["method"] != codexMethodReviewStart {
		t.Fatalf("3rd request = %v, want %s", req["method"], codexMethodReviewStart)
	}
	params, _ := req["params"].(map[string]any)
	if params == nil {
		t.Fatalf("review/start params missing: %v", req)
	}
	target, _ := params["target"].(map[string]any)
	if target == nil {
		t.Fatalf("review/start target must be the coerced target object, got %v", params["target"])
	}
	if got, _ := target["type"].(string); got != codexTargetUncommitted {
		t.Errorf("target.type = %q, want %q (REQ-CRT-006 shape)", got, codexTargetUncommitted)
	}
}
